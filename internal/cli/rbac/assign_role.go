package rbac

import (
	"fmt"

	"github.com/secretlyhq/secretly/internal/services"
	"github.com/spf13/cobra"
)

var assignRoleCmd = &cobra.Command{
	Use:   "assign-role",
	Short: "Assign a role to a user",
	Long:  "Assign a role to a user by email address",
	RunE:  runAssignRole,
}

var (
	userEmail string
	roleName  string
)

func init() {
	assignRoleCmd.Flags().StringVar(&userEmail, "user", "", "User email address (required)")
	assignRoleCmd.Flags().StringVar(&roleName, "role", "", "Role name to assign (required)")

	_ = assignRoleCmd.MarkFlagRequired("user")
	_ = assignRoleCmd.MarkFlagRequired("role")
}

func runAssignRole(cmd *cobra.Command, args []string) error {
	rbacService, err := services.NewRBACService()
	if err != nil {
		return fmt.Errorf("failed to initialize RBAC service: %w", err)
	}

	err = rbacService.AssignRoleToUser(userEmail, roleName)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	fmt.Printf("✅ Successfully assigned role '%s' to user '%s'\n", roleName, userEmail)
	return nil
}
