package lambdamiddleware_test

import (
	"context"
	"log"

	"github.com/easyCZ/aws-lambda-middleware"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

type HandlerPayload struct {
	Value string `json:"value"`
}

func ExampleMiddleware() {
	logger := log.Default()
	myLambdaHandler := func(ctx context.Context, payload HandlerPayload) error {
		logger.Printf("Handling lambda invocation for payload: %v\n", payload)

		return nil
	}

	handler := lambda.NewHandler(myLambdaHandler)
	handler = lambdamiddleware.Chain(handler, []lambdamiddleware.Middleware{
		LoggingMiddleware(logger),
		ContextExtendMiddleware,
	}...)

	// start your lambda
	lambda.StartHandler(handler)
}

// LoggingMiddleware is a constructor, it allows us to inject dependencies into this middleware.
func LoggingMiddleware(logger *log.Logger) lambdamiddleware.Middleware {
	// We return a Middleware which has access to our logger in a closure
	return func(next lambda.Handler) lambda.Handler {

		// Middleware must return a handler. We use the HandlerFunc helper to write a handler as a single function.
		return lambdamiddleware.HandlerFunc(func(ctx context.Context, bytes []byte) ([]byte, error) {
			// Extract invocation information from the context. This is populated from the lambda runtime and the internal lambda.Handler implementation.
			lctx, _ := lambdacontext.FromContext(ctx)

			// Print the request body, and the request ID.
			// You may want to attach this as structured fields to your *smart* logger.
			logger.Printf("Handing lambda invocation. Request ID: %s. Body: %s\n", lctx.AwsRequestID, string(bytes))

			// Trigger the next middleware in the chain. This may be the actual handler, or another Middleware.
			response, err := next.Invoke(ctx, bytes)

			// We have access to the response, we can choose to log it.
			if err != nil {
				logger.Printf("Finished lambda invocation with error: %s. Request ID: %s. Response: %s.", err.Error(), lctx.AwsRequestID, string(response))
			} else {
				logger.Printf("Finished lambda invocation with OK. Request ID: %s. Response: %s.", lctx.AwsRequestID, string(response))
			}

			// We return the response and the error back to the Middleware before us, or to the AWS Runtime.
			return response, err
		})
	}
}

// ContextExtendMiddleware is a middleware which extends the context with a value
// ContextExtendMiddleware is an example where we do not use a constructor to contruct the middleware, instead we implement the Middleware type directly.
func ContextExtendMiddleware(next lambda.Handler) lambda.Handler {
	// We return a HandlerFunc to get access to the request context.
	return lambdamiddleware.HandlerFunc(func(ctx context.Context, bytes []byte) ([]byte, error) {
		// Attach a custom value to the context.
		ctx = context.WithValue(ctx, "my_key", "my_value")

		// Directly invoke the next handler, without capturing the response and error - we do not need them in this case.
		return next.Invoke(ctx, bytes)
	})
}
