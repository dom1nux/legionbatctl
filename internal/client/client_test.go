package client

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/dom1nux/legionbatctl/internal/daemon"
	"github.com/dom1nux/legionbatctl/internal/protocol"
)

func TestNewClient(t *testing.T) {
	socketPath := "/tmp/test.sock"
	client := NewClient(socketPath)

	if client.GetSocketPath() != socketPath {
		t.Errorf("Expected socket path %s, got %s", socketPath, client.GetSocketPath())
	}

	if client.GetTimeout() != DefaultTimeout {
		t.Errorf("Expected default timeout %v, got %v", DefaultTimeout, client.GetTimeout())
	}

	if client.String() != fmt.Sprintf("legionbatctl Client{socket: %s, timeout: %v}", socketPath, DefaultTimeout) {
		t.Errorf("Unexpected string representation: %s", client.String())
	}
}

func TestNewClientWithDefaults(t *testing.T) {
	client := NewClient("")

	if client.GetSocketPath() != DefaultSocketPath {
		t.Errorf("Expected default socket path %s, got %s", DefaultSocketPath, client.GetSocketPath())
	}
}

func TestNewClientWithTimeout(t *testing.T) {
	socketPath := "/tmp/test.sock"
	timeout := 5 * time.Second
	client := NewClientWithTimeout(socketPath, timeout)

	if client.GetSocketPath() != socketPath {
		t.Errorf("Expected socket path %s, got %s", socketPath, client.GetSocketPath())
	}

	if client.GetTimeout() != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, client.GetTimeout())
	}
}

func TestClientSetTimeout(t *testing.T) {
	client := NewClient("/tmp/test.sock")
	newTimeout := 15 * time.Second

	client.SetTimeout(newTimeout)
	if client.GetTimeout() != newTimeout {
		t.Errorf("Expected timeout %v, got %v", newTimeout, client.GetTimeout())
	}
}

func TestClientIsDaemonRunning(t *testing.T) {
	tempDir := t.TempDir()
	socketPath := filepath.Join(tempDir, "test.sock")
	statePath := filepath.Join(tempDir, "test_state.json")

	client := NewClient(socketPath)

	// Should not be running initially
	if client.IsDaemonRunning() {
		t.Error("Expected daemon to not be running")
	}

	// Start daemon
	daemonInstance := daemon.NewDaemon(socketPath, statePath)
	if err := daemonInstance.Start(); err != nil {
		t.Fatalf("Failed to start daemon: %v", err)
	}
	defer daemonInstance.Stop()

	// Give daemon a moment to start
	time.Sleep(100 * time.Millisecond)

	// Should be running now
	if !client.IsDaemonRunning() {
		t.Error("Expected daemon to be running")
	}
}

func TestClientPing(t *testing.T) {
	tempDir := t.TempDir()
	socketPath := filepath.Join(tempDir, "test.sock")
	statePath := filepath.Join(tempDir, "test_state.json")

	client := NewClient(socketPath)

	// Should fail when daemon is not running
	if err := client.Ping(); err == nil {
		t.Error("Expected ping to fail when daemon is not running")
	}

	// Start daemon
	daemonInstance := daemon.NewDaemon(socketPath, statePath)
	if err := daemonInstance.Start(); err != nil {
		t.Fatalf("Failed to start daemon: %v", err)
	}
	defer daemonInstance.Stop()

	// Give daemon a moment to start
	time.Sleep(100 * time.Millisecond)

	// Should succeed when daemon is running
	if err := client.Ping(); err != nil {
		t.Errorf("Expected ping to succeed when daemon is running: %v", err)
	}
}

func TestCommandExecutor(t *testing.T) {
	client := NewClient("/tmp/nonexistent.sock")
	executor := NewCommandExecutor(client)

	if executor == nil {
		t.Error("Expected command executor to be created")
	}
}

func TestCommandExecutorExecuteEnable(t *testing.T) {
	tempDir := t.TempDir()
	socketPath := filepath.Join(tempDir, "test.sock")
	statePath := filepath.Join(tempDir, "test_state.json")

	client := NewClient(socketPath)
	executor := NewCommandExecutor(client)

	// Should fail when daemon is not running
	result := executor.ExecuteEnable()
	if result.Success {
		t.Error("Expected enable command to fail when daemon is not running")
	}

	if result.Error == "" {
		t.Error("Expected error message when daemon is not running")
	}

	// Start daemon
	daemonInstance := daemon.NewDaemon(socketPath, statePath)
	if err := daemonInstance.Start(); err != nil {
		t.Fatalf("Failed to start daemon: %v", err)
	}
	defer daemonInstance.Stop()

	// Give daemon a moment to start
	time.Sleep(100 * time.Millisecond)

	// Should succeed when daemon is running
	result = executor.ExecuteEnable()
	if !result.Success {
		t.Errorf("Expected enable command to succeed when daemon is running: %s", result.Error)
	}
}

func TestCommandExecutorExecuteSetThreshold(t *testing.T) {
	tempDir := t.TempDir()
	socketPath := filepath.Join(tempDir, "test.sock")
	statePath := filepath.Join(tempDir, "test_state.json")

	client := NewClient(socketPath)
	executor := NewCommandExecutor(client)

	// Start daemon
	daemonInstance := daemon.NewDaemon(socketPath, statePath)
	if err := daemonInstance.Start(); err != nil {
		t.Fatalf("Failed to start daemon: %v", err)
	}
	defer daemonInstance.Stop()

	// Give daemon a moment to start
	time.Sleep(100 * time.Millisecond)

	// Test valid threshold
	result := executor.ExecuteSetThreshold(80)
	if !result.Success {
		t.Errorf("Expected set_threshold command to succeed with valid threshold: %s", result.Error)
	}

	// Test invalid threshold
	result = executor.ExecuteSetThreshold(50)
	if result.Success {
		t.Error("Expected set_threshold command to fail with invalid threshold")
	}
}

func TestValidateThreshold(t *testing.T) {
	tests := []struct {
		threshold int
		expectErr bool
	}{
		{60, false},  // Minimum valid
		{80, false},  // Valid
		{100, false}, // Maximum valid
		{59, true},   // Below minimum
		{101, true},  // Above maximum
		{0, true},    // Invalid
		{-10, true},  // Invalid
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("threshold_%d", tt.threshold), func(t *testing.T) {
			err := protocol.ValidateThreshold(tt.threshold)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateThreshold(%d) error = %v, expectErr %v", tt.threshold, err, tt.expectErr)
			}
		})
	}
}

func TestGetThresholdRange(t *testing.T) {
	min, max, description := GetThresholdRange()

	if min != 60 {
		t.Errorf("Expected minimum threshold 60, got %d", min)
	}

	if max != 100 {
		t.Errorf("Expected maximum threshold 100, got %d", max)
	}

	if description == "" {
		t.Error("Expected non-empty description")
	}

	if !contains(description, "60-100%") {
		t.Errorf("Expected description to contain '60-100%%', got: %s", description)
	}
}

func TestFormatStatus(t *testing.T) {
	status := &protocol.StatusData{
		ConservationEnabled: true,
		Threshold:           80,
		CurrentMode:         "enabled",
		BatteryLevel:        75,
		ConservationMode:    false,
		Charging:            true,
		LastAction:          "enable",
		DaemonUptime:        "1h25m30s",
		HardwareSupported:   true,
	}

	formatted := FormatStatus(status)

	if !contains(formatted, "Conservation Management: enabled") {
		t.Error("Expected conservation enabled status in formatted output")
	}

	if !contains(formatted, "Charge Threshold: 80%") {
		t.Error("Expected threshold in formatted output")
	}

	if !contains(formatted, "Battery Level: 75%") {
		t.Error("Expected battery level in formatted output")
	}
}

func TestFormatEnableResult(t *testing.T) {
	// Test success result
	successResult := &CommandResult{
		Success: true,
		Message: "Battery management enabled",
	}

	formatted := FormatEnableResult(successResult)
	if !contains(formatted, "✓") {
		t.Error("Expected success indicator in formatted output")
	}

	// Test failure result
	failureResult := &CommandResult{
		Success: false,
		Error:   "Connection failed",
	}

	formatted = FormatEnableResult(failureResult)
	if !contains(formatted, "✗") {
		t.Error("Expected failure indicator in formatted output")
	}
}

func TestCheckDaemonConnection(t *testing.T) {
	tempDir := t.TempDir()
	socketPath := filepath.Join(tempDir, "test.sock")
	client := NewClient(socketPath)

	// Should fail when daemon is not running
	err := CheckDaemonConnection(client)
	if err == nil {
		t.Error("Expected error when daemon is not running")
	}

	if !contains(err.Error(), "daemon is not running") {
		t.Errorf("Expected 'daemon is not running' error, got: %s", err.Error())
	}

	// Start daemon
	daemonInstance := daemon.NewDaemon(socketPath, "")
	if err := daemonInstance.Start(); err != nil {
		t.Fatalf("Failed to start daemon: %v", err)
	}
	defer daemonInstance.Stop()

	// Give daemon a moment to start
	time.Sleep(100 * time.Millisecond)

	// Should succeed when daemon is running
	err = CheckDaemonConnection(client)
	if err != nil {
		t.Errorf("Expected no error when daemon is running: %v", err)
	}
}

func TestRetryOperation(t *testing.T) {
	attempts := 0
	maxAttempts := 3

	operation := func() error {
		attempts++
		if attempts < maxAttempts {
			return fmt.Errorf("attempt %d failed", attempts)
		}
		return nil
	}

	// Should succeed after retries
	err := RetryOperation(operation, maxAttempts, 10*time.Millisecond)
	if err != nil {
		t.Errorf("Expected operation to succeed after retries: %v", err)
	}

	// Verify all attempts were made
	if attempts != maxAttempts {
		t.Errorf("Expected %d attempts, got %d", maxAttempts, attempts)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr ||
		(len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
