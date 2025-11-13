package client

import (
	"fmt"
	"time"

	"github.com/dom1nux/legionbatctl/internal/protocol"
)

// CommandResult represents the result of a command execution
type CommandResult struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
}

// CommandExecutor provides high-level command execution with result formatting
type CommandExecutor struct {
	client *Client
}

// NewCommandExecutor creates a new command executor
func NewCommandExecutor(client *Client) *CommandExecutor {
	return &CommandExecutor{
		client: client,
	}
}

// Helper functions for creating command results
func newSuccessResult(message string, duration time.Duration) *CommandResult {
	return &CommandResult{
		Success:  true,
		Message:  message,
		Duration: duration,
	}
}

func newSuccessResultWithData(message string, data interface{}, duration time.Duration) *CommandResult {
	return &CommandResult{
		Success:  true,
		Message:  message,
		Data:     data,
		Duration: duration,
	}
}

func newFailureResult(message string, err error, duration time.Duration) *CommandResult {
	return &CommandResult{
		Success:  false,
		Message:  message,
		Error:    err.Error(),
		Duration: duration,
	}
}

// ExecuteEnable executes the enable command
func (e *CommandExecutor) ExecuteEnable() *CommandResult {
	start := time.Now()
	err := e.client.Enable()
	duration := time.Since(start)

	if err != nil {
		return newFailureResult("Failed to enable battery management", err, duration)
	}

	return newSuccessResult("Battery management enabled successfully", duration)
}

// ExecuteDisable executes the disable command
func (e *CommandExecutor) ExecuteDisable() *CommandResult {
	start := time.Now()
	err := e.client.Disable()
	duration := time.Since(start)

	if err != nil {
		return newFailureResult("Failed to disable battery management", err, duration)
	}

	return newSuccessResult("Battery management disabled successfully", duration)
}

// ExecuteSetThreshold executes the set_threshold command
func (e *CommandExecutor) ExecuteSetThreshold(threshold int) *CommandResult {
	start := time.Now()
	err := e.client.SetThreshold(threshold)
	duration := time.Since(start)

	if err != nil {
		return newFailureResult(fmt.Sprintf("Failed to set threshold to %d", threshold), err, duration)
	}

	return newSuccessResultWithData(
		fmt.Sprintf("Charge threshold set to %d%%", threshold),
		map[string]interface{}{"threshold": threshold},
		duration,
	)
}

// ExecuteStatus executes the status command
func (e *CommandExecutor) ExecuteStatus() *CommandResult {
	start := time.Now()
	status, err := e.client.GetStatus()
	duration := time.Since(start)

	if err != nil {
		return newFailureResult("Failed to get system status", err, duration)
	}

	return newSuccessResultWithData("System status retrieved successfully", status, duration)
}

// ExecuteDaemonStatus executes the daemon_status command
func (e *CommandExecutor) ExecuteDaemonStatus() *CommandResult {
	start := time.Now()
	status, err := e.client.GetDaemonStatus()
	duration := time.Since(start)

	if err != nil {
		return newFailureResult("Failed to get daemon status", err, duration)
	}

	return newSuccessResultWithData("Daemon status retrieved successfully", status, duration)
}

// FormatStatus formats status data for human-readable output
func FormatStatus(status *protocol.StatusData) string {
	output := "Battery Management Status:\n"
	output += fmt.Sprintf("  Conservation Management: %s\n", formatBool(status.ConservationEnabled))
	output += fmt.Sprintf("  Charge Threshold: %d%%\n", status.Threshold)
	output += fmt.Sprintf("  Current Mode: %s\n", status.CurrentMode)
	output += fmt.Sprintf("  Battery Level: %d%%\n", status.BatteryLevel)
	output += fmt.Sprintf("  Conservation Mode: %s\n", formatBool(status.ConservationMode))
	output += fmt.Sprintf("  Charging Status: %s\n", formatCharging(status.Charging))
	output += fmt.Sprintf("  Last Action: %s\n", status.LastAction)
	output += fmt.Sprintf("  Daemon Uptime: %s\n", status.DaemonUptime)
	output += fmt.Sprintf("  Hardware Supported: %s\n", formatBool(status.HardwareSupported))

	return output
}

// FormatDaemonStatus formats daemon status data for human-readable output
func FormatDaemonStatus(status *protocol.DaemonStatusData) string {
	output := "Daemon Status:\n"
	output += fmt.Sprintf("  Running: %s\n", formatBool(status.Running))
	output += fmt.Sprintf("  PID: %d\n", status.PID)
	output += fmt.Sprintf("  Uptime: %s\n", status.Uptime)
	output += fmt.Sprintf("  Version: %s\n", status.Version)
	output += fmt.Sprintf("  Socket Path: %s\n", status.SocketPath)
	output += fmt.Sprintf("  State File: %s\n", status.StateFile)

	return output
}

// FormatEnableResult formats the result of an enable command
func FormatEnableResult(result *CommandResult) string {
	if result.Success {
		return "✓ Battery management enabled. Conservation mode will be activated when battery reaches the threshold."
	} else {
		return fmt.Sprintf("✗ Failed to enable battery management: %s", result.Error)
	}
}

// FormatDisableResult formats the result of a disable command
func FormatDisableResult(result *CommandResult) string {
	if result.Success {
		return "✓ Battery management disabled. The battery will charge to 100%."
	} else {
		return fmt.Sprintf("✗ Failed to disable battery management: %s", result.Error)
	}
}

// FormatStatusResult formats the result of a status command
func FormatStatusResult(result *CommandResult) string {
	if result.Success {
		if status, ok := result.Data.(*protocol.StatusData); ok {
			return FormatStatus(status)
		}
		return result.Message
	} else {
		return fmt.Sprintf("✗ Failed to get status: %s", result.Error)
	}
}

// FormatSetThresholdResult formats the result of a set_threshold command
func FormatSetThresholdResult(result *CommandResult) string {
	if result.Success {
		if data, ok := result.Data.(map[string]interface{}); ok {
			if threshold, ok := data["threshold"].(int); ok {
				return fmt.Sprintf("✓ Charge threshold set to %d%%. Conservation mode will activate at this level.", threshold)
			}
		}
		return "✓ Charge threshold updated successfully."
	} else {
		return fmt.Sprintf("✗ Failed to set threshold: %s", result.Error)
	}
}

// formatBool formats a boolean value for display
func formatBool(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}

// formatCharging formats charging status for display
func formatCharging(charging bool) string {
	if charging {
		return "charging"
	}
	return "discharging"
}


// GetThresholdRange returns information about valid threshold range
func GetThresholdRange() (min, max int, description string) {
	return 60, 100, "Threshold must be between 60-100% due to hardware conservation mode limitation on Lenovo Legion Slim 7 (2021)"
}

// CheckDaemonConnection checks if the daemon is available and provides user-friendly error messages
func CheckDaemonConnection(client *Client) error {
	if !client.IsDaemonRunning() {
		return fmt.Errorf("daemon is not running. Start it with: sudo legionbatctl daemon")
	}

	if err := client.Ping(); err != nil {
		return fmt.Errorf("daemon is not responding. Check daemon status with: sudo systemctl status legionbatctl")
	}

	return nil
}

// RetryOperation executes an operation with retry logic
func RetryOperation(operation func() error, maxRetries int, delay time.Duration) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(delay)
		}

		if err := operation(); err != nil {
			lastErr = err
			continue
		}

		return nil
	}

	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}