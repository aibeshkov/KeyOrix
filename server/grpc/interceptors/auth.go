package interceptors

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UserContext represents the authenticated user context for gRPC
type UserContext struct {
	UserID      uint     `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

// contextKey is used for context keys to avoid collisions
type contextKey string

const (
	userContextKey contextKey = "grpc_user"
)

// AuthInterceptor returns a unary server interceptor for authentication
func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip authentication for certain methods
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// Extract and validate token
		userCtx, err := authenticateRequest(ctx)
		if err != nil {
			return nil, err
		}

		// Add user context to request context
		newCtx := context.WithValue(ctx, userContextKey, userCtx)
		return handler(newCtx, req)
	}
}

// StreamAuthInterceptor returns a stream server interceptor for authentication
func StreamAuthInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Skip authentication for certain methods
		if isPublicMethod(info.FullMethod) {
			return handler(srv, stream)
		}

		// Extract and validate token
		userCtx, err := authenticateRequest(stream.Context())
		if err != nil {
			return err
		}

		// Create a new stream with user context
		wrappedStream := &wrappedServerStream{
			ServerStream: stream,
			ctx:          context.WithValue(stream.Context(), userContextKey, userCtx),
		}

		return handler(srv, wrappedStream)
	}
}

// wrappedServerStream wraps grpc.ServerStream to override context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// authenticateRequest extracts and validates the authentication token
func authenticateRequest(ctx context.Context) (*UserContext, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "Missing metadata")
	}

	// Extract authorization header
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "Missing authorization header")
	}

	authHeader := authHeaders[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid authorization header format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return nil, status.Errorf(codes.Unauthenticated, "Missing token")
	}

	// TODO: Validate JWT token and extract user information
	// For now, we'll use a mock implementation
	userCtx, err := validateGRPCToken(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid or expired token")
	}

	return userCtx, nil
}

// validateGRPCToken validates a JWT token and returns user context
// TODO: Implement actual JWT validation
func validateGRPCToken(token string) (*UserContext, error) {
	// Mock implementation - replace with actual JWT validation
	if token == "valid-token" {
		return &UserContext{
			UserID:   1,
			Username: "admin",
			Email:    "admin@example.com",
			Roles:    []string{"admin", "user"},
			Permissions: []string{
				"secrets.read", "secrets.write", "secrets.delete",
				"users.read", "users.write", "users.delete",
				"roles.read", "roles.write", "roles.assign",
				"audit.read", "system.read",
			},
		}, nil
	}

	if token == "test-token" {
		return &UserContext{
			UserID:   2,
			Username: "testuser",
			Email:    "test@example.com",
			Roles:    []string{"viewer"},
			Permissions: []string{
				"secrets.read",
				"users.read",
			},
		}, nil
	}

	return nil, status.Errorf(codes.Unauthenticated, "Invalid token")
}

// isPublicMethod checks if a gRPC method is public (doesn't require authentication)
func isPublicMethod(method string) bool {
	publicMethods := []string{
		"/grpc.health.v1.Health/Check",
		"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
		// Add other public methods here
	}

	for _, publicMethod := range publicMethods {
		if method == publicMethod {
			return true
		}
	}

	return false
}

// GetUserFromGRPCContext extracts the user context from the gRPC request context
func GetUserFromGRPCContext(ctx context.Context) *UserContext {
	if userCtx, ok := ctx.Value(userContextKey).(*UserContext); ok {
		return userCtx
	}
	return nil
}

// GetUserContextKey returns the context key for user context (for testing)
func GetUserContextKey() contextKey {
	return userContextKey
}
