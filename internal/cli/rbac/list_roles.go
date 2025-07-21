package rbac

import (
	"fmt"

	"github.com/secretlyhq/secretly/internal/services"
	"github.com/spf13/cobra"
)

var listRolesCmd = &cobra.Command{
	Use:   "list-roles",
	Short: "List all available roles",
	Long:  "List all roles in the system",
	RunE:  runListRoles,
}

func runListRoles(cmd *cobra.Command, args []string) error {
	rbacService, err := services.NewRBACService()
	if err != nil {
		return fmt.Errorf("failed to initialize RBAC service: %w", err)
	}

	roles, err := rbacService.ListRoles()
	if err != nil {
		return fmt.Errorf("failed to list roles: %w", err)
	}

	if len(roles) == 0 {
		fmt.Println("No roles found")
		return nil
	}

	fmt.Println("Available roles:")
	for _, role := range roles {
		fmt.Printf("  - %s: %s\n", role.Name, role.Description)
	}

	return nil
}
