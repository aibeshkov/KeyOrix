package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/secretlyhq/secretly/internal/config"
	"github.com/secretlyhq/secretly/internal/i18n"
	"github.com/secretlyhq/secretly/server/grpc/interceptors"
	"github.com/secretlyhq/secretly/server/grpc/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

// TestSharingGRPCIntegration tests the complete gRPC sharing workflow
func TestSharingGRPCIntegration(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	// Create test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			GRPC: config.ServerInstanceConfig{
				Enabled: true,
				Port:    "9090",
			},
		},
	}

	// Setup gRPC server
	lis = bufconn.Listen(bufSize)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.AuthInterceptor),
	)

	// Register services (this would normally be done in the main server setup)
	// For testing, we'll create mock services
	shareService, err := services.NewShareService(nil) // Mock core service
	require.NoError(t, err)

	// Register the service
	services.RegisterShareServiceServer(server, shareService)

	go func() {
		if err := server.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	defer server.Stop()

	// Create client connection
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := services.NewShareServiceClient(conn)

	t.Run("Complete gRPC Sharing Workflow", func(t *testing.T) {
		// Create authenticated context
		user := &interceptors.UserContext{
			ID:          1,
			Username:    "testuser",
			Permissions: []string{"secrets.write", "secrets.read"},
		}
		authCtx := context.WithValue(ctx, interceptors.UserContextKey, user)

		var shareID uint32

		// Step 1: Share a secret
		t.Run("Share Secret", func(t *testing.T) {
			req := &services.ShareSecretRequest{
				SecretID:    1,
				RecipientID: 2,
				IsGroup:     false,
				Permission:  "read",
			}

			resp, err := client.ShareSecret(authCtx, req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, uint32(1), resp.SecretID)
			assert.Equal(t, uint32(2), resp.RecipientID)
			assert.Equal(t, "read", resp.Permission)
			assert.False(t, resp.IsGroup)

			shareID = resp.ID
		})

		// Step 2: List secret shares
		t.Run("List Secret Shares", func(t *testing.T) {
			req := &services.ListSecretSharesRequest{
				SecretID: 1,
			}

			resp, err := client.ListSecretShares(authCtx, req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Len(t, resp.Shares, 1)
			assert.Equal(t, shareID, resp.Shares[0].ID)
		})

		// Step 3: Update share permission
		t.Run("Update Share Permission", func(t *testing.T) {
			req := &services.UpdateSharePermissionRequest{
				ShareID:    shareID,
				Permission: "write",
			}

			resp, err := client.UpdateSharePermission(authCtx, req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, shareID, resp.ID)
			assert.Equal(t, "write", resp.Permission)
		})

		// Step 4: List user shares
		t.Run("List User Shares", func(t *testing.T) {
			req := &services.ListUserSharesRequest{
				Page:     1,
				PageSize: 10,
			}

			resp, err := client.ListUserShares(authCtx, req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.GreaterOrEqual(t, len(resp.Shares), 1)

			// Find our share
			var foundShare *services.ShareRecord
			for _, share := range resp.Shares {
				if share.ID == shareID {
					foundShare = share
					break
				}
			}
			require.NotNil(t, foundShare)
			assert.Equal(t, "write", foundShare.Permission)
		})

		// Step 5: List shared secrets
		t.Run("List Shared Secrets", func(t *testing.T) {
			// Create context for recipient
			recipientUser := &interceptors.UserContext{
				ID:          2,
				Username:    "recipient",
				Permissions: []string{"secrets.read"},
			}
			recipientCtx := context.WithValue(ctx, interceptors.UserContextKey, recipientUser)

			req := &services.ListSharedSecretsRequest{
				Page:     1,
				PageSize: 10,
			}

			resp, err := client.ListSharedSecrets(recipientCtx, req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Len(t, resp.Secrets, 1)
			assert.Equal(t, uint32(1), resp.Secrets[0].Id)
		})

		// Step 6: Revoke share
		t.Run("Revoke Share", func(t *testing.T) {
			req := &services.RevokeShareRequest{
				ShareID: shareID,
			}

			resp, err := client.RevokeShare(authCtx, req)
			require.NoError(t, err)
			assert.IsType(t, &emptypb.Empty{}, resp)
		})

		// Step 7: Verify share is revoked
		t.Run("Verify Share Revoked", func(t *testing.T) {
			req := &services.ListSecretSharesRequest{
				SecretID: 1,
			}

			resp, err := client.ListSecretShares(authCtx, req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Len(t, resp.Shares, 0)
		})
	})

	t.Run("Group Sharing via gRPC", func(t *testing.T) {
		// Create authenticated context
		user := &interceptors.UserContext{
			ID:          1,
			Username:    "testuser",
			Permissions: []string{"secrets.write", "secrets.read"},
		}
		authCtx := context.WithValue(ctx, interceptors.UserContextKey, user)

		var groupShareID uint32

		// Step 1: Share with group
		t.Run("Share with Group", func(t *testing.T) {
			req := &services.ShareSecretRequest{
				SecretID:    2,
				RecipientID: 1, // Group ID
				IsGroup:     true,
				Permission:  "read",
			}

			resp, err := client.ShareSecret(authCtx, req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, uint32(2), resp.SecretID)
			assert.Equal(t, uint32(1), resp.RecipientID)
			assert.True(t, resp.IsGroup)
			assert.Equal(t, "read", resp.Permission)

			groupShareID = resp.ID
		})

		// Step 2: Update group permission
		t.Run("Update Group Permission", func(t *testing.T) {
			req := &services.UpdateSharePermissionRequest{
				ShareID:    groupShareID,
				Permission: "write",
			}

			resp, err := client.UpdateSharePermission(authCtx, req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "write", resp.Permission)
		})

		// Step 3: Revoke group share
		t.Run("Revoke Group Share", func(t *testing.T) {
			req := &services.RevokeShareRequest{
				ShareID: groupShareID,
			}

			resp, err := client.RevokeShare(authCtx, req)
			require.NoError(t, err)
			assert.IsType(t, &emptypb.Empty{}, resp)
		})
	})

	t.Run("Permission Enforcement via gRPC", func(t *testing.T) {
		// Create owner context
		owner := &interceptors.UserContext{
			ID:          1,
			Username:    "owner",
			Permissions: []string{"secrets.write", "secrets.read"},
		}
		ownerCtx := context.WithValue(ctx, interceptors.UserContextKey, owner)

		// Create recipient context with limited permissions
		recipient := &interceptors.UserContext{
			ID:          2,
			Username:    "recipient",
			Permissions: []string{"secrets.read"}, // No write permission
		}
		recipientCtx := context.WithValue(ctx, interceptors.UserContextKey, recipient)

		var shareID uint32

		// Step 1: Owner shares secret
		t.Run("Owner Shares Secret", func(t *testing.T) {
			req := &services.ShareSecretRequest{
				SecretID:    3,
				RecipientID: 2,
				Permission:  "read",
			}

			resp, err := client.ShareSecret(ownerCtx, req)
			require.NoError(t, err)
			shareID = resp.ID
		})

		// Step 2: Recipient tries to update permission (should fail)
		t.Run("Recipient Cannot Update Permission", func(t *testing.T) {
			req := &services.UpdateSharePermissionRequest{
				ShareID:    shareID,
				Permission: "write",
			}

			_, err := client.UpdateSharePermission(recipientCtx, req)
			assert.Error(t, err)
		})

		// Step 3: Recipient tries to revoke share (should fail)
		t.Run("Recipient Cannot Revoke Share", func(t *testing.T) {
			req := &services.RevokeShareRequest{
				ShareID: shareID,
			}

			_, err := client.RevokeShare(recipientCtx, req)
			assert.Error(t, err)
		})

		// Step 4: Owner can still manage share
		t.Run("Owner Can Update Permission", func(t *testing.T) {
			req := &services.UpdateSharePermissionRequest{
				ShareID:    shareID,
				Permission: "write",
			}

			resp, err := client.UpdateSharePermission(ownerCtx, req)
			require.NoError(t, err)
			assert.Equal(t, "write", resp.Permission)
		})

		// Clean up
		t.Run("Owner Can Revoke Share", func(t *testing.T) {
			req := &services.RevokeShareRequest{
				ShareID: shareID,
			}

			resp, err := client.RevokeShare(ownerCtx, req)
			require.NoError(t, err)
			assert.IsType(t, &emptypb.Empty{}, resp)
		})
	})

	t.Run("Error Scenarios via gRPC", func(t *testing.T) {
		// Test unauthenticated access
		t.Run("Unauthenticated Access", func(t *testing.T) {
			req := &services.ShareSecretRequest{
				SecretID:    1,
				RecipientID: 2,
				Permission:  "read",
			}

			_, err := client.ShareSecret(ctx, req) // No auth context
			assert.Error(t, err)
		})

		// Test insufficient permissions
		t.Run("Insufficient Permissions", func(t *testing.T) {
			limitedUser := &interceptors.UserContext{
				ID:          3,
				Username:    "limited",
				Permissions: []string{"secrets.read"}, // No write permission
			}
			limitedCtx := context.WithValue(ctx, interceptors.UserContextKey, limitedUser)

			req := &services.ShareSecretRequest{
				SecretID:    1,
				RecipientID: 2,
				Permission:  "read",
			}

			_, err := client.ShareSecret(limitedCtx, req)
			assert.Error(t, err)
		})

		// Test invalid request data
		t.Run("Invalid Request Data", func(t *testing.T) {
			user := &interceptors.UserContext{
				ID:          1,
				Username:    "testuser",
				Permissions: []string{"secrets.write"},
			}
			authCtx := context.WithValue(ctx, interceptors.UserContextKey, user)

			req := &services.ShareSecretRequest{
				SecretID:    0, // Invalid secret ID
				RecipientID: 2,
				Permission:  "read",
			}

			_, err := client.ShareSecret(authCtx, req)
			assert.Error(t, err)
		})

		// Test non-existent share operations
		t.Run("Non-existent Share Operations", func(t *testing.T) {
			user := &interceptors.UserContext{
				ID:          1,
				Username:    "testuser",
				Permissions: []string{"secrets.write"},
			}
			authCtx := context.WithValue(ctx, interceptors.UserContextKey, user)

			// Update non-existent share
			updateReq := &services.UpdateSharePermissionRequest{
				ShareID:    99999,
				Permission: "write",
			}

			_, err := client.UpdateSharePermission(authCtx, updateReq)
			assert.Error(t, err)

			// Revoke non-existent share
			revokeReq := &services.RevokeShareRequest{
				ShareID: 99999,
			}

			_, err = client.RevokeShare(authCtx, revokeReq)
			assert.Error(t, err)
		})
	})
}

// TestSharingGRPCConcurrency tests concurrent gRPC operations
func TestSharingGRPCConcurrency(t *testing.T) {
	// Initialize i18n for testing
	err := i18n.InitializeForTesting()
	require.NoError(t, err)
	defer i18n.ResetForTesting()

	// Setup gRPC server (similar to above)
	lis = bufconn.Listen(bufSize)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.AuthInterceptor),
	)

	shareService, err := services.NewShareService(nil)
	require.NoError(t, err)

	services.RegisterShareServiceServer(server, shareService)

	go func() {
		if err := server.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	defer server.Stop()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := services.NewShareServiceClient(conn)

	t.Run("Concurrent Share Operations", func(t *testing.T) {
		const numGoroutines = 10
		const requestsPerGoroutine = 5

		results := make(chan error, numGoroutines*requestsPerGoroutine)

		user := &interceptors.UserContext{
			ID:          1,
			Username:    "testuser",
			Permissions: []string{"secrets.write"},
		}
		authCtx := context.WithValue(ctx, interceptors.UserContextKey, user)

		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				for j := 0; j < requestsPerGoroutine; j++ {
					req := &services.ShareSecretRequest{
						SecretID:    uint32(goroutineID + 1), // Different secrets
						RecipientID: uint32(goroutineID*requestsPerGoroutine + j + 10),
						Permission:  "read",
					}

					_, err := client.ShareSecret(authCtx, req)
					results <- err
				}
			}(i)
		}

		// Collect results
		successCount := 0
		for i := 0; i < numGoroutines*requestsPerGoroutine; i++ {
			select {
			case err := <-results:
				if err == nil {
					successCount++
				}
			case <-time.After(10 * time.Second):
				t.Fatal("Timeout waiting for concurrent requests")
			}
		}

		// At least 80% success rate
		expectedMinSuccess := int(float64(numGoroutines*requestsPerGoroutine) * 0.8)
		assert.GreaterOrEqual(t, successCount, expectedMinSuccess)
	})
}

// bufDialer is a helper function for testing with bufconn
func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}