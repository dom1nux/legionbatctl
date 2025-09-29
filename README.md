# Lenovo Legion Battery Control

A set of scripts to control battery charging behavior on Lenovo Legion laptops, helping to extend battery lifespan by limiting maximum charge.

## Overview

This project provides tools to manage the battery charging behavior of Lenovo Legion laptops by controlling the conservation mode feature. It allows you to:

- Set a maximum battery charge threshold (e.g., 80%)
- Enable/disable battery charge limiting
- Automatically maintain battery level within the configured threshold

By limiting the maximum charge level of your battery, you can significantly extend its lifespan, as keeping lithium-ion batteries at 100% charge for extended periods can cause faster degradation.

## Requirements

- Lenovo Legion laptop with conservation mode support
- Linux operating system
- Root/sudo privileges
- systemd (for automated mode)

## Installation

### Quick Installation

The easiest way to install is using the provided installation script:

```bash
git clone https://github.com/yourusername/legionbatctl.git
cd legionbatctl
sudo ./install.sh
```

### Manual Installation

For manual installation or more detailed instructions, see the [INSTALL.md](INSTALL.md) file.

## Usage

### Manual Control

The `legionbatctl` command provides several options for manually controlling battery charging:

```
sudo legionbatctl {status|toggle|enable|disable|set-threshold <percentage>}
```

Commands:
- `status` - Show current battery management and conservation mode status
- `toggle` - Toggle between threshold limit and 100% charging
- `enable` - Enable battery management (limit to configured threshold)
- `disable` - Disable battery management (allow charging to 100%)
- `set-threshold <percentage>` - Set battery charge threshold (20-100)

Examples:
```
# Check current status
sudo legionbatctl status

# Set charge threshold to 80%
sudo legionbatctl set-threshold 80

# Enable battery management
sudo legionbatctl enable

# Disable battery management
sudo legionbatctl disable
```

### Automated Mode

The automated mode runs in the background and automatically manages your battery charging based on the configured threshold:

- When battery level reaches the threshold, charging stops
- When battery level drops below the threshold, charging resumes
- The timer runs continuously to check the power state
- The service only activates when AC power is connected
- This ensures proper handling of charge thresholds regardless of when AC power is connected

The automated mode is controlled by systemd:

```
# Check status of the timer
sudo systemctl status legionbatctl.timer

# Temporarily stop automated mode
sudo systemctl stop legionbatctl.timer

# Restart automated mode
sudo systemctl start legionbatctl.timer

# Disable automated mode permanently
sudo systemctl disable legionbatctl.timer
```

## Configuration

The configuration file is located at `/etc/legionbatctl.conf` and contains the following settings:

```
CONSERVATION_ENABLED=1
CHARGE_THRESHOLD=80
```

- `CONSERVATION_ENABLED`: Set to 1 to enable battery management, 0 to disable
- `CHARGE_THRESHOLD`: The maximum battery charge percentage (20-100)

You can edit this file directly or use the `legionbatctl` commands to modify these settings.

## How It Works

### Conservation Mode

Lenovo Legion laptops provide a feature called "Conservation Mode" that can be enabled or disabled by writing to a system file:

```
/sys/bus/platform/drivers/ideapad_acpi/VPC2004:00/conservation_mode
```

When conservation mode is enabled (value 1), the laptop stops charging the battery even when plugged in. When disabled (value 0), normal charging behavior resumes.

### Manual Script

The `legionbatctl` script provides a user-friendly interface to control conservation mode and configure the desired charge threshold.

### Automated Script

The `legionbatctl-auto` script:
1. Reads the current battery level
2. Checks if it's above or below the configured threshold
3. Enables conservation mode if the battery level is at or above the threshold
4. Disables conservation mode if the battery level is below the threshold

This script is run every minute (when AC power is connected) by the systemd timer, ensuring your battery stays within the configured limits.

## Troubleshooting

### Script reports "Conservation mode control file not found"

This error occurs if your laptop model uses a different path for the conservation mode control. You may need to modify the `CONS_MODE` variable in both scripts to match your system.

To find the correct path, try:
```
find /sys -name conservation_mode
```

### Battery still charges to 100%

1. Check if the automated mode is running:
   ```
   sudo systemctl status legionbatctl.timer
   ```

2. Verify that conservation mode is working:
   ```
   cat /sys/bus/platform/drivers/ideapad_acpi/VPC2004:00/conservation_mode
   ```
   This should return 1 when enabled.

3. Check your configuration:
   ```
   cat /etc/legionbatctl.conf
   ```
   Ensure `CONSERVATION_ENABLED=1` and `CHARGE_THRESHOLD` is set to your desired value.

### Changes to configuration don't take effect

After manually editing the configuration file, you may need to restart the timer:
```
sudo systemctl restart legionbatctl.timer
```

## Uninstallation

To uninstall the utility, you can use the provided uninstallation script:

```bash
sudo ./uninstall.sh
```

For manual uninstallation or more detailed instructions, see the [INSTALL.md](INSTALL.md) file.

## License

[Insert your license information here]

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
