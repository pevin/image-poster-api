package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"rest"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/oklog/ulid/v2"
	"github.com/pevin/image-poster-api/comment"
	dynamodbInternal "github.com/pevin/image-poster-api/lib/aws/dynamodb"
)

type requestHeader struct {
	UserID string `json:"user-id"`
}

type requestBody struct {
	Content string `json:"content"`
	PostID  string `json:"post_id"`
}

func Handle(ctx context.Context, req events.APIGatewayProxyRequest) (res events.APIGatewayProxyResponse, err error) {
	var header requestHeader
	headerBytes, err := json.Marshal(req.Headers)
	if err != nil {
		return
	}
	err = json.Unmarshal(headerBytes, &header)
	if err != nil {
		return
	}

	if len(header.UserID) == 0 {
		res = rest.BadRequestResponse("user-id is required in header.")
		return
	}

	var body requestBody
	err = json.Unmarshal([]byte(req.Body), &body)
	if err != nil {
		return
	}

	// todo: validate if post ID exist.

	// Create comment
	id := ulid.Make().String()
	comment := comment.Comment{
		ID:        id,
		Content:   body.Content,
		PostID:    body.PostID,
		Creator:   header.UserID,
		CreatedAt: time.Now(),
	}
	av, err := comment.ToItem()
	if err != nil {
		fmt.Printf("Got error marshalling concessionaire: %s", err)
		return
	}

	dynamodbClient := dynamodbInternal.GetClient()
	tableName := os.Getenv("TABLE_NAME")

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = dynamodbClient.PutItem(input)
	if err != nil {
		fmt.Printf("Got error calling PutItem: %s", err)
		return
	}

	res = rest.OkResponse(comment, "Comment created successfully.")
	return
}
