# Agent Guidelines for legionbatctl

## Build and Test Commands

### Building
```bash
# Build the binary (standard)
make build

# Manual build
go build -o build/legionbatctl ./cmd/legionbatctl

# Build with version info
go build -ldflags "-X main.version=$(VERSION)" ./cmd/legionbatctl
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/daemon -v
go test ./internal/client -v
go test ./internal/state -v
go test ./internal/protocol -v

# Run a single test function
go test ./internal/state -v -run TestNewManager
go test ./internal/state -v -run TestStateManager_GetSet

# Run tests matching a pattern
go test ./internal/state -v -run TestStateManager/.*
```

### Other Commands
```bash
# Clean build artifacts
make clean

# Install and start service (requires root)
sudo make install

# Check service status
make status

# View daemon logs
make logs
```

## Code Style Guidelines

### Imports
- Standard library imports first, grouped together
- Third-party imports after, grouped together
- Internal imports (`github.com/dom1nux/legionbatctl/...`) last
- No blank lines between import groups
- Remove unused imports

Example:
```go
import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/dom1nux/legionbatctl/internal/state"
)
```

### Formatting
- Use `gofmt` to format code (standard Go formatting)
- Use tabs for indentation (Go standard)
- Maximum line length: ~100-120 characters (flexible)
- Use blank lines between related functions (1 blank line)

### Types and Structs
- Exported types use PascalCase
- Unexported types use camelCase
- Add JSON tags for structs that will be serialized
- Document struct fields that aren't self-explanatory
- Keep structs focused and single-purpose

Example:
```go
type State struct {
	ConservationEnabled bool `json:"conservation_enabled"`
	ChargeThreshold     int  `json:"charge_threshold"`
	BatteryLevel        int  `json:"battery_level"`
}

type Manager struct {
	statePath string
	mutex     sync.RWMutex
	state     *State
}
```

### Naming Conventions
- **Packages**: lowercase, single word, concise (e.g., `state`, `daemon`, `client`)
- **Constants**: PascalCase for exported, UPPER_SNAKE_CASE for global constants
- **Variables**: camelCase
- **Functions**: PascalCase for exported, camelCase for unexported
- **Interfaces**: PascalCase, typically end with `-er` suffix (e.g., `Writer`, `Reader`)
- **Acronyms**: capitalize first letter only (e.g., `XmlWriter`, not `XMLWriter`)

### Error Handling
- Use custom error types for package-specific errors
- Define package-level error variables for common errors
- Wrap errors with context using `fmt.Errorf` and `%w`
- Check errors immediately after operations
- Return errors, don't panic in application code

Example:
```go
// Custom error type
type StateError struct {
	Message string
}

func (e *StateError) Error() string {
	return e.Message
}

// Package-level error variables
var (
	ErrInvalidThreshold = NewStateError("threshold must be between 60 and 100")
	ErrInvalidPID       = NewStateError("PID must be positive")
)

// Error wrapping
if err := d.stateManager.Load(); err != nil {
	return fmt.Errorf("failed to load state: %w", err)
}
```

### Concurrency
- Use `sync.RWMutex` for read-heavy scenarios (read lock for Getters, write lock for Setters)
- Always defer mutex unlocking
- Use channels for goroutine communication
- Be careful with shared state across goroutines

Example:
```go
func (m *Manager) GetState() State {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return *m.state
}

func (m *Manager) UpdateState(updateFn func(*State)) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	updateFn(m.state)
	return m.saveStateAtomic()
}
```

### Testing
- Use `t.TempDir()` for temporary files/directories
- Use table-driven tests for multiple test cases
- Use `t.Run()` for subtests
- Name test functions with `Test` prefix (e.g., `TestNewManager`)
- Use descriptive test names that explain what is being tested
- Test both success and failure paths
- Use `assert` patterns consistently

Example:
```go
func TestStateManager_GetSet(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "test_state.json")
	manager := NewManager(statePath)

	if manager.GetChargeThreshold() != 0 {
		t.Errorf("Expected threshold 0, got %d", manager.GetChargeThreshold())
	}

	err := manager.SetChargeThreshold(80)
	if err != nil {
		t.Errorf("Unexpected error setting threshold: %v", err)
	}
}
```

### Project Structure
- `cmd/`: Main entry points (e.g., `cmd/legionbatctl/main.go`)
- `internal/`: Internal packages not imported outside this project
  - `cli/`: Command-line interface and commands
  - `client/`: Socket client for daemon communication
  - `daemon/`: Background daemon and battery monitoring
  - `protocol/`: Message types and communication protocol
  - `state/`: State management and persistence
- `pkg/`: Public packages (if any)
- `systemd/`: Systemd service files

### Constants and Configuration
- Define constants at package level
- Use descriptive names (e.g., `DefaultSocketPath`, `DefaultTimeout`)
- Group related constants together
- Use `const` block for multiple constants

Example:
```go
const (
	DefaultSocketPath = "/var/run/legionbatctl.sock"
	DefaultStatePath  = "/etc/legionbatctl.state"
	DefaultTimeout    = 10 * time.Second
)
```

### File Organization
- One struct per file is common, but not required
- Keep related functions in the same file
- Separate files for different concerns (e.g., `errors.go`, `types.go`)
- Use descriptive file names

### Comments
- Exported functions must have comments explaining behavior
- Keep comments concise and focused on "what" and "why", not "how"
- Use godoc format (comments start with function name)
- Avoid inline comments for self-explanatory code

Example:
```go
// GetState returns a copy of the current state (thread-safe)
func (m *Manager) GetState() State {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	state := *m.state
	return state
}
```

### Additional Guidelines
- Use environment variables for configuration that needs to be overridden (e.g., `SOCKET_PATH`, `STATE_PATH`)
- Prefer returning copies of state instead of references to prevent mutation
- Use atomic file operations for persistence (write to temp file, then rename)
- Implement graceful shutdown for long-running processes
- Use time.Duration constants instead of magic numbers for timeouts

Example:
```go
// atomic write
func (m *Manager) saveStateAtomic() error {
	tmpPath := m.statePath + ".tmp"
	if err := json.MarshalToFile(m.state, tmpPath); err != nil {
		return err
	}
	return os.Rename(tmpPath, m.statePath)
}
```
