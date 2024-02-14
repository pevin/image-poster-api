package handler_test

import (
	"context"
	"rest/request"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pevin/image-poster-api/handlers/create_post/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockedS3Uploader struct {
	mock.Mock
}

func (m *mockedS3Uploader) Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*s3manager.UploadOutput), args.Error(1)
}

type mockedMultipartRequest struct {
	mock.Mock
}

func (m *mockedMultipartRequest) GetMultipartValues(req events.APIGatewayProxyRequest, fileFieldName string) (mv request.MultipartValues, err error) {
	args := m.Called(req, fileFieldName)
	return args.Get(0).(request.MultipartValues), args.Error(1)
}

func TestHandle(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		req := events.APIGatewayProxyRequest{
			Headers: map[string]string{"user-id": "test-user-id"},
		}

		mockUploader := new(mockedS3Uploader)
		mockUploader.On("Upload", mock.Anything).Return(&s3manager.UploadOutput{}, nil)
		mockMultipart := new(mockedMultipartRequest)
		mockMultipart.On("GetMultipartValues", req, "image").Return(request.MultipartValues{}, nil)

		handler := handler.New(mockUploader, "test-bucket", mockMultipart)

		res, err := handler.Handle(context.TODO(), req)
		require.NoError(t, err)
		assert.Equal(t, 200, res.StatusCode)
	})
	t.Run("Returns bad request if user-id doesn't exist", func(t *testing.T) {
		req := events.APIGatewayProxyRequest{}

		mockUploader := new(mockedS3Uploader)
		mockUploader.On("Upload", mock.Anything).Return(&s3manager.UploadOutput{}, nil)
		mockMultipart := new(mockedMultipartRequest)
		mockMultipart.On("GetMultipartValues", req, "image").Return(request.MultipartValues{}, nil)

		handler := handler.New(mockUploader, "test-bucket", mockMultipart)

		res, err := handler.Handle(context.TODO(), req)
		require.NoError(t, err)
		assert.Equal(t, 400, res.StatusCode)
	})
}
