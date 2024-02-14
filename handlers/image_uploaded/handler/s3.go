package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/pevin/image-poster-api/post"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/disintegration/imaging"
)

type s3Client interface {
	GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error)
}

type uploader interface {
	Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

type dynamodbClient interface {
	PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error)
}

type ImageUploadedS3Handler struct {
	s3Client       s3Client
	s3Uploader     uploader
	s3PublicBucket string
	dynamodbClient dynamodbClient
	tableName      string
}

func New(
	s3Client *s3.S3,
	s3Uploader *s3manager.Uploader,
	s3Bucket string,
	ddbClient *dynamodb.DynamoDB,
	tableName string) *ImageUploadedS3Handler {
	return &ImageUploadedS3Handler{
		s3Client:       s3Client,
		s3Uploader:     s3Uploader,
		s3PublicBucket: s3Bucket,
		dynamodbClient: ddbClient,
		tableName:      tableName,
	}
}

func (h *ImageUploadedS3Handler) Handle(ctx context.Context, event events.S3Event) (err error) {
	for _, r := range event.Records {
		rErr := h.handleRecord(r)
		if rErr != nil {
			err = rErr
		}
	}
	return
}

type Metadata struct {
	Caption string `json:"x-amz-meta-caption"`
	UserID  string `json:"x-amz-meta-user"`
}

func (h *ImageUploadedS3Handler) handleRecord(r events.S3EventRecord) (err error) {
	k := r.S3.Object.Key
	id := strings.Split(k, ".")[0]

	s3Obj, err := h.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(r.S3.Bucket.Name),
		Key:    aws.String(k),
	})

	// get caption and user ID
	var metadata Metadata
	metadataBytes, err := json.Marshal(s3Obj.Metadata)
	if err != nil {
		fmt.Printf("Got error in marshal metadata: %v", err)
		return
	}
	err = json.Unmarshal(metadataBytes, &metadata)
	if err != nil {
		fmt.Printf("Got error in unmarshal metadata: %v", err)
		return
	}
	if len(metadata.UserID) == 0 {
		fmt.Printf("User ID doesn't exist in metadata s3 object.")
		return
	}

	if err != nil {
		fmt.Printf("Error while getting s3 object: %v", err)
		return
	}

	img, err := imaging.Decode(s3Obj.Body)
	if err != nil {
		fmt.Printf("Got error in image decoding: %v", err)
		return
	}

	img600 := imaging.Resize(img, 600, 600, imaging.Lanczos)
	buff := bytes.NewBuffer([]byte{})
	err = imaging.Encode(buff, img600, imaging.JPEG)
	if err != nil {
		fmt.Printf("Got error in encoding image: %v", err)
		return
	}

	filename := id + "." + imaging.JPEG.String()
	uploaded, err := h.s3Uploader.Upload(&s3manager.UploadInput{
		Bucket:             aws.String(h.s3PublicBucket),
		Key:                aws.String(filename),
		Body:               buff,
		ContentDisposition: aws.String("inline"),
		ContentType:        aws.String("image/jpeg"),
	})

	url := uploaded.Location

	// Create post
	post := post.Post{
		ID:            id,
		Caption:       metadata.Caption,
		Image:         url,
		Creator:       metadata.UserID,
		CreatedAt:     r.EventTime,
		TotalComments: 0,
	}
	av, err := post.ToItem()
	if err != nil {
		fmt.Printf("Got error marshalling concessionaire: %s", err)
		return
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(h.tableName),
	}

	_, err = h.dynamodbClient.PutItem(input)
	if err != nil {
		fmt.Printf("Got error calling PutItem: %s", err)
		return
	}

	return
}
