package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	statePath := "/tmp/test_state.json"
	manager := NewManager(statePath)

	if manager.statePath != statePath {
		t.Errorf("Expected state path %s, got %s", statePath, manager.statePath)
	}

	if manager.state == nil {
		t.Error("Expected state to be initialized")
	}
}

func TestStateManager_GetSet(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "test_state.json")
	manager := NewManager(statePath)

	// Test initial state
	if manager.GetConservationEnabled() {
		t.Error("Expected conservation to be disabled by default")
	}

	if manager.GetChargeThreshold() != 0 {
		t.Errorf("Expected threshold 0, got %d", manager.GetChargeThreshold())
	}

	// Test setting values
	err := manager.SetChargeThreshold(80)
	if err != nil {
		t.Errorf("Unexpected error setting threshold: %v", err)
	}

	if manager.GetChargeThreshold() != 80 {
		t.Errorf("Expected threshold 80, got %d", manager.GetChargeThreshold())
	}

	// Test enabling conservation
	err = manager.EnableConservation()
	if err != nil {
		t.Errorf("Unexpected error enabling conservation: %v", err)
	}

	if !manager.GetConservationEnabled() {
		t.Error("Expected conservation to be enabled")
	}

	state := manager.GetState()
	if state.CurrentMode != "enabled" {
		t.Errorf("Expected mode 'enabled', got %s", state.CurrentMode)
	}

	// Test disabling conservation
	err = manager.DisableConservation()
	if err != nil {
		t.Errorf("Unexpected error disabling conservation: %v", err)
	}

	if manager.GetConservationEnabled() {
		t.Error("Expected conservation to be disabled")
	}
}

func TestStateManager_UpdateBatteryInfo(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "test_state.json")
	manager := NewManager(statePath)

	// Set valid threshold first
	manager.state.ChargeThreshold = 80

	err := manager.UpdateBatteryInfo(75, true, true)
	if err != nil {
		t.Errorf("Unexpected error updating battery info: %v", err)
	}

	if manager.GetBatteryLevel() != 75 {
		t.Errorf("Expected battery level 75, got %d", manager.GetBatteryLevel())
	}

	if !manager.GetConservationMode() {
		t.Error("Expected conservation mode to be enabled")
	}

	if !manager.IsCharging() {
		t.Error("Expected charging to be true")
	}
}

func TestStateManager_ShouldEnableDisableConservation(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "test_state.json")
	manager := NewManager(statePath)

	// Set up state: management enabled, charging, threshold 80
	manager.state.ConservationEnabled = true
	manager.state.ChargeThreshold = 80

	// Test should enable when battery >= threshold
	manager.state.BatteryLevel = 85
	manager.state.Charging = true
	if !manager.ShouldEnableConservation() {
		t.Error("Should enable conservation when battery >= threshold")
	}

	// Test should disable when battery < threshold
	manager.state.BatteryLevel = 75
	if !manager.ShouldDisableConservation() {
		t.Error("Should disable conservation when battery < threshold")
	}

	// Test should not enable when not charging
	manager.state.BatteryLevel = 85
	manager.state.Charging = false
	if manager.ShouldEnableConservation() {
		t.Error("Should not enable conservation when not charging")
	}

	// Test should not do anything when management disabled
	manager.state.ConservationEnabled = false
	manager.state.Charging = true
	if manager.ShouldEnableConservation() {
		t.Error("Should not enable conservation when management disabled")
	}
	if manager.ShouldDisableConservation() {
		t.Error("Should not disable conservation when management disabled")
	}
}

func TestStateManager_Persistence(t *testing.T) {
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "test_state.json")

	// Create manager and set some state
	manager1 := NewManager(statePath)
	manager1.state.ConservationEnabled = true
	manager1.state.ChargeThreshold = 85
	manager1.state.BatteryLevel = 75
	manager1.state.CurrentMode = "enabled"

	// Save state
	err := manager1.Save()
	if err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Create new manager and load state
	manager2 := NewManager(statePath)
	err = manager2.Load()
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	// Verify state was loaded correctly
	state := manager2.GetState()
	if !state.ConservationEnabled {
		t.Error("Expected conservation enabled to be loaded")
	}

	if state.ChargeThreshold != 85 {
		t.Errorf("Expected threshold 85, got %d", state.ChargeThreshold)
	}

	if state.BatteryLevel != 75 {
		t.Errorf("Expected battery level 75, got %d", state.BatteryLevel)
	}

	if state.CurrentMode != "enabled" {
		t.Errorf("Expected mode 'enabled', got %s", state.CurrentMode)
	}
}

func TestStateManager_LoadDefaultState(t *testing.T) {
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "nonexistent_state.json")

	manager := NewManager(statePath)
	err := manager.Load()
	if err != nil {
		t.Errorf("Unexpected error loading non-existent state: %v", err)
	}

	state := manager.GetState()
	if state.ChargeThreshold != 80 {
		t.Errorf("Expected default threshold 80, got %d", state.ChargeThreshold)
	}

	if state.ConservationEnabled {
		t.Error("Expected conservation disabled by default")
	}

	if state.CurrentMode != "unknown" {
		t.Errorf("Expected default mode 'unknown', got %s", state.CurrentMode)
	}
}

func TestStateManager_BackupRestore(t *testing.T) {
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "test_state.json")

	manager := NewManager(statePath)
	manager.state.ConservationEnabled = true
	manager.state.ChargeThreshold = 85

	// Save state and create backup
	err := manager.Save()
	if err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	err = manager.Backup()
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Modify state
	manager.state.ConservationEnabled = false
	manager.state.ChargeThreshold = 90
	manager.Save()

	// Restore from backup
	err = manager.Restore()
	if err != nil {
		t.Fatalf("Failed to restore from backup: %v", err)
	}

	// Verify restored state
	state := manager.GetState()
	if !state.ConservationEnabled {
		t.Error("Expected conservation enabled after restore")
	}

	if state.ChargeThreshold != 85 {
		t.Errorf("Expected threshold 85 after restore, got %d", state.ChargeThreshold)
	}
}

func TestStateManager_Validate(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "test_state.json")
	manager := NewManager(statePath)

	// Test valid state
	manager.state.ChargeThreshold = 80
	manager.state.BatteryLevel = 50
	manager.state.CurrentMode = "enabled"
	manager.state.PID = 1234 // Set valid PID
	err := manager.Validate()
	if err != nil {
		t.Errorf("Unexpected validation error: %v", err)
	}

	// Test invalid threshold
	manager.state.ChargeThreshold = 50
	err = manager.Validate()
	if err != ErrInvalidThreshold {
		t.Errorf("Expected ErrInvalidThreshold, got %v", err)
	}

	// Test invalid battery level
	manager.state.ChargeThreshold = 80
	manager.state.BatteryLevel = 150
	err = manager.Validate()
	if err != ErrInvalidBatteryLevel {
		t.Errorf("Expected ErrInvalidBatteryLevel, got %v", err)
	}

	// Test invalid mode
	manager.state.BatteryLevel = 50
	manager.state.ChargeThreshold = 80 // Reset to valid
	manager.state.PID = 1234           // Reset to valid
	manager.state.CurrentMode = "invalid"
	err = manager.Validate()
	if err != ErrInvalidMode {
		t.Errorf("Expected ErrInvalidMode, got %v", err)
	}
}

func TestStateManager_Reset(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "test_state.json")
	manager := NewManager(statePath)

	// Set some state
	manager.state.ConservationEnabled = true
	manager.state.ChargeThreshold = 90
	manager.state.BatteryLevel = 80

	// Reset
	err := manager.Reset()
	if err != nil {
		t.Errorf("Unexpected error resetting state: %v", err)
	}

	state := manager.GetState()
	if state.ConservationEnabled {
		t.Error("Expected conservation disabled after reset")
	}

	if state.ChargeThreshold != 80 {
		t.Errorf("Expected default threshold 80 after reset, got %d", state.ChargeThreshold)
	}

	if state.BatteryLevel != 0 {
		t.Errorf("Expected battery level 0 after reset, got %d", state.BatteryLevel)
	}
}

func TestStateManager_UpdateState(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "test_state.json")
	manager := NewManager(statePath)

	// Test atomic update
	err := manager.UpdateState(func(s *State) {
		s.ConservationEnabled = true
		s.ChargeThreshold = 85
		s.BatteryLevel = 75
	})
	if err != nil {
		t.Errorf("Unexpected error updating state: %v", err)
	}

	state := manager.GetState()
	if !state.ConservationEnabled {
		t.Error("Expected conservation enabled after update")
	}

	if state.ChargeThreshold != 85 {
		t.Errorf("Expected threshold 85 after update, got %d", state.ChargeThreshold)
	}

	if state.BatteryLevel != 75 {
		t.Errorf("Expected battery level 75 after update, got %d", state.BatteryLevel)
	}
}

func TestStateManager_Uptime(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "test_state.json")
	manager := NewManager(statePath)

	// Initially no start time, uptime should be 0
	uptime := manager.GetUptime()
	if uptime != 0 {
		t.Errorf("Expected uptime 0, got %v", uptime)
	}

	// Set start time
	startTime := time.Now().Add(-10 * time.Second)
	manager.state.StartTime = startTime

	uptime = manager.GetUptime()
	if uptime < 9*time.Second || uptime > 11*time.Second {
		t.Errorf("Expected uptime around 10 seconds, got %v", uptime)
	}
}

func TestStateManager_Exists(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "test_state.json")
	manager := NewManager(statePath)

	// Initially should not exist
	if manager.Exists() {
		t.Error("Expected state file to not exist initially")
	}

	// Create file
	file, err := os.Create(statePath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	// Now should exist
	if !manager.Exists() {
		t.Error("Expected state file to exist")
	}
}

func TestStateManager_Remove(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "test_state.json")
	manager := NewManager(statePath)

	// Create file
	file, err := os.Create(statePath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	// Verify it exists
	if !manager.Exists() {
		t.Error("Expected state file to exist")
	}

	// Remove it
	err = manager.Remove()
	if err != nil {
		t.Errorf("Unexpected error removing state file: %v", err)
	}

	// Verify it's gone
	if manager.Exists() {
		t.Error("Expected state file to be removed")
	}
}

func TestStateErrors(t *testing.T) {
	tests := []struct {
		name string
		err  *StateError
		want string
	}{
		{"invalid threshold", ErrInvalidThreshold, "threshold must be between 60 and 100"},
		{"invalid battery level", ErrInvalidBatteryLevel, "battery level must be between 0 and 100"},
		{"invalid PID", ErrInvalidPID, "PID must be positive"},
		{"invalid mode", ErrInvalidMode, "invalid current mode"},
		{"no backup", ErrNoBackup, "no backup file found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("StateError.Error() = %v, want %v", tt.err.Error(), tt.want)
			}

			if !IsStateError(tt.err) {
				t.Error("Expected IsStateError to return true")
			}

			var regularErr error = tt.err
			if !IsStateError(regularErr) {
				t.Error("Expected IsStateError to work with error interface")
			}
		})
	}
}
