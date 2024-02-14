package main

import (
	"context"
	"os"
	"rest/request"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pevin/image-poster-api/handlers/create_post/handler"
	"github.com/pevin/image-poster-api/lib/aws/s3"
)

func main() {
	// init lambda handler
	h := initHandler()
	lambda.StartWithOptions(h.Handle, lambda.WithContext(context.Background()))
}

func initHandler() *handler.CreatePostAPIGatewayHandler {
	s3Uploader := s3.GetS3Uploader()
	bucket := os.Getenv("S3_BUCKET_NAME")
	mr := request.NewMultipartRequest()
	return handler.New(s3Uploader, bucket, mr)
}
