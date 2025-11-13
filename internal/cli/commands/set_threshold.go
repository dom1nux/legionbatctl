package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// NewSetThresholdCommand creates the set-threshold command
func NewSetThresholdCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-threshold <percentage>",
		Short: "Set battery charge threshold (20-100)",
		Long: `Set the maximum battery charge threshold. When battery management is enabled,
the system will stop charging once the battery reaches this percentage by
enabling conservation mode.

For optimal battery health, thresholds between 70-85% are recommended.
Your hardware's native conservation mode limit is 60%, but this utility
allows you to effectively achieve higher limits.`,
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

	if threshold < 20 || threshold > 100 {
		return fmt.Errorf("threshold must be between 20 and 100")
	}

	// TODO: Implement actual threshold setting logic
	fmt.Printf("Charge threshold set to %d%%\n", threshold)
	return nil
}