package rbac

import (
	"fmt"

	"github.com/secretlyhq/secretly/internal/services"
	"github.com/spf13/cobra"
)

var listPermissionsCmd = &cobra.Command{
	Use:   "list-permissions",
	Short: "List all permissions for a user",
	Long:  "List all permissions assigned to a user through their roles",
	RunE:  runListPermissions,
}

var listPermissionsUserEmail string

func init() {
	listPermissionsCmd.Flags().StringVar(&listPermissionsUserEmail, "user", "", "User email address (required)")
	_ = listPermissionsCmd.MarkFlagRequired("user")
}

func runListPermissions(cmd *cobra.Command, args []string) error {
	rbacService, err := services.NewRBACService()
	if err != nil {
		return fmt.Errorf("failed to initialize RBAC service: %w", err)
	}

	permissions, err := rbacService.ListUserPermissions(listPermissionsUserEmail)
	if err != nil {
		return fmt.Errorf("failed to list user permissions: %w", err)
	}

	if len(permissions) == 0 {
		fmt.Printf("No permissions found for user '%s'\n", listPermissionsUserEmail)
		return nil
	}

	fmt.Printf("Permissions for user '%s':\n", listPermissionsUserEmail)

	// Group permissions by resource
	resourceMap := make(map[string][]services.Permission)
	for _, perm := range permissions {
		resourceMap[perm.Resource] = append(resourceMap[perm.Resource], perm)
	}

	for resource, perms := range resourceMap {
		fmt.Printf("\n📁 %s:\n", resource)
		for _, perm := range perms {
			fmt.Printf("  - %s (%s): %s\n", perm.Name, perm.Action, perm.Description)
		}
	}

	return nil
}
