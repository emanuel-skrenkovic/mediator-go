package mediator

import (
	"context"
	"testing"
)

func TestNewMediatorDoesNotReturnNil(t *testing.T) {
	m := NewMediator()
	if m == nil {
		t.Errorf("NewMediator returned nil")
	}
}

type TestHandler[TRequest any, TResponse any] struct{}

var _ RequestHandler[string, string] = &TestHandler[string, string]{}

func (_ *TestHandler[TRequest, TResponse]) Handle(_ context.Context, request string) (string, error) {
	return request, nil
}

func TestMediatorStoresRegisteredHandler(t *testing.T) {
	m := NewMediator()
	testHandler := TestHandler[string, string]{}

	err := RegisterNewRequestHandler[string, string](m, &testHandler)
	if err != nil {
		t.Errorf("failed to register handler")
	}

	if len(m.handlers) != 1 {
		t.Errorf("failed to store handler")
	}
}

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

func TestMediatorPipelineBehavior(t *testing.T) {
	// Arrange
	m := NewMediator()
	testHandler := TestHandler[string, string]{}

	err := RegisterNewRequestHandler[string, string](m, &testHandler)
	if err != nil {
		t.Errorf("failed to register handler")
	}

	appendValueInner := "append1"
	pipelineBehaviorInner := TestPipelineBehavior{valueToAppend: appendValueInner}
	m.RegisterPipelineBehavior(&pipelineBehaviorInner)

	appendValueOuter := "append2"
	pipelineBehaviorOuter := TestPipelineBehavior{valueToAppend: appendValueOuter}
	m.RegisterPipelineBehavior(&pipelineBehaviorOuter)

	ctx := context.Background()
	request := "Hello, World!"

	// Act
	response, err := Send[string, string](m, ctx, request)

	// Assert
	if err != nil {
		t.Errorf("failed to send")
	}

	expected := request + appendValueInner + appendValueOuter
	if response != expected {
		t.Errorf("unexpected response %s", response)
	}
}
