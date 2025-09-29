#!/bin/bash

# Lenovo Legion Battery Control - Installation Script

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root"
    echo "Please run with: sudo $0"
    exit 1
fi

echo "Installing Lenovo Legion Battery Control..."

# Check if directories exist and create them if they don't
if [ ! -d "/usr/local/bin" ]; then
    echo "Creating /usr/local/bin directory..."
    mkdir -p /usr/local/bin
fi

if [ ! -d "/usr/local/share/man/man1" ]; then
    echo "Creating /usr/local/share/man/man1 directory..."
    mkdir -p /usr/local/share/man/man1
fi

# Install scripts
echo "Installing scripts..."
cp src/legionbatctl /usr/local/bin/
cp src/legionbatctl-auto /usr/local/bin/
chmod +x /usr/local/bin/legionbatctl
chmod +x /usr/local/bin/legionbatctl-auto

# Install systemd service files
echo "Installing systemd service files..."
cp src/legionbatctl.service /etc/systemd/system/
cp src/legionbatctl.timer /etc/systemd/system/

# Install man page
echo "Installing man page..."
cp src/legionbatctl.1 /usr/local/share/man/man1/
mandb > /dev/null 2>&1

# Check for conservation mode file
CONS_MODE="/sys/bus/platform/drivers/ideapad_acpi/VPC2004:00/conservation_mode"
if [ ! -f "$CONS_MODE" ]; then
    echo "Warning: Conservation mode file not found at default path: $CONS_MODE"
    echo "Searching for conservation mode file..."
    
    FOUND_PATH=$(find /sys -name conservation_mode 2>/dev/null | head -n 1)
    
    if [ -n "$FOUND_PATH" ]; then
        echo "Found conservation mode file at: $FOUND_PATH"
        echo "Updating scripts to use this path..."
        
        # Update the path in the scripts
        sed -i "s|CONS_MODE=\"/sys/bus/platform/drivers/ideapad_acpi/VPC2004:00/conservation_mode\"|CONS_MODE=\"$FOUND_PATH\"|g" /usr/local/bin/legionbatctl
        sed -i "s|CONS_MODE=\"/sys/bus/platform/drivers/ideapad_acpi/VPC2004:00/conservation_mode\"|CONS_MODE=\"$FOUND_PATH\"|g" /usr/local/bin/legionbatctl-auto
    else
        echo "Warning: Could not find conservation mode file in /sys"
        echo "Your laptop model may not be compatible with this utility"
        echo "Installation will continue, but the utility may not work correctly"
    fi
fi

# Create initial configuration
echo "Creating initial configuration..."
if [ ! -f "/etc/legionbatctl.conf" ]; then
    cat > "/etc/legionbatctl.conf" << EOF
CONSERVATION_ENABLED=1
CHARGE_THRESHOLD=80
EOF
    chmod 644 "/etc/legionbatctl.conf"
fi

# Enable and start systemd timer
echo "Enabling and starting systemd timer..."
systemctl daemon-reload
systemctl enable legionbatctl.timer
systemctl start legionbatctl.timer

echo "Installation complete!"
echo
echo "You can now use the legionbatctl command to control battery charging."
echo "For example: legionbatctl status"
echo
echo "The automated battery management is now active and will maintain"
echo "your battery at the configured threshold (default: 80%)."
echo
echo "For more information, see the man page: man legionbatctl"
echo "Or read the documentation in README.md and INSTALL.md"
