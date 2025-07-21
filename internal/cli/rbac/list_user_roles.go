package rbac

import (
	"fmt"

	"github.com/secretlyhq/secretly/internal/services"
	"github.com/spf13/cobra"
)

var listUserRolesCmd = &cobra.Command{
	Use:   "list-user-roles",
	Short: "List roles assigned to a user",
	Long:  "List all roles assigned to a specific user by email address",
	RunE:  runListUserRoles,
}

var listUserEmail string

func init() {
	listUserRolesCmd.Flags().StringVar(&listUserEmail, "user", "", "User email address (required)")
	_ = listUserRolesCmd.MarkFlagRequired("user")
}

func runListUserRoles(cmd *cobra.Command, args []string) error {
	rbacService, err := services.NewRBACService()
	if err != nil {
		return fmt.Errorf("failed to initialize RBAC service: %w", err)
	}

	roles, err := rbacService.ListUserRoles(listUserEmail)
	if err != nil {
		return fmt.Errorf("failed to list user roles: %w", err)
	}

	if len(roles) == 0 {
		fmt.Printf("No roles assigned to user '%s'\n", listUserEmail)
		return nil
	}

	fmt.Printf("Roles assigned to user '%s':\n", listUserEmail)
	for _, role := range roles {
		fmt.Printf("  - %s: %s\n", role.Name, role.Description)
	}

	return nil
}
