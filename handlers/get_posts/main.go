package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pevin/image-poster-api/handlers/get_posts/handler"
)

func main() {
	// init lambda handler
	lambda.StartWithOptions(handler.Handle, lambda.WithContext(context.Background()))
}
