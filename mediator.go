package mediator

import (
	"context"
	"fmt"
	"reflect"
)

var (
	_ Sender    = NewMediator()
	_ Publisher = NewMediator()
)

type RequestHandlerFunc func(ctx context.Context, request interface{}) (interface{}, error)

type RequestHandler[TRequest any, TResponse any] interface {
	Handle(ctx context.Context, request TRequest) (TResponse, error)
}

type Sender interface {
	Send(ctx context.Context, request interface{}) (interface{}, error)
}

type Publisher interface {
	Publish(ctx context.Context, notification interface{})
}

type Mediator struct {
	handlers map[reflect.Type]interface{}
}

func NewMediator() *Mediator {
	return &Mediator{
		handlers: make(map[reflect.Type]interface{}),
	}
}

func (m *Mediator) RegisterNewRequestHandler() {
	panic("TODO: implement Mediator.RegisterNewRequestHandler")
}

func (m *Mediator) Send(ctx context.Context, request interface{}) (interface{}, error) {
	panic("TODO: implement Mediator.Send")
}

func (m *Mediator) Publish(ctx context.Context, notification interface{}) {
	panic("TODO: implement Mediator.Publish")
}

func Send[TRequest any, TResponse any](m *Mediator, ctx context.Context, request TRequest) (TResponse, error) {
	requestType := reflect.TypeOf(request)

	var response TResponse
	handler, registered := m.handlers[requestType]
	if !registered {
		return response, fmt.Errorf("handlers for type %s is not registered", requestType.String())
	}

	typedRequestHandler, ok := handler.(RequestHandler[TRequest, TResponse])
	if !ok {
		return response, fmt.Errorf("failed to convert converter %s", reflect.TypeOf(handler).String())
	}

	// TODO: implement PipelineBehavior

	return typedRequestHandler.Handle(ctx, request)
}

func Publish[TNotification any](m *Mediator, notification TNotification) {
	panic("TODO: implement Publish")
}
