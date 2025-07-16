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

// Create добавляет новую сессию в базу
func (r *sessionRepo) Create(session *models.Session) error {
	return r.db.Create(session).Error
}

// GetByToken возвращает сессию по токену
func (r *sessionRepo) GetByToken(token string) (*models.Session, error) {
	var session models.Session
	err := r.db.Where("token = ?", token).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// DeleteExpired удаляет все сессии с истекшим сроком действия
func (r *sessionRepo) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.Session{}).Error
}
