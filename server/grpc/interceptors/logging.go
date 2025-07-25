package interceptors

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

// LoggingInterceptor returns a unary server interceptor for logging
func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Get client info
		clientIP := "unknown"
		if p, ok := peer.FromContext(ctx); ok {
			clientIP = p.Addr.String()
		}

		// Call the handler
		resp, err := handler(ctx, req)

		// Log the request
		duration := time.Since(start)
		status := "OK"
		if err != nil {
			status = "ERROR"
		}

		log.Printf(
			"gRPC [%s] %s - %s - %v - %s",
			info.FullMethod,
			status,
			duration,
			clientIP,
			getErrorMessage(err),
		)

		// Log slow requests
		if duration > 1*time.Second {
			log.Printf("SLOW gRPC REQUEST: %s took %v", info.FullMethod, duration)
		}

		return resp, err
	}
}

// StreamLoggingInterceptor returns a stream server interceptor for logging
func StreamLoggingInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		// Get client info
		clientIP := "unknown"
		if p, ok := peer.FromContext(stream.Context()); ok {
			clientIP = p.Addr.String()
		}

		// Call the handler
		err := handler(srv, stream)

		// Log the stream
		duration := time.Since(start)
		status := "OK"
		if err != nil {
			status = "ERROR"
		}

		log.Printf(
			"gRPC STREAM [%s] %s - %v - %s - %s",
			info.FullMethod,
			status,
			duration,
			clientIP,
			getErrorMessage(err),
		)

		return err
	}
}

// getErrorMessage extracts error message safely
func getErrorMessage(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
