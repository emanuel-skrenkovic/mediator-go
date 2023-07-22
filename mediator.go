package mediator

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

type RequestHandlerFunc func(ctx context.Context, request any) (any, error)

type RequestHandler[TRequest any, TResponse any] interface {
	Handle(ctx context.Context, request TRequest) (TResponse, error)
}

type NotificationHandler[TNotification any] interface {
	Handle(ctx context.Context, notification TNotification) error
}

type PipelineBehavior interface {
	Handle(ctx context.Context, request any, next RequestHandlerFunc) (any, error)
}

var (
	requestHandlers      map[reflect.Type]any   = make(map[reflect.Type]any)
	notificationHandlers map[reflect.Type][]any = make(map[reflect.Type][]any)
	pipelineBehaviors                           = []PipelineBehavior{}
)

func RegisterRequestHandler[TRequest any, TResponse any](
	handler RequestHandler[TRequest, TResponse],
) error {
	var request TRequest
	requestType := reflect.TypeOf(request)

	if _, contains := requestHandlers[requestType]; contains {
		return fmt.Errorf("handler for request type '%s' is already registered", requestType.String())
	}

	requestHandlers[requestType] = handler
	return nil
}

func RegisterPipelineBehavior(pipelineBehavior PipelineBehavior) {
	pipelineBehaviors = append(pipelineBehaviors, pipelineBehavior)
}

func RegisterNotificationHandler[TNotification any](handler NotificationHandler[TNotification]) {
	var notification TNotification
	notificationType := reflect.TypeOf(notification)

	notificationHandlers[notificationType] = append(
		notificationHandlers[notificationType],
		handler,
	)
}

func Send[TRequest any, TResponse any](ctx context.Context, request TRequest) (TResponse, error) {
	requestType := reflect.TypeOf(request)

	var response TResponse
	handler, registered := requestHandlers[requestType]
	if !registered {
		return response, fmt.Errorf(
			"request handler for request type '%s' is not registered", requestType.String(),
		)
	}

	typedRequestHandler, ok := handler.(RequestHandler[TRequest, TResponse])
	if !ok {
		return response, fmt.Errorf(
			"failed to convert handler '%s' to typed handler 'RequestHandler[%s, %s]'",
			typeName(handler),
			typeName(request),
			typeName(response),
		)
	}

	numBehaviors := len(pipelineBehaviors)
	if numBehaviors < 1 {
		return typedRequestHandler.Handle(ctx, request)
	}

	var behavior RequestHandlerFunc = func(ctx context.Context, req any) (any, error) {
		typedRequest, ok := req.(TRequest)
		if !ok {
			return response, fmt.Errorf(
				"incorrect request type expected '%s' got '%s'",
				typeName(req),
				typeName(request),
			)
		}
		return typedRequestHandler.Handle(ctx, typedRequest)
	}

	for i := numBehaviors - 1; i >= 0; i-- {
		pipeline := pipelineBehaviors[i]

		// Create new behavior through a func to avoid infinite loops of self-reference.
		// Passing in the parameters through the function avoids that.
		behavior = func(pipelineBehavior PipelineBehavior, next RequestHandlerFunc) RequestHandlerFunc {
			return func(ctx context.Context, request any) (any, error) {
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
			"failed to convert response of type '%s' to type '%s'",
			typeName(untypedResponse),
			typeName(response),
		)
	}

	return response, nil
}

func Publish[TNotification any](ctx context.Context, notification TNotification) error {
	notificationType := reflect.TypeOf(notification)

	handlers := notificationHandlers[notificationType]

	if len(handlers) < 1 {
		return nil
	}

	var aggregateError error
	for _, handler := range handlers {
		typedHandler, _ := handler.(NotificationHandler[TNotification])
		if err := typedHandler.Handle(ctx, notification); err != nil {
			// Poor "substitute" for actual aggregate errors.
			handleErr := fmt.Errorf(
				"failed to execute notification handler '%s' with error: %w",
				typeName(typedHandler),
				err,
			)
			aggregateError = errors.Join(handleErr, aggregateError)
		}
	}

	return aggregateError
}

func typeName(obj any) string {
	return reflect.TypeOf(obj).String()
}
