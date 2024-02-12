package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"rest"
	"rest/request"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/oklog/ulid/v2"
)

type Header struct {
	UserID string `json:"user-id"`
}

func Handle(ctx context.Context, req events.APIGatewayProxyRequest) (res events.APIGatewayProxyResponse, err error) {
	var header Header
	headerBytes, err := json.Marshal(req.Headers)
	if err != nil {
		return
	}
	err = json.Unmarshal(headerBytes, &header)
	if err != nil {
		return
	}

	if len(header.UserID) == 0 {
		// bad request
		res = rest.BadRequestResponse("user-id is required in header.")
		return
	}

	mv, err := request.GetMultipartValues(req, "image")
	if err != nil {
		return
	}

	// init uploader client
	awsRegion := os.Getenv("APP_AWS_REGION")
	conf := &aws.Config{Region: aws.String(awsRegion)}
	sess, err := session.NewSession(conf)
	if err != nil {
		fmt.Println("Error creating aws session: ", err)
		panic(err)
	}
	uploader := s3manager.NewUploader(sess)

	s3Bucket := os.Getenv("S3_BUCKET_NAME")
	id := ulid.Make().String()
	filename := id + "." + mv.FileExtension
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:             aws.String(s3Bucket),
		Key:                aws.String(filename),
		Body:               mv.Body,
		ContentDisposition: aws.String("inline"),
		ContentType:        aws.String(mv.ContentType),
		Metadata: map[string]*string{
			"x-amz-meta-caption": aws.String(mv.Caption),
			"x-amz-meta-user":    aws.String(header.UserID),
		},
	})
	if err != nil {
		return
	}

	res = rest.EmptyOkResponse("Request successful.")
	return
}
