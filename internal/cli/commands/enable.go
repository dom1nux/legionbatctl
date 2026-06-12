package commands

import (
	"fmt"

	"github.com/dom1nux/legionbatctl/internal/client"
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
	// Create client with default socket path
	c := client.NewClient("")

	// Create command executor
	executor := client.NewCommandExecutor(c)

	// Execute enable command
	result := executor.ExecuteEnable()

	// Format and output result
	output := client.FormatEnableResult(result)
	fmt.Print(output)

	if !result.Success {
		return fmt.Errorf("enable command failed: %s", result.Error)
	}

	return nil
}
