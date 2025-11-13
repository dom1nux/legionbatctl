# LegionBatCTL Go Conversion Plan

## Project Overview

Convert the existing bash-based Lenovo Legion battery control utility to Go to improve reliability, performance, and maintainability while preserving the unique functionality of hacking around the 60% conservation mode limit on the Lenovo Legion Slim 7 15ACH6 (2021).

### Current Situation
- Hardware conservation mode is fixed at 60% charge limit
- Goal is to achieve 80% charge limit through software control
- Current solution: Enable conservation mode when battery reaches 80%, disable when below 80%
- Uses systemd timer with 1-minute intervals

### Go Conversion Benefits
- Better timing precision for critical threshold management
- Improved error handling and recovery
- Cross-compilation support
- Enhanced logging and debugging capabilities
- More robust hardware interaction

## Project Structure

```
legionbatctl-go/
├── cmd/
│   ├── legionbatctl/          # Main CLI binary
│   │   └── main.go
│   └── legionbatctl-auto/     # Auto-mode binary
│       └── main.go
├── internal/
│   ├── config/                # Configuration management
│   │   ├── config.go
│   │   └── config_test.go
│   ├── battery/               # Battery operations
│   │   ├── battery.go
│   │   ├── battery_test.go
│   │   └── types.go
│   ├── conservation/          # Conservation mode control
│   │   ├── conservation.go
│   │   ├── conservation_test.go
│   │   └── hardware.go
│   ├── power/                 # Power source detection
│   │   ├── power.go
│   │   └── power_test.go
│   └── logger/                # Logging system
│       ├── logger.go
│       └── logger_test.go
├── pkg/
│   ├── cli/                   # Shared CLI utilities
│   │   ├── root.go
│   │   ├── status.go
│   │   ├── toggle.go
│   │   ├── enable.go
│   │   ├── disable.go
│   │   └── threshold.go
│   └── version/               # Version information
│       └── version.go
├── systemd/                   # Updated systemd files
│   ├── legionbatctl.service
│   ├── legionbatctl.timer
│   └── README.md
├── man/
│   └── legionbatctl.1
├── scripts/
│   ├── install.sh
│   └── uninstall.sh
├── docs/
│   ├── DESIGN.md
│   ├── API.md
│   └── TROUBLESHOOTING.md
├── test/
│   ├── integration/
│   └── mocks/
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── LICENSE
└── CHANGELOG.md
```

## Core Components

### 1. Configuration Management (`internal/config/`)

**Purpose**: Handle configuration file operations with atomic updates

**Key Features**:
- Read/write `/etc/legionbatctl.conf`
- Backward compatibility with existing config format
- Atomic configuration updates with file locking
- Default value management
- Configuration validation

**Interface**:
```go
type Config struct {
    ConservationEnabled bool `json:"conservation_enabled"`
    ChargeThreshold     int  `json:"charge_threshold"`
    LogLevel           string `json:"log_level"`
}

type Manager interface {
    Load() (*Config, error)
    Save(*Config) error
    UpdateConservationEnabled(bool) error
    UpdateChargeThreshold(int) error
}
```

### 2. Battery Operations (`internal/battery/`)

**Purpose**: Safe battery level reading and validation

**Key Features**:
- Auto-discovery of battery path (BAT0, BAT1, etc.)
- Battery level validation
- Charging status detection
- Error handling for invalid readings
- Trend analysis (optional)

**Interface**:
```go
type BatteryInfo struct {
    Level    int    `json:"level"`
    Status   string `json:"status"` // Charging, Discharging, Full
    Path     string `json:"path"`
}

type Manager interface {
    GetBatteryInfo() (*BatteryInfo, error)
    GetBatteryLevel() (int, error)
    IsCharging() (bool, error)
    IsValidLevel(int) bool
}
```

### 3. Conservation Control (`internal/conservation/`)

**Purpose**: Hardware interaction for conservation mode

**Key Features**:
- Auto-discovery of conservation mode path
- Safe write operations with verification
- Hardware capability detection
- Latency handling for hardware response
- Retry logic with exponential backoff

**Interface**:
```go
type Controller interface {
    IsSupported() bool
    GetStatus() (bool, error)
    Enable() error
    Disable() error
    SetMode(bool) error
    VerifyMode(bool) error
}

type HardwareInfo struct {
    Supported    bool   `json:"supported"`
    Path        string `json:"path"`
    CurrentMode bool   `json:"current_mode"`
}
```

### 4. Power Detection (`internal/power/`)

**Purpose**: Detect AC power status for smart operation

**Key Features**:
- AC power detection
- Power source switching detection
- Optimize operations based on power state

**Interface**:
```go
type PowerSource int

const (
    PowerUnknown PowerSource = iota
    PowerBattery
    PowerAC
)

type Detector interface {
    GetPowerSource() (PowerSource, error)
    IsOnAC() (bool, error)
    WatchPowerChanges() <-chan PowerSource
}
```

### 5. Logging System (`internal/logger/`)

**Purpose**: Structured logging with multiple outputs

**Key Features**:
- Configurable log levels
- System logging (syslog)
- File logging for debugging
- Contextual information
- Performance metrics

## Implementation Phases

### Phase 1: Core Foundation (Week 1)

**Priority**: High
**Timeline**: 3-4 days

1. **Project Setup**
   - Initialize Go module
   - Set up directory structure
   - Configure CI/CD pipeline
   - Create Makefile

2. **Basic Configuration System**
   - Implement configuration file reading
   - Add backward compatibility
   - Create configuration validation
   - Write unit tests

3. **Hardware Discovery**
   - Implement battery path discovery
   - Implement conservation mode discovery
   - Add hardware capability detection
   - Create mock interfaces for testing

4. **Basic CLI Framework**
   - Set up Cobra CLI framework
   - Implement basic help and version commands
   - Add error handling framework

**Deliverables**:
- Working CLI with help/version
- Configuration system
- Hardware discovery
- Unit tests (80%+ coverage)

### Phase 2: Core Functionality (Week 2)

**Priority**: High
**Timeline**: 4-5 days

1. **Battery Operations**
   - Implement safe battery reading
   - Add charging status detection
   - Create battery level validation
   - Add trend analysis (if time permits)

2. **Conservation Control**
   - Implement conservation mode reading/writing
   - Add verification logic
   - Create retry mechanism
   - Handle hardware latency

3. **Main CLI Commands**
   - Implement `status` command
   - Implement `enable` command
   - Implement `disable` command
   - Implement `toggle` command

4. **Configuration Commands**
   - Implement `set-threshold` command
   - Add configuration validation
   - Create atomic config updates

**Deliverables**:
- Fully functional main CLI
- All manual commands working
- Comprehensive error handling
- Integration tests

### Phase 3: Auto Mode & Intelligence (Week 3)

**Priority**: High
**Timeline**: 4-5 days

1. **Auto Mode Binary**
   - Create `legionbatctl-auto` binary
   - Implement core logic with proper timing
   - Add AC power detection
   - Create debouncing logic

2. **Enhanced Logic**
   - Implement proper threshold crossing detection
   - Add charging state validation
   - Create smart timing intervals
   - Add failure recovery

3. **Power Management**
   - Implement AC power detection
   - Add power state change monitoring
   - Create operation optimization

4. **Logging & Monitoring**
   - Implement structured logging
   - Add performance metrics
   - Create debug mode
   - Add systemd integration

**Deliverables**:
- Working auto mode binary
- Enhanced battery management logic
- Comprehensive logging
- Systemd integration

### Phase 4: Polish & Integration (Week 4)

**Priority**: Medium
**Timeline**: 3-4 days

1. **System Integration**
   - Update systemd service files
   - Modify install/uninstall scripts
   - Update man page
   - Test installation process

2. **Testing & QA**
   - Comprehensive integration testing
   - Performance testing
   - Edge case testing
   - Hardware compatibility testing

3. **Documentation**
   - Update README
   - Create design document
   - Write troubleshooting guide
   - Document API/interfaces

4. **Release Preparation**
   - Create build scripts
   - Set up release pipeline
   - Create packages
   - Write changelog

**Deliverables**:
- Production-ready binaries
- Updated system integration
- Complete documentation
- Release packages

## Critical Design Decisions

### 1. Threshold Management Logic

**Current Issue**: Precise timing needed for 80% threshold
**Go Solution**: Enhanced logic with trend detection

```go
func shouldEnableConservation(currentLevel, threshold int, isCharging bool) bool {
    if !isCharging {
        return false // Don't enable if not charging
    }

    // Enable at or above threshold
    return currentLevel >= threshold
}

func shouldDisableConservation(currentLevel, threshold int, isCharging bool) bool {
    // Disable if below threshold AND charging is needed
    return currentLevel < threshold && isCharging
}
```

### 2. Timer Frequency

**Current**: 1-minute intervals
**Go Enhancement**: Adaptive timing

```go
func getCheckInterval(batteryLevel, threshold int) time.Duration {
    difference := math.Abs(float64(batteryLevel - threshold))

    if difference < 5 { // Within 5% of threshold
        return 30 * time.Second // Check every 30 seconds
    } else if difference < 15 { // Within 15% of threshold
        return 1 * time.Minute // Check every minute
    }

    return 2 * time.Minute // Check every 2 minutes
}
```

### 3. Error Handling Strategy

**Principle**: Fail safe and recover gracefully

```go
type Manager struct {
    config   *config.Manager
    battery  *battery.Manager
 conserve *conservation.Controller
    logger   *logger.Logger
}

func (m *Manager) ManageBattery() error {
    // Always check prerequisites first
    if !m.conserve.IsSupported() {
        return errors.New("conservation mode not supported")
    }

    if !m.isOnACPower() {
        m.logger.Info("On battery power, skipping management")
        return nil
    }

    // Core logic with error recovery
    return m.executeManagementCycle()
}
```

### 4. Hardware Compatibility

**Strategy**: Auto-discovery with fallback

```go
func discoverConservationPath() (string, error) {
    candidates := []string{
        "/sys/bus/platform/drivers/ideapad_acpi/VPC2004:00/conservation_mode",
        "/sys/devices/platform/VPC2004:00/conservation_mode",
        // Add more known paths as needed
    }

    for _, path := range candidates {
        if _, err := os.Stat(path); err == nil {
            return path, nil
        }
    }

    return "", errors.New("conservation mode file not found")
}
```

## Testing Strategy

### 1. Unit Testing
- Each package has comprehensive unit tests
- Mock interfaces for hardware interactions
- Test coverage target: 85%+

### 2. Integration Testing
- Test full battery management cycle
- Test configuration persistence
- Test CLI command integration

### 3. Hardware Testing
- Test on actual Lenovo Legion hardware
- Test with different battery levels
- Test conservation mode timing

### 4. Edge Case Testing
- Test with invalid configuration
- Test with missing hardware files
- Test with rapid power state changes

## Deployment Strategy

### 1. Migration Path
1. Build Go binaries alongside existing bash scripts
2. Test Go version in parallel with bash version
3. Switch to Go version after validation period
4. Remove bash scripts

### 2. Backward Compatibility
- Maintain same CLI interface
- Preserve configuration file format
- Keep same systemd service interface

### 3. Rollback Plan
- Keep bash scripts as backup
- Create quick uninstall procedure
- Document manual rollback steps

## Performance Considerations

### 1. Memory Usage
- Minimal memory footprint (< 10MB)
- Efficient string handling
- Proper resource cleanup

### 2. CPU Usage
- Low CPU overhead
- Efficient polling intervals
- Smart power state detection

### 3. I/O Optimization
- Batch system file reads
- Caching for frequently accessed data
- Atomic file operations

## Security Considerations

### 1. Privilege Management
- Run with minimal required privileges
- Validate all file paths
- Secure temporary file handling

### 2. Input Validation
- Validate configuration values
- Sanitize all user inputs
- Prevent path traversal

### 3. File Permissions
- Secure configuration file permissions
- Proper ownership of system files
- Audit logging for privileged operations

## Success Criteria

1. **Functional Parity**: All existing functionality works in Go version
2. **Improved Reliability**: Better error handling and recovery
3. **Enhanced Performance**: More precise timing and lower resource usage
4. **Maintainability**: Clean, testable, well-documented code
5. **Hardware Compatibility**: Works with target Legion Slim 7 hardware
6. **Backward Compatibility**: Seamless migration from bash version

## Risks & Mitigations

### 1. Hardware Timing Issues
**Risk**: Go version may have different timing characteristics
**Mitigation**: Extensive testing on target hardware, adjustable timing parameters

### 2. Permission Issues
**Risk**: Go binary may have different permission requirements
**Mitigation**: Test with various permission scenarios, clear documentation

### 3. Configuration Compatibility
**Risk**: Go config handling may differ from bash version
**Mitigation**: Comprehensive testing of configuration scenarios

### 4. System Integration
**Risk**: Systemd integration may need adjustments
**Mitigation**: Careful testing of service files and timers

## Future Enhancements

### 1. GUI Interface
- Optional GUI for battery management
- System tray integration
- Real-time battery monitoring

### 2. Advanced Features
- Multiple battery support
- Charging profiles
- Usage statistics
- Battery health monitoring

### 3. Cross-Platform Support
- Support for other Lenovo models
- Support for different Linux distributions
- Potential Windows/macOS support

### 4. Monitoring & Analytics
- Battery degradation tracking
- Usage pattern analysis
- Optimization suggestions

---

This plan provides a comprehensive roadmap for converting the legionbatctl utility to Go while addressing the unique requirements of working around the 60% conservation mode limit on the Lenovo Legion Slim 7 15ACH6.