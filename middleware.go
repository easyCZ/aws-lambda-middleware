package aws_lambda_middleware

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
)

var (
	// Compile type check that HandlerFunc implements lambda.Handler interface
	_ lambda.Handler = (*HandlerFunc)(nil)
)

// Middleware represents a Middleware for a lambda.Handler
type Middleware func(next lambda.Handler) lambda.Handler

// HandlerFunc is a helper to define a lambda.Handler as a function
type HandlerFunc func(context.Context, []byte) ([]byte, error)

func (f HandlerFunc) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	return f(ctx, payload)
}

// Chain creates a middleware chain.
// Middlewares are executed in the order specified
func Chain(handler lambda.Handler, middlewares ...Middleware) lambda.Handler {
	if len(middlewares) == 0 {
		return handler
	}

	chain := middlewares[len(middlewares)-1](handler)

	// Middlewares must be wrapped in reverse order, to ensure first middleware is invoked first.
	for i := len(middlewares) - 2; i >= 0; i-- {
		chain = middlewares[i](chain)
	}

	return chain
}
