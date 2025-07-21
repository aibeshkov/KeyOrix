package services

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/secretlyhq/secretly/internal/config"
	"github.com/secretlyhq/secretly/internal/encryption"
	"github.com/secretlyhq/secretly/internal/i18n"
	"github.com/secretlyhq/secretly/internal/storage/models"
	"gorm.io/gorm"
)

// AuthService handles authentication operations with encryption
type AuthService struct {
	db         *gorm.DB
	encryption *encryption.AuthEncryption
}

// NewAuthService creates a new authentication service
func NewAuthService(db *gorm.DB, cfg *config.EncryptionConfig, baseDir string) (*AuthService, error) {
	authEncryption := encryption.NewAuthEncryption(cfg, baseDir, db)
	
	// Initialize encryption if enabled
	if err := authEncryption.Initialize(); err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("ErrorInitializationFailed", nil), err)
	}

	return &AuthService{
		db:         db,
		encryption: authEncryption,
	}, nil
}

// APIClientService handles API client operations
type APIClientService struct {
	*AuthService
}

// CreateAPIClient creates a new API client with encrypted secret
func (s *APIClientService) CreateAPIClient(name, description string, scopes []string) (*models.APIClient, string, error) {
	// Generate client ID and secret
	clientID, err := generateSecureToken(32)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", i18n.T("ErrorTokenGeneration", nil), err)
	}

	clientSecret, err := generateSecureToken(64)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", i18n.T("ErrorTokenGeneration", nil), err)
	}

	// Create API client
	client := &models.APIClient{
		Name:        name,
		Description: description,
		ClientID:    clientID,
		Scopes:      joinScopes(scopes),
		IsActive:    true,
		CreatedAt:   time.Now(),
	}

	// Store with encrypted secret
	if err := s.encryption.StoreEncryptedAPIClient(client, clientSecret); err != nil {
		return nil, "", fmt.Errorf("%s: %w", i18n.T("ErrorStorageFailed", nil), err)
	}

	return client, clientSecret, nil
}

// ValidateAPIClient validates an API client's credentials
func (s *APIClientService) ValidateAPIClient(clientID, clientSecret string) (*models.APIClient, error) {
	var client models.APIClient
	if err := s.db.Where("client_id = ? AND is_active = ?", clientID, true).First(&client).Error; err != nil {
		return nil, fmt.Errorf("%s", i18n.T("ErrorInvalidCredentials", nil))
	}

	// Retrieve and validate client secret
	storedSecret, err := s.encryption.RetrieveAPIClientSecret(clientID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("ErrorInternalServer", nil), err)
	}

	if storedSecret != clientSecret {
		return nil, fmt.Errorf("%s", i18n.T("ErrorInvalidCredentials", nil))
	}

	return &client, nil
}

// RevokeAPIClient deactivates an API client
func (s *APIClientService) RevokeAPIClient(clientID string) error {
	result := s.db.Model(&models.APIClient{}).
		Where("client_id = ?", clientID).
		Update("is_active", false)

	if result.Error != nil {
		return fmt.Errorf("%s: %w", i18n.T("ErrorInternalServer", nil), result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("%s", i18n.T("ErrorNotFound", nil))
	}

	return nil
}

// SessionService handles session operations
type SessionService struct {
	*AuthService
}

// CreateSession creates a new session with encrypted token
func (s *SessionService) CreateSession(userID uint, expiresAt *time.Time) (*models.Session, string, error) {
	// Generate session token
	sessionToken, err := generateSecureToken(64)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", i18n.T("ErrorTokenGeneration", nil), err)
	}

	// Create session
	session := &models.Session{
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}

	// Store with encrypted token
	if err := s.encryption.StoreEncryptedSession(session, sessionToken); err != nil {
		return nil, "", fmt.Errorf("%s: %w", i18n.T("ErrorStorageFailed", nil), err)
	}

	return session, sessionToken, nil
}

// ValidateSession validates a session token
func (s *SessionService) ValidateSession(sessionToken string) (*models.Session, error) {
	// For validation, we need to find the session by comparing encrypted tokens
	// This is a limitation of encrypted storage - we can't directly query encrypted fields
	// In practice, you might want to use a hash-based approach for lookups
	
	var sessions []models.Session
	if err := s.db.Where("expires_at IS NULL OR expires_at > ?", time.Now()).Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("ErrorRetrievalFailed", nil), err)
	}

	// Check each session (this is not optimal for large datasets)
	for _, session := range sessions {
		isValid, err := s.encryption.ValidateEncryptedToken(
			session.EncryptedSessionToken,
			[]byte(session.SessionTokenMetadata),
			sessionToken,
		)
		if err != nil {
			continue // Skip invalid sessions
		}
		if isValid {
			return &session, nil
		}
	}

	return nil, fmt.Errorf("%s", i18n.T("ErrorUnauthorized", nil))
}

// RevokeSession removes a session
func (s *SessionService) RevokeSession(sessionID uint) error {
	result := s.db.Delete(&models.Session{}, sessionID)
	if result.Error != nil {
		return fmt.Errorf("%s: %w", i18n.T("ErrorInternalServer", nil), result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("%s", i18n.T("ErrorNotFound", nil))
	}

	return nil
}

// CleanupExpiredSessions removes expired sessions
func (s *SessionService) CleanupExpiredSessions() error {
	result := s.db.Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Delete(&models.Session{})

	if result.Error != nil {
		return fmt.Errorf("%s: %w", i18n.T("ErrorCleanupFailed", nil), result.Error)
	}

	return nil
}

// APITokenService handles API token operations
type APITokenService struct {
	*AuthService
}

// CreateAPIToken creates a new API token with encrypted token
func (s *APITokenService) CreateAPIToken(clientID uint, userID *uint, scope string, expiresAt *time.Time) (*models.APIToken, string, error) {
	// Generate API token
	apiToken, err := generateSecureToken(64)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", i18n.T("ErrorTokenGeneration", nil), err)
	}

	// Create API token
	token := &models.APIToken{
		ClientID:  clientID,
		UserID:    userID,
		Scope:     scope,
		Revoked:   false,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	// Store with encrypted token
	if err := s.encryption.StoreEncryptedAPIToken(token, apiToken); err != nil {
		return nil, "", fmt.Errorf("%s: %w", i18n.T("ErrorStorageFailed", nil), err)
	}

	return token, apiToken, nil
}

// ValidateAPIToken validates an API token
func (s *APITokenService) ValidateAPIToken(apiToken string) (*models.APIToken, error) {
	// Similar to session validation, we need to check all active tokens
	var tokens []models.APIToken
	if err := s.db.Where("revoked = ? AND (expires_at IS NULL OR expires_at > ?)", 
		false, time.Now()).Find(&tokens).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("ErrorRetrievalFailed", nil), err)
	}

	// Check each token
	for _, token := range tokens {
		storedToken, err := s.encryption.DecryptAPIToken(token.EncryptedToken, []byte(token.TokenMetadata))
		if err != nil {
			continue // Skip invalid tokens
		}
		if storedToken == apiToken {
			return &token, nil
		}
	}

	return nil, fmt.Errorf("%s", i18n.T("ErrorUnauthorized", nil))
}

// RevokeAPIToken revokes an API token
func (s *APITokenService) RevokeAPIToken(tokenID uint) error {
	result := s.db.Model(&models.APIToken{}).
		Where("id = ?", tokenID).
		Update("revoked", true)

	if result.Error != nil {
		return fmt.Errorf("%s: %w", i18n.T("ErrorInternalServer", nil), result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("%s", i18n.T("ErrorNotFound", nil))
	}

	return nil
}

// PasswordResetService handles password reset operations
type PasswordResetService struct {
	*AuthService
}

// CreatePasswordResetToken creates a new password reset token
func (s *PasswordResetService) CreatePasswordResetToken(userID uint, expiresAt time.Time) (*models.PasswordReset, string, error) {
	// Generate reset token
	resetToken, err := generateSecureToken(64)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", i18n.T("ErrorTokenGeneration", nil), err)
	}

	// Encrypt the token
	encryptedToken, metadata, err := s.encryption.EncryptPasswordResetToken(resetToken)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", i18n.T("ErrorEncryptionFailed", nil), err)
	}

	// Create password reset record
	passwordReset := &models.PasswordReset{
		UserID:         userID,
		EncryptedToken: encryptedToken,
		TokenMetadata:  metadata,
		ExpiresAt:      &expiresAt,
		CreatedAt:      time.Now(),
	}

	// Store in database
	if err := s.db.Create(passwordReset).Error; err != nil {
		return nil, "", fmt.Errorf("%s: %w", i18n.T("ErrorStorageFailed", nil), err)
	}

	return passwordReset, resetToken, nil
}

// ValidatePasswordResetToken validates a password reset token
func (s *PasswordResetService) ValidatePasswordResetToken(resetToken string) (*models.PasswordReset, error) {
	// Retrieve all active reset tokens
	var resets []models.PasswordReset
	if err := s.db.Where("expires_at > ?", time.Now()).Find(&resets).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("ErrorRetrievalFailed", nil), err)
	}

	// Check each token
	for _, reset := range resets {
		storedToken, err := s.encryption.DecryptPasswordResetToken(reset.EncryptedToken, []byte(reset.TokenMetadata))
		if err != nil {
			continue // Skip invalid tokens
		}
		if storedToken == resetToken {
			return &reset, nil
		}
	}

	return nil, fmt.Errorf("%s", i18n.T("ErrorInvalidResetToken", nil))
}

// ConsumePasswordResetToken validates and removes a password reset token
func (s *PasswordResetService) ConsumePasswordResetToken(resetToken string) (*models.PasswordReset, error) {
	// Validate the token first
	reset, err := s.ValidatePasswordResetToken(resetToken)
	if err != nil {
		return nil, err
	}

	// Remove the token to prevent reuse
	if err := s.db.Delete(reset).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("ErrorStorageFailed", nil), err)
	}

	return reset, nil
}

// Helper functions

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// joinScopes joins scope strings
func joinScopes(scopes []string) string {
	result := ""
	for i, scope := range scopes {
		if i > 0 {
			result += " "
		}
		result += scope
	}
	return result
}

// GetAuthEncryptionStatus returns the encryption status for authentication data
func (s *AuthService) GetAuthEncryptionStatus() map[string]interface{} {
	return s.encryption.GetAuthEncryptionStatus()
}

// RotateAuthKeys rotates encryption keys for all authentication data
func (s *AuthService) RotateAuthKeys() error {
	return s.encryption.RotateAuthEncryption()
}