package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/secretlyhq/secretly/internal/i18n"
	"github.com/secretlyhq/secretly/internal/storage/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type RBACService struct {
	db *gorm.DB
}

// RBACAuditLog represents an audit log entry for RBAC operations
type RBACAuditLog struct {
	ID           uint   `gorm:"primaryKey"`
	Action       string `gorm:"not null"` // 'ASSIGN_ROLE', 'REMOVE_ROLE', 'CREATE_ROLE', 'DELETE_ROLE'
	ActorUserID  *uint  `gorm:"index"`
	TargetUserID *uint  `gorm:"index"`
	RoleID       *uint  `gorm:"index"`
	NamespaceID  *uint  `gorm:"index"`
	Details      string // JSON with additional details
	CreatedAt    time.Time
}

// Permission represents a fine-grained permission
type Permission struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"unique;not null"`
	Description string
	Resource    string `gorm:"not null;index"`
	Action      string `gorm:"not null;index"`
	CreatedAt   time.Time
}

// RolePermission represents the many-to-many relationship between roles and permissions
type RolePermission struct {
	RoleID       uint `gorm:"primaryKey"`
	PermissionID uint `gorm:"primaryKey"`
}

func NewRBACService() (*RBACService, error) {
	db, err := gorm.Open(sqlite.Open("secretly.db"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("ErrorDatabaseConnection", nil), err)
	}

	return &RBACService{db: db}, nil
}

func (s *RBACService) AssignRoleToUser(userEmail, roleName string) error {
	// Find user by email
	var user models.User
	if err := s.db.Where("email = ?", userEmail).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("%s", i18n.T("ErrorUserNotFound", nil))
		}
		return fmt.Errorf("%s: %w", i18n.T("ErrorInternalServer", nil), err)
	}

	// Find role by name
	var role models.Role
	if err := s.db.Where("name = ?", roleName).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("%s", i18n.T("ErrorRoleNotFound", nil))
		}
		return fmt.Errorf("%s: %w", i18n.T("ErrorInternalServer", nil), err)
	}

	// Check if user already has this role
	var existingUserRole models.UserRole
	err := s.db.Where("user_id = ? AND role_id = ?", user.ID, role.ID).First(&existingUserRole).Error
	if err == nil {
		return fmt.Errorf("%s", i18n.T("ErrorRoleAlreadyAssigned", nil))
	}
	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("%s: %w", i18n.T("ErrorInternalServer", nil), err)
	}

	// Create new user role assignment
	userRole := models.UserRole{
		UserID: user.ID,
		RoleID: role.ID,
	}

	if err := s.db.Create(&userRole).Error; err != nil {
		return fmt.Errorf("%s: %w", i18n.T("ErrorStorageFailed", nil), err)
	}

	// Log the role assignment
	s.logRBACAction("ASSIGN_ROLE", nil, &user.ID, &role.ID, nil, map[string]interface{}{
		"user_email": userEmail,
		"role_name":  roleName,
	})

	return nil
}

func (s *RBACService) ListRoles() ([]models.Role, error) {
	var roles []models.Role
	if err := s.db.Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("ErrorRetrievalFailed", nil), err)
	}
	return roles, nil
}

func (s *RBACService) ListUserRoles(userEmail string) ([]models.Role, error) {
	// Find user by email
	var user models.User
	if err := s.db.Where("email = ?", userEmail).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%s", i18n.T("ErrorUserNotFound", nil))
		}
		return nil, fmt.Errorf("%s: %w", i18n.T("ErrorInternalServer", nil), err)
	}

	// Get user's roles
	var roles []models.Role
	if err := s.db.Table("roles").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", user.ID).
		Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("ErrorRetrievalFailed", nil), err)
	}

	return roles, nil
}

func (s *RBACService) RemoveRoleFromUser(userEmail, roleName string) error {
	// Find user by email
	var user models.User
	if err := s.db.Where("email = ?", userEmail).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("%s", i18n.T("ErrorUserNotFound", nil))
		}
		return fmt.Errorf("%s: %w", i18n.T("ErrorInternalServer", nil), err)
	}

	// Find role by name
	var role models.Role
	if err := s.db.Where("name = ?", roleName).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("%s", i18n.T("ErrorRoleNotFound", nil))
		}
		return fmt.Errorf("%s: %w", i18n.T("ErrorInternalServer", nil), err)
	}

	// Check if user has this role
	var userRole models.UserRole
	err := s.db.Where("user_id = ? AND role_id = ?", user.ID, role.ID).First(&userRole).Error
	if err == gorm.ErrRecordNotFound {
		return fmt.Errorf("%s", i18n.T("ErrorRoleNotAssigned", nil))
	}
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("ErrorInternalServer", nil), err)
	}

	// Remove the role assignment
	if err := s.db.Where("user_id = ? AND role_id = ?", user.ID, role.ID).Delete(&models.UserRole{}).Error; err != nil {
		return fmt.Errorf("%s: %w", i18n.T("ErrorStorageFailed", nil), err)
	}

	// Log the role removal
	s.logRBACAction("REMOVE_ROLE", nil, &user.ID, &role.ID, nil, map[string]interface{}{
		"user_email": userEmail,
		"role_name":  roleName,
	})

	return nil
}

// logRBACAction logs RBAC operations for audit purposes
func (s *RBACService) logRBACAction(action string, actorUserID, targetUserID, roleID, namespaceID *uint, details map[string]interface{}) {
	detailsJSON, _ := json.Marshal(details)

	auditLog := RBACAuditLog{
		Action:       action,
		ActorUserID:  actorUserID,
		TargetUserID: targetUserID,
		RoleID:       roleID,
		NamespaceID:  namespaceID,
		Details:      string(detailsJSON),
		CreatedAt:    time.Now(),
	}

	// Log the audit entry (ignore errors to not break the main operation)
	s.db.Table("rbac_audit_log").Create(&auditLog)
}

// GetRBACAuditLogs retrieves RBAC audit logs with optional filtering
func (s *RBACService) GetRBACAuditLogs(limit int, offset int) ([]RBACAuditLog, error) {
	var logs []RBACAuditLog
	query := s.db.Table("rbac_audit_log").Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("ErrorRetrievalFailed", nil), err)
	}

	return logs, nil
}

// HasPermission checks if a user has a specific permission
func (s *RBACService) HasPermission(userEmail, permissionName string) (bool, error) {
	// Find user by email
	var user models.User
	if err := s.db.Where("email = ?", userEmail).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, fmt.Errorf("%s", i18n.T("ErrorUserNotFound", nil))
		}
		return false, fmt.Errorf("%s: %w", i18n.T("ErrorInternalServer", nil), err)
	}

	// Check if user has the permission through any of their roles
	var count int64
	err := s.db.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Where("user_roles.user_id = ? AND permissions.name = ?", user.ID, permissionName).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("%s: %w", i18n.T("ErrorInternalServer", nil), err)
	}

	return count > 0, nil
}

// ListUserPermissions returns all permissions for a user
func (s *RBACService) ListUserPermissions(userEmail string) ([]Permission, error) {
	// Find user by email
	var user models.User
	if err := s.db.Where("email = ?", userEmail).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%s", i18n.T("ErrorUserNotFound", nil))
		}
		return nil, fmt.Errorf("%s: %w", i18n.T("ErrorInternalServer", nil), err)
	}

	// Get user's permissions through their roles
	var permissions []Permission
	err := s.db.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Where("user_roles.user_id = ?", user.ID).
		Group("permissions.id").
		Find(&permissions).Error

	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("ErrorRetrievalFailed", nil), err)
	}

	return permissions, nil
}
