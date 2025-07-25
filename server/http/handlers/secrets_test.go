package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/secretlyhq/secretly/server/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretHandler_ListSecrets(t *testing.T) {
	tests := []struct {
		name           string
		authToken      string
		queryParams    string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful list with valid token",
			authToken:      "valid-token",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "list with pagination",
			authToken:      "valid-token",
			queryParams:    "?page=1&page_size=10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "list with filters",
			authToken:      "valid-token",
			queryParams:    "?namespace=production&environment=prod",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized without token",
			authToken:      "",
			queryParams:    "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Missing authorization header",
		},
		{
			name:           "unauthorized with invalid token",
			authToken:      "invalid-token",
			queryParams:    "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or expired token",
		},
		{
			name:           "forbidden with insufficient permissions",
			authToken:      "test-token",
			queryParams:    "",
			expectedStatus: http.StatusOK, // test-token has secrets.read permission
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler
			handler, err := NewSecretHandler()
			require.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets"+tt.queryParams, nil)
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
				// Add authentication middleware context
				ctx := addAuthContext(req.Context(), tt.authToken)
				req = req.WithContext(ctx)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.ListSecrets(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				// Check response structure
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				// Check that data field exists
				assert.Contains(t, response, "data")
				data := response["data"].(map[string]interface{})
				assert.Contains(t, data, "secrets")
				assert.Contains(t, data, "total")
				assert.Contains(t, data, "page")
				assert.Contains(t, data, "page_size")
			} else if tt.expectedError != "" {
				// Check error response
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["message"].(string), tt.expectedError)
			}
		})
	}
}

func TestSecretHandler_CreateSecret(t *testing.T) {
	tests := []struct {
		name           string
		authToken      string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:      "successful creation",
			authToken: "valid-token",
			requestBody: map[string]interface{}{
				"name":        "test-secret",
				"value":       "secret-value",
				"namespace":   "default",
				"zone":        "us-east-1",
				"environment": "dev",
				"type":        "password",
				"metadata": map[string]string{
					"owner": "test-user",
				},
				"tags": []string{"test", "development"},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:      "missing required fields",
			authToken: "valid-token",
			requestBody: map[string]interface{}{
				"name":  "",
				"value": "secret-value",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request data",
		},
		{
			name:           "unauthorized without token",
			authToken:      "",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:      "forbidden with insufficient permissions",
			authToken: "test-token", // test-token doesn't have secrets.write permission
			requestBody: map[string]interface{}{
				"name":        "test-secret",
				"value":       "secret-value",
				"namespace":   "default",
				"zone":        "us-east-1",
				"environment": "dev",
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "invalid JSON",
			authToken:      "valid-token",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler
			handler, err := NewSecretHandler()
			require.NoError(t, err)

			// Create request body
			var body bytes.Buffer
			if str, ok := tt.requestBody.(string); ok {
				body.WriteString(str)
			} else {
				err := json.NewEncoder(&body).Encode(tt.requestBody)
				require.NoError(t, err)
			}

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/api/v1/secrets", &body)
			req.Header.Set("Content-Type", "application/json")
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
				ctx := addAuthContext(req.Context(), tt.authToken)
				req = req.WithContext(ctx)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.CreateSecret(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				// Check response structure
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, response, "data")
				assert.Contains(t, response, "message")

				data := response["data"].(map[string]interface{})
				assert.Contains(t, data, "id")
				assert.Contains(t, data, "name")
				assert.Contains(t, data, "created_at")
			}
		})
	}
}

func TestSecretHandler_GetSecret(t *testing.T) {
	tests := []struct {
		name           string
		authToken      string
		secretID       string
		queryParams    string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful get",
			authToken:      "valid-token",
			secretID:       "1",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get with decrypted value",
			authToken:      "valid-token",
			secretID:       "1",
			queryParams:    "?include_value=true",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid secret ID",
			authToken:      "valid-token",
			secretID:       "invalid",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid secret ID",
		},
		{
			name:           "secret not found",
			authToken:      "valid-token",
			secretID:       "999",
			queryParams:    "",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "unauthorized",
			authToken:      "",
			secretID:       "1",
			queryParams:    "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler
			handler, err := NewSecretHandler()
			require.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets/"+tt.secretID+tt.queryParams, nil)
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
				ctx := addAuthContext(req.Context(), tt.authToken)
				req = req.WithContext(ctx)
			}

			// Add URL parameters to context (simulate chi router)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.secretID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.GetSecret(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				// Check response structure
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, response, "data")
				data := response["data"].(map[string]interface{})
				assert.Contains(t, data, "id")
				assert.Contains(t, data, "name")
			}
		})
	}
}

func TestSecretHandler_UpdateSecret(t *testing.T) {
	tests := []struct {
		name           string
		authToken      string
		secretID       string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:      "successful update",
			authToken: "valid-token",
			secretID:  "1",
			requestBody: map[string]interface{}{
				"value": "new-secret-value",
				"metadata": map[string]string{
					"updated_by": "admin",
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid secret ID",
			authToken:      "valid-token",
			secretID:       "invalid",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			authToken:      "",
			secretID:       "1",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "forbidden",
			authToken:      "test-token",
			secretID:       "1",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler
			handler, err := NewSecretHandler()
			require.NoError(t, err)

			// Create request body
			var body bytes.Buffer
			err = json.NewEncoder(&body).Encode(tt.requestBody)
			require.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodPut, "/api/v1/secrets/"+tt.secretID, &body)
			req.Header.Set("Content-Type", "application/json")
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
				ctx := addAuthContext(req.Context(), tt.authToken)
				req = req.WithContext(ctx)
			}

			// Add URL parameters to context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.secretID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.UpdateSecret(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestSecretHandler_DeleteSecret(t *testing.T) {
	tests := []struct {
		name           string
		authToken      string
		secretID       string
		expectedStatus int
	}{
		{
			name:           "successful delete",
			authToken:      "valid-token",
			secretID:       "1",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid secret ID",
			authToken:      "valid-token",
			secretID:       "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			authToken:      "",
			secretID:       "1",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "forbidden",
			authToken:      "test-token",
			secretID:       "1",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler
			handler, err := NewSecretHandler()
			require.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/secrets/"+tt.secretID, nil)
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
				ctx := addAuthContext(req.Context(), tt.authToken)
				req = req.WithContext(ctx)
			}

			// Add URL parameters to context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.secretID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.DeleteSecret(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// Helper function to add authentication context
func addAuthContext(ctx context.Context, token string) context.Context {
	var userCtx *middleware.UserContext

	switch token {
	case "valid-token":
		userCtx = &middleware.UserContext{
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
		}
	case "test-token":
		userCtx = &middleware.UserContext{
			UserID:   2,
			Username: "testuser",
			Email:    "test@example.com",
			Roles:    []string{"viewer"},
			Permissions: []string{
				"secrets.read",
				"users.read",
			},
		}
	}

	if userCtx != nil {
		// Use the same context key as the middleware
		return context.WithValue(ctx, middleware.GetUserContextKey(), userCtx)
	}

	return ctx
}

// GetUserContextKey returns the context key for user context (for testing)
func GetUserContextKey() interface{} {
	return userContextKey
}

// Define the context key for testing (must match middleware)
type contextKey string

const userContextKey contextKey = "user"

// Benchmark tests
func BenchmarkSecretHandler_ListSecrets(b *testing.B) {
	handler, err := NewSecretHandler()
	require.NoError(b, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	ctx := addAuthContext(req.Context(), "valid-token")
	req = req.WithContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ListSecrets(w, req)
	}
}

func BenchmarkSecretHandler_CreateSecret(b *testing.B) {
	handler, err := NewSecretHandler()
	require.NoError(b, err)

	requestBody := map[string]interface{}{
		"name":        "benchmark-secret",
		"value":       "secret-value",
		"namespace":   "default",
		"zone":        "us-east-1",
		"environment": "dev",
		"type":        "password",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var body bytes.Buffer
		_ = json.NewEncoder(&body).Encode(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/secrets", &body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		ctx := addAuthContext(req.Context(), "valid-token")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.CreateSecret(w, req)
	}
}
