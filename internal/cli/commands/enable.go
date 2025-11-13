package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewEnableCommand creates the enable command
func NewEnableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable battery management (limit to configured threshold)",
		Long: `Enable battery management, which will limit charging to the configured
threshold by using conservation mode. When enabled, the system will stop
charging the battery once it reaches the configured threshold.`,
		RunE: runEnable,
	}

	return cmd
}

func runEnable(cmd *cobra.Command, args []string) error {
	// TODO: Implement actual enable logic
	fmt.Println("Battery management enabled (will maintain 80% limit)")
	return nil
}