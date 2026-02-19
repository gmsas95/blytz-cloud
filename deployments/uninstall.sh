#!/bin/bash
set -e

echo "Blytz Uninstallation Script"
echo "============================"

if [ "$EUID" -ne 0 ]; then 
    echo "Please run as root (use sudo)"
    exit 1
fi

echo "Stopping blytz service..."
systemctl stop blytz 2>/dev/null || true
systemctl disable blytz 2>/dev/null || true

echo "Removing systemd service..."
rm -f /etc/systemd/system/blytz.service
systemctl daemon-reload

echo "Removing blytz user..."
userdel blytz 2>/dev/null || true

echo ""
echo "Blytz has been uninstalled."
echo ""
echo "The following data has NOT been removed:"
echo "  - /opt/blytz/ (contains customer data and database)"
echo ""
echo "To remove all data, run: sudo rm -rf /opt/blytz"
