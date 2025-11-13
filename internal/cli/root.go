package cli

import (
	"github.com/spf13/cobra"
	"github.com/dom1nux/legionbatctl/internal/cli/commands"
	"github.com/dom1nux/legionbatctl/pkg/version"
)

// Run initializes and runs the CLI application
func Run() error {
	rootCmd := &cobra.Command{
		Use:   "legionbatctl",
		Short: "Lenovo Legion Battery Control Utility",
		Long: `LegionBatCTL is a utility for controlling battery charging behavior on Lenovo Legion laptops.
It helps extend battery lifespan by managing conservation mode to maintain battery levels
within configured thresholds.

This is particularly useful for laptops with fixed conservation mode limits (e.g., 60%),
allowing you to effectively achieve higher charge limits (e.g., 80%).`,
		Version: version.GetVersionInfo().String(),
	}

	// Add global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().String("config", "/etc/legionbatctl.conf", "Path to configuration file")

	// Add subcommands
	rootCmd.AddCommand(commands.NewStatusCommand())
	rootCmd.AddCommand(commands.NewEnableCommand())
	rootCmd.AddCommand(commands.NewDisableCommand())
	rootCmd.AddCommand(commands.NewSetThresholdCommand())
	rootCmd.AddCommand(commands.NewAutoCommand()) // For manual testing

	// Set completion
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Customize help output
	rootCmd.SetUsageTemplate(usageTemplate())
	cobra.EnableCommandSorting = false

	return rootCmd.Execute()
}

// usageTemplate returns a custom usage template
func usageTemplate() string {
	return `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
}