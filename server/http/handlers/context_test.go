package handlers

import (
	"context"
	"testing"

	"github.com/secretlyhq/secretly/server/middleware"
	"github.com/stretchr/testify/assert"
)

func TestContextSetup(t *testing.T) {
	// Test that we can set and retrieve user context correctly
	userCtx := &middleware.UserContext{
		UserID:      1,
		Username:    "admin",
		Email:       "admin@example.com",
		Permissions: []string{"secrets.read"},
	}

	// Test with the same context key as middleware
	ctx := context.WithValue(context.Background(), middleware.GetUserContextKey(), userCtx)

	// Try to retrieve it using middleware function
	retrievedCtx := middleware.GetUserFromContext(ctx)

	assert.NotNil(t, retrievedCtx)
	assert.Equal(t, userCtx.UserID, retrievedCtx.UserID)
	assert.Equal(t, userCtx.Username, retrievedCtx.Username)

	// Test the helper function
	ctx2 := addAuthContext(context.Background(), "valid-token")
	retrievedCtx2 := middleware.GetUserFromContext(ctx2)

	assert.NotNil(t, retrievedCtx2)
	assert.Equal(t, uint(1), retrievedCtx2.UserID)
	assert.Equal(t, "admin", retrievedCtx2.Username)
}
