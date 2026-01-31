package grpc

import (
	"context"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// authUnaryInterceptor adds authentication to unary RPC calls
func authUnaryInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Add API key to metadata
		ctx = addAuthMetadata(ctx)

		// Log the RPC call
		log.Printf("Invoking RPC: %s", method)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// authStreamInterceptor adds authentication to streaming RPC calls
func authStreamInterceptor() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		// Add API key to metadata
		ctx = addAuthMetadata(ctx)

		// Log the stream call
		log.Printf("Opening stream: %s", method)

		return streamer(ctx, desc, cc, method, opts...)
	}
}

// addAuthMetadata adds authentication metadata to the context
func addAuthMetadata(ctx context.Context) context.Context {
	// Get API key from environment or config
	apiKey := os.Getenv("ISPAGENT_API_KEY")
	if apiKey == "" {
		apiKey = "dev-api-key" // Fallback for development
	}

	md := metadata.New(map[string]string{
		"authorization": "Bearer " + apiKey,
		"user-agent":    "ispagent/1.0",
	})

	return metadata.NewOutgoingContext(ctx, md)
}
