package services

import (
	"context"
	"testing"

	"github.com/secretlyhq/secretly/internal/i18n"
	"github.com/secretlyhq/secretly/server/grpc/interceptors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Helper function to create context with user
func createUserContext(userCtx *interceptors.UserContext) context.Context {
	if userCtx == nil {
		return context.Background()
	}
	return context.WithValue(context.Background(), interceptors.GetUserContextKey(), userCtx)
}

func TestNewSecretService(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	service, err := NewSecretService()
	require.NoError(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.secretService)
}

func TestSecretGRPCService_CreateSecret(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	service, err := NewSecretService()
	require.NoError(t, err)

	tests := []struct {
		name           string
		userCtx        *interceptors.UserContext
		request        *CreateSecretRequest
		expectedError  codes.Code
		expectResponse bool
	}{
		{
			name: "successful creation with admin user",
			userCtx: &interceptors.UserContext{
				UserID:   1,
				Username: "admin",
				Permissions: []string{"secrets.write"},
			},
			request: &CreateSecretRequest{
				Name:        "test-secret",
				Value:       "secret-value",
				Namespace:   "default",
				Zone:        "us-east-1",
				Environment: "dev",
				Type:        "password",
				Metadata: map[string]string{
					"owner": "test-user",
				},
				Tags: []string{"test", "development"},
			},
			expectedError:  codes.OK,
			expectResponse: true,
		},
		{
			name:           "unauthenticated user",
			userCtx:        nil,
			request:        &CreateSecretRequest{},
			expectedError:  codes.Unauthenticated,
			expectResponse: false,
		},
		{
			name: "insufficient permissions",
			userCtx: &interceptors.UserContext{
				UserID:   2,
				Username: "user",
				Permissions: []string{"secrets.read"}, // missing secrets.write
			},
			request: &CreateSecretRequest{
				Name:        "test-secret",
				Value:       "secret-value",
				Namespace:   "default",
				Zone:        "us-east-1",
				Environment: "dev",
			},
			expectedError:  codes.PermissionDenied,
			expectResponse: false,
		},
		{
			name: "missing required fields",
			userCtx: &interceptors.UserContext{
				UserID:   1,
				Username: "admin",
				Permissions: []string{"secrets.write"},
			},
			request: &CreateSecretRequest{
				Name:  "", // missing name
				Value: "secret-value",
			},
			expectedError:  codes.InvalidArgument,
			expectResponse: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create context with user
			ctx := createUserContext(tt.userCtx)

			// Call service method
			response, err := service.CreateSecret(ctx, tt.request)

			// Check error code
			if tt.expectedError != codes.OK {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedError, st.Code())
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				if tt.expectResponse {
					require.NotNil(t, response)
					assert.NotZero(t, response.Id)
					assert.Equal(t, tt.request.Name, response.Name)
					assert.Equal(t, tt.request.Namespace, response.Namespace)
					assert.Equal(t, tt.request.Zone, response.Zone)
					assert.Equal(t, tt.request.Environment, response.Environment)
				}
			}
		})
	}
}

func TestSecretGRPCService_GetSecret(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	service, err := NewSecretService()
	require.NoError(t, err)

	tests := []struct {
		name           string
		userCtx        *interceptors.UserContext
		request        *GetSecretRequest
		expectedError  codes.Code
		expectResponse bool
	}{
		{
			name: "successful get with admin user",
			userCtx: &interceptors.UserContext{
				UserID:   1,
				Username: "admin",
				Permissions: []string{"secrets.read"},
			},
			request: &GetSecretRequest{
				Id:           1,
				IncludeValue: false,
			},
			expectedError:  codes.OK,
			expectResponse: true,
		},
		{
			name:           "unauthenticated user",
			userCtx:        nil,
			request:        &GetSecretRequest{Id: 1},
			expectedError:  codes.Unauthenticated,
			expectResponse: false,
		},
		{
			name: "insufficient permissions",
			userCtx: &interceptors.UserContext{
				UserID:   2,
				Username: "user",
				Permissions: []string{"users.read"}, // missing secrets.read
			},
			request: &GetSecretRequest{Id: 1},
			expectedError:  codes.PermissionDenied,
			expectResponse: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create context with user
			ctx := createUserContext(tt.userCtx)

			// Call service method
			response, err := service.GetSecret(ctx, tt.request)

			// Check error code
			if tt.expectedError != codes.OK {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedError, st.Code())
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				if tt.expectResponse {
					require.NotNil(t, response)
					assert.Equal(t, tt.request.Id, response.Id)
					assert.NotEmpty(t, response.Name)
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkSecretGRPCService_CreateSecret(b *testing.B) {
	service, err := NewSecretService()
	require.NoError(b, err)

	userCtx := &interceptors.UserContext{
		UserID:   1,
		Username: "admin",
		Permissions: []string{"secrets.write"},
	}

	ctx := createUserContext(userCtx)

	request := &CreateSecretRequest{
		Name:        "benchmark-secret",
		Value:       "secret-value",
		Namespace:   "default",
		Zone:        "us-east-1",
		Environment: "dev",
		Type:        "password",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.CreateSecret(ctx, request)
	}
}

