#!/bin/bash
# Uninstall OpenPanel
# This script stops services, removes docker containers, directories, and cleanup configuration files.
# Run as root.

if [ "$(id -u)" != "0" ]; then
    echo "Please run this script as root."
    exit 1
fi

read -p "WARNING: This will completely uninstall OpenPanel and remove all its data. Continue? (yes/no): " CONFIRM
if [ "$CONFIRM" != "yes" ]; then
    echo "Uninstall aborted."
    exit 0
fi

echo "Stopping OpenPanel systemd services..."
if command -v systemctl >/dev/null 2>&1; then
    systemctl stop admin 2>/dev/null
    systemctl stop watcher 2>/dev/null
    systemctl stop floatingip 2>/dev/null
else
    echo "Warning: systemctl not found. Skipping service stop commands."
fi

echo "Stopping docker containers..."
if [ -f /root/docker-compose.yml ]; then
    docker compose -f /root/docker-compose.yml down --volumes
fi

echo "Removing installed directories and configuration files..."
rm -rf /etc/openpanel
rm -rf /usr/local/admin
rm -rf /usr/local/opencli

echo "Removing systemd service files..."
if command -v systemctl >/dev/null 2>&1; then
    rm -f /etc/systemd/system/admin.service
    rm -f /etc/systemd/system/watcher.service
    rm -f /etc/systemd/system/floatingip.service
    systemctl daemon-reload
else
    echo "Warning: systemctl not found. Skipping removal of systemd service files."
fi

echo "Checking for and removing swap file created by OpenPanel..."
if [ -f /swapfile ]; then
    swapoff /swapfile
    sed -i '/\/swapfile/d' /etc/fstab
    rm -f /swapfile
fi

echo "Removing docker volume for MySQL (if exists)..."
docker volume rm root_openadmin_mysql 2>/dev/null

echo "Removing install lock file..."
rm -f /root/openpanel_install.lock

echo "OpenPanel has been uninstalled successfully. A system reboot may be required."

# Analyzer: Check uninstall outcome
analyze_uninstall() {
    echo "Analyzing uninstall results..."
    [ -d /etc/openpanel ] && echo "[FAIL] Directory /etc/openpanel still exists." || echo "[OK] Directory /etc/openpanel removed."
    [ -d /usr/local/admin ] && echo "[FAIL] Directory /usr/local/admin still exists." || echo "[OK] Directory /usr/local/admin removed."
    [ -d /usr/local/opencli ] && echo "[FAIL] Directory /usr/local/opencli still exists." || echo "[OK] Directory /usr/local/opencli removed."
    for svc in admin watcher floatingip; do
        [ -f /etc/systemd/system/${svc}.service ] && echo "[FAIL] Systemd service file ${svc}.service still exists." || echo "[OK] Systemd service file ${svc}.service removed."
    done
    [ -f /swapfile ] && echo "[FAIL] Swap file /swapfile still exists." || echo "[OK] Swap file /swapfile removed."
    if docker volume ls | awk '{print $2}' | grep -qw root_openadmin_mysql; then
        echo "[FAIL] Docker volume root_openadmin_mysql still exists."
    else
        echo "[OK] Docker volume root_openadmin_mysql removed."
    fi
    [ -f /root/openpanel_install.lock ] && echo "[FAIL] Install lock file still exists." || echo "[OK] Install lock file removed."
}
analyze_uninstall
