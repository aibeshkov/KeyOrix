package services

import (
	"context"
	"testing"
	"time"

	"github.com/secretlyhq/secretly/internal/core"
	"github.com/secretlyhq/secretly/internal/storage/models"
	"github.com/secretlyhq/secretly/server/grpc/interceptors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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

func TestShareGRPCService_ShareSecret(t *testing.T) {
	// Setup
	mockCore := new(MockCoreService)
	service, err := NewShareService(mockCore)
	require.NoError(t, err)

	// Create test context with user
	user := &interceptors.UserContext{
		ID:          1,
		Username:    "testuser",
		Permissions: []string{"secrets.write"},
	}
	ctx := context.WithValue(context.Background(), interceptors.UserContextKey, user)

	// Test data
	req := &ShareSecretRequest{
		SecretID:    1,
		RecipientID: 2,
		IsGroup:     false,
		Permission:  "read",
	}
	shareRecord := &models.ShareRecord{
		ID:          1,
		SecretID:    1,
		OwnerID:     1,
		RecipientID: 2,
		IsGroup:     false,
		Permission:  "read",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Mock expectations
	mockCore.On("ShareSecret", mock.Anything, mock.MatchedBy(func(req *core.ShareSecretRequest) bool {
		return req.SecretID == 1 && req.RecipientID == 2 && req.Permission == "read"
	})).Return(shareRecord, nil)

	// Execute
	result, err := service.ShareSecret(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, uint32(1), result.ID)
	assert.Equal(t, uint32(1), result.SecretID)
	assert.Equal(t, uint32(2), result.RecipientID)
	assert.Equal(t, "read", result.Permission)
	mockCore.AssertExpectations(t)
}

func TestShareGRPCService_ShareSecretWithGroup(t *testing.T) {
	// Setup
	mockCore := new(MockCoreService)
	service, err := NewShareService(mockCore)
	require.NoError(t, err)

	// Create test context with user
	user := &interceptors.UserContext{
		ID:          1,
		Username:    "testuser",
		Permissions: []string{"secrets.write"},
	}
	ctx := context.WithValue(context.Background(), interceptors.UserContextKey, user)

	// Test data
	req := &ShareSecretRequest{
		SecretID:    1,
		RecipientID: 2,
		IsGroup:     true,
		Permission:  "read",
	}
	shareRecord := &models.ShareRecord{
		ID:          1,
		SecretID:    1,
		OwnerID:     1,
		RecipientID: 2,
		IsGroup:     true,
		Permission:  "read",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Mock expectations
	mockCore.On("ShareSecretWithGroup", mock.Anything, mock.MatchedBy(func(req *core.GroupShareSecretRequest) bool {
		return req.SecretID == 1 && req.GroupID == 2 && req.Permission == "read"
	})).Return(shareRecord, nil)

	// Execute
	result, err := service.ShareSecret(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, uint32(1), result.ID)
	assert.Equal(t, uint32(1), result.SecretID)
	assert.Equal(t, uint32(2), result.RecipientID)
	assert.Equal(t, true, result.IsGroup)
	assert.Equal(t, "read", result.Permission)
	mockCore.AssertExpectations(t)
}

func TestShareGRPCService_ListSecretShares(t *testing.T) {
	// Setup
	mockCore := new(MockCoreService)
	service, err := NewShareService(mockCore)
	require.NoError(t, err)

	// Create test context with user
	user := &interceptors.UserContext{
		ID:          1,
		Username:    "testuser",
		Permissions: []string{"secrets.read"},
	}
	ctx := context.WithValue(context.Background(), interceptors.UserContextKey, user)

	// Test data
	req := &ListSecretSharesRequest{
		SecretID: 1,
	}
	shares := []*models.ShareRecord{
		{
			ID:          1,
			SecretID:    1,
			OwnerID:     1,
			RecipientID: 2,
			IsGroup:     false,
			Permission:  "read",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          2,
			SecretID:    1,
			OwnerID:     1,
			RecipientID: 3,
			IsGroup:     false,
			Permission:  "write",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// Mock expectations
	mockCore.On("ListSecretShares", mock.Anything, uint(1)).Return(shares, nil)

	// Execute
	result, err := service.ListSecretShares(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 2, len(result.Shares))
	assert.Equal(t, uint32(1), result.Shares[0].ID)
	assert.Equal(t, uint32(2), result.Shares[1].ID)
	mockCore.AssertExpectations(t)
}

func TestShareGRPCService_UpdateSharePermission(t *testing.T) {
	// Setup
	mockCore := new(MockCoreService)
	service, err := NewShareService(mockCore)
	require.NoError(t, err)

	// Create test context with user
	user := &interceptors.UserContext{
		ID:          1,
		Username:    "testuser",
		Permissions: []string{"secrets.write"},
	}
	ctx := context.WithValue(context.Background(), interceptors.UserContextKey, user)

	// Test data
	req := &UpdateSharePermissionRequest{
		ShareID:    1,
		Permission: "write",
	}
	shareRecord := &models.ShareRecord{
		ID:          1,
		SecretID:    1,
		OwnerID:     1,
		RecipientID: 2,
		IsGroup:     false,
		Permission:  "write", // Updated permission
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Mock expectations
	mockCore.On("UpdateSharePermission", mock.Anything, mock.MatchedBy(func(req *core.UpdateShareRequest) bool {
		return req.ShareID == 1 && req.Permission == "write"
	})).Return(shareRecord, nil)

	// Execute
	result, err := service.UpdateSharePermission(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, uint32(1), result.ID)
	assert.Equal(t, "write", result.Permission)
	mockCore.AssertExpectations(t)
}

func TestShareGRPCService_RevokeShare(t *testing.T) {
	// Setup
	mockCore := new(MockCoreService)
	service, err := NewShareService(mockCore)
	require.NoError(t, err)

	// Create test context with user
	user := &interceptors.UserContext{
		ID:          1,
		Username:    "testuser",
		Permissions: []string{"secrets.write"},
	}
	ctx := context.WithValue(context.Background(), interceptors.UserContextKey, user)

	// Test data
	req := &RevokeShareRequest{
		ShareID: 1,
	}

	// Mock expectations
	mockCore.On("RevokeShare", mock.Anything, uint(1), "testuser").Return(nil)

	// Execute
	result, err := service.RevokeShare(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.IsType(t, &emptypb.Empty{}, result)
	mockCore.AssertExpectations(t)
}

func TestShareGRPCService_ListUserShares(t *testing.T) {
	// Setup
	mockCore := new(MockCoreService)
	service, err := NewShareService(mockCore)
	require.NoError(t, err)

	// Create test context with user
	user := &interceptors.UserContext{
		ID:          1,
		Username:    "testuser",
		Permissions: []string{"secrets.read"},
	}
	ctx := context.WithValue(context.Background(), interceptors.UserContextKey, user)

	// Test data
	req := &ListUserSharesRequest{
		Page:     1,
		PageSize: 10,
	}
	shares := []*models.ShareRecord{
		{
			ID:          1,
			SecretID:    1,
			OwnerID:     1,
			RecipientID: 2,
			IsGroup:     false,
			Permission:  "read",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// Mock expectations
	mockCore.On("ListSharesByUser", mock.Anything, uint(1)).Return(shares, nil)

	// Execute
	result, err := service.ListUserShares(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, len(result.Shares))
	assert.Equal(t, uint32(1), result.Shares[0].ID)
	mockCore.AssertExpectations(t)
}

func TestShareGRPCService_ListSharedSecrets(t *testing.T) {
	// Setup
	mockCore := new(MockCoreService)
	service, err := NewShareService(mockCore)
	require.NoError(t, err)

	// Create test context with user
	user := &interceptors.UserContext{
		ID:          1,
		Username:    "testuser",
		Permissions: []string{"secrets.read"},
	}
	ctx := context.WithValue(context.Background(), interceptors.UserContextKey, user)

	// Test data
	req := &ListSharedSecretsRequest{
		Page:     1,
		PageSize: 10,
	}
	secrets := []*models.SecretNode{
		{
			ID:        1,
			Name:      "shared-secret",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Mock expectations
	mockCore.On("ListSharedSecrets", mock.Anything, uint(1)).Return(secrets, nil)

	// Execute
	result, err := service.ListSharedSecrets(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, len(result.Secrets))
	assert.Equal(t, uint32(1), result.Secrets[0].Id)
	mockCore.AssertExpectations(t)
}

func TestShareGRPCService_Unauthorized(t *testing.T) {
	// Setup
	mockCore := new(MockCoreService)
	service, err := NewShareService(mockCore)
	require.NoError(t, err)

	// Create context without user
	ctx := context.Background()

	// Test cases
	testCases := []struct {
		name     string
		testFunc func() error
	}{
		{
			name: "ShareSecret",
			testFunc: func() error {
				_, err := service.ShareSecret(ctx, &ShareSecretRequest{})
				return err
			},
		},
		{
			name: "ListSecretShares",
			testFunc: func() error {
				_, err := service.ListSecretShares(ctx, &ListSecretSharesRequest{})
				return err
			},
		},
		{
			name: "ListUserShares",
			testFunc: func() error {
				_, err := service.ListUserShares(ctx, &ListUserSharesRequest{})
				return err
			},
		},
		{
			name: "ListSharedSecrets",
			testFunc: func() error {
				_, err := service.ListSharedSecrets(ctx, &ListSharedSecretsRequest{})
				return err
			},
		},
		{
			name: "UpdateSharePermission",
			testFunc: func() error {
				_, err := service.UpdateSharePermission(ctx, &UpdateSharePermissionRequest{})
				return err
			},
		},
		{
			name: "RevokeShare",
			testFunc: func() error {
				_, err := service.RevokeShare(ctx, &RevokeShareRequest{})
				return err
			},
		},
	}

	// Execute and assert
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.testFunc()
			require.Error(t, err)
			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, codes.Unauthenticated, st.Code())
		})
	}
}