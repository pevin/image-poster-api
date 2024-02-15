package event

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/pevin/image-poster-api/post"
)

type ModifyPostEventHandler struct {
	dbClient  dynamodbClient
	tableName string
}

func (h *ModifyPostEventHandler) Handle(r events.DynamoDBEventRecord) error {
	id := r.Change.NewImage["id"].String()
	totalCommentsOld := r.Change.OldImage["total_comments"].Number()
	totalCommentsNew := r.Change.NewImage["total_comments"].Number()

	if totalCommentsNew == totalCommentsOld {
		// skip
		return nil
	}

	totalComments, err := strconv.Atoi(totalCommentsNew)
	if err != nil {
		fmt.Printf("Got error converting total comments %v to int: %v", totalCommentsNew, err)
		return err
	}
	p := post.Post{
		ID:            id,
		TotalComments: totalComments,
	}
	gsi1SK := p.GetGSI1SK()

	updateInput := &dynamodb.UpdateItemInput{
		TableName:        aws.String(h.tableName),
		Key:              p.ToKey(),
		ReturnValues:     aws.String("NONE"),
		UpdateExpression: aws.String("SET GSI1SK = :sk"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":sk": {
				S: aws.String(gsi1SK),
			},
		},
	}
	_, err = h.dbClient.UpdateItem(updateInput)
	if err != nil {
		fmt.Printf("Got error updating item using %+v input: %v", updateInput, err)
	}
	return err
}
