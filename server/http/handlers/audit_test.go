package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/secretlyhq/secretly/internal/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAuditLogs(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	tests := []struct {
		name           string
		authToken      string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "successful audit logs retrieval",
			authToken:      "valid-token",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "audit logs with pagination",
			authToken:      "valid-token",
			queryParams:    "?page=1&page_size=10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "audit logs with filters",
			authToken:      "valid-token",
			queryParams:    "?action=CREATE_SECRET&resource=secret&user_id=1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "audit logs with time range",
			authToken:      "valid-token",
			queryParams:    "?start_time=2024-01-01T00:00:00Z&end_time=2024-12-31T23:59:59Z",
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
			authToken:      "test-token", // test-token doesn't have audit.read permission
			queryParams:    "",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/logs"+tt.queryParams, nil)
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
				ctx := addAuthContext(req.Context(), tt.authToken)
				req = req.WithContext(ctx)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			GetAuditLogs(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				// Check response structure
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, response, "data")
				data := response["data"].(map[string]interface{})

				// Verify audit logs structure
				assert.Contains(t, data, "logs")
				assert.Contains(t, data, "page")
				assert.Contains(t, data, "page_size")
				assert.Contains(t, data, "total")
				assert.Contains(t, data, "total_pages")
				assert.Contains(t, data, "filters")

				// Verify logs array
				logs := data["logs"].([]interface{})
				if len(logs) > 0 {
					log := logs[0].(map[string]interface{})
					assert.Contains(t, log, "id")
					assert.Contains(t, log, "user_id")
					assert.Contains(t, log, "username")
					assert.Contains(t, log, "action")
					assert.Contains(t, log, "resource")
					assert.Contains(t, log, "details")
					assert.Contains(t, log, "ip_address")
					assert.Contains(t, log, "user_agent")
					assert.Contains(t, log, "success")
					assert.Contains(t, log, "timestamp")
				}

				// Verify filters structure
				filters := data["filters"].(map[string]interface{})
				assert.Contains(t, filters, "action")
				assert.Contains(t, filters, "resource")
				assert.Contains(t, filters, "user_id")
				assert.Contains(t, filters, "start_time")
				assert.Contains(t, filters, "end_time")
			}
		})
	}
}

func TestGetRBACAuditLogs(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	tests := []struct {
		name           string
		authToken      string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "successful RBAC audit logs retrieval",
			authToken:      "valid-token",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "RBAC audit logs with pagination",
			authToken:      "valid-token",
			queryParams:    "?page=1&page_size=20",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "RBAC audit logs with filters",
			authToken:      "valid-token",
			queryParams:    "?action=CREATE_USER&target_type=user&user_id=1",
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
			authToken:      "test-token", // test-token doesn't have audit.read permission
			queryParams:    "",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/rbac-logs"+tt.queryParams, nil)
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
				ctx := addAuthContext(req.Context(), tt.authToken)
				req = req.WithContext(ctx)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			GetRBACAuditLogs(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				// Check response structure
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, response, "data")
				data := response["data"].(map[string]interface{})

				// Verify RBAC audit logs structure
				assert.Contains(t, data, "logs")
				assert.Contains(t, data, "page")
				assert.Contains(t, data, "page_size")
				assert.Contains(t, data, "total")
				assert.Contains(t, data, "total_pages")
				assert.Contains(t, data, "filters")

				// Verify logs array
				logs := data["logs"].([]interface{})
				if len(logs) > 0 {
					log := logs[0].(map[string]interface{})
					assert.Contains(t, log, "id")
					assert.Contains(t, log, "user_id")
					assert.Contains(t, log, "username")
					assert.Contains(t, log, "action")
					assert.Contains(t, log, "target_type")
					assert.Contains(t, log, "target_id")
					assert.Contains(t, log, "target_name")
					assert.Contains(t, log, "details")
					assert.Contains(t, log, "ip_address")
					assert.Contains(t, log, "success")
					assert.Contains(t, log, "timestamp")
				}

				// Verify filters structure
				filters := data["filters"].(map[string]interface{})
				assert.Contains(t, filters, "action")
				assert.Contains(t, filters, "target_type")
				assert.Contains(t, filters, "user_id")
			}
		})
	}
}

// Test query parameter parsing
func TestAuditLogsQueryParameterParsing(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	tests := []struct {
		name        string
		queryParams string
		description string
	}{
		{
			name:        "valid pagination parameters",
			queryParams: "?page=2&page_size=50",
			description: "Should parse pagination parameters correctly",
		},
		{
			name:        "invalid pagination parameters",
			queryParams: "?page=invalid&page_size=999",
			description: "Should handle invalid pagination gracefully",
		},
		{
			name:        "filter by action",
			queryParams: "?action=CREATE_SECRET",
			description: "Should filter by action",
		},
		{
			name:        "filter by resource",
			queryParams: "?resource=secret",
			description: "Should filter by resource type",
		},
		{
			name:        "filter by user ID",
			queryParams: "?user_id=123",
			description: "Should filter by user ID",
		},
		{
			name:        "filter by time range",
			queryParams: "?start_time=2024-01-01T00:00:00Z&end_time=2024-01-31T23:59:59Z",
			description: "Should filter by time range",
		},
		{
			name:        "invalid time format",
			queryParams: "?start_time=invalid-date&end_time=also-invalid",
			description: "Should handle invalid time formats gracefully",
		},
		{
			name:        "combined filters",
			queryParams: "?action=CREATE_SECRET&resource=secret&user_id=1&page=1&page_size=10",
			description: "Should handle multiple filters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/logs"+tt.queryParams, nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			ctx := addAuthContext(req.Context(), "valid-token")
			req = req.WithContext(ctx)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			GetAuditLogs(w, req)

			// Should always return 200 OK for valid auth
			assert.Equal(t, http.StatusOK, w.Code)

			// Check response structure
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Contains(t, response, "data")
			data := response["data"].(map[string]interface{})
			assert.Contains(t, data, "logs")
			assert.Contains(t, data, "filters")
		})
	}
}

// Test concurrent access to audit endpoints
func TestAuditEndpointsConcurrency(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	const numGoroutines = 10
	const numRequests = 5

	tests := []struct {
		name    string
		handler http.HandlerFunc
		path    string
	}{
		{
			name:    "audit logs concurrency",
			handler: GetAuditLogs,
			path:    "/api/v1/audit/logs",
		},
		{
			name:    "RBAC audit logs concurrency",
			handler: GetRBACAuditLogs,
			path:    "/api/v1/audit/rbac-logs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := make(chan int, numGoroutines*numRequests)

			// Launch multiple goroutines
			for i := 0; i < numGoroutines; i++ {
				go func() {
					for j := 0; j < numRequests; j++ {
						req := httptest.NewRequest(http.MethodGet, tt.path, nil)
						req.Header.Set("Authorization", "Bearer valid-token")
						ctx := addAuthContext(req.Context(), "valid-token")
						req = req.WithContext(ctx)

						w := httptest.NewRecorder()
						tt.handler(w, req)
						results <- w.Code
					}
				}()
			}

			// Collect results
			successCount := 0
			for i := 0; i < numGoroutines*numRequests; i++ {
				code := <-results
				if code == http.StatusOK {
					successCount++
				}
			}

			// All requests should succeed
			assert.Equal(t, numGoroutines*numRequests, successCount)
		})
	}
}

// Benchmark tests for audit handlers
func BenchmarkGetAuditLogs(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/logs", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	ctx := addAuthContext(req.Context(), "valid-token")
	req = req.WithContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		GetAuditLogs(w, req)
	}
}

func BenchmarkGetRBACAuditLogs(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/rbac-logs", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	ctx := addAuthContext(req.Context(), "valid-token")
	req = req.WithContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		GetRBACAuditLogs(w, req)
	}
}

// Test audit logs data consistency
func TestAuditLogsDataConsistency(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/logs", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	ctx := addAuthContext(req.Context(), "valid-token")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	GetAuditLogs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].(map[string]interface{})
	logs := data["logs"].([]interface{})
	total := int(data["total"].(float64))
	page := int(data["page"].(float64))
	pageSize := int(data["page_size"].(float64))
	totalPages := int(data["total_pages"].(float64))

	// Verify pagination consistency
	assert.GreaterOrEqual(t, total, len(logs))
	assert.GreaterOrEqual(t, page, 1)
	assert.GreaterOrEqual(t, pageSize, 1)
	assert.GreaterOrEqual(t, totalPages, 1)

	// Verify each log entry has required fields
	for _, logEntry := range logs {
		log := logEntry.(map[string]interface{})
		assert.Contains(t, log, "id")
		assert.Contains(t, log, "user_id")
		assert.Contains(t, log, "username")
		assert.Contains(t, log, "action")
		assert.Contains(t, log, "resource")
		assert.Contains(t, log, "success")
		assert.Contains(t, log, "timestamp")

		// Verify data types
		assert.IsType(t, float64(0), log["id"])
		assert.IsType(t, float64(0), log["user_id"])
		assert.IsType(t, "", log["username"])
		assert.IsType(t, "", log["action"])
		assert.IsType(t, "", log["resource"])
		assert.IsType(t, true, log["success"])
		assert.IsType(t, "", log["timestamp"])
	}
}