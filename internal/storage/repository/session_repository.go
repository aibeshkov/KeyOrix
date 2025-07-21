package repository

import (
	"time"

	"github.com/secretlyhq/secretly/internal/storage/models"
	"gorm.io/gorm"
)

type SessionRepository interface {
	Create(session *models.Session) error
	GetByToken(token string) (*models.Session, error)
	DeleteExpired() error
}

type sessionRepo struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepo{db}
}

// Create adds a new session to the database
func (r *sessionRepo) Create(session *models.Session) error {
	return r.db.Create(session).Error
}

// GetByToken returns a session by token
func (r *sessionRepo) GetByToken(token string) (*models.Session, error) {
	var session models.Session
	err := r.db.Where("session_token = ?", token).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// DeleteExpired deletes all expired sessions
func (r *sessionRepo) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.Session{}).Error
}
