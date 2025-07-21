package rbac

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/secretlyhq/secretly/internal/storage/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	testDBFile = "test_rbac.db"
	mainDBFile = "secretly.db"
)

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func setupTestDatabase(t *testing.T) *gorm.DB {
	// Create temporary database file
	tmpDB := testDBFile
	t.Cleanup(func() {
		_ = os.Remove(tmpDB)
	})

	db, err := gorm.Open(sqlite.Open(tmpDB), &gorm.Config{})
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

	// Create test data
	user := models.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	role1 := models.Role{Name: "admin", Description: "Administrator role"}
	role2 := models.Role{Name: "viewer", Description: "Read-only access role"}
	if err := db.Create(&role1).Error; err != nil {
		t.Fatalf("Failed to create role1: %v", err)
	}
	if err := db.Create(&role2).Error; err != nil {
		t.Fatalf("Failed to create role2: %v", err)
	}

	return db
}

func TestAssignRoleCommand(t *testing.T) {
	// Setup test database
	db := setupTestDatabase(t)
	defer db.Exec("DROP TABLE IF EXISTS users, roles, user_roles")

	// Override the database path for testing
	originalDB := mainDBFile
	testDB := testDBFile

	// Create a backup of the original database if it exists
	if _, err := os.Stat(originalDB); err == nil {
		data, err := os.ReadFile(originalDB)
		if err == nil {
			defer func() { _ = os.WriteFile(originalDB, data, 0600) }()
		}
	}

	// Copy test database to the expected location
	testData, err := os.ReadFile(testDB)
	if err != nil {
		t.Fatalf("Failed to read test database: %v", err)
	}
	if err := os.WriteFile(originalDB, testData, 0600); err != nil {
		t.Fatalf("Failed to write original database: %v", err)
	}

	defer func() { _ = os.Remove(originalDB) }()

	t.Run("successful role assignment", func(t *testing.T) {
		cmd := assignRoleCmd
		cmd.SetArgs([]string{"--user", "test@example.com", "--role", "admin"})

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		err := cmd.Execute()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !contains(buf.String(), "Successfully assigned role 'admin' to user 'test@example.com'") {
			t.Errorf("Expected success message in output: %s", buf.String())
		}
	})

	t.Run("user not found", func(t *testing.T) {
		cmd := assignRoleCmd
		cmd.SetArgs([]string{"--user", "nonexistent@example.com", "--role", "admin"})

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		err := cmd.Execute()
		if err == nil {
			t.Error("Expected error for non-existent user")
		}
		if err != nil && !contains(err.Error(), "not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})
}

func TestListRolesCommand(t *testing.T) {
	// Setup test database
	db := setupTestDatabase(t)
	defer db.Exec("DROP TABLE IF EXISTS users, roles, user_roles")

	// Override the database path for testing
	originalDB := mainDBFile
	testDB := testDBFile

	// Copy test database to the expected location
	testData, err := os.ReadFile(testDB)
	if err != nil {
		t.Fatalf("Failed to read test database: %v", err)
	}
	if err := os.WriteFile(originalDB, testData, 0600); err != nil {
		t.Fatalf("Failed to write original database: %v", err)
	}

	defer func() { _ = os.Remove(originalDB) }()

	t.Run("list available roles", func(t *testing.T) {
		cmd := listRolesCmd

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		err := cmd.Execute()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		output := buf.String()
		if !contains(output, "Available roles:") {
			t.Errorf("Expected 'Available roles:' in output: %s", output)
		}
		if !contains(output, "admin") {
			t.Errorf("Expected 'admin' in output: %s", output)
		}
		if !contains(output, "viewer") {
			t.Errorf("Expected 'viewer' in output: %s", output)
		}
	})
}

func TestListUserRolesCommand(t *testing.T) {
	// Setup test database
	db := setupTestDatabase(t)
	defer db.Exec("DROP TABLE IF EXISTS users, roles, user_roles")

	// Assign a role to the test user
	var user models.User
	var role models.Role
	if err := db.Where("email = ?", "test@example.com").First(&user).Error; err != nil {
		t.Fatalf("Failed to find user: %v", err)
	}
	if err := db.Where("name = ?", "admin").First(&role).Error; err != nil {
		t.Fatalf("Failed to find role: %v", err)
	}

	userRole := models.UserRole{UserID: user.ID, RoleID: role.ID}
	if err := db.Create(&userRole).Error; err != nil {
		t.Fatalf("Failed to create user role: %v", err)
	}

	// Override the database path for testing
	originalDB := mainDBFile
	testDB := testDBFile

	// Copy test database to the expected location
	testData, err := os.ReadFile(testDB)
	if err != nil {
		t.Fatalf("Failed to read test database: %v", err)
	}
	if err := os.WriteFile(originalDB, testData, 0600); err != nil {
		t.Fatalf("Failed to write original database: %v", err)
	}

	defer func() { _ = os.Remove(originalDB) }()

	t.Run("list user roles", func(t *testing.T) {
		cmd := listUserRolesCmd
		cmd.SetArgs([]string{"--user", "test@example.com"})

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		err := cmd.Execute()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		output := buf.String()
		if !contains(output, "Roles assigned to user 'test@example.com':") {
			t.Errorf("Expected user roles message in output: %s", output)
		}
		if !contains(output, "admin") {
			t.Errorf("Expected 'admin' in output: %s", output)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		cmd := listUserRolesCmd
		cmd.SetArgs([]string{"--user", "nonexistent@example.com"})

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		err := cmd.Execute()
		if err == nil {
			t.Error("Expected error for non-existent user")
		}
		if err != nil && !contains(err.Error(), "not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})
}

func TestRemoveRoleCommand(t *testing.T) {
	// Setup test database
	db := setupTestDatabase(t)
	defer db.Exec("DROP TABLE IF EXISTS users, roles, user_roles")

	// Assign a role to the test user first
	var user models.User
	var role models.Role
	if err := db.Where("email = ?", "test@example.com").First(&user).Error; err != nil {
		t.Fatalf("Failed to find user: %v", err)
	}
	if err := db.Where("name = ?", "admin").First(&role).Error; err != nil {
		t.Fatalf("Failed to find role: %v", err)
	}

	userRole := models.UserRole{UserID: user.ID, RoleID: role.ID}
	if err := db.Create(&userRole).Error; err != nil {
		t.Fatalf("Failed to create user role: %v", err)
	}

	// Override the database path for testing
	originalDB := mainDBFile
	testDB := testDBFile

	// Copy test database to the expected location
	testData, err := os.ReadFile(testDB)
	if err != nil {
		t.Fatalf("Failed to read test database: %v", err)
	}
	if err := os.WriteFile(originalDB, testData, 0600); err != nil {
		t.Fatalf("Failed to write original database: %v", err)
	}

	defer func() { _ = os.Remove(originalDB) }()

	t.Run("successful role removal", func(t *testing.T) {
		cmd := removeRoleCmd
		cmd.SetArgs([]string{"--user", "test@example.com", "--role", "admin"})

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		err := cmd.Execute()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !contains(buf.String(), "Successfully removed role 'admin' from user 'test@example.com'") {
			t.Errorf("Expected success message in output: %s", buf.String())
		}
	})

	t.Run("role not assigned", func(t *testing.T) {
		cmd := removeRoleCmd
		cmd.SetArgs([]string{"--user", "test@example.com", "--role", "viewer"})

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		err := cmd.Execute()
		if err == nil {
			t.Error("Expected error for role not assigned")
		}
		if err != nil && !contains(err.Error(), "does not have role") {
			t.Errorf("Expected 'does not have role' error, got: %v", err)
		}
	})
}
