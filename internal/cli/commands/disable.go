package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/dom1nux/legionbatctl/internal/client"
)

// NewDisableCommand creates the disable command
func NewDisableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable battery management (allow charging to 100%)",
		Long: `Disable battery management, allowing the battery to charge to 100%.
This disables the automatic threshold management and allows normal
charging behavior.`,
		RunE: runDisable,
	}

	return cmd
}

func runDisable(cmd *cobra.Command, args []string) error {
	// Create client with default socket path
	c := client.NewClient("")

	// Create command executor
	executor := client.NewCommandExecutor(c)

	// Execute disable command
	result := executor.ExecuteDisable()

	// Format and output result
	output := client.FormatDisableResult(result)
	fmt.Print(output)

	if !result.Success {
		return fmt.Errorf(result.Error)
	}

	return nil
}