package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/dom1nux/legionbatctl/internal/client"
)

// NewStatusCommand creates the status command
func NewStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current battery management and conservation mode status",
		Long: `Display the current status of battery management, conservation mode, and
charge threshold settings. This shows both the hardware conservation mode
status and the software battery management configuration.`,
		RunE: runStatus,
	}

	return cmd
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Create client with default socket path
	c := client.NewClient("")

	// Create command executor
	executor := client.NewCommandExecutor(c)

	// Execute status command
	result := executor.ExecuteStatus()

	// Format and output result
	output := client.FormatStatusResult(result)
	fmt.Print(output)

	if !result.Success {
		return fmt.Errorf(result.Error)
	}

	return nil
}