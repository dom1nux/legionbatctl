package state

import "fmt"

// Common state management errors
var (
	ErrInvalidThreshold    = NewStateError("threshold must be between 60 and 100")
	ErrInvalidBatteryLevel = NewStateError("battery level must be between 0 and 100")
	ErrInvalidPID          = NewStateError("PID must be positive")
	ErrInvalidMode         = NewStateError("invalid current mode")
	ErrNoBackup            = NewStateError("no backup file found")
)

// StateError represents a state management error
type StateError struct {
	Message string
}

func NewStateError(message string) *StateError {
	return &StateError{Message: message}
}

func (e *StateError) Error() string {
	return e.Message
}

// IsStateError checks if an error is a StateError
func IsStateError(err error) bool {
	_, ok := err.(*StateError)
	return ok
}

// WrapStateError wraps an error with state context
func WrapStateError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("state %s: %w", context, err)
}
