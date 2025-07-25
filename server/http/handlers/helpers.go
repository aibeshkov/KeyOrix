package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// sendSuccess sends a successful JSON response
func sendSuccess(w http.ResponseWriter, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	response := map[string]interface{}{
		"data": data,
	}
	if message != "" {
		response["message"] = message
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// sendError sends an error JSON response
func sendError(w http.ResponseWriter, errorType, message string, statusCode int, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":   errorType,
		"message": message,
		"code":    statusCode,
	}
	if details != nil {
		response["details"] = details
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON error response: %v", err)
	}
}
