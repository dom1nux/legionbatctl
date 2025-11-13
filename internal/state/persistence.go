package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Load loads the state from file
func (m *Manager) Load() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// If file doesn't exist, create default state
	if _, err := os.Stat(m.statePath); os.IsNotExist(err) {
		m.state = createDefaultState()
		return m.saveStateAtomic()
	}

	// Read existing state file
	data, err := os.ReadFile(m.statePath)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		// If JSON is corrupted, create default state
		m.state = createDefaultState()
		return fmt.Errorf("failed to unmarshal state file, using defaults: %w", err)
	}

	// Validate loaded state
	m.state = &state
	if err := m.validateState(); err != nil {
		// If state is invalid, create default state
		m.state = createDefaultState()
		return fmt.Errorf("invalid state file, using defaults: %w", err)
	}

	return nil
}

// Save saves the current state to file (requires write lock)
func (m *Manager) Save() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.saveStateAtomic()
}

// saveStateAtomic saves the state atomically using temp file + rename
func (m *Manager) saveStateAtomic() error {
	// Validate state before saving
	if err := m.validateState(); err != nil {
		return fmt.Errorf("invalid state: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(m.statePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Create temporary file
	tempPath := m.statePath + ".tmp"
	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer file.Close()

	// Write JSON with indentation for readability
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(m.state); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to write state to temp file: %w", err)
	}

	// Ensure data is written to disk
	if err := file.Sync(); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	// Close file before rename
	if err := file.Close(); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, m.statePath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	// Set appropriate permissions
	if err := os.Chmod(m.statePath, 0644); err != nil {
		return fmt.Errorf("failed to set permissions on state file: %w", err)
	}

	return nil
}

// validateState validates the current state (requires read lock)
func (m *Manager) validateState() error {
	return validateStateFields(m.state)
}

// createDefaultState creates a default state
func createDefaultState() *State {
	return &State{
		ConservationEnabled: false,
		ChargeThreshold:     80, // Default threshold for battery health
		CurrentMode:         "unknown",
		LastAction:          "init",
		LastActionTime:      time.Now(),
		BatteryLevel:        0,
		ConservationMode:    false,
		Charging:            false,
		PID:                 0,
		StartTime:           time.Now(),
	}
}

// Backup creates a backup of the current state file
func (m *Manager) Backup() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if _, err := os.Stat(m.statePath); os.IsNotExist(err) {
		return nil // No file to backup
	}

	backupPath := m.statePath + ".backup"
	data, err := os.ReadFile(m.statePath)
	if err != nil {
		return fmt.Errorf("failed to read state file for backup: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}

// Restore restores state from backup
func (m *Manager) Restore() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	backupPath := m.statePath + ".backup"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return ErrNoBackup
	}

	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to unmarshal backup file: %w", err)
	}

	m.state = &state
	return m.saveStateAtomic()
}

// Remove removes the state file
func (m *Manager) Remove() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if err := os.Remove(m.statePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove state file: %w", err)
	}

	// Also remove backup if it exists
	backupPath := m.statePath + ".backup"
	os.Remove(backupPath) // Ignore error

	return nil
}

// Exists checks if the state file exists
func (m *Manager) Exists() bool {
	_, err := os.Stat(m.statePath)
	return !os.IsNotExist(err)
}

// GetStatePath returns the state file path
func (m *Manager) GetStatePath() string {
	return m.statePath
}