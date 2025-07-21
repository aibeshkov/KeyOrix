package services

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/secretlyhq/secretly/internal/config"
	"github.com/secretlyhq/secretly/internal/i18n"
	"github.com/secretlyhq/secretly/internal/storage/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuthServiceTest(t *testing.T) (*AuthService, *gorm.DB, func()) {
	// Initialize i18n system for tests
	cfg := &config.Config{
		Locale: config.LocaleConfig{
			Language:         "en",
			FallbackLanguage: "en",
		},
	}
	
	// Reset global state and initialize i18n
	i18n.ResetForTesting()
	err := i18n.Initialize(cfg)
	require.NoError(t, err)

	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "auth_service_test")
	require.NoError(t, err)

	// Create test database
	db, err := gorm.Open(sqlite.Open(filepath.Join(tempDir, "test.db")), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate tables
	err = db.AutoMigrate(
		&models.User{},
		&models.APIClient{},
		&models.Session{},
		&models.APIToken{},
		&models.PasswordReset{},
	)
	require.NoError(t, err)

	// Create encryption config with encryption disabled for simpler testing
	encCfg := &config.EncryptionConfig{
		Enabled: false,
	}

	// Create auth service
	authService, err := NewAuthService(db, encCfg, tempDir)
	require.NoError(t, err)

	// Cleanup function
	cleanup := func() {
		_ = os.RemoveAll(tempDir)
	}

	return authService, db, cleanup
}

func TestAPIClientService_CreateAPIClient(t *testing.T) {
	authService, db, cleanup := setupAuthServiceTest(t)
	defer cleanup()

	clientService := &APIClientService{AuthService: authService}

	// Create API client
	client, clientSecret, err := clientService.CreateAPIClient(
		"Test Client",
		"Test API Client Description",
		[]string{"read", "write"},
	)

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotEmpty(t, clientSecret)
	assert.Equal(t, "Test Client", client.Name)
	assert.Equal(t, "Test API Client Description", client.Description)
	assert.Equal(t, "read write", client.Scopes)
	assert.True(t, client.IsActive)
	assert.NotEmpty(t, client.ClientID)

	// Verify client was stored in database
	var storedClient models.APIClient
	err = db.First(&storedClient, client.ID).Error
	require.NoError(t, err)
	assert.Equal(t, client.Name, storedClient.Name)
	assert.NotEmpty(t, storedClient.EncryptedClientSecret)
}

func TestAPIClientService_ValidateAPIClient(t *testing.T) {
	authService, _, cleanup := setupAuthServiceTest(t)
	defer cleanup()

	clientService := &APIClientService{AuthService: authService}

	// Create API client
	client, clientSecret, err := clientService.CreateAPIClient(
		"Validation Test Client",
		"Client for validation testing",
		[]string{"read"},
	)
	require.NoError(t, err)

	tests := []struct {
		name         string
		clientID     string
		clientSecret string
		expectError  bool
	}{
		{
			name:         "valid credentials",
			clientID:     client.ClientID,
			clientSecret: clientSecret,
			expectError:  false,
		},
		{
			name:         "invalid client ID",
			clientID:     "invalid-client-id",
			clientSecret: clientSecret,
			expectError:  true,
		},
		{
			name:         "invalid client secret",
			clientID:     client.ClientID,
			clientSecret: "invalid-secret",
			expectError:  true,
		},
		{
			name:         "empty credentials",
			clientID:     "",
			clientSecret: "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validatedClient, err := clientService.ValidateAPIClient(tt.clientID, tt.clientSecret)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, validatedClient)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, validatedClient)
				assert.Equal(t, client.ID, validatedClient.ID)
				assert.Equal(t, client.ClientID, validatedClient.ClientID)
			}
		})
	}
}

func TestAPIClientService_RevokeAPIClient(t *testing.T) {
	authService, db, cleanup := setupAuthServiceTest(t)
	defer cleanup()

	clientService := &APIClientService{AuthService: authService}

	// Create API client
	client, clientSecret, err := clientService.CreateAPIClient(
		"Revoke Test Client",
		"Client for revocation testing",
		[]string{"read"},
	)
	require.NoError(t, err)

	// Verify client is active
	validatedClient, err := clientService.ValidateAPIClient(client.ClientID, clientSecret)
	require.NoError(t, err)
	assert.True(t, validatedClient.IsActive)

	// Revoke client
	err = clientService.RevokeAPIClient(client.ClientID)
	require.NoError(t, err)

	// Verify client is no longer active
	_, err = clientService.ValidateAPIClient(client.ClientID, clientSecret)
	assert.Error(t, err)

	// Verify client is marked as inactive in database
	var storedClient models.APIClient
	err = db.First(&storedClient, client.ID).Error
	require.NoError(t, err)
	assert.False(t, storedClient.IsActive)
}

func TestSessionService_CreateSession(t *testing.T) {
	authService, db, cleanup := setupAuthServiceTest(t)
	defer cleanup()

	sessionService := &SessionService{AuthService: authService}

	// Create user for session
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	expiresAt := time.Now().Add(24 * time.Hour)

	// Create session
	session, sessionToken, err := sessionService.CreateSession(user.ID, &expiresAt)

	require.NoError(t, err)
	assert.NotNil(t, session)
	assert.NotEmpty(t, sessionToken)
	assert.Equal(t, user.ID, session.UserID)
	assert.NotNil(t, session.ExpiresAt)

	// Verify session was stored in database
	var storedSession models.Session
	err = db.First(&storedSession, session.ID).Error
	require.NoError(t, err)
	assert.Equal(t, user.ID, storedSession.UserID)
	assert.NotEmpty(t, storedSession.EncryptedSessionToken)
}

func TestSessionService_ValidateSession(t *testing.T) {
	authService, db, cleanup := setupAuthServiceTest(t)
	defer cleanup()

	sessionService := &SessionService{AuthService: authService}

	// Create user
	user := &models.User{
		Username: "sessionuser",
		Email:    "session@example.com",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	expiresAt := time.Now().Add(24 * time.Hour)

	// Create session
	session, sessionToken, err := sessionService.CreateSession(user.ID, &expiresAt)
	require.NoError(t, err)

	// Validate session
	validatedSession, err := sessionService.ValidateSession(sessionToken)
	require.NoError(t, err)
	assert.NotNil(t, validatedSession)
	assert.Equal(t, session.ID, validatedSession.ID)
	assert.Equal(t, user.ID, validatedSession.UserID)

	// Test invalid session token
	_, err = sessionService.ValidateSession("invalid-session-token")
	assert.Error(t, err)
}

func TestSessionService_RevokeSession(t *testing.T) {
	authService, db, cleanup := setupAuthServiceTest(t)
	defer cleanup()

	sessionService := &SessionService{AuthService: authService}

	// Create user
	user := &models.User{
		Username: "revokeuser",
		Email:    "revoke@example.com",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	expiresAt := time.Now().Add(24 * time.Hour)

	// Create session
	session, sessionToken, err := sessionService.CreateSession(user.ID, &expiresAt)
	require.NoError(t, err)

	// Verify session exists
	_, err = sessionService.ValidateSession(sessionToken)
	require.NoError(t, err)

	// Revoke session
	err = sessionService.RevokeSession(session.ID)
	require.NoError(t, err)

	// Verify session no longer exists
	_, err = sessionService.ValidateSession(sessionToken)
	assert.Error(t, err)

	// Verify session was deleted from database
	var count int64
	db.Model(&models.Session{}).Where("id = ?", session.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestAPITokenService_CreateAPIToken(t *testing.T) {
	authService, db, cleanup := setupAuthServiceTest(t)
	defer cleanup()

	tokenService := &APITokenService{AuthService: authService}

	// Create API client first
	client := &models.APIClient{
		Name:     "Token Test Client",
		ClientID: "token-test-client",
		IsActive: true,
	}
	err := db.Create(client).Error
	require.NoError(t, err)

	// Create user
	user := &models.User{
		Username: "tokenuser",
		Email:    "token@example.com",
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Create API token
	token, apiToken, err := tokenService.CreateAPIToken(client.ID, &user.ID, "read write", &expiresAt)

	require.NoError(t, err)
	assert.NotNil(t, token)
	assert.NotEmpty(t, apiToken)
	assert.Equal(t, client.ID, token.ClientID)
	assert.Equal(t, user.ID, *token.UserID)
	assert.Equal(t, "read write", token.Scope)
	assert.False(t, token.Revoked)

	// Verify token was stored in database
	var storedToken models.APIToken
	err = db.First(&storedToken, token.ID).Error
	require.NoError(t, err)
	assert.NotEmpty(t, storedToken.EncryptedToken)
}

func TestAPITokenService_ValidateAPIToken(t *testing.T) {
	authService, db, cleanup := setupAuthServiceTest(t)
	defer cleanup()

	tokenService := &APITokenService{AuthService: authService}

	// Create API client
	client := &models.APIClient{
		Name:     "Validate Token Client",
		ClientID: "validate-token-client",
		IsActive: true,
	}
	err := db.Create(client).Error
	require.NoError(t, err)

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Create API token
	token, apiToken, err := tokenService.CreateAPIToken(client.ID, nil, "read", &expiresAt)
	require.NoError(t, err)

	// Validate token
	validatedToken, err := tokenService.ValidateAPIToken(apiToken)
	require.NoError(t, err)
	assert.NotNil(t, validatedToken)
	assert.Equal(t, token.ID, validatedToken.ID)
	assert.Equal(t, client.ID, validatedToken.ClientID)

	// Test invalid token
	_, err = tokenService.ValidateAPIToken("invalid-api-token")
	assert.Error(t, err)
}

func TestPasswordResetService_CreatePasswordResetToken(t *testing.T) {
	authService, db, cleanup := setupAuthServiceTest(t)
	defer cleanup()

	resetService := &PasswordResetService{AuthService: authService}

	// Create user
	user := &models.User{
		Username: "resetuser",
		Email:    "reset@example.com",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	expiresAt := time.Now().Add(1 * time.Hour)

	// Create password reset token
	reset, resetToken, err := resetService.CreatePasswordResetToken(user.ID, expiresAt)

	require.NoError(t, err)
	assert.NotNil(t, reset)
	assert.NotEmpty(t, resetToken)
	assert.Equal(t, user.ID, reset.UserID)
	assert.NotNil(t, reset.ExpiresAt)

	// Verify reset token was stored in database
	var storedReset models.PasswordReset
	err = db.First(&storedReset, reset.ID).Error
	require.NoError(t, err)
	assert.NotEmpty(t, storedReset.EncryptedToken)
}

func TestPasswordResetService_ValidateAndConsumeToken(t *testing.T) {
	authService, db, cleanup := setupAuthServiceTest(t)
	defer cleanup()

	resetService := &PasswordResetService{AuthService: authService}

	// Create user
	user := &models.User{
		Username: "consumeuser",
		Email:    "consume@example.com",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	expiresAt := time.Now().Add(1 * time.Hour)

	// Create password reset token
	reset, resetToken, err := resetService.CreatePasswordResetToken(user.ID, expiresAt)
	require.NoError(t, err)

	// Validate token
	validatedReset, err := resetService.ValidatePasswordResetToken(resetToken)
	require.NoError(t, err)
	assert.Equal(t, reset.ID, validatedReset.ID)

	// Consume token
	consumedReset, err := resetService.ConsumePasswordResetToken(resetToken)
	require.NoError(t, err)
	assert.Equal(t, reset.ID, consumedReset.ID)

	// Verify token can't be used again
	_, err = resetService.ValidatePasswordResetToken(resetToken)
	assert.Error(t, err)

	// Verify token was deleted from database
	var count int64
	db.Model(&models.PasswordReset{}).Where("id = ?", reset.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestAuthService_GetEncryptionStatus(t *testing.T) {
	authService, _, cleanup := setupAuthServiceTest(t)
	defer cleanup()

	status := authService.GetAuthEncryptionStatus()

	assert.Contains(t, status, "enabled")
	assert.Contains(t, status, "initialized")
	// When encryption is disabled, both should be false
	assert.False(t, status["enabled"].(bool))
	assert.False(t, status["initialized"].(bool))
}

func TestGenerateSecureToken(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"short token", 16},
		{"medium token", 32},
		{"long token", 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := generateSecureToken(tt.length)
			require.NoError(t, err)
			assert.NotEmpty(t, token)
			
			// Base64 URL encoding increases length, so check it's reasonable
			assert.Greater(t, len(token), tt.length)
		})
	}
}

func TestJoinScopes(t *testing.T) {
	tests := []struct {
		name     string
		scopes   []string
		expected string
	}{
		{
			name:     "single scope",
			scopes:   []string{"read"},
			expected: "read",
		},
		{
			name:     "multiple scopes",
			scopes:   []string{"read", "write", "delete"},
			expected: "read write delete",
		},
		{
			name:     "empty scopes",
			scopes:   []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinScopes(tt.scopes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests
func BenchmarkAPIClientService_CreateAPIClient(b *testing.B) {
	authService, _, cleanup := setupAuthServiceTest(&testing.T{})
	defer cleanup()

	clientService := &APIClientService{AuthService: authService}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := clientService.CreateAPIClient(
			"Benchmark Client",
			"Benchmark Description",
			[]string{"read"},
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSessionService_CreateSession(b *testing.B) {
	authService, db, cleanup := setupAuthServiceTest(&testing.T{})
	defer cleanup()

	sessionService := &SessionService{AuthService: authService}

	// Create user
	user := &models.User{Username: "benchuser", Email: "bench@example.com"}
	db.Create(user)

	expiresAt := time.Now().Add(24 * time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := sessionService.CreateSession(user.ID, &expiresAt)
		if err != nil {
			b.Fatal(err)
		}
	}
}