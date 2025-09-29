#!/bin/bash

# Lenovo Legion Battery Control - Uninstallation Script

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root"
    echo "Please run with: sudo $0"
    exit 1
fi

echo "Uninstalling Lenovo Legion Battery Control..."

# Stop and disable systemd timer and service
echo "Stopping and disabling systemd timer and service..."
systemctl stop legionbatctl.timer 2>/dev/null
systemctl disable legionbatctl.timer 2>/dev/null
systemctl stop legionbatctl.service 2>/dev/null
systemctl disable legionbatctl.service 2>/dev/null

# Remove systemd service files
echo "Removing systemd service files..."
rm -f /etc/systemd/system/legionbatctl.service
rm -f /etc/systemd/system/legionbatctl.timer
systemctl daemon-reload

# Remove scripts
echo "Removing scripts..."
rm -f /usr/local/bin/legionbatctl
rm -f /usr/local/bin/legionbatctl-auto

# Remove man page
echo "Removing man page..."
rm -f /usr/local/share/man/man1/legionbatctl.1
mandb > /dev/null 2>&1

# Ask about configuration file
echo
echo "Do you want to remove the configuration file (/etc/legionbatctl.conf)?"
echo "This will delete your saved settings."
read -p "Remove configuration file? [y/N]: " remove_config

if [[ "$remove_config" =~ ^[Yy]$ ]]; then
    echo "Removing configuration file..."
    rm -f /etc/legionbatctl.conf
    echo "Configuration file removed."
else
    echo "Configuration file preserved."
fi

echo
echo "Uninstallation complete!"
echo
echo "The Lenovo Legion Battery Control utility has been removed from your system."
echo "If you want to reinstall it in the future, you can run the install.sh script again."
