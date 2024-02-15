package event

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/pevin/image-poster-api/comment"
	"github.com/pevin/image-poster-api/post"
)

type InsertCommentEventHandler struct {
	dbClient  dynamodbClient
	tableName string
}

func (h *InsertCommentEventHandler) Handle(r events.DynamoDBEventRecord) error {
	// Get latest comments.
	pk := r.Change.Keys["PK"].String()
	sk := r.Change.Keys["SK"].String()
	keyCond := expression.Key("PK").Equal(expression.Value(pk)).
		And(expression.Key("SK").BeginsWith(comment.GetSKPrefix()))

	builder := expression.NewBuilder().WithKeyCondition(keyCond)

	expr, err := builder.Build()
	if err != nil {
		fmt.Printf("Got error building dynamodb expression: %v", err)
		return err
	}

	input := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		TableName:                 aws.String(h.tableName),
		ScanIndexForward:          aws.Bool(false),
		Limit:                     aws.Int64(2), // Number of latest comments to be appended in Post item
	}

	result, err := h.dbClient.Query(input)
	if err != nil {
		fmt.Printf("Got error on dynamodb query: %v", err)
		return err
	}
	latestComments := []*dynamodb.AttributeValue{}

	for _, i := range result.Items {
		c, itemErr := comment.FromItem(i)
		if itemErr != nil {
			fmt.Printf("Got error on unmarshal comment: %v", itemErr)
			return itemErr
		}
		pc := post.PostComment{
			ID:        c.ID,
			Content:   c.Content,
			CreatedAt: c.CreatedAt,
			Creator:   c.Creator,
		}
		av, itemErr := dynamodbattribute.MarshalMap(pc)
		if itemErr != nil {
			fmt.Printf("Got error on marshal map post comment: %v", itemErr)
			return itemErr
		}

		latestComments = append(latestComments, &dynamodb.AttributeValue{
			M: av,
		})
	}

	// update post to increment total comments and update latest 2 comments
	c := comment.FromKey(pk, sk)
	p := post.Post{
		ID: c.PostID,
	}

	updateInput := &dynamodb.UpdateItemInput{
		TableName:        aws.String(h.tableName),
		Key:              p.ToKey(),
		ReturnValues:     aws.String("NONE"),
		UpdateExpression: aws.String("SET total_comments = total_comments + :one, latest_comments = :lc"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":one": {
				N: aws.String("1"),
			},
			":lc": {
				L: latestComments,
			},
		},
	}
	_, err = h.dbClient.UpdateItem(updateInput)
	if err != nil {
		fmt.Printf("Got error updating item using %+v input: %v", updateInput, err)
	}
	return err
}
