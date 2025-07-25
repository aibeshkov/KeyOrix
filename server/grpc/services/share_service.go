package services

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/secretlyhq/secretly/internal/core"
	"github.com/secretlyhq/secretly/internal/storage/models"
	"github.com/secretlyhq/secretly/internal/utils/safeconv"
	"github.com/secretlyhq/secretly/server/grpc/interceptors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ShareGRPCService implements the gRPC share service
type ShareGRPCService struct {
	coreService *core.SecretlyCore
	// TODO: Add UnimplementedShareServiceServer when proto is generated
}

// NewShareService creates a new share gRPC service
func NewShareService(coreService *core.SecretlyCore) (*ShareGRPCService, error) {
	return &ShareGRPCService{
		coreService: coreService,
	}, nil
}

// ShareSecretRequest represents a gRPC share secret request
type ShareSecretRequest struct {
	SecretID    uint32 `json:"secret_id"`
	RecipientID uint32 `json:"recipient_id"`
	IsGroup     bool   `json:"is_group"`
	Permission  string `json:"permission"`
}

// ShareRecord represents a gRPC share record response
type ShareRecord struct {
	ID          uint32                 `json:"id"`
	SecretID    uint32                 `json:"secret_id"`
	OwnerID     uint32                 `json:"owner_id"`
	RecipientID uint32                 `json:"recipient_id"`
	IsGroup     bool                   `json:"is_group"`
	Permission  string                 `json:"permission"`
	CreatedAt   *timestamppb.Timestamp `json:"created_at"`
	UpdatedAt   *timestamppb.Timestamp `json:"updated_at"`
}

// ShareSecret shares a secret with another user or group via gRPC
func (s *ShareGRPCService) ShareSecret(ctx context.Context, req *ShareSecretRequest) (*ShareRecord, error) {
	// Get user from context
	user := interceptors.GetUserFromGRPCContext(ctx)
	if user == nil {
		return nil, status.Errorf(codes.Unauthenticated, "User not authenticated")
	}

	// Check permissions
	if !s.hasPermission(user.Permissions, "secrets.write") {
		return nil, status.Errorf(codes.PermissionDenied, "Insufficient permissions for sharing secrets")
	}

	log.Printf("gRPC ShareSecret called by user %s for secret ID %d", user.Username, req.SecretID)

	// Validate request
	if req.SecretID == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Secret ID is required")
	}
	if req.RecipientID == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Recipient ID is required")
	}
	if req.Permission != "read" && req.Permission != "write" {
		return nil, status.Errorf(codes.InvalidArgument, "Permission must be 'read' or 'write'")
	}

	var shareRecord *models.ShareRecord
	var err error
	
	// Handle group sharing differently
	if req.IsGroup {
		// Convert to group share service request
		groupReq := &core.GroupShareSecretRequest{
			SecretID:   uint(req.SecretID),
			GroupID:    uint(req.RecipientID),
			Permission: req.Permission,
			SharedBy:   user.UserID,
		}
		
		// Call service for group sharing
		shareRecord, err = s.coreService.ShareSecretWithGroup(ctx, groupReq)
	} else {
		// Convert to user share service request
		userReq := &core.ShareSecretRequest{
			SecretID:    uint(req.SecretID),
			RecipientID: uint(req.RecipientID),
			IsGroup:     false,
			Permission:  req.Permission,
			SharedBy:    user.UserID,
		}
		
		// Call service for user sharing
		shareRecord, err = s.coreService.ShareSecret(ctx, userReq)
	}
	
	// Handle errors
	if err != nil {
		log.Printf("Error sharing secret via gRPC: %v", err)
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "Secret not found")
		} else if strings.Contains(err.Error(), "not authorized") {
			return nil, status.Errorf(codes.PermissionDenied, "Not authorized to share this secret")
		} else {
			return nil, status.Errorf(codes.Internal, "Failed to share secret")
		}
	}

	// Convert response
	return s.convertToGRPCShareRecord(shareRecord), nil
}

// ListSecretSharesRequest represents a gRPC list secret shares request
type ListSecretSharesRequest struct {
	SecretID uint32 `json:"secret_id"`
}

// ListSharesResponse represents a gRPC list shares response
type ListSharesResponse struct {
	Shares     []*ShareRecord `json:"shares"`
	Total      uint32         `json:"total"`
	Page       uint32         `json:"page"`
	PageSize   uint32         `json:"page_size"`
	TotalPages uint32         `json:"total_pages"`
}

// ListSecretShares lists all shares for a secret via gRPC
func (s *ShareGRPCService) ListSecretShares(ctx context.Context, req *ListSecretSharesRequest) (*ListSharesResponse, error) {
	// Get user from context
	user := interceptors.GetUserFromGRPCContext(ctx)
	if user == nil {
		return nil, status.Errorf(codes.Unauthenticated, "User not authenticated")
	}

	// Check permissions
	if !s.hasPermission(user.Permissions, "secrets.read") {
		return nil, status.Errorf(codes.PermissionDenied, "Insufficient permissions for viewing shares")
	}

	log.Printf("gRPC ListSecretShares called by user %s for secret ID %d", user.Username, req.SecretID)

	// Validate request
	if req.SecretID == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Secret ID is required")
	}

	// Call service
	shares, err := s.coreService.ListSecretShares(ctx, uint(req.SecretID))
	if err != nil {
		log.Printf("Error listing secret shares via gRPC: %v", err)
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "Secret not found")
		} else {
			return nil, status.Errorf(codes.Internal, "Failed to list secret shares")
		}
	}

	// Convert response
	grpcShares := make([]*ShareRecord, len(shares))
	for i, share := range shares {
		grpcShares[i] = s.convertToGRPCShareRecord(share)
	}

	return &ListSharesResponse{
		Shares:     grpcShares,
		Total:      uint32(len(grpcShares)),
		Page:       1,
		PageSize:   uint32(len(grpcShares)),
		TotalPages: 1,
	}, nil
}

// ListUserSharesRequest represents a gRPC list user shares request
type ListUserSharesRequest struct {
	Page     uint32 `json:"page"`
	PageSize uint32 `json:"page_size"`
}

// ListUserShares lists all shares created by a user via gRPC
func (s *ShareGRPCService) ListUserShares(ctx context.Context, req *ListUserSharesRequest) (*ListSharesResponse, error) {
	// Get user from context
	user := interceptors.GetUserFromGRPCContext(ctx)
	if user == nil {
		return nil, status.Errorf(codes.Unauthenticated, "User not authenticated")
	}

	// Check permissions
	if !s.hasPermission(user.Permissions, "secrets.read") {
		return nil, status.Errorf(codes.PermissionDenied, "Insufficient permissions for viewing shares")
	}

	log.Printf("gRPC ListUserShares called by user %s", user.Username)

	// Validate pagination
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// Call service
	shares, err := s.coreService.ListSharesByUser(ctx, user.UserID)
	if err != nil {
		log.Printf("Error listing user shares via gRPC: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to list user shares")
	}

	// Convert response
	grpcShares := make([]*ShareRecord, len(shares))
	for i, share := range shares {
		grpcShares[i] = s.convertToGRPCShareRecord(share)
	}

	return &ListSharesResponse{
		Shares:     grpcShares,
		Total:      uint32(len(grpcShares)),
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: uint32((len(grpcShares) + int(req.PageSize) - 1) / int(req.PageSize)),
	}, nil
}

// ListSharedSecretsRequest represents a gRPC list shared secrets request
type ListSharedSecretsRequest struct {
	Page     uint32 `json:"page"`
	PageSize uint32 `json:"page_size"`
}

// ListSharedSecretsResponse represents a gRPC list shared secrets response
type ListSharedSecretsResponse struct {
	Secrets    []*SecretResponse `json:"secrets"`
	Total      uint32            `json:"total"`
	Page       uint32            `json:"page"`
	PageSize   uint32            `json:"page_size"`
	TotalPages uint32            `json:"total_pages"`
}

// ListSharedSecrets lists all secrets shared with a user via gRPC
func (s *ShareGRPCService) ListSharedSecrets(ctx context.Context, req *ListSharedSecretsRequest) (*ListSecretsResponse, error) {
	// Get user from context
	user := interceptors.GetUserFromGRPCContext(ctx)
	if user == nil {
		return nil, status.Errorf(codes.Unauthenticated, "User not authenticated")
	}

	// Check permissions
	if !s.hasPermission(user.Permissions, "secrets.read") {
		return nil, status.Errorf(codes.PermissionDenied, "Insufficient permissions for viewing shared secrets")
	}

	log.Printf("gRPC ListSharedSecrets called by user %s", user.Username)

	// Validate pagination
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// Call service
	secrets, err := s.coreService.ListSharedSecrets(ctx, user.UserID)
	if err != nil {
		log.Printf("Error listing shared secrets via gRPC: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to list shared secrets")
	}

	// Convert response
	grpcSecrets := make([]*SecretResponse, len(secrets))
	for i, secret := range secrets {
		grpcSecrets[i] = s.convertToGRPCSecretResponse(secret)
	}

	return &ListSecretsResponse{
		Secrets:    grpcSecrets,
		Total:      int64(len(grpcSecrets)),
		Page:       int32(req.Page),
		PageSize:   int32(req.PageSize),
		TotalPages: int32((len(grpcSecrets) + int(req.PageSize) - 1) / int(req.PageSize)),
	}, nil
}

// UpdateSharePermissionRequest represents a gRPC update share permission request
type UpdateSharePermissionRequest struct {
	ShareID    uint32 `json:"share_id"`
	Permission string `json:"permission"`
}

// UpdateSharePermission updates the permission level of a share via gRPC
func (s *ShareGRPCService) UpdateSharePermission(ctx context.Context, req *UpdateSharePermissionRequest) (*ShareRecord, error) {
	// Get user from context
	user := interceptors.GetUserFromGRPCContext(ctx)
	if user == nil {
		return nil, status.Errorf(codes.Unauthenticated, "User not authenticated")
	}

	// Check permissions
	if !s.hasPermission(user.Permissions, "secrets.write") {
		return nil, status.Errorf(codes.PermissionDenied, "Insufficient permissions for updating shares")
	}

	log.Printf("gRPC UpdateSharePermission called by user %s for share ID %d", user.Username, req.ShareID)

	// Validate request
	if req.ShareID == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Share ID is required")
	}
	if req.Permission != "read" && req.Permission != "write" {
		return nil, status.Errorf(codes.InvalidArgument, "Permission must be 'read' or 'write'")
	}

	// Convert to service request
	serviceReq := &core.UpdateShareRequest{
		ShareID:    uint(req.ShareID),
		Permission: req.Permission,
		UpdatedBy:  user.UserID,
	}

	// Call service
	shareRecord, err := s.coreService.UpdateSharePermission(ctx, serviceReq)
	if err != nil {
		log.Printf("Error updating share permission via gRPC: %v", err)
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "Share not found")
		} else if strings.Contains(err.Error(), "not authorized") {
			return nil, status.Errorf(codes.PermissionDenied, "Not authorized to update this share")
		} else {
			return nil, status.Errorf(codes.Internal, "Failed to update share permission")
		}
	}

	// Convert response
	return s.convertToGRPCShareRecord(shareRecord), nil
}

// RevokeShareRequest represents a gRPC revoke share request
type RevokeShareRequest struct {
	ShareID uint32 `json:"share_id"`
}

// RevokeShare revokes a share via gRPC
func (s *ShareGRPCService) RevokeShare(ctx context.Context, req *RevokeShareRequest) (*emptypb.Empty, error) {
	// Get user from context
	user := interceptors.GetUserFromGRPCContext(ctx)
	if user == nil {
		return nil, status.Errorf(codes.Unauthenticated, "User not authenticated")
	}

	// Check permissions
	if !s.hasPermission(user.Permissions, "secrets.write") {
		return nil, status.Errorf(codes.PermissionDenied, "Insufficient permissions for revoking shares")
	}

	log.Printf("gRPC RevokeShare called by user %s for share ID %d", user.Username, req.ShareID)

	// Validate request
	if req.ShareID == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Share ID is required")
	}

	// Call service
	err := s.coreService.RevokeShare(ctx, uint(req.ShareID), user.UserID)
	if err != nil {
		log.Printf("Error revoking share via gRPC: %v", err)
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "Share not found")
		} else if strings.Contains(err.Error(), "not authorized") {
			return nil, status.Errorf(codes.PermissionDenied, "Not authorized to revoke this share")
		} else {
			return nil, status.Errorf(codes.Internal, "Failed to revoke share")
		}
	}

	return &emptypb.Empty{}, nil
}

// RemoveSelfFromShareRequest represents a gRPC remove self from share request
type RemoveSelfFromShareRequest struct {
	SecretID uint32 `json:"secret_id"`
}

// RemoveSelfFromShare allows a user to remove themselves from a shared secret via gRPC
func (s *ShareGRPCService) RemoveSelfFromShare(ctx context.Context, req *RemoveSelfFromShareRequest) (*emptypb.Empty, error) {
	// Get user from context
	user := interceptors.GetUserFromGRPCContext(ctx)
	if user == nil {
		return nil, status.Errorf(codes.Unauthenticated, "User not authenticated")
	}

	// Check permissions
	if !s.hasPermission(user.Permissions, "secrets.read") {
		return nil, status.Errorf(codes.PermissionDenied, "Insufficient permissions for removing self from shares")
	}

	log.Printf("gRPC RemoveSelfFromShare called by user %s for secret ID %d", user.Username, req.SecretID)

	// Validate request
	if req.SecretID == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Secret ID is required")
	}

	// Call service
	err := s.coreService.RemoveSelfFromShare(ctx, uint(req.SecretID), user.UserID)
	if err != nil {
		log.Printf("Error removing self from share via gRPC: %v", err)
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "Share not found")
		} else {
			return nil, status.Errorf(codes.Internal, "Failed to remove self from share")
		}
	}

	return &emptypb.Empty{}, nil
}

// Helper methods

// hasPermission checks if user has a specific permission
func (s *ShareGRPCService) hasPermission(permissions []string, required string) bool {
	for _, perm := range permissions {
		if perm == required {
			return true
		}
	}
	return false
}

// convertToGRPCShareRecord converts a storage share record to a gRPC share record
func (s *ShareGRPCService) convertToGRPCShareRecord(share *models.ShareRecord) *ShareRecord {
	return &ShareRecord{
		ID: func() uint32 {
			id, err := safeconv.UintToUint32(share.ID)
			if err != nil {
				log.Printf("Warning: ID conversion overflow for share %d: %v", share.ID, err)
				return 0
			}
			return id
		}(),
		SecretID: func() uint32 {
			id, err := safeconv.UintToUint32(share.SecretID)
			if err != nil {
				log.Printf("Warning: SecretID conversion overflow for share %d: %v", share.ID, err)
				return 0
			}
			return id
		}(),
		OwnerID: func() uint32 {
			id, err := safeconv.UintToUint32(share.OwnerID)
			if err != nil {
				log.Printf("Warning: OwnerID conversion overflow for share %d: %v", share.ID, err)
				return 0
			}
			return id
		}(),
		RecipientID: func() uint32 {
			id, err := safeconv.UintToUint32(share.RecipientID)
			if err != nil {
				log.Printf("Warning: RecipientID conversion overflow for share %d: %v", share.ID, err)
				return 0
			}
			return id
		}(),
		IsGroup:    share.IsGroup,
		Permission: share.Permission,
		CreatedAt:  timestamppb.New(share.CreatedAt),
		UpdatedAt:  timestamppb.New(share.UpdatedAt),
	}
}

// convertToGRPCSecretResponse converts a storage secret node to a gRPC secret response
func (s *ShareGRPCService) convertToGRPCSecretResponse(secret *models.SecretNode) *SecretResponse {
	var maxReads *int32
	if secret.MaxReads != nil {
		maxReadsInt32, err := safeconv.IntToInt32(*secret.MaxReads)
		if err != nil {
			log.Printf("Warning: MaxReads conversion overflow for secret %d: %v", secret.ID, err)
			maxReadsInt32 = 0
		}
		maxReads = &maxReadsInt32
	}

	return &SecretResponse{
		Id: func() uint32 {
			id, err := safeconv.UintToUint32(secret.ID)
			if err != nil {
				log.Printf("Warning: ID conversion overflow for secret %d: %v", secret.ID, err)
				return 0
			}
			return id
		}(),
		Name:        secret.Name,
		Namespace:   fmt.Sprintf("%d", secret.NamespaceID), // TODO: Convert ID to name
		Zone:        fmt.Sprintf("%d", secret.ZoneID),      // TODO: Convert ID to name
		Environment: fmt.Sprintf("%d", secret.EnvironmentID), // TODO: Convert ID to name
		Type:        secret.Type,
		MaxReads:    maxReads,
		Metadata:    make(map[string]string), // TODO: Implement metadata
		Tags:        []string{},              // TODO: Implement tags
		CreatedBy:   secret.CreatedBy,
		CreatedAt:   secret.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   secret.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Version:     1, // TODO: Implement proper versioning
	}
}