package state

import (
	"sync"
	"time"
)

// State represents the current runtime state of the battery management system
type State struct {
	// Configuration
	ConservationEnabled bool `json:"conservation_enabled"`
	ChargeThreshold     int  `json:"charge_threshold"`

	// Runtime State
	CurrentMode    string    `json:"current_mode"` // "enabled", "disabled", "unknown"
	LastAction     string    `json:"last_action"`  // "enable", "disable", "set_threshold", "auto"
	LastActionTime time.Time `json:"last_action_time"`

	// Battery Information
	BatteryLevel     int  `json:"battery_level"`
	ConservationMode bool `json:"conservation_mode"` // Hardware conservation mode state
	Charging         bool `json:"charging"`

	// Daemon Information
	PID       int       `json:"pid"`
	StartTime time.Time `json:"start_time"`
}

// Manager manages the state with thread-safe operations and persistence
type Manager struct {
	statePath string
	mutex     sync.RWMutex
	state     *State
}

// NewManager creates a new state manager
func NewManager(statePath string) *Manager {
	return &Manager{
		statePath: statePath,
		state: &State{
			CurrentMode: "unknown", // Initialize with valid default
		},
	}
}

// GetState returns a copy of the current state (thread-safe)
func (m *Manager) GetState() State {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Return a copy to prevent external modification
	state := *m.state
	return state
}

// GetBatteryLevel returns the current battery level
func (m *Manager) GetBatteryLevel() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.state.BatteryLevel
}

// GetConservationEnabled returns whether conservation management is enabled
func (m *Manager) GetConservationEnabled() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.state.ConservationEnabled
}

// GetChargeThreshold returns the current charge threshold
func (m *Manager) GetChargeThreshold() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.state.ChargeThreshold
}

// GetConservationMode returns the hardware conservation mode state
func (m *Manager) GetConservationMode() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.state.ConservationMode
}

// IsCharging returns whether the battery is currently charging
func (m *Manager) IsCharging() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.state.Charging
}

// UpdateState performs an atomic update of the state
func (m *Manager) UpdateState(updateFn func(*State)) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	updateFn(m.state)
	return m.saveStateAtomic()
}

// EnableConservation enables battery management
func (m *Manager) EnableConservation() error {
	return m.UpdateState(func(s *State) {
		s.ConservationEnabled = true
		s.CurrentMode = "enabled"
		s.LastAction = "enable"
		s.LastActionTime = time.Now()
	})
}

// DisableConservation disables battery management
func (m *Manager) DisableConservation() error {
	return m.UpdateState(func(s *State) {
		s.ConservationEnabled = false
		s.CurrentMode = "disabled"
		s.LastAction = "disable"
		s.LastActionTime = time.Now()
	})
}

// SetChargeThreshold sets the charge threshold
func (m *Manager) SetChargeThreshold(threshold int) error {
	return m.UpdateState(func(s *State) {
		s.ChargeThreshold = threshold
		s.LastAction = "set_threshold"
		s.LastActionTime = time.Now()
	})
}

// UpdateBatteryInfo updates battery-related information
func (m *Manager) UpdateBatteryInfo(level int, conservationMode, charging bool) error {
	return m.UpdateState(func(s *State) {
		s.BatteryLevel = level
		s.ConservationMode = conservationMode
		s.Charging = charging
		s.LastAction = "auto"
		s.LastActionTime = time.Now()
	})
}

// SetDaemonInfo sets daemon-related information
func (m *Manager) SetDaemonInfo(pid int) error {
	return m.UpdateState(func(s *State) {
		s.PID = pid
		s.StartTime = time.Now()
	})
}

// ShouldEnableConservation determines if conservation mode should be enabled
func (m *Manager) ShouldEnableConservation() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Only enable if management is enabled AND on AC power AND battery >= threshold
	return m.state.ConservationEnabled &&
		m.state.Charging &&
		m.state.BatteryLevel >= m.state.ChargeThreshold
}

// ShouldDisableConservation determines if conservation mode should be disabled
func (m *Manager) ShouldDisableConservation() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Only disable if management is enabled AND on AC power AND battery < threshold
	return m.state.ConservationEnabled &&
		m.state.Charging &&
		m.state.BatteryLevel < m.state.ChargeThreshold
}

// GetUptime returns the daemon uptime
func (m *Manager) GetUptime() time.Duration {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.state.StartTime.IsZero() {
		return 0
	}
	return time.Since(m.state.StartTime)
}

// validateStateFields validates state field values (internal helper)
func validateStateFields(state *State) error {
	// Validate threshold
	if state.ChargeThreshold < 60 || state.ChargeThreshold > 100 {
		return ErrInvalidThreshold
	}

	// Validate battery level
	if state.BatteryLevel < 0 || state.BatteryLevel > 100 {
		return ErrInvalidBatteryLevel
	}

	// Validate current mode
	validModes := map[string]bool{
		"enabled":  true,
		"disabled": true,
		"unknown":  true,
	}
	if !validModes[state.CurrentMode] {
		return ErrInvalidMode
	}

	return nil
}

// Validate validates the current state
func (m *Manager) Validate() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Validate common state fields
	if err := validateStateFields(m.state); err != nil {
		return err
	}

	// Validate PID (only for full validation)
	if m.state.PID <= 0 {
		return ErrInvalidPID
	}

	return nil
}

// Reset resets the state to defaults
func (m *Manager) Reset() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.state = &State{
		ConservationEnabled: false,
		ChargeThreshold:     80, // Default threshold
		CurrentMode:         "unknown",
		BatteryLevel:        0,
		ConservationMode:    false,
		Charging:            false,
	}

	return m.saveStateAtomic()
}
