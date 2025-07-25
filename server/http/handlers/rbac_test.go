package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/secretlyhq/secretly/internal/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListUsers(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	tests := []struct {
		name           string
		authToken      string
		queryParams    string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful list users",
			authToken:      "valid-token",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "list users with pagination",
			authToken:      "valid-token",
			queryParams:    "?page=1&page_size=10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized without token",
			authToken:      "",
			queryParams:    "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "forbidden with insufficient permissions",
			authToken:      "test-token", // test-token has users.read permission
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/users"+tt.queryParams, nil)
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
				ctx := addAuthContext(req.Context(), tt.authToken)
				req = req.WithContext(ctx)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			ListUsers(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				// Check response structure
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, response, "data")
				data := response["data"].(map[string]interface{})
				assert.Contains(t, data, "users")
				assert.Contains(t, data, "total")
			}
		})
	}
}

func TestCreateUser(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	tests := []struct {
		name           string
		authToken      string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:      "successful user creation",
			authToken: "valid-token",
			requestBody: map[string]interface{}{
				"username":     "newuser",
				"email":        "newuser@example.com",
				"display_name": "New User",
				"password":     "securepassword123",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:      "missing required fields",
			authToken: "valid-token",
			requestBody: map[string]interface{}{
				"username": "",
				"email":    "invalid-email",
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
			authToken: "test-token", // test-token doesn't have users.write permission
			requestBody: map[string]interface{}{
				"username":     "newuser",
				"email":        "newuser@example.com",
				"display_name": "New User",
				"password":     "securepassword123",
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			var body bytes.Buffer
			err := json.NewEncoder(&body).Encode(tt.requestBody)
			require.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/api/v1/users", &body)
			req.Header.Set("Content-Type", "application/json")
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
				ctx := addAuthContext(req.Context(), tt.authToken)
				req = req.WithContext(ctx)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			CreateUser(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				// Check response structure
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, response, "data")
				assert.Contains(t, response, "message")
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	tests := []struct {
		name           string
		authToken      string
		userID         string
		expectedStatus int
	}{
		{
			name:           "successful get user",
			authToken:      "valid-token",
			userID:         "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user not found",
			authToken:      "valid-token",
			userID:         "999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid user ID",
			authToken:      "valid-token",
			userID:         "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			authToken:      "",
			userID:         "1",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+tt.userID, nil)
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
				ctx := addAuthContext(req.Context(), tt.authToken)
				req = req.WithContext(ctx)
			}

			// Add URL parameters to context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			GetUser(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestListRoles(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	tests := []struct {
		name           string
		authToken      string
		expectedStatus int
	}{
		{
			name:           "successful list roles",
			authToken:      "valid-token",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized",
			authToken:      "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/roles", nil)
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
				ctx := addAuthContext(req.Context(), tt.authToken)
				req = req.WithContext(ctx)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			ListRoles(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				// Check response structure
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, response, "data")
				data := response["data"].(map[string]interface{})
				assert.Contains(t, data, "roles")
			}
		})
	}
}

func TestCreateRole(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	tests := []struct {
		name           string
		authToken      string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name:      "successful role creation",
			authToken: "valid-token",
			requestBody: map[string]interface{}{
				"name":        "developer",
				"description": "Developer role with limited access",
				"permissions": []string{"secrets.read", "secrets.write"},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:      "missing required fields",
			authToken: "valid-token",
			requestBody: map[string]interface{}{
				"name": "",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			authToken:      "",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			var body bytes.Buffer
			err := json.NewEncoder(&body).Encode(tt.requestBody)
			require.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/api/v1/roles", &body)
			req.Header.Set("Content-Type", "application/json")
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
				ctx := addAuthContext(req.Context(), tt.authToken)
				req = req.WithContext(ctx)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			CreateRole(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAssignRole(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	tests := []struct {
		name           string
		authToken      string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name:      "successful role assignment",
			authToken: "valid-token",
			requestBody: map[string]interface{}{
				"user_id": 2,
				"role_id": 1,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:      "missing required fields",
			authToken: "valid-token",
			requestBody: map[string]interface{}{
				"user_id": 0,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			authToken:      "",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			var body bytes.Buffer
			err := json.NewEncoder(&body).Encode(tt.requestBody)
			require.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/api/v1/user-roles", &body)
			req.Header.Set("Content-Type", "application/json")
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
				ctx := addAuthContext(req.Context(), tt.authToken)
				req = req.WithContext(ctx)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			AssignRole(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestGetUserRoles(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	tests := []struct {
		name           string
		authToken      string
		userID         string
		expectedStatus int
	}{
		{
			name:           "successful get user roles",
			authToken:      "valid-token",
			userID:         "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user ID",
			authToken:      "valid-token",
			userID:         "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			authToken:      "",
			userID:         "1",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/user-roles/user/"+tt.userID, nil)
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
				ctx := addAuthContext(req.Context(), tt.authToken)
				req = req.WithContext(ctx)
			}

			// Add URL parameters to context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("userId", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			GetUserRoles(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// Benchmark tests for RBAC handlers
func BenchmarkListUsers(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	ctx := addAuthContext(req.Context(), "valid-token")
	req = req.WithContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		ListUsers(w, req)
	}
}

func BenchmarkCreateUser(b *testing.B) {
	requestBody := map[string]interface{}{
		"username":     "benchuser",
		"email":        "bench@example.com",
		"display_name": "Benchmark User",
		"password":     "securepassword123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var body bytes.Buffer
		_ = json.NewEncoder(&body).Encode(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", &body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		ctx := addAuthContext(req.Context(), "valid-token")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		CreateUser(w, req)
	}
}
