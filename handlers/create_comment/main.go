package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pevin/image-poster-api/handlers/create_comment/handler"
	"github.com/pevin/image-poster-api/lib/aws/dynamodb"
)

func main() {
	// init lambda handler
	h := initHandler()
	lambda.StartWithOptions(h.Handle, lambda.WithContext(context.Background()))
}

func initHandler() *handler.CreateCommentAPIGatewayHandler {
	dynamodbClient := dynamodb.GetClient()
	tableName := os.Getenv("TABLE_NAME")
	return handler.New(dynamodbClient, tableName)
}
