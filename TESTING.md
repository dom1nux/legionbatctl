# Testing Guide

This guide provides comprehensive instructions for testing the LegionBatCTL daemon safely on your machine.

## Quick Start

The easiest way to test is using the provided test script:

```bash
# Build and start daemon in background
./test-daemon.sh build
./test-daemon.sh background

# Check status
./test-daemon.sh status

# Test CLI commands
./test-daemon.sh cli

# View logs
./test-daemon.sh logs

# Stop daemon
./test-daemon.sh stop
```

## Testing Environment

The test script uses isolated paths to avoid conflicts with any production installation:

- **Socket**: `/tmp/legionbatctl-test.sock`
- **State**: `/tmp/legionbatctl-test.state`
- **PID**: `/tmp/legionbatctl-test.pid`
- **Logs**: `./test-daemon.log`

## Test Script Commands

| Command | Description |
|---------|-------------|
| `./test-daemon.sh build` | Build the binary |
| `./test-daemon.sh foreground` | Run daemon in foreground (for debugging) |
| `./test-daemon.sh background` | Start daemon in background |
| `./test-daemon.sh stop` | Stop background daemon |
| `./test-daemon.sh status` | Show daemon status |
| `./test-daemon.sh logs` | Show live logs |
| `./test-daemon.sh cli` | Test CLI commands |
| `./test-daemon.sh cleanup` | Clean up test artifacts |
| `./test-daemon.sh install-service` | Install systemd test service |
| `./test-daemon.sh help` | Show help |

## Manual Testing

If you prefer to test manually without the script:

### 1. Environment Variables

Set environment variables for testing:

```bash
export SOCKET_PATH="/tmp/legionbatctl-test.sock"
export STATE_PATH="/tmp/legionbatctl-test.state"
```

### 2. Start Daemon

```bash
# Build
go build -o legionbatctl ./cmd/legionbatctl

# Start in foreground (for debugging)
./legionbatctl daemon

# Or start in background
nohup ./legionbatctl daemon > test-daemon.log 2>&1 &
```

### 3. Test CLI Commands

```bash
# Status
SOCKET_PATH="/tmp/legionbatctl-test.sock" ./legionbatctl status

# Set threshold
SOCKET_PATH="/tmp/legionbatctl-test.sock" ./legionbatctl set-threshold 80

# Enable management
SOCKET_PATH="/tmp/legionbatctl-test.sock" ./legionbatctl enable

# Disable management
SOCKET_PATH="/tmp/legionbatctl-test.sock" ./legionbatctl disable
```

## Systemd Service Testing

For more realistic testing, you can install the test systemd service:

```bash
# Install test service
./test-daemon.sh install-service

# Start the service
sudo systemctl start legionbatctl-test.service

# Check status
sudo systemctl status legionbatctl-test.service

# View logs
sudo journalctl -u legionbatctl-test.service -f

# Stop service
sudo systemctl stop legionbatctl-test.service
```

## Troubleshooting

### Daemon Fails to Start

1. **Check logs**:
   ```bash
   ./test-daemon.sh logs
   # or
   cat test-daemon.log
   ```

2. **Common issues**:
   - **Permission denied**: Make sure you have permissions to write to `/tmp/`
   - **Socket in use**: Run `./test-daemon.sh cleanup` first
   - **Binary not found**: Run `./test-daemon.sh build`

3. **Hardware compatibility**:
   ```bash
   # Check if conservation mode file exists
   ls -la /sys/bus/platform/drivers/ideapad_acpi/VPC2004:00/conservation_mode

   # Check if you can write to it
   echo 1 | sudo tee /sys/bus/platform/drivers/ideapad_acpi/VPC2004:00/conservation_mode
   ```

### CLI Commands Fail

1. **Check if daemon is running**:
   ```bash
   ./test-daemon.sh status
   ```

2. **Check socket connectivity**:
   ```bash
   # Test socket connection
   nc -U /tmp/legionbatctl-test.sock
   ```

3. **Verify environment variables**:
   ```bash
   echo $SOCKET_PATH
   echo $STATE_PATH
   ```

### Permission Issues

If you encounter permission errors, you might need to run some commands with sudo:

```bash
# Test hardware access
sudo ./legionbatctl daemon

# Or check hardware manually
sudo cat /sys/bus/platform/drivers/ideapad_acpi/VPC2004:00/conservation_mode
```

## Test Scenarios

### Basic Functionality

1. **Start daemon**: `./test-daemon.sh background`
2. **Check status**: `./test-daemon.sh status`
3. **Set threshold**: `SOCKET_PATH="/tmp/legionbatctl-test.sock" ./legionbatctl set-threshold 80`
4. **Enable management**: `SOCKET_PATH="/tmp/legionbatctl-test.sock" ./legionbatctl enable`
5. **Check status again**: `./test-daemon.sh status`

### Error Handling

1. **Test invalid threshold**:
   ```bash
   SOCKET_PATH="/tmp/legionbatctl-test.sock" ./legionbatctl set-threshold 50
   ```

2. **Test with daemon stopped**:
   ```bash
   ./test-daemon.sh stop
   SOCKET_PATH="/tmp/legionbatctl-test.sock" ./legionbatctl status
   ```

### System Integration

1. **Install systemd service**: `./test-daemon.sh install-service`
2. **Start service**: `sudo systemctl start legionbatctl-test.service`
3. **Check logs**: `sudo journalctl -u legionbatctl-test.service -f`
4. **Test CLI**: `./legionbatctl status` (should connect to running daemon)

## Collecting Debug Information

If you encounter issues, please collect the following information:

1. **System information**:
   ```bash
   uname -a
   cat /etc/os-release
   ```

2. **Hardware information**:
   ```bash
   ls -la /sys/bus/platform/drivers/ideapad_acpi/VPC2004:00/
   cat /sys/bus/platform/drivers/ideapad_acpi/VPC2004:00/conservation_mode
   ```

3. **Daemon logs**:
   ```bash
   cat test-daemon.log
   ```

4. **Test output**:
   ```bash
   ./test-daemon.sh status > test-output.txt 2>&1
   ```

5. **Systemd logs** (if using service):
   ```bash
   sudo journalctl -u legionbatctl-test.service --since "1 hour ago" > systemd-log.txt
   ```

## Cleanup

After testing, clean up all artifacts:

```bash
# Stop daemon
./test-daemon.sh stop

# Clean up files
./test-daemon.sh cleanup

# Remove systemd test service
sudo systemctl stop legionbatctl-test.service
sudo systemctl disable legionbatctl-test.service
sudo rm /etc/systemd/system/legionbatctl-test.service
sudo systemctl daemon-reload
```

## Contributing Test Results

When reporting issues or contributing, please include:

1. **Steps to reproduce**
2. **Expected vs actual behavior**
3. **Relevant log output**
4. **System and hardware information**
5. **Any error messages**

This helps ensure that issues can be properly diagnosed and resolved.