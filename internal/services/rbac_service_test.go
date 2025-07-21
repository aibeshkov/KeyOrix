package services

import (
	"strings"
	"testing"

	"github.com/secretlyhq/secretly/internal/storage/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.UserRole{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

func TestRBACService_AssignRoleToUser(t *testing.T) {
	db := setupTestDB(t)
	service := &RBACService{db: db}

	// Create test data
	user := models.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	role := models.Role{
		Name:        "admin",
		Description: "Administrator role",
	}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}

	t.Run("successful role assignment", func(t *testing.T) {
		err := service.AssignRoleToUser("test@example.com", "admin")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Verify assignment was created
		var userRole models.UserRole
		err = db.Where("user_id = ? AND role_id = ?", user.ID, role.ID).First(&userRole).Error
		if err != nil {
			t.Errorf("Expected to find user role, got error: %v", err)
		}
		if userRole.UserID != user.ID {
			t.Errorf("Expected UserID %d, got %d", user.ID, userRole.UserID)
		}
		if userRole.RoleID != role.ID {
			t.Errorf("Expected RoleID %d, got %d", role.ID, userRole.RoleID)
		}
	})

	t.Run("duplicate role assignment", func(t *testing.T) {
		err := service.AssignRoleToUser("test@example.com", "admin")
		if err == nil {
			t.Error("Expected error for duplicate role assignment")
		}
		if err != nil && !contains(err.Error(), "Role already assigned") {
			t.Errorf("Expected 'Role already assigned' error, got: %v", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		err := service.AssignRoleToUser("nonexistent@example.com", "admin")
		if err == nil {
			t.Error("Expected error for non-existent user")
		}
		if err != nil && !contains(err.Error(), "not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})

	t.Run("role not found", func(t *testing.T) {
		err := service.AssignRoleToUser("test@example.com", "nonexistent")
		if err == nil {
			t.Error("Expected error for non-existent role")
		}
		if err != nil && !contains(err.Error(), "not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})
}

func TestRBACService_RemoveRoleFromUser(t *testing.T) {
	db := setupTestDB(t)
	service := &RBACService{db: db}

	// Create test data
	user := models.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	role := models.Role{
		Name:        "admin",
		Description: "Administrator role",
	}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}

	userRole := models.UserRole{
		UserID: user.ID,
		RoleID: role.ID,
	}
	if err := db.Create(&userRole).Error; err != nil {
		t.Fatalf("Failed to create user role: %v", err)
	}

	t.Run("successful role removal", func(t *testing.T) {
		err := service.RemoveRoleFromUser("test@example.com", "admin")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Verify assignment was removed
		var count int64
		db.Model(&models.UserRole{}).Where("user_id = ? AND role_id = ?", user.ID, role.ID).Count(&count)
		if count != 0 {
			t.Errorf("Expected 0 user roles, got %d", count)
		}
	})

	t.Run("role not assigned to user", func(t *testing.T) {
		err := service.RemoveRoleFromUser("test@example.com", "admin")
		if err == nil {
			t.Error("Expected error for role not assigned to user")
		}
		if err != nil && !contains(err.Error(), "Role not assigned") {
			t.Errorf("Expected 'Role not assigned' error, got: %v", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		err := service.RemoveRoleFromUser("nonexistent@example.com", "admin")
		if err == nil {
			t.Error("Expected error for non-existent user")
		}
		if err != nil && !contains(err.Error(), "not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})

	t.Run("role not found", func(t *testing.T) {
		err := service.RemoveRoleFromUser("test@example.com", "nonexistent")
		if err == nil {
			t.Error("Expected error for non-existent role")
		}
		if err != nil && !contains(err.Error(), "not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})
}

func TestRBACService_ListRoles(t *testing.T) {
	db := setupTestDB(t)
	service := &RBACService{db: db}

	t.Run("empty roles list", func(t *testing.T) {
		roles, err := service.ListRoles()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if len(roles) != 0 {
			t.Errorf("Expected empty roles list, got %d roles", len(roles))
		}
	})

	t.Run("list existing roles", func(t *testing.T) {
		// Create test roles
		role1 := models.Role{Name: "admin", Description: "Administrator"}
		role2 := models.Role{Name: "viewer", Description: "Read-only access"}
		if err := db.Create(&role1).Error; err != nil {
			t.Fatalf("Failed to create role1: %v", err)
		}
		if err := db.Create(&role2).Error; err != nil {
			t.Fatalf("Failed to create role2: %v", err)
		}

		roles, err := service.ListRoles()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if len(roles) != 2 {
			t.Errorf("Expected 2 roles, got %d", len(roles))
		}

		roleNames := make([]string, len(roles))
		for i, role := range roles {
			roleNames[i] = role.Name
		}

		foundAdmin := false
		foundViewer := false
		for _, name := range roleNames {
			if name == "admin" {
				foundAdmin = true
			}
			if name == "viewer" {
				foundViewer = true
			}
		}
		if !foundAdmin {
			t.Error("Expected to find 'admin' role")
		}
		if !foundViewer {
			t.Error("Expected to find 'viewer' role")
		}
	})
}

func TestRBACService_ListUserRoles(t *testing.T) {
	db := setupTestDB(t)
	service := &RBACService{db: db}

	// Create test data
	user := models.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	role1 := models.Role{Name: "admin", Description: "Administrator"}
	role2 := models.Role{Name: "viewer", Description: "Read-only access"}
	if err := db.Create(&role1).Error; err != nil {
		t.Fatalf("Failed to create role1: %v", err)
	}
	if err := db.Create(&role2).Error; err != nil {
		t.Fatalf("Failed to create role2: %v", err)
	}

	t.Run("user with no roles", func(t *testing.T) {
		roles, err := service.ListUserRoles("test@example.com")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if len(roles) != 0 {
			t.Errorf("Expected empty roles list, got %d roles", len(roles))
		}
	})

	t.Run("user with multiple roles", func(t *testing.T) {
		// Assign roles to user
		userRole1 := models.UserRole{UserID: user.ID, RoleID: role1.ID}
		userRole2 := models.UserRole{UserID: user.ID, RoleID: role2.ID}
		if err := db.Create(&userRole1).Error; err != nil {
			t.Fatalf("Failed to create userRole1: %v", err)
		}
		if err := db.Create(&userRole2).Error; err != nil {
			t.Fatalf("Failed to create userRole2: %v", err)
		}

		roles, err := service.ListUserRoles("test@example.com")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if len(roles) != 2 {
			t.Errorf("Expected 2 roles, got %d", len(roles))
		}

		roleNames := make([]string, len(roles))
		for i, role := range roles {
			roleNames[i] = role.Name
		}

		foundAdmin := false
		foundViewer := false
		for _, name := range roleNames {
			if name == "admin" {
				foundAdmin = true
			}
			if name == "viewer" {
				foundViewer = true
			}
		}
		if !foundAdmin {
			t.Error("Expected to find 'admin' role")
		}
		if !foundViewer {
			t.Error("Expected to find 'viewer' role")
		}
	})

	t.Run("user not found", func(t *testing.T) {
		_, err := service.ListUserRoles("nonexistent@example.com")
		if err == nil {
			t.Error("Expected error for non-existent user")
		}
		if err != nil && !contains(err.Error(), "not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})
}
