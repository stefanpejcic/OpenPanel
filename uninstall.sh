#!/bin/bash
set -euo pipefail

# Constants
GREEN='\033[0;32m'
RED='\033[0;31m'
RESET='\033[0m'

# Functions
remove_files_and_directories() {
    echo "Removing OpenPanel files and directories..."
    rm -rf /etc/openpanel /usr/local/admin /usr/local/opencli /var/log/openpanel || {
        echo -e "${RED}Failed to remove some files or directories.${RESET}"
        exit 1
    }
    echo -e "${GREEN}Files and directories removed successfully.${RESET}"
}

stop_and_disable_services() {
    echo "Stopping and disabling OpenPanel services..."
    systemctl stop admin watcher || true
    systemctl disable admin watcher || true
    echo -e "${GREEN}Services stopped and disabled.${RESET}"
}

remove_systemd_units() {
    echo "Removing systemd unit files..."
    rm -f /etc/systemd/system/admin.service /etc/systemd/system/watcher.service || true
    systemctl daemon-reload
    echo -e "${GREEN}Systemd unit files removed.${RESET}"
}

remove_cron_jobs() {
    echo "Removing cron jobs..."
    rm -f /etc/cron.d/openpanel || true
    echo -e "${GREEN}Cron jobs removed.${RESET}"
}

# Main execution
echo "Starting OpenPanel uninstallation..."
stop_and_disable_services
remove_files_and_directories
remove_systemd_units
remove_cron_jobs
echo -e "${GREEN}OpenPanel uninstallation completed successfully.${RESET}"
