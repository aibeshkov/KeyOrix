package rbac

import (
	"fmt"

	"github.com/secretlyhq/secretly/internal/services"
	"github.com/spf13/cobra"
)

var removeRoleCmd = &cobra.Command{
	Use:   "remove-role",
	Short: "Remove a role from a user",
	Long:  "Remove a role assignment from a user by email address",
	RunE:  runRemoveRole,
}

var (
	removeUserEmail string
	removeRoleName  string
)

func init() {
	removeRoleCmd.Flags().StringVar(&removeUserEmail, "user", "", "User email address (required)")
	removeRoleCmd.Flags().StringVar(&removeRoleName, "role", "", "Role name to remove (required)")

	_ = removeRoleCmd.MarkFlagRequired("user")
	_ = removeRoleCmd.MarkFlagRequired("role")
}

func runRemoveRole(cmd *cobra.Command, args []string) error {
	rbacService, err := services.NewRBACService()
	if err != nil {
		return fmt.Errorf("failed to initialize RBAC service: %w", err)
	}

	err = rbacService.RemoveRoleFromUser(removeUserEmail, removeRoleName)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	fmt.Printf("✅ Successfully removed role '%s' from user '%s'\n", removeRoleName, removeUserEmail)
	return nil
}
