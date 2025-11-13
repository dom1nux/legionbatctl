package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dom1nux/legionbatctl/internal/daemon"
	"github.com/dom1nux/legionbatctl/internal/cli"
)

func main() {
	// Fast path for daemon mode - check if first argument is "daemon"
	if len(os.Args) > 1 && os.Args[1] == "daemon" {
		// Use default paths for systemd service
		socketPath := "/var/run/legionbatctl.sock"
		statePath := "/etc/legionbatctl.state"

		if err := daemon.RunDaemon(socketPath, statePath); err != nil {
			fmt.Fprintf(os.Stderr, "Daemon failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Full CLI initialization for interactive use
	if err := cli.Run(); err != nil {
		// Handle help/version flags specially
		if strings.Contains(err.Error(), "help requested") ||
		   strings.Contains(err.Error(), "version") {
			return // These are not errors, just exit cleanly
		}

		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}