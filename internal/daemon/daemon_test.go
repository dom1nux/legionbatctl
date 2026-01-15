package daemon

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewDaemon(t *testing.T) {
	daemon := NewDaemon("/tmp/test.sock", "/tmp/test_state.json")

	if daemon.socketPath != "/tmp/test.sock" {
		t.Errorf("Expected socket path '/tmp/test.sock', got %s", daemon.socketPath)
	}

	if daemon.statePath != "/tmp/test_state.json" {
		t.Errorf("Expected state path '/tmp/test_state.json', got %s", daemon.statePath)
	}

	if daemon.running {
		t.Error("Expected daemon to not be running initially")
	}

	if daemon.checkInterval != 30*time.Second {
		t.Errorf("Expected default check interval 30s, got %v", daemon.checkInterval)
	}
}

func TestNewDaemonDefaultPaths(t *testing.T) {
	daemon := NewDaemon("", "")

	if daemon.socketPath != DefaultSocketPath {
		t.Errorf("Expected default socket path %s, got %s", DefaultSocketPath, daemon.socketPath)
	}

	if daemon.statePath != DefaultStatePath {
		t.Errorf("Expected default state path %s, got %s", DefaultStatePath, daemon.statePath)
	}
}

func TestDaemonStartStop(t *testing.T) {
	tempDir := t.TempDir()
	socketPath := filepath.Join(tempDir, "test.sock")
	statePath := filepath.Join(tempDir, "test_state.json")

	daemon := NewDaemon(socketPath, statePath)

	// Start daemon
	err := daemon.Start()
	if err != nil {
		t.Fatalf("Failed to start daemon: %v", err)
	}

	if !daemon.IsRunning() {
		t.Error("Expected daemon to be running")
	}

	// Check if socket file was created
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		t.Error("Expected socket file to be created")
	}

	// Check if state file was created
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Error("Expected state file to be created")
	}

	// Stop daemon
	err = daemon.Stop()
	if err != nil {
		t.Fatalf("Failed to stop daemon: %v", err)
	}

	if daemon.IsRunning() {
		t.Error("Expected daemon to be stopped")
	}

	// Check if socket file was removed
	if _, err := os.Stat(socketPath); !os.IsNotExist(err) {
		t.Error("Expected socket file to be removed")
	}
}

func TestDaemonStartAlreadyRunning(t *testing.T) {
	tempDir := t.TempDir()
	socketPath := filepath.Join(tempDir, "test.sock")
	statePath := filepath.Join(tempDir, "test_state.json")

	daemon := NewDaemon(socketPath, statePath)

	// Start daemon
	err := daemon.Start()
	if err != nil {
		t.Fatalf("Failed to start daemon: %v", err)
	}

	// Try to start again
	err = daemon.Start()
	if err == nil {
		t.Error("Expected error when starting already running daemon")
	}

	// Clean up
	daemon.Stop()
}

func TestDaemonGetters(t *testing.T) {
	daemon := NewDaemon("/tmp/test.sock", "/tmp/test_state.json")

	if daemon.GetPID() != os.Getpid() {
		t.Errorf("Expected PID %d, got %d", os.Getpid(), daemon.GetPID())
	}

	if daemon.GetVersion() != "dev" {
		t.Errorf("Expected version 'dev', got %s", daemon.GetVersion())
	}

	if daemon.GetSocketPath() != "/tmp/test.sock" {
		t.Errorf("Expected socket path '/tmp/test.sock', got %s", daemon.GetSocketPath())
	}

	if daemon.GetStatePath() != "/tmp/test_state.json" {
		t.Errorf("Expected state path '/tmp/test_state.json', got %s", daemon.GetStatePath())
	}
}

func TestDaemonSetCheckInterval(t *testing.T) {
	daemon := NewDaemon("/tmp/test.sock", "/tmp/test_state.json")

	// Test setting valid interval
	daemon.SetCheckInterval(45 * time.Second)
	if daemon.GetCheckInterval() != 45*time.Second {
		t.Errorf("Expected check interval 45s, got %v", daemon.GetCheckInterval())
	}

	// Test setting minimum interval
	daemon.SetCheckInterval(5 * time.Second)
	if daemon.GetCheckInterval() != 10*time.Second {
		t.Errorf("Expected minimum interval 10s, got %v", daemon.GetCheckInterval())
	}

	// Test setting maximum interval
	daemon.SetCheckInterval(15 * time.Minute)
	if daemon.GetCheckInterval() != 10*time.Minute {
		t.Errorf("Expected maximum interval 10m, got %v", daemon.GetCheckInterval())
	}
}

func TestMonitoringStatus(t *testing.T) {
	daemon := NewDaemon("/tmp/test.sock", "/tmp/test_state.json")

	status := daemon.GetMonitoringStatus()

	if status.Enabled {
		t.Error("Expected monitoring to be disabled initially")
	}

	if status.Interval != daemon.GetCheckInterval() {
		t.Errorf("Expected interval %v, got %v", daemon.GetCheckInterval(), status.Interval)
	}
}

func TestGetNextCheckTime(t *testing.T) {
	daemon := NewDaemon("/tmp/test.sock", "/tmp/test_state.json")
	daemon.SetCheckInterval(30 * time.Second)

	nextCheck := daemon.GetNextCheckTime()
	expected := time.Now().Add(30 * time.Second)

	// Allow 1 second tolerance
	diff := nextCheck.Sub(expected)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("Expected next check around %v, got %v (diff: %v)", expected, nextCheck, diff)
	}
}
