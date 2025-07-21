package rbac

import (
	"fmt"

	"github.com/secretlyhq/secretly/internal/services"
	"github.com/spf13/cobra"
)

var checkPermissionCmd = &cobra.Command{
	Use:   "check-permission",
	Short: "Check if a user has a specific permission",
	Long:  "Check if a user has a specific permission by email address and permission name",
	RunE:  runCheckPermission,
}

var (
	checkUserEmail      string
	checkPermissionName string
)

func init() {
	checkPermissionCmd.Flags().StringVar(&checkUserEmail, "user", "", "User email address (required)")
	checkPermissionCmd.Flags().StringVar(&checkPermissionName, "permission", "", "Permission name to check (required)")

	_ = checkPermissionCmd.MarkFlagRequired("user")
	_ = checkPermissionCmd.MarkFlagRequired("permission")
}

func runCheckPermission(cmd *cobra.Command, args []string) error {
	rbacService, err := services.NewRBACService()
	if err != nil {
		return fmt.Errorf("failed to initialize RBAC service: %w", err)
	}

	hasPermission, err := rbacService.HasPermission(checkUserEmail, checkPermissionName)
	if err != nil {
		return fmt.Errorf("failed to check permission: %w", err)
	}

	if hasPermission {
		fmt.Printf("✅ User '%s' has permission '%s'\n", checkUserEmail, checkPermissionName)
	} else {
		fmt.Printf("❌ User '%s' does NOT have permission '%s'\n", checkUserEmail, checkPermissionName)
	}

	return nil
}
