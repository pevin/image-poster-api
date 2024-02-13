package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"rest"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	dynamodbInternal "github.com/pevin/image-poster-api/lib/aws/dynamodb"
	"github.com/pevin/image-poster-api/post"
)

type queryParam struct {
	Cursor string `json:"cursor"`
	Limit  int64  `json:"limit,string"`
}

type responseBody struct {
	Data []post.Post `json:"data"`
	Next string      `json:"next"`
}

func Handle(ctx context.Context, req events.APIGatewayProxyRequest) (res events.APIGatewayProxyResponse, err error) {
	qpBytes, err := json.Marshal(req.QueryStringParameters)
	if err != nil {
		fmt.Printf("Got error marshaling query parameters: %v", err)
		return
	}
	params := queryParam{Limit: 10}
	err = json.Unmarshal(qpBytes, &params)
	if err != nil {
		fmt.Printf("Error unmarshaling raw query param: %v", err)
		return
	}

	builder := expression.NewBuilder()
	keyCond := expression.Key("GSI1PK").Equal(expression.Value("POST"))

	// apply cursor if exist
	if params.Cursor != "" {
		keyCond = keyCond.And(expression.Key("GSI1SK").LessThan(expression.Value(params.Cursor)))
	}

	builder = builder.WithKeyCondition(keyCond)

	expr, err := builder.Build()
	if err != nil {
		fmt.Printf("Got error building dynamodb expression: %v", err)
		return
	}

	tableName := os.Getenv("TABLE_NAME")
	dynamodbClient := dynamodbInternal.GetClient()
	input := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		TableName:                 aws.String(tableName),
		ScanIndexForward:          aws.Bool(false),
		IndexName:                 aws.String("GSI1"),
		Limit:                     aws.Int64(params.Limit),
	}

	result, err := dynamodbClient.Query(input)

	list := []post.Post{}

	for _, item := range result.Items {
		p, itemErr := post.FromItem(item)
		if itemErr != nil {
			err = itemErr
			return
		}
		list = append(list, p)
	}

	lek := map[string]string{}
	dynamodbattribute.UnmarshalMap(result.LastEvaluatedKey, &lek)

	respBody := responseBody{
		Data: list,
		Next: lek["GSI1SK"],
	}

	res = rest.OkResponse(respBody, "Fetched posts successful.")
	return
}
