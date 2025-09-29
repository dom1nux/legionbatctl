# Installation Guide for Lenovo Legion Battery Control

This guide provides detailed instructions for installing and configuring the Lenovo Legion Battery Control utility on your system.

## Prerequisites

- A Lenovo Legion laptop with conservation mode support
- Linux operating system
- Root/sudo privileges
- systemd (for automated mode)

## Checking Compatibility

Before installation, verify that your laptop supports the conservation mode feature:

```bash
# Check if the conservation mode file exists
ls /sys/bus/platform/drivers/ideapad_acpi/VPC2004:00/conservation_mode
```

If the file doesn't exist, try to find it with:

```bash
find /sys -name conservation_mode
```

If you find the file at a different path, you'll need to modify the scripts after installation to use the correct path.

## Installation Methods

### Method 1: Manual Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/legionbatctl.git
   cd legionbatctl
   ```

2. Install the scripts and service files:
   ```bash
   sudo cp src/legionbatctl /usr/local/bin/
   sudo cp src/legionbatctl-auto /usr/local/bin/
   sudo cp src/legionbatctl.service /etc/systemd/system/
   sudo cp src/legionbatctl.timer /etc/systemd/system/
   sudo chmod +x /usr/local/bin/legionbatctl
   sudo chmod +x /usr/local/bin/legionbatctl-auto
   ```

3. Install the man page:
   ```bash
   sudo mkdir -p /usr/local/share/man/man1
   sudo cp src/legionbatctl.1 /usr/local/share/man/man1/
   sudo mandb
   ```

4. Create the initial configuration file:
   ```bash
   sudo legionbatctl status
   ```
   This will automatically create the configuration file with default settings if it doesn't exist.

5. Enable and start the timer for automated mode:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable legionbatctl.timer
   sudo systemctl start legionbatctl.timer
   ```

### Method 2: Using the Installation Script

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/legionbatctl.git
   cd legionbatctl
   ```

2. Run the installation script:
   ```bash
   sudo ./install.sh
   ```

The installation script will:
- Copy all necessary files to the appropriate locations
- Set the correct permissions
- Create the initial configuration file
- Enable and start the systemd timer
- Automatically detect and use the correct conservation mode path for your laptop model

## Post-Installation Configuration

### Setting Your Preferred Charge Threshold

After installation, you can set your preferred battery charge threshold:

```bash
sudo legionbatctl set-threshold 80
```

Replace `80` with your desired percentage (between 20 and 100).

### Enabling Battery Management

Ensure battery management is enabled:

```bash
sudo legionbatctl enable
```

### Verifying Installation

1. Check the status of the battery management:
   ```bash
   sudo legionbatctl status
   ```

2. Verify that the systemd timer is running:
   ```bash
   sudo systemctl status legionbatctl.timer
   ```

## Customizing the Configuration

The configuration file is located at `/etc/legionbatctl.conf`. You can edit it directly:

```bash
sudo nano /etc/legionbatctl.conf
```

Default configuration:
```
CONSERVATION_ENABLED=1
CHARGE_THRESHOLD=80
```

After making changes, restart the timer:
```bash
sudo systemctl restart legionbatctl.timer
```

## Adapting to Different Laptop Models

If your conservation mode file is at a different path, you need to modify both scripts:

1. Edit the main script:
   ```bash
   sudo nano /usr/local/bin/legionbatctl
   ```
   Update the `CONS_MODE` variable with the correct path.

2. Edit the auto script:
   ```bash
   sudo nano /usr/local/bin/legionbatctl-auto
   ```
   Update the `CONS_MODE` variable with the correct path.

3. Restart the timer:
   ```bash
   sudo systemctl restart legionbatctl.timer
   ```

## Uninstallation

### Method 1: Using the Uninstallation Script

The easiest way to uninstall is to use the provided uninstallation script:

```bash
sudo ./uninstall.sh
```

The script will:
- Stop and disable the systemd timer and service
- Remove all installed files
- Ask if you want to remove the configuration file

### Method 2: Manual Uninstallation

If you prefer to uninstall manually:

1. Stop and disable the systemd timer:
   ```bash
   sudo systemctl stop legionbatctl.timer
   sudo systemctl disable legionbatctl.timer
   ```

2. Remove the service and timer files:
   ```bash
   sudo rm /etc/systemd/system/legionbatctl.service
   sudo rm /etc/systemd/system/legionbatctl.timer
   sudo systemctl daemon-reload
   ```

3. Remove the scripts:
   ```bash
   sudo rm /usr/local/bin/legionbatctl
   sudo rm /usr/local/bin/legionbatctl-auto
   ```

4. Remove the man page:
   ```bash
   sudo rm /usr/local/share/man/man1/legionbatctl.1
   sudo mandb
   ```

5. Remove the configuration file (optional):
   ```bash
   sudo rm /etc/legionbatctl.conf
   ```

## Troubleshooting

### Permission Issues

If you encounter permission issues:

```bash
# Ensure the scripts are executable
sudo chmod +x /usr/local/bin/legionbatctl
sudo chmod +x /usr/local/bin/legionbatctl-auto

# Ensure the configuration file has the correct permissions
sudo chmod 644 /etc/legionbatctl.conf
```

### Service Not Running

If the automated service isn't running:

```bash
# Check the service status
sudo systemctl status legionbatctl.service

# Check the timer status
sudo systemctl status legionbatctl.timer

# Check the journal for errors
sudo journalctl -u legionbatctl.service
