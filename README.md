# LegionBatCTL

A sophisticated battery management utility for Lenovo Legion laptops that extends battery lifespan through intelligent charge threshold control.

## Overview

LegionBatCTL is a Go-based battery management solution designed specifically for Lenovo Legion laptops. It solves a critical hardware limitation where the native conservation mode is fixed at 60% charge, allowing users to achieve higher charge limits (e.g., 80%) through intelligent threshold management.

> **Note**: This project was originally developed as a bash-based solution and completely rewritten in Go to provide better performance, maintainability, and architectural flexibility.

## The Problem: Hardware Limitations

### The Challenge

Lenovo Legion Slim 7 15ACH6 (2021) and similar models have a critical limitation:
- **Native conservation mode**: Fixed at 60% charge limit
- **User need**: Desire for higher charge limits (75-85%) for optimal battery health
- **Gap**: No built-in way to achieve custom charge thresholds

### The Solution Approach

Instead of accepting the 60% limitation, I developed an intelligent workaround:

1. **Monitor battery levels continuously** when on AC power
2. **Enable conservation mode at the desired threshold** (e.g., 80%)
3. **Disable conservation mode when below threshold** to allow charging
4. **Maintain state persistence** across reboots and service restarts

This approach effectively "hacks" the 60% hardware limit by using conservation mode as a switch rather than a fixed limit.

## Architecture

### Dual-Mode Design

LegionBatCTL implements a sophisticated single-binary architecture that operates in two modes:

```
┌─────────────────┐    Unix Socket    ┌─────────────────┐
│   CLI Client    │◄─────────────────►│   Daemon Mode   │
│  (User-facing)  │                   │  (Background)   │
└─────────────────┘                   └─────────────────┘
         │                                     │
         ▼                                     ▼
┌─────────────────┐                   ┌─────────────────┐
│  Command Input  │                   │   Battery       │
│  - status       │                   │   Monitoring    │
│  - enable       │                   │   - Adaptive    │
│  - disable      │                   │     intervals   │
│  - set-threshold│                   │   - Hardware     │
└─────────────────┘                   │     control     │
         │                             └─────────────────┘
         ▼                                     │
┌─────────────────┐                             ▼
│ Formatted Output│                   ┌─────────────────┐
│ - Human readable│                   │  State          │
│ - Success/error │                   │  Persistence    │
└─────────────────┘                   │  - JSON files   │
                                      │  - Atomic ops   │
                                      └─────────────────┘
```

### Core Components

1. **Protocol Package** (`internal/protocol/`)
   - Message types and validation for client-daemon communication
   - JSON-based request/response protocol over Unix socket
   - Status data structures for battery and daemon information

2. **State Management** (`internal/state/`)
   - Thread-safe state management with mutex protection
   - Atomic file operations for persistence across reboots
   - Support for conservation mode settings and threshold configuration

3. **Daemon Framework** (`internal/daemon/`)
   - Continuous battery monitoring with adaptive intervals (15s-2min)
   - Hardware-aware conservation mode control
   - Graceful shutdown and systemd integration
   - Request handlers for all CLI operations

4. **Client Package** (`internal/client/`)
   - Socket client with timeout and retry mechanisms
   - High-level command execution with result formatting
   - Connection health checking and error handling

5. **CLI Interface** (`internal/cli/`)
   - User-friendly command-line interface using Cobra
   - Commands: status, enable, disable, set-threshold
   - Formatted output for human readability

## Installation

### Prerequisites

- Lenovo Legion laptop with conservation mode support
- Linux operating system
- Go 1.22+ (for building from source) - can be installed via mise
- Root/sudo privileges
- systemd (for daemon mode)

### Quick Installation with Makefile

The project includes a simplified Makefile that works with mise-based Go installations:

```bash
# Clone the repository
git clone https://github.com/dom1nux/legionbatctl.git
cd legionbatctl

# Build binary (uses mise for Go)
make build

# Install and start daemon (requires root)
sudo make install

# Check status
make status
```

The Makefile separates build and installation phases:
- **Build**: Done as regular user using your mise Go installation
- **Install**: Done as root without requiring Go to be installed system-wide

### Manual Installation

If you prefer manual installation:

```bash
git clone https://github.com/dom1nux/legionbatctl.git
cd legionbatctl
go build -o legionbatctl ./cmd/legionbatctl
sudo cp legionbatctl /usr/bin/
sudo chmod +x /usr/bin/legionbatctl
sudo cp systemd/legionbatctl.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable legionbatctl.service
sudo systemctl start legionbatctl.service
```

### Verification

```bash
# Check daemon status
sudo systemctl status legionbatctl.service

# Check battery management status
legionbatctl status
```

## Usage

### CLI Commands

The `legionbatctl` command provides a comprehensive interface for battery management:

```bash
# Check current battery and daemon status
legionbatctl status

# Enable battery management with current threshold
legionbatctl enable

# Disable battery management (charge to 100%)
legionbatctl disable

# Set custom charge threshold (60-100%)
legionbatctl set-threshold 80

# Run in daemon mode (usually handled by systemd)
sudo legionbatctl daemon
```

### Status Output Example

```
Battery Management Status:
  Conservation Management: enabled
  Charge Threshold: 80%
  Current Mode: enabled
  Battery Level: 75%
  Conservation Mode: false
  Charging Status: charging
  Last Action: enable
  Daemon Uptime: 2h15m30s
  Hardware Supported: true
```

## Technical Deep Dive

### Adaptive Battery Monitoring

The daemon implements intelligent monitoring that adapts to system state:

- **Active charging**: 15-second intervals for responsiveness
- **Idle states**: 30-second intervals for efficiency
- **Stable states**: Up to 2-minute intervals to reduce overhead
- **AC power detection**: Only monitors when connected to AC power

### State Persistence

The daemon maintains state across reboots through:

- **Atomic file operations**: Prevents corruption during writes
- **JSON format**: Human-readable and easily editable
- **Backup mechanism**: Automatic backup of previous state
- **Validation**: Ensures state integrity on load

### Hardware Integration

Direct integration with Lenovo's conservation mode hardware interface:

```bash
# Conservation mode control file
/sys/bus/platform/drivers/ideapad_acpi/VPC2004:00/conservation_mode

# Values:
# 0 = Normal charging (disabled)
# 1 = Stop charging (enabled)
```

### Error Handling and Resilience

Comprehensive error handling throughout the system:

- **Socket communication**: Timeout and retry mechanisms
- **File operations**: Atomic writes with rollback capability
- **Hardware interaction**: Validation and graceful degradation
- **Daemon lifecycle**: Proper cleanup and resource management

## Makefile Commands

The simplified Makefile provides all essential operations:

```bash
# Build binary (as user, with mise)
make build

# Install and start service (as root)
sudo make install

# Check service and CLI status
make status

# View live daemon logs
make logs

# Restart daemon
sudo make restart

# Uninstall completely
sudo make uninstall

# Clean build artifacts
make clean

# Local testing
make dev          # Build and test CLI locally
make help         # Show all available commands
```

## Configuration

### Default Configuration

The daemon uses sensible defaults but can be customized:

- **Socket path**: `/var/run/legionbatctl.sock`
- **State file**: `/etc/legionbatctl.state`
- **PID file**: `/var/run/legionbatctl.pid`

### Threshold Validation

Hardware constraints require threshold validation:

- **Minimum**: 60% (hardware conservation mode limit)
- **Maximum**: 100% (full charge)
- **Recommended**: 75-85% for optimal battery health

### Hardware Detection Improvements

The daemon now uses AC adapter status for more reliable power detection:

```bash
# AC adapter detection
cat /sys/class/power_supply/ADP1/online

# Battery level
cat /sys/class/power_supply/BAT0/capacity

# Conservation mode status
cat /sys/bus/platform/drivers/ideapad_acpi/VPC2004:00/conservation_mode
```

This fixes issues where conservation mode causes the battery to report "Not charging" despite being connected to power.

## Development

### Project Structure

```
legionbatctl/
├── cmd/legionbatctl/          # Main entry point
├── internal/
│   ├── cli/                   # CLI interface and commands
│   ├── client/                # Socket client implementation
│   ├── daemon/                # Daemon framework and monitoring
│   ├── protocol/              # Communication protocol
│   └── state/                 # State management and persistence
├── systemd/                   # Systemd service files
└── pkg/                       # Public packages (if needed)
```

### Testing

Run the comprehensive test suite:

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/daemon -v
go test ./internal/client -v
go test ./internal/state -v
```

### Building

```bash
# Build for current platform
go build ./cmd/legionbatctl

# Build with version info
go build -ldflags "-X main.version=1.0.0" ./cmd/legionbatctl

# Build for release
CGO_ENABLED=0 go build -ldflags "-s -w" ./cmd/legionbatctl
```

## Troubleshooting

### Daemon Issues

```bash
# Check daemon status
sudo systemctl status legionbatctl.service

# View daemon logs
sudo journalctl -u legionbatctl.service -f

# Test daemon connectivity
legionbatctl status
```

### Hardware Compatibility

If conservation mode control fails:

```bash
# Find conservation mode file
find /sys -name conservation_mode 2>/dev/null

# Check permissions
ls -la /sys/bus/platform/drivers/ideapad_acpi/VPC2004:00/conservation_mode

# Test manual control
echo 1 | sudo tee /sys/bus/platform/drivers/ideapad_acpi/VPC2004:00/conservation_mode
echo 0 | sudo tee /sys/bus/platform/drivers/ideapad_acpi/VPC2004:00/conservation_mode
```

### Socket Issues

```bash
# Check if socket exists
ls -la /var/run/legionbatctl.sock

# Check permissions
sudo lsof /var/run/legionbatctl.sock

# Test socket connection
nc -U /var/run/legionbatctl.sock
```

## Performance

### Resource Usage

The daemon is designed for minimal resource impact:

- **Memory usage**: <10MB typical
- **CPU usage**: <0.1% when idle
- **Disk I/O**: Minimal, only on state changes
- **Network**: No external network connections

### Monitoring Performance

```bash
# Monitor resource usage
top -p $(cat /var/run/legionbatctl.pid)

# Check daemon uptime
legionbatctl status | grep "Daemon Uptime"

# Monitor state changes
watch -n 5 cat /etc/legionbatctl.state
```

## Contributing

### Development Guidelines

1. **Follow Go conventions**: Use standard Go formatting and idioms
2. **Test thoroughly**: Include unit tests for all new functionality
3. **Document changes**: Update README and code comments
4. **Error handling**: Use proper error wrapping and context
5. **Thread safety**: Ensure all shared state is properly protected

### Submitting Changes

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request with detailed description

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- **Lenovo**: For providing the conservation mode hardware interface
- **Go community**: For excellent tools and libraries
- **Systemd**: For robust service management
- **Cobra**: For powerful CLI framework

---

**Note**: This project was developed to solve a real-world hardware limitation and demonstrates practical problem-solving in systems programming. The architecture showcases modern Go development practices including concurrent programming, inter-process communication, and system integration.