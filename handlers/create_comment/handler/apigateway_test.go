package handler_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/pevin/image-poster-api/handlers/create_comment/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockedDynamodbClient struct {
	mock.Mock
}

func (mr *mockedDynamodbClient) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	args := mr.Called(input)
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func TestHandle(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		body := map[string]string{
			"content": "test content",
			"post_id": "test-post-id-123",
		}
		bodyBytes, err := json.Marshal(body)
		require.NoError(t, err)
		req := events.APIGatewayProxyRequest{
			Headers: map[string]string{"user-id": "test-user-id"},
			Body:    string(bodyBytes),
		}

		client := new(mockedDynamodbClient)
		client.On("PutItem", mock.Anything).Return(&dynamodb.PutItemOutput{}, nil)
		h := handler.New(client, "table-name")

		res, err := h.Handle(context.TODO(), req)
		require.NoError(t, err)

		assert.Equal(t, 200, res.StatusCode)
	})
	t.Run("failed: no user returns 400 status code", func(t *testing.T) {
		req := events.APIGatewayProxyRequest{}

		client := new(mockedDynamodbClient)
		client.On("PutItem", mock.Anything).Return(&dynamodb.PutItemOutput{}, nil)
		h := handler.New(client, "table-name")

		res, err := h.Handle(context.TODO(), req)
		require.NoError(t, err)

		assert.Equal(t, 400, res.StatusCode)
	})
}
