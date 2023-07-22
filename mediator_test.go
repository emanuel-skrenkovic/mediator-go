package mediator

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
)


// Handler tests

type TestRequest struct {
	value string
}

type TestHandler[TRequest any, TResponse any] struct {
	executed bool
}

func (h *TestHandler[TRequest, TResponse]) Handle(_ context.Context, request TRequest) (TRequest, error) {
	h.executed = true
	return request, nil
}

func Test_RegisterRequestHandler_Stores_Registered_Handler(t *testing.T) {
	// Arrange
	testHandler := TestHandler[string, string]{}

	// Act
	err := RegisterRequestHandler[string, string](&testHandler)

	// Assert
	if err != nil {
		t.Errorf("failed to register handler: %s", err.Error())
	}

	if len(requestHandlers) != 1 {
		t.Errorf(
			"mediator.requestHandlers length '%d' does not match the expected length '%d' after registering a handler",
			len(requestHandlers),
			1,
		)
	}

	requestHandlers = make(map[reflect.Type]any)
}

func Test_RegisterRequestHandler_Stores_Registered_Handler_Of_Struct_Type(t *testing.T) {
	// Arrange
	testHandler := TestHandler[TestRequest, TestRequest]{}

	// Act
	err := RegisterRequestHandler[TestRequest, TestRequest](&testHandler)

	// Assert
	if err != nil {
		t.Errorf("failed to register handler: %s", err.Error())
	}

	if len(requestHandlers) != 1 {
		t.Errorf(
			"mediator.requestHandlers length '%d' does not match the expected length '%d' after registering a handler",
			len(requestHandlers),
			1,
		)
	}

	requestHandlers = make(map[reflect.Type]any)
}

func Test_RegisterRequestHandler_Returns_Error_When_Registering_Handler_For_Already_Registered_Request_Type(t *testing.T) {
	// Arrange
	testHandler := TestHandler[string, string]{}
	err := RegisterRequestHandler[string, string](&testHandler)
	if err != nil {
		t.Errorf("failed to register handler: %s", err.Error())
	}

	testHandler2 := TestHandler[string, string]{}

	// Act
	err = RegisterRequestHandler[string, string](&testHandler2)

	// Assert
	if err == nil {
		t.Errorf(
			"did not receive expected error when registering a handler with request type that already has a registered handler",
		)
	}

	if len(requestHandlers) != 1 {
		t.Errorf(
			"mediator.pipelineBehaviors length '%d' does not match the expected value '%d' after trying to register a second pipeline behavior with same request type",
			len(pipelineBehaviors),
			1,
		)
	}

	requestHandlers = make(map[reflect.Type]any)
}

func Test_Send_Executes_Registered_Handler_For_Given_Request_Type(t *testing.T) {
	// Arrange
	testHandler := TestHandler[string, string]{}
	_ = RegisterRequestHandler[string, string](&testHandler)

	// Act
	_, err := Send[string, string](context.Background(), "value")

	// Assert
	if err != nil {
		t.Errorf("unexpected error occurred on Send: %s", err.Error())
	}

	if !testHandler.executed {
		t.Error("failed to execute registered handler")
	}

	requestHandlers = make(map[reflect.Type]any)
}

func Test_Send_Executes_Registered_Handler_For_Given_Request_Type_With_Struct_Request_Type(t *testing.T) {
	// Arrange
	testHandler := TestHandler[TestRequest, TestRequest]{}
	_ = RegisterRequestHandler[TestRequest, TestRequest](&testHandler)

	// Act
	_, err := Send[TestRequest, TestRequest](context.Background(), TestRequest{value: "value"})

	// Assert
	if err != nil {
		t.Errorf("unexpected error occurred on Send: %s", err.Error())
	}

	if !testHandler.executed {
		t.Error("failed to execute registered handler")
	}

	requestHandlers = make(map[reflect.Type]any)
}

func Test_Send_Returns_Error_When_Sending_Request_Type_Without_A_Matching_Registered_Handler(t *testing.T) {
	// Arrange
	testHandler := TestHandler[TestRequest, TestRequest]{}
	_ = RegisterRequestHandler[TestRequest, TestRequest](&testHandler)

	// Act
	_, err := Send[string, string](context.Background(), "value")

	// Assert
	if err == nil {
		t.Error("did not receive expected error on sending a request without registering a matching handler")
	}

	requestHandlers = make(map[reflect.Type]any)
}

// Pipeline tests

type TestPipelineBehavior struct {
	valueToAppend string
}

func (b *TestPipelineBehavior) Handle(
	ctx context.Context,
	request interface{},
	next RequestHandlerFunc,
) (interface{}, error) {
	value := request.(string)
	value += b.valueToAppend
	return next(ctx, value)
}

func Test_RegisterPipelineBehavior_Appends_Behavior_Instance_To_Mediator(t *testing.T) {
	// Arrange
	testHandler := TestHandler[string, string]{}

	err := RegisterRequestHandler[string, string](&testHandler)
	if err != nil {
		t.Errorf("failed to register handler %s", err.Error())
	}

	pipelineBehaviorInner := TestPipelineBehavior{}
	RegisterPipelineBehavior(&pipelineBehaviorInner)

	if len(pipelineBehaviors) != 1 {
		t.Errorf(
			"mediator.pipelineBehaviors length '%d' does not match the expected value '%d' after registering a pipeline behavior",
			len(pipelineBehaviors),
			1,
		)
	}

	requestHandlers = make(map[reflect.Type]any)
	pipelineBehaviors = make([]PipelineBehavior, 0)
}

func Test_PipelineBehavior_First_Registered_Gets_Executed_First(t *testing.T) {
	// Arrange
	testHandler := TestHandler[string, string]{}

	err := RegisterRequestHandler[string, string](&testHandler)
	if err != nil {
		t.Errorf("failed to register handler %s", err.Error())
	}

	appendValueInner := "append1"
	pipelineBehaviorInner := TestPipelineBehavior{valueToAppend: appendValueInner}
	RegisterPipelineBehavior(&pipelineBehaviorInner)

	appendValueOuter := "append2"
	pipelineBehaviorOuter := TestPipelineBehavior{valueToAppend: appendValueOuter}
	RegisterPipelineBehavior(&pipelineBehaviorOuter)

	ctx := context.Background()
	request := "Hello, World!"

	// Act
	response, err := Send[string, string](ctx, request)

	// Assert
	if err != nil {
		t.Errorf("failed to send")
	}

	if !testHandler.executed {
		t.Errorf("failed to execute handler")
	}

	expected := request + appendValueInner + appendValueOuter
	if response != expected {
		t.Errorf("unexpected response %s", response)
	}

	requestHandlers = make(map[reflect.Type]any)
	pipelineBehaviors = make([]PipelineBehavior, 0)
}

type TestContextPipelineBehavior struct {
	onHandle func(ctx context.Context, request interface{}) (context.Context, interface{})
}

func (h *TestContextPipelineBehavior) Handle(ctx context.Context, request interface{}, next RequestHandlerFunc) (interface{}, error) {
	if h.onHandle != nil {
		ctx, request = h.onHandle(ctx, request)
	}
	return next(ctx, request)
}

type TestContextHandler[TRequest any, TResponse any] struct {
	onHandle        func(ctx context.Context, request TRequest)
	receivedMessage TRequest
}

func (h *TestContextHandler[TRequest, TResponse]) Handle(ctx context.Context, request TRequest) (TRequest, error) {
	h.receivedMessage = request

	if h.onHandle != nil {
		h.onHandle(ctx, request)
	}

	return request, nil
}

type CTXKey string

func Test_PipelineBehavior_Context_Preserved_Over_Execution(t *testing.T) {
	// Arrange
	const (
		contextKey CTXKey = "key"
		contextVal string = "value"
	)

	testHandler := TestContextHandler[string, string]{onHandle: func(ctx context.Context, _ string) {
		if val := ctx.Value(contextKey); val != contextVal {
			t.Errorf(
				"failed to receive value from context expected %s found %s",
				contextVal,
				val,
			)
		}
	}}

	err := RegisterRequestHandler[string, string](&testHandler)
	if err != nil {
		t.Errorf("failed to register handler %s", err.Error())
	}

	onPipeline := func(ctx context.Context, request interface{}) (context.Context, interface{}) {
		return context.WithValue(ctx, contextKey, "value"), request
	}

	pipelineBehaviorInner := TestContextPipelineBehavior{onHandle: onPipeline}
	RegisterPipelineBehavior(&pipelineBehaviorInner)

	pipelineBehaviorOuter := TestContextPipelineBehavior{}
	RegisterPipelineBehavior(&pipelineBehaviorOuter)

	ctx := context.Background()
	request := "Hello, World!"

	// Act
	_, err = Send[string, string](ctx, request)

	// Assert
	if err != nil {
		t.Errorf("failed to send")
	}

	pipelineBehaviors = make([]PipelineBehavior, 0)
}

// Notification tests

type TestNotificationHandler[TNotification any] struct {
	executedWithNotification TNotification
	err                      error
}

func (h *TestNotificationHandler[TNotification]) Handle(_ context.Context, notification TNotification) error {
	if h.err != nil {
		return h.err
	}

	h.executedWithNotification = notification
	return nil
}

func Test_RegisterNotificationHandler_Appends_Handler_To_Mediator(t *testing.T) {
	// Arrange
	handler := TestNotificationHandler[string]{}

	// Act
	RegisterNotificationHandler[string](&handler)

	// Assert
	var notification string
	notificationHandlersNum := len(notificationHandlers[reflect.TypeOf(notification)])

	if notificationHandlersNum != 1 {
		t.Errorf("expected '%d' handlers, found '%d'", 1, notificationHandlersNum)
	}

	notificationHandlers = make(map[reflect.Type][]any)
}

func Test_RegisterNotificationHandler_Appends_Multiple_Handlers_For_Same_Notification_Type_To_Mediator(t *testing.T) {
	// Arrange
	handlers := []TestNotificationHandler[string]{
		{},
		{},
		{},
		{},
		{},
	}

	// Act
	for _, handler := range handlers {
		RegisterNotificationHandler[string](&handler)
	}

	// Assert
	var notification string
	notificationHandlersNum := len(notificationHandlers[reflect.TypeOf(notification)])

	if notificationHandlersNum != len(handlers) {
		t.Errorf("expected '%d' handlers, found '%d'", 1, notificationHandlersNum)
	}

	notificationHandlers = make(map[reflect.Type][]any)
}

func Test_Publish_Publishes_To_All_Registered_Notification_Handlers(t *testing.T) {
	// Arrange
	handlers := []TestNotificationHandler[string]{
		{},
		{},
		{},
		{},
		{},
	}

	for i := 0; i < len(handlers); i++ {
		RegisterNotificationHandler[string](&handlers[i])
	}

	notification := "value"

	// Act
	err := Publish[string](context.Background(), notification)

	// Assert
	if err != nil {
		t.Errorf("received unexpected error on publish: %s", err.Error())
	}

	for _, handler := range handlers {
		if handler.executedWithNotification != notification {
			t.Errorf(
				"failed to execute all registered notification handlers expected '%s' found '%s'",
				notification,
				handler.executedWithNotification,
			)
		}
	}

	notificationHandlers = make(map[reflect.Type][]any)
}

func Test_Publish_Continues_Publishing_After_Receiving_Error_From_Handler(t *testing.T) {
	// Arrange
	expectedErr := fmt.Errorf("EXPLOSIONS, MAYHEM, OTHER BAD STUFF")
	handlers := []TestNotificationHandler[string]{
		{},
		{err: expectedErr},
		{},
		{},
		{},
	}

	for i := 0; i < len(handlers); i++ {
		RegisterNotificationHandler[string](&handlers[i])
	}

	notification := "value"

	// Act
	err := Publish[string](context.Background(), notification)

	// Assert
	if err == nil {
		t.Error("did not receive expected error on publish")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf(
			"received error of unexpected type expected '%s' received '%s'",
			expectedErr.Error(),
			err.Error(),
		)
	}

	for _, handler := range handlers {
		if handler.executedWithNotification != notification && handler.err == nil {
			t.Errorf(
				"failed to execute all registered notification handlers expected '%s' found '%s'",
				notification,
				handler.executedWithNotification,
			)
		}
	}

	notificationHandlers = make(map[reflect.Type][]any)
}

func Test_Publish_Does_Not_Return_Error_If_No_Handlers_Found(t *testing.T) {
	// Arrange
	notification := "value"

	// Act
	err := Publish[string](context.Background(), notification)

	// Assert
	if err != nil {
		t.Errorf("unexpected error occurred on Publish: %s", err.Error())
	}
}
