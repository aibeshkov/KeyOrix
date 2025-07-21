package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/secretlyhq/secretly/internal/config"
	"github.com/secretlyhq/secretly/internal/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests for the complete HTTP server
func TestHTTPServerIntegration(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	// Create test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			HTTP: config.ServerInstanceConfig{
				Enabled:        true,
				Port:           "8080",
				SwaggerEnabled: true,
			},
		},
	}

	// Create router
	router, err := NewRouter(cfg)
	require.NoError(t, err)

	// Create test server
	server := httptest.NewServer(router)
	defer server.Close()

	// Test cases for complete workflow
	t.Run("Complete Secret Management Workflow", func(t *testing.T) {
		client := &http.Client{Timeout: 10 * time.Second}
		baseURL := server.URL

		// Step 1: Health check (no auth required)
		t.Run("Health Check", func(t *testing.T) {
			resp, err := client.Get(baseURL + "/health")
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

			var health map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&health)
			require.NoError(t, err)

			assert.Equal(t, "healthy", health["status"])
			assert.Contains(t, health, "timestamp")
			assert.Contains(t, health, "uptime")
		})

		// Step 2: Try to access protected endpoint without auth (should fail)
		t.Run("Unauthorized Access", func(t *testing.T) {
			resp, err := client.Get(baseURL + "/api/v1/secrets")
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})

		// Step 3: List secrets with valid auth
		var secretID uint
		t.Run("List Secrets", func(t *testing.T) {
			req, err := http.NewRequest("GET", baseURL+"/api/v1/secrets", nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer valid-token")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assert.Contains(t, response, "data")
			data := response["data"].(map[string]interface{})
			assert.Contains(t, data, "secrets")
			assert.Contains(t, data, "total")
		})

		// Step 4: Create a new secret
		t.Run("Create Secret", func(t *testing.T) {
			secretData := map[string]interface{}{
				"name":        "integration-test-secret",
				"value":       "super-secret-value",
				"namespace":   "test",
				"zone":        "us-west-2",
				"environment": "integration",
				"type":        "password",
				"metadata": map[string]string{
					"test":  "integration",
					"owner": "test-suite",
				},
				"tags": []string{"integration", "test", "automated"},
			}

			body, err := json.Marshal(secretData)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", baseURL+"/api/v1/secrets", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer valid-token")
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusCreated, resp.StatusCode)

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assert.Contains(t, response, "data")
			assert.Contains(t, response, "message")

			data := response["data"].(map[string]interface{})
			assert.Contains(t, data, "id")
			assert.Equal(t, "integration-test-secret", data["name"])
			assert.Equal(t, "test", data["namespace"])

			// Store secret ID for later tests
			secretID = uint(data["id"].(float64))
		})

		// Step 5: Get the created secret
		t.Run("Get Secret", func(t *testing.T) {
			req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/secrets/%d", baseURL, secretID), nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer valid-token")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assert.Contains(t, response, "data")
			data := response["data"].(map[string]interface{})
			assert.Equal(t, float64(secretID), data["id"])
			assert.Equal(t, "integration-test-secret", data["name"])
		})

		// Step 6: Update the secret
		t.Run("Update Secret", func(t *testing.T) {
			updateData := map[string]interface{}{
				"value": "updated-secret-value",
				"metadata": map[string]string{
					"test":       "integration",
					"owner":      "test-suite",
					"updated_by": "integration-test",
				},
			}

			body, err := json.Marshal(updateData)
			require.NoError(t, err)

			req, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/secrets/%d", baseURL, secretID), bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer valid-token")
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assert.Contains(t, response, "data")
			assert.Contains(t, response, "message")
		})

		// Step 7: Get secret versions
		t.Run("Get Secret Versions", func(t *testing.T) {
			req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/secrets/%d/versions", baseURL, secretID), nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer valid-token")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assert.Contains(t, response, "data")
			data := response["data"].(map[string]interface{})
			assert.Contains(t, data, "versions")
		})

		// Step 8: Delete the secret
		t.Run("Delete Secret", func(t *testing.T) {
			req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/secrets/%d", baseURL, secretID), nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer valid-token")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		})

		// Step 9: Verify secret is deleted
		t.Run("Verify Secret Deleted", func(t *testing.T) {
			req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/secrets/%d", baseURL, secretID), nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer valid-token")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		})
	})

	// Test RBAC workflow
	t.Run("RBAC Management Workflow", func(t *testing.T) {
		client := &http.Client{Timeout: 10 * time.Second}
		baseURL := server.URL

		// Test user management
		t.Run("User Management", func(t *testing.T) {
			// List users
			req, err := http.NewRequest("GET", baseURL+"/api/v1/users", nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer valid-token")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			// Create user
			userData := map[string]interface{}{
				"username":     "integration-user",
				"email":        "integration@test.com",
				"display_name": "Integration Test User",
				"password":     "securepassword123",
			}

			body, err := json.Marshal(userData)
			require.NoError(t, err)

			req, err = http.NewRequest("POST", baseURL+"/api/v1/users", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer valid-token")
			req.Header.Set("Content-Type", "application/json")

			resp, err = client.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		})

		// Test role management
		t.Run("Role Management", func(t *testing.T) {
			// List roles
			req, err := http.NewRequest("GET", baseURL+"/api/v1/roles", nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer valid-token")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			// Create role
			roleData := map[string]interface{}{
				"name":        "integration-role",
				"description": "Role created by integration test",
				"permissions": []string{"secrets.read", "users.read"},
			}

			body, err := json.Marshal(roleData)
			require.NoError(t, err)

			req, err = http.NewRequest("POST", baseURL+"/api/v1/roles", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer valid-token")
			req.Header.Set("Content-Type", "application/json")

			resp, err = client.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		})
	})

	// Test system endpoints
	t.Run("System Information", func(t *testing.T) {
		client := &http.Client{Timeout: 10 * time.Second}
		baseURL := server.URL

		// Get system info
		t.Run("System Info", func(t *testing.T) {
			req, err := http.NewRequest("GET", baseURL+"/api/v1/system/info", nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer valid-token")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assert.Contains(t, response, "data")
			data := response["data"].(map[string]interface{})
			assert.Contains(t, data, "version")
			assert.Contains(t, data, "go_version")
			assert.Contains(t, data, "features")
		})

		// Get metrics
		t.Run("System Metrics", func(t *testing.T) {
			req, err := http.NewRequest("GET", baseURL+"/api/v1/system/metrics", nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer valid-token")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assert.Contains(t, response, "data")
			data := response["data"].(map[string]interface{})
			assert.Contains(t, data, "memory")
			assert.Contains(t, data, "goroutines")
			assert.Contains(t, data, "gc")
		})
	})

	// Test audit endpoints
	t.Run("Audit Logs", func(t *testing.T) {
		client := &http.Client{Timeout: 10 * time.Second}
		baseURL := server.URL

		// Get audit logs
		t.Run("General Audit Logs", func(t *testing.T) {
			req, err := http.NewRequest("GET", baseURL+"/api/v1/audit/logs", nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer valid-token")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assert.Contains(t, response, "data")
			data := response["data"].(map[string]interface{})
			assert.Contains(t, data, "logs")
			assert.Contains(t, data, "total")
		})

		// Get RBAC audit logs
		t.Run("RBAC Audit Logs", func(t *testing.T) {
			req, err := http.NewRequest("GET", baseURL+"/api/v1/audit/rbac-logs", nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer valid-token")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assert.Contains(t, response, "data")
			data := response["data"].(map[string]interface{})
			assert.Contains(t, data, "logs")
			assert.Contains(t, data, "total")
		})
	})
}

// Test error scenarios
func TestHTTPServerErrorScenarios(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	cfg := &config.Config{
		Server: config.ServerConfig{
			HTTP: config.ServerInstanceConfig{
				Enabled: true,
				Port:    "8080",
			},
		},
	}

	router, err := NewRouter(cfg)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	baseURL := server.URL

	t.Run("Authentication Errors", func(t *testing.T) {
		// Missing authorization header
		resp, err := client.Get(baseURL + "/api/v1/secrets")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		// Invalid token
		req, err := http.NewRequest("GET", baseURL+"/api/v1/secrets", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer invalid-token")

		resp, err = client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		// Malformed authorization header
		req, err = http.NewRequest("GET", baseURL+"/api/v1/secrets", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "InvalidFormat token")

		resp, err = client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Authorization Errors", func(t *testing.T) {
		// Test user trying to delete (insufficient permissions)
		req, err := http.NewRequest("DELETE", baseURL+"/api/v1/secrets/1", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer test-token") // test-token lacks delete permission

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Validation Errors", func(t *testing.T) {
		// Invalid JSON
		req, err := http.NewRequest("POST", baseURL+"/api/v1/secrets", bytes.NewBufferString("{invalid json}"))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer valid-token")
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		// Missing required fields
		invalidSecret := map[string]interface{}{
			"name":  "", // empty name
			"value": "test",
		}

		body, err := json.Marshal(invalidSecret)
		require.NoError(t, err)

		req, err = http.NewRequest("POST", baseURL+"/api/v1/secrets", bytes.NewBuffer(body))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer valid-token")
		req.Header.Set("Content-Type", "application/json")

		resp, err = client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Not Found Errors", func(t *testing.T) {
		// Non-existent secret
		req, err := http.NewRequest("GET", baseURL+"/api/v1/secrets/99999", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer valid-token")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		// Non-existent endpoint
		req, err = http.NewRequest("GET", baseURL+"/api/v1/nonexistent", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer valid-token")

		resp, err = client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

// Performance and load testing
func TestHTTPServerPerformance(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	cfg := &config.Config{
		Server: config.ServerConfig{
			HTTP: config.ServerInstanceConfig{
				Enabled: true,
				Port:    "8080",
			},
		},
	}

	router, err := NewRouter(cfg)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("Concurrent Requests", func(t *testing.T) {
		const numGoroutines = 50
		const requestsPerGoroutine = 10

		results := make(chan int, numGoroutines*requestsPerGoroutine)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				client := &http.Client{Timeout: 10 * time.Second}
				for j := 0; j < requestsPerGoroutine; j++ {
					req, err := http.NewRequest("GET", server.URL+"/health", nil)
					if err != nil {
						results <- 0
						continue
					}

					resp, err := client.Do(req)
					if err != nil {
						results <- 0
						continue
					}
					_ = resp.Body.Close()
					results <- resp.StatusCode
				}
			}()
		}

		successCount := 0
		for i := 0; i < numGoroutines*requestsPerGoroutine; i++ {
			code := <-results
			if code == http.StatusOK {
				successCount++
			}
		}

		// At least 95% success rate
		expectedMinSuccess := int(float64(numGoroutines*requestsPerGoroutine) * 0.95)
		assert.GreaterOrEqual(t, successCount, expectedMinSuccess)
	})

	t.Run("Response Time", func(t *testing.T) {
		client := &http.Client{Timeout: 10 * time.Second}

		// Measure response time for health check
		start := time.Now()
		resp, err := client.Get(server.URL + "/health")
		duration := time.Since(start)

		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Less(t, duration, 100*time.Millisecond) // Should respond within 100ms
	})
}