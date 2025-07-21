package rbac

import (
	"fmt"
	"time"

	"github.com/secretlyhq/secretly/internal/services"
	"github.com/spf13/cobra"
)

var auditLogsCmd = &cobra.Command{
	Use:   "audit-logs",
	Short: "View RBAC audit logs",
	Long:  "View audit logs for RBAC operations (role assignments, removals, etc.)",
	RunE:  runAuditLogs,
}

var (
	auditLimit  int
	auditOffset int
)

func init() {
	auditLogsCmd.Flags().IntVar(&auditLimit, "limit", 50, "Maximum number of logs to retrieve")
	auditLogsCmd.Flags().IntVar(&auditOffset, "offset", 0, "Number of logs to skip")
}

func runAuditLogs(cmd *cobra.Command, args []string) error {
	rbacService, err := services.NewRBACService()
	if err != nil {
		return fmt.Errorf("failed to initialize RBAC service: %w", err)
	}

	logs, err := rbacService.GetRBACAuditLogs(auditLimit, auditOffset)
	if err != nil {
		return fmt.Errorf("failed to get audit logs: %w", err)
	}

	if len(logs) == 0 {
		fmt.Println("No audit logs found")
		return nil
	}

	fmt.Printf("RBAC Audit Logs (showing %d entries):\n\n", len(logs))

	for _, log := range logs {
		fmt.Printf("🔍 %s | %s | %s\n",
			log.CreatedAt.Format(time.RFC3339),
			log.Action,
			log.Details)

		if log.ActorUserID != nil {
			fmt.Printf("   Actor User ID: %d\n", *log.ActorUserID)
		}
		if log.TargetUserID != nil {
			fmt.Printf("   Target User ID: %d\n", *log.TargetUserID)
		}
		if log.RoleID != nil {
			fmt.Printf("   Role ID: %d\n", *log.RoleID)
		}
		if log.NamespaceID != nil {
			fmt.Printf("   Namespace ID: %d\n", *log.NamespaceID)
		}
		fmt.Println()
	}

	return nil
}
