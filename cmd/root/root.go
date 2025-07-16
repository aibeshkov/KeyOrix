package root

import (
	"github.com/spf13/cobra"
)

var version = "dev" // value will be overwritten via ldflags

var RootCmd = &cobra.Command{
	Use:     "secretly",
	Short:   "Secretly - Secure secrets management CLI",
	Version: version, // 💡 automatically adds --version flag
}
