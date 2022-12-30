package mediator

import (
	"context"
	"testing"
)

func Benchmark_Send(b *testing.B) {
	// Arrange
	m := NewMediator()

	testHandler := TestHandler[TestRequest, TestRequest]{}
	_ = RegisterRequestHandler[TestRequest, TestRequest](m, &testHandler)

	request := TestRequest{value: "some inconspicuous value"}
	ctx := context.Background()

	// Act
	for i := 0; i < b.N; i++ {
		_, _ = Send[TestRequest, TestRequest](m, ctx, request)
	}
}

type BenchmarkPipelineBehavior struct{}

func (b *BenchmarkPipelineBehavior) Handle(
	ctx context.Context,
	request interface{},
	next RequestHandlerFunc,
) (interface{}, error) {
	return next(ctx, request)
}

func Benchmark_Send_With_Pipeline_Behaviors(b *testing.B) {
	// Arrange
	m := NewMediator()
	testHandler := TestHandler[TestRequest, TestRequest]{}

	_ = RegisterRequestHandler[TestRequest, TestRequest](m, &testHandler)

	pipelines := []BenchmarkPipelineBehavior{
		{},
		{},
		{},
		{},
		{},
	}

	for i := 0; i < len(pipelines); i++ {
		m.RegisterPipelineBehavior(&pipelines[i])
	}

	request := TestRequest{value: "some inconspicuous value"}
	ctx := context.Background()

	// Act
	for i := 0; i < b.N; i++ {
		_, _ = Send[TestRequest, TestRequest](m, ctx, request)
	}
}
