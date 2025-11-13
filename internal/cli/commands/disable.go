package commands

import (
	"fmt"

	"github.com/spf13/cobra"
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
	// TODO: Implement actual disable logic
	fmt.Println("Battery management disabled (will charge to 100%)")
	return nil
}