package protocol

import "time"

// Message represents a communication message between CLI and daemon
type Message struct {
	Type     string     `json:"type"`     // "request", "response"
	ID       string     `json:"id"`       // Unique request ID
	Request  *Request   `json:"request,omitempty"`
	Response *Response  `json:"response,omitempty"`
}

// Request represents a command request from CLI to daemon
type Request struct {
	Command string                 `json:"command"` // "enable", "disable", "status", "set_threshold", "daemon_status"
	Params  map[string]interface{} `json:"params"`
}

// Response represents a response from daemon to CLI
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Command constants
const (
	CmdEnable       = "enable"
	CmdDisable      = "disable"
	CmdStatus       = "status"
	CmdSetThreshold = "set_threshold"
	CmdDaemonStatus = "daemon_status"
)

// StatusData represents the data returned by status command
type StatusData struct {
	ConservationEnabled bool      `json:"conservation_enabled"`
	Threshold          int       `json:"threshold"`
	CurrentMode        string    `json:"current_mode"`
	BatteryLevel       int       `json:"battery_level"`
	ConservationMode   bool      `json:"conservation_mode"`
	Charging           bool      `json:"charging"`
	LastAction         string    `json:"last_action"`
	LastActionTime     time.Time `json:"last_action_time"`
	DaemonUptime       string    `json:"daemon_uptime"`
	HardwareSupported  bool      `json:"hardware_supported"`
}

// EnableData represents the data returned by enable command
type EnableData struct {
	Message     string `json:"message"`
	Threshold   int    `json:"threshold"`
	CurrentMode string `json:"current_mode"`
}

// DisableData represents the data returned by disable command
type DisableData struct {
	Message     string `json:"message"`
	CurrentMode string `json:"current_mode"`
}

// SetThresholdData represents the data returned by set_threshold command
type SetThresholdData struct {
	Message   string `json:"message"`
	Threshold int    `json:"threshold"`
}

// DaemonStatusData represents the data returned by daemon_status command
type DaemonStatusData struct {
	Running     bool   `json:"running"`
	PID         int    `json:"pid"`
	Uptime      string `json:"uptime"`
	Version     string `json:"version"`
	SocketPath  string `json:"socket_path"`
	StateFile   string `json:"state_file"`
}

// IsValidCommand checks if a command string is valid
func IsValidCommand(cmd string) bool {
	validCommands := map[string]bool{
		CmdEnable:       true,
		CmdDisable:      true,
		CmdStatus:       true,
		CmdSetThreshold: true,
		CmdDaemonStatus: true,
	}
	return validCommands[cmd]
}

// ValidateThreshold validates a threshold value
func ValidateThreshold(threshold int) error {
	if threshold < 60 || threshold > 100 {
		return ErrInvalidThreshold
	}
	return nil
}

// Common errors
var (
	ErrInvalidThreshold = NewError("threshold must be between 60 and 100")
	ErrDaemonNotRunning = NewError("daemon not running")
	ErrHardwareNotSupported = NewError("hardware not supported")
	ErrPermissionDenied = NewError("permission denied")
	ErrInvalidCommand = NewError("invalid command")
)

// Error represents a protocol error
type Error struct {
	Message string
}

func NewError(message string) *Error {
	return &Error{Message: message}
}

func (e *Error) Error() string {
	return e.Message
}