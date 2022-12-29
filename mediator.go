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

type PipelineBehavior interface {
	Handle(ctx context.Context, request interface{}, next RequestHandlerFunc) (interface{}, error)
}

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
	handlers          map[reflect.Type]interface{}
	pipelineBehaviors []PipelineBehavior
}

func NewMediator() *Mediator {
	return &Mediator{
		handlers: make(map[reflect.Type]interface{}),
	}
}

func RegisterNewRequestHandler[TRequest any, TResponse any](
	m *Mediator,
	handler RequestHandler[TRequest, TResponse],
) error {
	var request TRequest
	requestType := reflect.TypeOf(request)

	if _, contains := m.handlers[requestType]; contains {
		return fmt.Errorf("handler for %s already registered", requestType.String())
	}

	m.handlers[requestType] = handler

	return nil
}

func (m *Mediator) RegisterPipelineBehavior(pipelineBehavior PipelineBehavior) {
	m.pipelineBehaviors = append(m.pipelineBehaviors, pipelineBehavior)
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

	numBehaviors := len(m.pipelineBehaviors)
	if numBehaviors < 1 {
		return typedRequestHandler.Handle(ctx, request)
	}

	var behavior RequestHandlerFunc = func(ctx context.Context, req interface{}) (interface{}, error) {
		typedRequest, ok := req.(TRequest)
		if !ok {
			return response, fmt.Errorf(
				"incorrect request type expected %s got %s",
				reflect.TypeOf(req).String(),
				reflect.TypeOf(request).String(),
			)
		}
		return typedRequestHandler.Handle(ctx, typedRequest)
	}

	for i := numBehaviors - 1; i >= 0; i-- {
		pipeline := m.pipelineBehaviors[i]

		behavior = func(pipelineBehavior PipelineBehavior, next RequestHandlerFunc) RequestHandlerFunc {
			return func(ctx context.Context, request interface{}) (interface{}, error) {
				return pipeline.Handle(ctx, request, next)
			}
		}(pipeline, behavior)
	}

	untypedResponse, err := behavior(ctx, request)
	if err != nil {
		return response, err
	}

	response, ok = untypedResponse.(TResponse)
	if !ok {
		return response, fmt.Errorf(
			"failed to convert response of type %s to type %s",
			reflect.TypeOf(untypedResponse).String(),
			reflect.TypeOf(response).String(),
		)
	}

	return response, nil
}

func Publish[TNotification any](m *Mediator, notification TNotification) {
	panic("TODO: implement Publish")
}
