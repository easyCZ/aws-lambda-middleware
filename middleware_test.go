package aws_lambda_middleware

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/stretchr/testify/require"
)

func TestChain_NoMiddlewares(t *testing.T) {
	expectedResponse := "invoked"
	handler := lambda.NewHandler(func() (string, error) {
		return expectedResponse, nil
	})

	chained := Chain(handler)

	resp, err := chained.Invoke(context.Background(), []byte("test"))
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(`"%s"`, expectedResponse), string(resp)) // lambda.NewHandler will wrap plain string responses in quotes
}

func TestChain_ExecutionOrdering(t *testing.T) {
	var responses string
	m1 := Middleware(func(handler lambda.Handler) lambda.Handler {
		return HandlerFunc(func(ctx context.Context, bytes []byte) ([]byte, error) {
			responses += "first,"
			return handler.Invoke(ctx, bytes)
		})
	})

	m2 := Middleware(func(handler lambda.Handler) lambda.Handler {
		return HandlerFunc(func(ctx context.Context, bytes []byte) ([]byte, error) {
			responses += "second,"
			return handler.Invoke(ctx, bytes)
		})
	})

	m3 := Middleware(func(handler lambda.Handler) lambda.Handler {
		return HandlerFunc(func(ctx context.Context, bytes []byte) ([]byte, error) {
			responses += "third"
			return handler.Invoke(ctx, bytes)
		})
	})

	expectedResponse := "invoked"
	handler := lambda.NewHandler(func() (string, error) {
		return expectedResponse, nil
	})

	chained := Chain(handler, m1, m2, m3)

	resp, err := chained.Invoke(context.Background(), []byte("test"))
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(`"%s"`, expectedResponse), string(resp)) // lambda.NewHandler will wrap plain string responses in quotes

	require.Equal(t, "first,second,third", responses)
}
