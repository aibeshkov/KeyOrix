package repository

import (
	"github.com/secretlyhq/secretly/internal/storage/models"
	"gorm.io/gorm"
)

type AuditRepository interface {
	LogEvent(event *models.AuditEvent) error
	ListByUser(userID uint) ([]models.AuditEvent, error)
}

type auditRepo struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepo{db}
}

// LogEvent saves an audit event to the database
func (r *auditRepo) LogEvent(event *models.AuditEvent) error {
	return r.db.Create(event).Error
}

// ListByUser returns a list of audit events for a user by userID
func (r *auditRepo) ListByUser(userID uint) ([]models.AuditEvent, error) {
	var events []models.AuditEvent
	err := r.db.Where("user_id = ?", userID).Find(&events).Error
	return events, err
}
