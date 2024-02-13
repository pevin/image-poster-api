package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	dynamodbInternal "github.com/pevin/image-poster-api/lib/aws/dynamodb"
	s3Internal "github.com/pevin/image-poster-api/lib/aws/s3"
	"github.com/pevin/image-poster-api/post"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/disintegration/imaging"
)

var (
	s3Client       *s3.S3
	s3Uploader     *s3manager.Uploader
	s3PublicBucket string
	dynamodbClient *dynamodb.DynamoDB
	tableName      string
)

func Handle(ctx context.Context, event events.S3Event) (err error) {
	s3Client = s3Internal.GetS3Client()
	s3Uploader = s3Internal.GetS3Uploader()
	s3PublicBucket = os.Getenv("S3_PUBLIC_BUCKET_NAME")
	dynamodbClient = dynamodbInternal.GetClient()
	tableName = os.Getenv("TABLE_NAME")

	for _, r := range event.Records {
		rErr := handleRecord(r)
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

func handleRecord(r events.S3EventRecord) (err error) {
	k := r.S3.Object.Key
	id := strings.Split(k, ".")[0]

	s3Obj, err := s3Client.GetObject(&s3.GetObjectInput{
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
	uploaded, err := s3Uploader.Upload(&s3manager.UploadInput{
		Bucket:             aws.String(s3PublicBucket),
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
		TableName: aws.String(tableName),
	}

	_, err = dynamodbClient.PutItem(input)
	if err != nil {
		fmt.Printf("Got error calling PutItem: %s", err)
		return
	}

	return
}
