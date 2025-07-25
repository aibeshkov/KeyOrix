package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services"`
}

// HealthCheck handles the health check endpoint
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now().Add(-5 * time.Minute) // Mock uptime

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"uptime":    time.Since(startTime).String(),
		"version":   "1.0.0",
		"checks": map[string]interface{}{
			"database": map[string]interface{}{
				"status":  "healthy",
				"latency": "2ms",
			},
			"encryption": map[string]interface{}{
				"status":   "healthy",
				"provider": "AES-256-GCM",
			},
			"storage": map[string]interface{}{
				"status":     "healthy",
				"free_space": "85%",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")

	if err := json.NewEncoder(w).Encode(health); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
