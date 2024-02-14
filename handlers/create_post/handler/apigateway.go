package handler

import (
	"context"
	"encoding/json"
	"rest"
	"rest/request"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/oklog/ulid/v2"
)

type requestHeader struct {
	UserID string `json:"user-id"`
}

type uploader interface {
	Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

type multipartRequest interface {
	GetMultipartValues(req events.APIGatewayProxyRequest, fileFieldName string) (mv request.MultipartValues, err error)
}

type CreatePostAPIGatewayHandler struct {
	s3Uploader       uploader
	bucket           string
	multipartRequest multipartRequest
}

func New(su uploader, bucket string, mr multipartRequest) *CreatePostAPIGatewayHandler {
	return &CreatePostAPIGatewayHandler{s3Uploader: su, bucket: bucket, multipartRequest: mr}
}

func (h *CreatePostAPIGatewayHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (res events.APIGatewayProxyResponse, err error) {
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
		// bad request
		res = rest.BadRequestResponse("user-id is required in header.")
		return
	}

	mv, err := h.multipartRequest.GetMultipartValues(req, "image")
	if err != nil {
		return
	}

	// todo: validate file extension

	// todo: validate image size

	// init uploader client

	id := ulid.Make().String()
	filename := id + "." + mv.FileExtension
	_, err = h.s3Uploader.Upload(&s3manager.UploadInput{
		Bucket:             aws.String(h.bucket),
		Key:                aws.String(filename),
		Body:               mv.Body,
		ContentDisposition: aws.String("inline"),
		ContentType:        aws.String(mv.ContentType),
		Metadata: map[string]*string{
			"x-amz-meta-caption": aws.String(mv.Caption),
			"x-amz-meta-user":    aws.String(header.UserID),
		},
	})

	if err != nil {
		return
	}

	res = rest.EmptyOkResponse("Request successful.")
	return
}
