package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/secretlyhq/secretly/internal/core"
	"github.com/secretlyhq/secretly/internal/i18n"
	"github.com/secretlyhq/secretly/internal/storage/models"
	"github.com/secretlyhq/secretly/server/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockCoreService is a mock implementation of the core service
type MockCoreService struct {
	mock.Mock
}

func (m *MockCoreService) ShareSecret(ctx context.Context, req *core.ShareSecretRequest) (*models.ShareRecord, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ShareRecord), args.Error(1)
}

func (m *MockCoreService) ShareSecretWithGroup(ctx context.Context, req *core.GroupShareSecretRequest) (*models.ShareRecord, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ShareRecord), args.Error(1)
}

func (m *MockCoreService) UpdateSharePermission(ctx context.Context, req *core.UpdateShareRequest) (*models.ShareRecord, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ShareRecord), args.Error(1)
}

func (m *MockCoreService) RevokeShare(ctx context.Context, shareID uint, revokedBy string) error {
	args := m.Called(ctx, shareID, revokedBy)
	return args.Error(0)
}

func (m *MockCoreService) ListSecretShares(ctx context.Context, secretID uint) ([]*models.ShareRecord, error) {
	args := m.Called(ctx, secretID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ShareRecord), args.Error(1)
}

func (m *MockCoreService) ListSharesByUser(ctx context.Context, userID uint) ([]*models.ShareRecord, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ShareRecord), args.Error(1)
}

func (m *MockCoreService) ListSharedSecrets(ctx context.Context, userID uint) ([]*models.SecretNode, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.SecretNode), args.Error(1)
}

func TestShareSecret(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	// Create mock core service
	mockCore := new(MockCoreService)
	
	// Create handler
	handler, err := NewShareHandler(mockCore)
	require.NoError(t, err)
	
	// Create test request
	reqBody := map[string]interface{}{
		"recipient_id": 2,
		"is_group":     false,
		"permission":   "read",
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)
	
	req := httptest.NewRequest("POST", "/api/v1/secrets/1/share", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	
	// Add user context
	user := &middleware.UserContext{
		ID:       1,
		Username: "testuser",
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	
	// Set up chi router context with URL parameter
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	// Create response recorder
	rr := httptest.NewRecorder()
	
	// Mock expectations
	shareRecord := &models.ShareRecord{
		ID:          1,
		SecretID:    1,
		OwnerID:     1,
		RecipientID: 2,
		IsGroup:     false,
		Permission:  "read",
	}
	mockCore.On("ShareSecret", mock.Anything, mock.MatchedBy(func(req *core.ShareSecretRequest) bool {
		return req.SecretID == 1 && req.RecipientID == 2 && req.Permission == "read"
	})).Return(shareRecord, nil)
	
	// Call handler
	handler.ShareSecret(rr, req)
	
	// Assert response
	assert.Equal(t, http.StatusCreated, rr.Code)
	
	var response SuccessResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	// Verify mock expectations
	mockCore.AssertExpectations(t)
}

func TestShareSecretWithGroup(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	// Create mock core service
	mockCore := new(MockCoreService)
	
	// Create handler
	handler, err := NewShareHandler(mockCore)
	require.NoError(t, err)
	
	// Create test request
	reqBody := map[string]interface{}{
		"recipient_id": 2,
		"is_group":     true,
		"permission":   "read",
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)
	
	req := httptest.NewRequest("POST", "/api/v1/secrets/1/share", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	
	// Add user context
	user := &middleware.UserContext{
		ID:       1,
		Username: "testuser",
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	
	// Set up chi router context with URL parameter
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	// Create response recorder
	rr := httptest.NewRecorder()
	
	// Mock expectations
	shareRecord := &models.ShareRecord{
		ID:          1,
		SecretID:    1,
		OwnerID:     1,
		RecipientID: 2,
		IsGroup:     true,
		Permission:  "read",
	}
	mockCore.On("ShareSecretWithGroup", mock.Anything, mock.MatchedBy(func(req *core.GroupShareSecretRequest) bool {
		return req.SecretID == 1 && req.GroupID == 2 && req.Permission == "read"
	})).Return(shareRecord, nil)
	
	// Call handler
	handler.ShareSecret(rr, req)
	
	// Assert response
	assert.Equal(t, http.StatusCreated, rr.Code)
	
	var response SuccessResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	// Verify mock expectations
	mockCore.AssertExpectations(t)
}

func TestListSecretShares(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	// Create mock core service
	mockCore := new(MockCoreService)
	
	// Create handler
	handler, err := NewShareHandler(mockCore)
	require.NoError(t, err)
	
	// Create test request
	req := httptest.NewRequest("GET", "/api/v1/secrets/1/shares", nil)
	
	// Add user context
	user := &middleware.UserContext{
		ID:       1,
		Username: "testuser",
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	
	// Set up chi router context with URL parameter
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	// Create response recorder
	rr := httptest.NewRecorder()
	
	// Mock expectations
	shares := []*models.ShareRecord{
		{
			ID:          1,
			SecretID:    1,
			OwnerID:     1,
			RecipientID: 2,
			IsGroup:     false,
			Permission:  "read",
		},
	}
	mockCore.On("ListSecretShares", mock.Anything, uint(1)).Return(shares, nil)
	
	// Call handler
	handler.ListSecretShares(rr, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response SuccessResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	// Verify mock expectations
	mockCore.AssertExpectations(t)
}

func TestUpdateSharePermission(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	// Create mock core service
	mockCore := new(MockCoreService)
	
	// Create handler
	handler, err := NewShareHandler(mockCore)
	require.NoError(t, err)
	
	// Create test request
	reqBody := map[string]interface{}{
		"permission": "write",
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)
	
	req := httptest.NewRequest("PUT", "/api/v1/shares/1", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	
	// Add user context
	user := &middleware.UserContext{
		ID:       1,
		Username: "testuser",
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	
	// Set up chi router context with URL parameter
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	// Create response recorder
	rr := httptest.NewRecorder()
	
	// Mock expectations
	shareRecord := &models.ShareRecord{
		ID:          1,
		SecretID:    1,
		OwnerID:     1,
		RecipientID: 2,
		IsGroup:     false,
		Permission:  "write", // Updated permission
	}
	mockCore.On("UpdateSharePermission", mock.Anything, mock.MatchedBy(func(req *core.UpdateShareRequest) bool {
		return req.ShareID == 1 && req.Permission == "write"
	})).Return(shareRecord, nil)
	
	// Call handler
	handler.UpdateSharePermission(rr, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response SuccessResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	// Verify mock expectations
	mockCore.AssertExpectations(t)
}

func TestRevokeShare(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	// Create mock core service
	mockCore := new(MockCoreService)
	
	// Create handler
	handler, err := NewShareHandler(mockCore)
	require.NoError(t, err)
	
	// Create test request
	req := httptest.NewRequest("DELETE", "/api/v1/shares/1", nil)
	
	// Add user context
	user := &middleware.UserContext{
		ID:       1,
		Username: "testuser",
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	
	// Set up chi router context with URL parameter
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	// Create response recorder
	rr := httptest.NewRecorder()
	
	// Mock expectations
	mockCore.On("RevokeShare", mock.Anything, uint(1), "testuser").Return(nil)
	
	// Call handler
	handler.RevokeShare(rr, req)
	
	// Assert response
	assert.Equal(t, http.StatusNoContent, rr.Code)
	
	// Verify mock expectations
	mockCore.AssertExpectations(t)
}

func TestListShares(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	// Create mock core service
	mockCore := new(MockCoreService)
	
	// Create handler
	handler, err := NewShareHandler(mockCore)
	require.NoError(t, err)
	
	// Create test request
	req := httptest.NewRequest("GET", "/api/v1/shares", nil)
	
	// Add user context
	user := &middleware.UserContext{
		ID:       1,
		Username: "testuser",
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	
	// Create response recorder
	rr := httptest.NewRecorder()
	
	// Mock expectations
	shares := []*models.ShareRecord{
		{
			ID:          1,
			SecretID:    1,
			OwnerID:     1,
			RecipientID: 2,
			IsGroup:     false,
			Permission:  "read",
		},
	}
	mockCore.On("ListSharesByUser", mock.Anything, uint(1)).Return(shares, nil)
	
	// Call handler
	handler.ListShares(rr, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response SuccessResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	// Verify mock expectations
	mockCore.AssertExpectations(t)
}

func TestListSharedSecrets(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	// Create mock core service
	mockCore := new(MockCoreService)
	
	// Create handler
	handler, err := NewShareHandler(mockCore)
	require.NoError(t, err)
	
	// Create test request
	req := httptest.NewRequest("GET", "/api/v1/shared-secrets", nil)
	
	// Add user context
	user := &middleware.UserContext{
		ID:       1,
		Username: "testuser",
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	
	// Create response recorder
	rr := httptest.NewRecorder()
	
	// Mock expectations
	secrets := []*models.SecretNode{
		{
			ID:   1,
			Name: "shared-secret",
		},
	}
	mockCore.On("ListSharedSecrets", mock.Anything, uint(1)).Return(secrets, nil)
	
	// Call handler
	handler.ListSharedSecrets(rr, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response SuccessResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	// Verify mock expectations
	mockCore.AssertExpectations(t)
}