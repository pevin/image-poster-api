package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pevin/image-poster-api/handlers/image_uploaded/handler"
	"github.com/pevin/image-poster-api/lib/aws/dynamodb"
	"github.com/pevin/image-poster-api/lib/aws/s3"
)

func main() {
	// init lambda handler
	h := initHandler()
	lambda.StartWithOptions(h.Handle, lambda.WithContext(context.Background()))
}

func initHandler() *handler.ImageUploadedS3Handler {
	s3Client := s3.GetS3Client()
	s3Uploader := s3.GetS3Uploader()
	s3PublicBucket := os.Getenv("S3_PUBLIC_BUCKET_NAME")
	dynamodbClient := dynamodb.GetClient()
	tableName := os.Getenv("TABLE_NAME")
	return handler.New(s3Client, s3Uploader, s3PublicBucket, dynamodbClient, tableName)
}
