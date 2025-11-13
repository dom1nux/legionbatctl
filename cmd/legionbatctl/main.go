package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dom1nux/legionbatctl/internal/auto"
	"github.com/dom1nux/legionbatctl/internal/cli"
)

func main() {
	// Fast path for auto mode - check if first argument is "auto"
	if len(os.Args) > 1 && os.Args[1] == "auto" {
		// For auto mode, we want minimal overhead and simple error handling
		if err := auto.Run(); err != nil {
			// Log to stderr for systemd compatibility
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
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