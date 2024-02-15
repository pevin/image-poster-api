package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pevin/image-poster-api/handlers/cdc/handler"
	"github.com/pevin/image-poster-api/handlers/cdc/handler/event"
	"github.com/pevin/image-poster-api/lib/aws/dynamodb"
)

func main() {
	// init lambda handler
	h := initHandler()
	lambda.StartWithOptions(h.Handle, lambda.WithContext(context.Background()))
}

func initHandler() *handler.DynamoDBStreamHandler {
	dynamodbClient := dynamodb.GetClient()
	tableName := os.Getenv("TABLE_NAME")
	factory := event.NewFactory(dynamodbClient, tableName)
	return handler.New(factory)
}
