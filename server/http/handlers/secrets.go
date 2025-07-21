package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/secretlyhq/secretly/internal/i18n"
	"github.com/secretlyhq/secretly/server/middleware"
	"github.com/secretlyhq/secretly/server/services"
	"github.com/secretlyhq/secretly/server/validation"
)

// SecretHandler handles secret-related HTTP requests
type SecretHandler struct {
	secretService *services.SecretServiceWrapper
	validator     *validation.Validator
}

// NewSecretHandler creates a new secret handler
func NewSecretHandler() (*SecretHandler, error) {
	secretService, err := services.NewSecretServiceWrapper()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("ErrorInitializationFailed", nil), err)
	}

	return &SecretHandler{
		secretService: secretService,
		validator:     validation.NewValidator(),
	}, nil
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Details interface{} `json:"details,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

// ListSecrets handles GET /api/v1/secrets
func (h *SecretHandler) ListSecrets(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userCtx := middleware.GetUserFromContext(r.Context())
	if userCtx == nil {
		h.sendError(w, "Unauthorized", "User context not found", http.StatusUnauthorized, nil)
		return
	}

	// Parse and validate query parameters
	req := &services.ListSecretsRequest{
		Namespace:   strings.TrimSpace(r.URL.Query().Get("namespace")),
		Zone:        strings.TrimSpace(r.URL.Query().Get("zone")),
		Environment: strings.TrimSpace(r.URL.Query().Get("environment")),
		Type:        strings.TrimSpace(r.URL.Query().Get("type")),
		Page:        1,
		PageSize:    20,
	}

	// Parse tags
	if tagsParam := r.URL.Query().Get("tags"); tagsParam != "" {
		req.Tags = strings.Split(tagsParam, ",")
		for i, tag := range req.Tags {
			req.Tags[i] = strings.TrimSpace(tag)
		}
	}

	// Parse pagination
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			req.PageSize = pageSize
		}
	}

	// Validate request
	if err := h.validator.Validate(req); err != nil {
		h.sendError(w, "ValidationError", "Invalid request parameters", http.StatusBadRequest, err)
		return
	}

	// Call service
	response, err := h.secretService.ListSecrets(r.Context(), req, userCtx.UserID)
	if err != nil {
		log.Printf("Error listing secrets: %v", err)
		h.sendError(w, "InternalError", "Failed to list secrets", http.StatusInternalServerError, nil)
		return
	}

	// Send response
	h.sendSuccess(w, response, "")
}

// CreateSecret handles POST /api/v1/secrets
func (h *SecretHandler) CreateSecret(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userCtx := middleware.GetUserFromContext(r.Context())
	if userCtx == nil {
		h.sendError(w, "Unauthorized", "User context not found", http.StatusUnauthorized, nil)
		return
	}

	// Parse request body
	var req services.SecretCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "InvalidJSON", "Invalid JSON in request body", http.StatusBadRequest, nil)
		return
	}

	// Validate request
	if err := h.validator.Validate(&req); err != nil {
		h.sendError(w, "ValidationError", "Invalid request data", http.StatusBadRequest, err)
		return
	}

	// Parse expiration if provided
	// Note: req.Expiration is already a *time.Time, no parsing needed

	// Call service
	response, err := h.secretService.CreateSecret(r.Context(), &req, userCtx.UserID)
	if err != nil {
		log.Printf("Error creating secret: %v", err)
		if strings.Contains(err.Error(), "already exists") {
			h.sendError(w, "ConflictError", "Secret with this name already exists", http.StatusConflict, nil)
		} else {
			h.sendError(w, "InternalError", "Failed to create secret", http.StatusInternalServerError, nil)
		}
		return
	}

	// Send response
	w.WriteHeader(http.StatusCreated)
	h.sendSuccess(w, response, i18n.T("SuccessSecretCreated", nil))
}

// GetSecret handles GET /api/v1/secrets/{id}
func (h *SecretHandler) GetSecret(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userCtx := middleware.GetUserFromContext(r.Context())
	if userCtx == nil {
		h.sendError(w, "Unauthorized", "User context not found", http.StatusUnauthorized, nil)
		return
	}

	// Parse ID parameter
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.sendError(w, "InvalidParameter", "Invalid secret ID", http.StatusBadRequest, nil)
		return
	}

	// Check if requesting decrypted value
	includeValue := r.URL.Query().Get("include_value") == "true"

	// Call service
	response, err := h.secretService.GetSecret(r.Context(), uint(id), includeValue, userCtx.UserID)
	if err != nil {
		log.Printf("Error getting secret: %v", err)
		if strings.Contains(err.Error(), "not found") {
			h.sendError(w, "NotFound", "Secret not found", http.StatusNotFound, nil)
		} else if strings.Contains(err.Error(), "permission denied") {
			h.sendError(w, "Forbidden", "Access denied", http.StatusForbidden, nil)
		} else {
			h.sendError(w, "InternalError", "Failed to get secret", http.StatusInternalServerError, nil)
		}
		return
	}

	// Send response
	h.sendSuccess(w, response, "")
}

// UpdateSecret handles PUT /api/v1/secrets/{id}
func (h *SecretHandler) UpdateSecret(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userCtx := middleware.GetUserFromContext(r.Context())
	if userCtx == nil {
		h.sendError(w, "Unauthorized", "User context not found", http.StatusUnauthorized, nil)
		return
	}

	// Parse ID parameter
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.sendError(w, "InvalidParameter", "Invalid secret ID", http.StatusBadRequest, nil)
		return
	}

	// Parse request body
	var req services.SecretUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "InvalidJSON", "Invalid JSON in request body", http.StatusBadRequest, nil)
		return
	}

	// Validate request
	if err := h.validator.Validate(&req); err != nil {
		h.sendError(w, "ValidationError", "Invalid request data", http.StatusBadRequest, err)
		return
	}

	// Call service
	response, err := h.secretService.UpdateSecret(r.Context(), uint(id), &req, userCtx.UserID)
	if err != nil {
		log.Printf("Error updating secret: %v", err)
		if strings.Contains(err.Error(), "not found") {
			h.sendError(w, "NotFound", "Secret not found", http.StatusNotFound, nil)
		} else if strings.Contains(err.Error(), "permission denied") {
			h.sendError(w, "Forbidden", "Access denied", http.StatusForbidden, nil)
		} else {
			h.sendError(w, "InternalError", "Failed to update secret", http.StatusInternalServerError, nil)
		}
		return
	}

	// Send response
	h.sendSuccess(w, response, i18n.T("SuccessSecretUpdated", nil))
}

// DeleteSecret handles DELETE /api/v1/secrets/{id}
func (h *SecretHandler) DeleteSecret(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userCtx := middleware.GetUserFromContext(r.Context())
	if userCtx == nil {
		h.sendError(w, "Unauthorized", "User context not found", http.StatusUnauthorized, nil)
		return
	}

	// Parse ID parameter
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.sendError(w, "InvalidParameter", "Invalid secret ID", http.StatusBadRequest, nil)
		return
	}

	// Call service
	err = h.secretService.DeleteSecret(r.Context(), uint(id), userCtx.UserID)
	if err != nil {
		log.Printf("Error deleting secret: %v", err)
		if strings.Contains(err.Error(), "not found") {
			h.sendError(w, "NotFound", "Secret not found", http.StatusNotFound, nil)
		} else if strings.Contains(err.Error(), "permission denied") {
			h.sendError(w, "Forbidden", "Access denied", http.StatusForbidden, nil)
		} else {
			h.sendError(w, "InternalError", "Failed to delete secret", http.StatusInternalServerError, nil)
		}
		return
	}

	// Send response
	w.WriteHeader(http.StatusNoContent)
}

// GetSecretVersions handles GET /api/v1/secrets/{id}/versions
func (h *SecretHandler) GetSecretVersions(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userCtx := middleware.GetUserFromContext(r.Context())
	if userCtx == nil {
		h.sendError(w, "Unauthorized", "User context not found", http.StatusUnauthorized, nil)
		return
	}

	// Parse ID parameter
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.sendError(w, "InvalidParameter", "Invalid secret ID", http.StatusBadRequest, nil)
		return
	}

	// Call service
	versions, err := h.secretService.GetSecretVersions(r.Context(), uint(id), userCtx.UserID)
	if err != nil {
		log.Printf("Error getting secret versions: %v", err)
		if strings.Contains(err.Error(), "not found") {
			h.sendError(w, "NotFound", "Secret not found", http.StatusNotFound, nil)
		} else if strings.Contains(err.Error(), "permission denied") {
			h.sendError(w, "Forbidden", "Access denied", http.StatusForbidden, nil)
		} else {
			h.sendError(w, "InternalError", "Failed to get secret versions", http.StatusInternalServerError, nil)
		}
		return
	}

	// Send response
	response := map[string]interface{}{
		"versions": versions,
	}
	h.sendSuccess(w, response, "")
}

// Helper methods for consistent response handling

// sendSuccess sends a successful JSON response
func (h *SecretHandler) sendSuccess(w http.ResponseWriter, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	
	response := SuccessResponse{
		Data:    data,
		Message: message,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// sendError sends an error JSON response
func (h *SecretHandler) sendError(w http.ResponseWriter, errorType, message string, statusCode int, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(statusCode)
	
	response := ErrorResponse{
		Error:   errorType,
		Message: message,
		Code:    statusCode,
		Details: details,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON error response: %v", err)
	}
}

