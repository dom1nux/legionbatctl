package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/dom1nux/legionbatctl/internal/client"
)

// NewSetThresholdCommand creates the set-threshold command
func NewSetThresholdCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-threshold <percentage>",
		Short: "Set battery charge threshold (60-100)",
		Long: `Set the maximum battery charge threshold. When battery management is enabled,
the system will stop charging once the battery reaches this percentage by
enabling conservation mode.

NOTE: Due to hardware limitations on Lenovo Legion Slim 7 (2021), the threshold
must be between 60-100%. The native conservation mode is fixed at 60%, but this
utility allows you to effectively achieve higher charge limits.

For optimal battery health, thresholds between 75-85% are recommended.`,
		Args: cobra.ExactArgs(1),
		RunE: runSetThreshold,
	}

	return cmd
}

func runSetThreshold(cmd *cobra.Command, args []string) error {
	threshold, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid threshold value: %s", args[0])
	}

	// Create client with default socket path
	c := client.NewClient("")

	// Create command executor
	executor := client.NewCommandExecutor(c)

	// Execute set threshold command
	result := executor.ExecuteSetThreshold(threshold)

	// Format and output result
	output := client.FormatSetThresholdResult(result)
	fmt.Print(output)

	if !result.Success {
		return fmt.Errorf(result.Error)
	}

	return nil
}