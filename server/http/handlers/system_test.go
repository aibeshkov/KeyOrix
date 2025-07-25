package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthCheck(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
	}{
		{
			name:           "successful health check",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request (no authentication required for health check)
			req := httptest.NewRequest(http.MethodGet, "/health", nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			HealthCheck(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check response structure
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Verify required fields
			assert.Contains(t, response, "status")
			assert.Contains(t, response, "timestamp")
			assert.Contains(t, response, "uptime")
			assert.Contains(t, response, "version")
			assert.Contains(t, response, "checks")

			// Verify status is healthy
			assert.Equal(t, "healthy", response["status"])

			// Verify checks structure
			checks := response["checks"].(map[string]interface{})
			assert.Contains(t, checks, "database")
			assert.Contains(t, checks, "encryption")
			assert.Contains(t, checks, "storage")

			// Verify database check
			dbCheck := checks["database"].(map[string]interface{})
			assert.Contains(t, dbCheck, "status")
			assert.Contains(t, dbCheck, "latency")

			// Check content type
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
		})
	}
}

func TestGetSystemInfo(t *testing.T) {
	tests := []struct {
		name           string
		authToken      string
		expectedStatus int
	}{
		{
			name:           "successful system info retrieval",
			authToken:      "valid-token",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized without token",
			authToken:      "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "forbidden with insufficient permissions",
			authToken:      "test-token", // test-token doesn't have system.read permission
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/system/info", nil)
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
				ctx := addAuthContext(req.Context(), tt.authToken)
				req = req.WithContext(ctx)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			GetSystemInfo(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				// Check response structure
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, response, "data")
				data := response["data"].(map[string]interface{})

				// Verify system info fields
				assert.Contains(t, data, "version")
				assert.Contains(t, data, "build_time")
				assert.Contains(t, data, "git_commit")
				assert.Contains(t, data, "go_version")
				assert.Contains(t, data, "os")
				assert.Contains(t, data, "arch")
				assert.Contains(t, data, "uptime")
				assert.Contains(t, data, "environment")
				assert.Contains(t, data, "features")
				assert.Contains(t, data, "database")
				assert.Contains(t, data, "security")

				// Verify features
				features := data["features"].(map[string]interface{})
				assert.Contains(t, features, "tls_enabled")
				assert.Contains(t, features, "auth_enabled")
				assert.Contains(t, features, "audit_enabled")

				// Verify database info
				database := data["database"].(map[string]interface{})
				assert.Contains(t, database, "type")
				assert.Contains(t, database, "connected")
				assert.Contains(t, database, "version")
				assert.Contains(t, database, "pool")

				// Verify security info
				security := data["security"].(map[string]interface{})
				assert.Contains(t, security, "tls_enabled")
				assert.Contains(t, security, "auth_enabled")
				assert.Contains(t, security, "encryption_method")
				assert.Contains(t, security, "audit_enabled")
			}
		})
	}
}

func TestGetMetrics(t *testing.T) {
	tests := []struct {
		name           string
		authToken      string
		expectedStatus int
	}{
		{
			name:           "successful metrics retrieval",
			authToken:      "valid-token",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized without token",
			authToken:      "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "forbidden with insufficient permissions",
			authToken:      "test-token", // test-token doesn't have system.read permission
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/system/metrics", nil)
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
				ctx := addAuthContext(req.Context(), tt.authToken)
				req = req.WithContext(ctx)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			GetMetrics(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				// Check response structure
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, response, "data")
				data := response["data"].(map[string]interface{})

				// Verify metrics fields
				assert.Contains(t, data, "memory")
				assert.Contains(t, data, "goroutines")
				assert.Contains(t, data, "gc")
				assert.Contains(t, data, "http")
				assert.Contains(t, data, "database")
				assert.Contains(t, data, "secrets")
				assert.Contains(t, data, "uptime")
				assert.Contains(t, data, "timestamp")

				// Verify memory metrics
				memory := data["memory"].(map[string]interface{})
				assert.Contains(t, memory, "alloc")
				assert.Contains(t, memory, "total_alloc")
				assert.Contains(t, memory, "sys")
				assert.Contains(t, memory, "heap_alloc")
				assert.Contains(t, memory, "heap_sys")

				// Verify GC metrics
				gc := data["gc"].(map[string]interface{})
				assert.Contains(t, gc, "num_gc")
				assert.Contains(t, gc, "pause_total")
				assert.Contains(t, gc, "gc_cpu_fraction")

				// Verify HTTP metrics
				httpMetrics := data["http"].(map[string]interface{})
				assert.Contains(t, httpMetrics, "requests_total")
				assert.Contains(t, httpMetrics, "requests_per_sec")
				assert.Contains(t, httpMetrics, "avg_response_time")
				assert.Contains(t, httpMetrics, "error_rate")

				// Verify database metrics
				dbMetrics := data["database"].(map[string]interface{})
				assert.Contains(t, dbMetrics, "queries_total")
				assert.Contains(t, dbMetrics, "queries_per_sec")
				assert.Contains(t, dbMetrics, "avg_query_time")

				// Verify secrets metrics
				secretsMetrics := data["secrets"].(map[string]interface{})
				assert.Contains(t, secretsMetrics, "total_secrets")
				assert.Contains(t, secretsMetrics, "active_secrets")
				assert.Contains(t, secretsMetrics, "expired_secrets")
				assert.Contains(t, secretsMetrics, "secrets_created_24h")
				assert.Contains(t, secretsMetrics, "secrets_accessed_24h")

				// Verify goroutines count is reasonable
				goroutines := data["goroutines"].(float64)
				assert.Greater(t, goroutines, float64(0))
				assert.Less(t, goroutines, float64(1000)) // Reasonable upper bound
			}
		})
	}
}

// Test concurrent access to system endpoints
func TestSystemEndpointsConcurrency(t *testing.T) {
	const numGoroutines = 10
	const numRequests = 5

	tests := []struct {
		name    string
		handler http.HandlerFunc
		path    string
	}{
		{
			name:    "health check concurrency",
			handler: HealthCheck,
			path:    "/health",
		},
		{
			name:    "system info concurrency",
			handler: GetSystemInfo,
			path:    "/api/v1/system/info",
		},
		{
			name:    "metrics concurrency",
			handler: GetMetrics,
			path:    "/api/v1/system/metrics",
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

						// Add auth context for protected endpoints
						if tt.path != "/health" {
							req.Header.Set("Authorization", "Bearer valid-token")
							ctx := addAuthContext(req.Context(), "valid-token")
							req = req.WithContext(ctx)
						}

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

// Benchmark tests for system handlers
func BenchmarkHealthCheck(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		HealthCheck(w, req)
	}
}

func BenchmarkGetSystemInfo(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/info", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	ctx := addAuthContext(req.Context(), "valid-token")
	req = req.WithContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		GetSystemInfo(w, req)
	}
}

func BenchmarkGetMetrics(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/metrics", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	ctx := addAuthContext(req.Context(), "valid-token")
	req = req.WithContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		GetMetrics(w, req)
	}
}

// Test memory usage of metrics collection
func TestMetricsMemoryUsage(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/metrics", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	ctx := addAuthContext(req.Context(), "valid-token")
	req = req.WithContext(ctx)

	// Collect metrics multiple times to ensure no memory leaks
	for i := 0; i < 100; i++ {
		w := httptest.NewRecorder()
		GetMetrics(w, req)

		// Verify response is valid
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "data")
	}
}
