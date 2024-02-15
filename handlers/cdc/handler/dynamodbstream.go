package handler

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/pevin/image-poster-api/handlers/cdc/handler/event"
)

type eventHandlerFactory interface {
	GetEventHandler(r events.DynamoDBEventRecord) (event.EventHandler, error)
}

type DynamoDBStreamHandler struct {
	factory eventHandlerFactory
}

func New(factory eventHandlerFactory) *DynamoDBStreamHandler {
	return &DynamoDBStreamHandler{factory: factory}
}

func (h *DynamoDBStreamHandler) Handle(ctx context.Context, event events.DynamoDBEvent) error {
	for _, r := range event.Records {
		eventHandler, factoryErr := h.factory.GetEventHandler(r)
		if factoryErr != nil {
			// skip because unsupported
			continue
		}
		handlerErr := eventHandler.Handle(r)
		if handlerErr != nil {
			// log for now
			fmt.Printf("Got error when handling event: %v", handlerErr)
		}
	}
	return nil
}
