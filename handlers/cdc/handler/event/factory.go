package event

import (
	"errors"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/pevin/image-poster-api/comment"
	"github.com/pevin/image-poster-api/post"
)

type dynamodbClient interface {
	Query(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error)
	UpdateItem(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error)
}

type EventHandler interface {
	Handle(events.DynamoDBEventRecord) error
}

type Factory struct {
	dynamodbClient dynamodbClient
	tableName      string
}

func NewFactory(dbClient dynamodbClient, tableName string) *Factory {
	return &Factory{
		dynamodbClient: dbClient,
		tableName:      tableName,
	}
}

func (f *Factory) GetEventHandler(r events.DynamoDBEventRecord) (EventHandler, error) {
	pk := r.Change.Keys["PK"].String()
	sk := r.Change.Keys["SK"].String()
	pkPrefix := strings.Split(pk, "#")[0]
	skPrefix := strings.Split(sk, "#")[0]
	switch r.EventName {
	case "INSERT":
		{
			// check if comment
			if pkPrefix == comment.GetPKPrefix() && skPrefix == comment.GetSKPrefix() {
				return &InsertCommentEventHandler{dbClient: f.dynamodbClient, tableName: f.tableName}, nil
			}
		}
	case "MODIFY":
		{
			// check if POST
			if pkPrefix == post.GetPKPrefix() && skPrefix == post.GetSKPrefix() {
				return &ModifyPostEventHandler{dbClient: f.dynamodbClient, tableName: f.tableName}, nil
			}
		}
	}
	return nil, errors.New("Unsupported event.")
}
