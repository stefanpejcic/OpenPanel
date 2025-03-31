#!/bin/bash
################################################################################
#
# Uninstall OpenPanel
# This script removes OpenPanel, its services, and related configurations.
#
# Supported OS:            Ubuntu, Debian, AlmaLinux, RockyLinux, CentOS
# Supported Architecture:  x86_64, aarch64
#
# Usage:                   bash unstall.sh
# Author:                  Adapted from install.sh
# Created:                 30.03.2025
#
################################################################################

# ======================================================================
# Constants
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
RESET='\033[0m'

# ======================================================================
# Helper Functions

print_header() {
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
    echo -e "   ____                         _____                      _  "
    echo -e "  / __ \                       |  __ \                    | | "
    echo -e " | |  | | _ __    ___  _ __    | |__) | __ _  _ __    ___ | | "
    echo -e " | |  | || '_ \  / _ \| '_ \   |  ___/ / _\" || '_ \ / _  \| | "
    echo -e " | |__| || |_) ||  __/| | | |  | |    | (_| || | | ||  __/| | "
    echo -e "  \____/ | .__/  \___||_| |_|  |_|     \__,_||_| |_| \___||_| "
    echo -e "         | |                                                  "
    echo -e "         |_|                                   ${RED}Uninstalling OpenPanel${RESET} "
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
}

radovan() {
    echo -e "${RED}UNINSTALLATION FAILED${RESET}"
    echo ""
    echo -e "Error: $1" >&2
    exit 1
}

# ======================================================================
# Uninstallation Steps

stop_and_remove_services() {
    echo "Stopping and disabling OpenPanel services..."
    systemctl stop admin watcher floatingip || true
    systemctl disable admin watcher floatingip || true
    rm -f /etc/systemd/system/admin.service
    rm -f /etc/systemd/system/watcher.service
    rm -f /etc/systemd/system/floatingip.service
    systemctl daemon-reload
    echo -e "[${GREEN} OK ${RESET}] OpenPanel services stopped and removed."
}

remove_docker_containers() {
    echo "Removing Docker containers and volumes..."
    docker compose -f /root/docker-compose.yml down || true
    docker volume rm root_openadmin_mysql || true
    echo -e "[${GREEN} OK ${RESET}] Docker containers and volumes removed."
}

remove_files_and_directories() {
    echo "Removing OpenPanel files and directories..."
    rm -rf /etc/openpanel
    rm -rf /usr/local/admin
    rm -rf /usr/local/opencli
    rm -f /usr/local/bin/opencli
    rm -f /etc/profile.d/welcome.sh
    echo -e "[${GREEN} OK ${RESET}] OpenPanel files and directories removed."
}

remove_firewall_rules() {
    echo "Removing firewall rules..."
    if command -v csf > /dev/null 2>&1; then
        csf -f || true
        systemctl stop csf || true
        systemctl disable csf || true
        rm -rf /etc/csf || true
        echo -e "[${GREEN} OK ${RESET}] ConfigServer Firewall (CSF) removed."
    elif command -v ufw > /dev/null 2>&1; then
        ufw disable || true
        echo -e "[${GREEN} OK ${RESET}] Uncomplicated Firewall (UFW) disabled."
    fi
}

remove_swap() {
    echo "Removing swap file..."
    if grep -q '/swapfile' /etc/fstab; then
        sed -i '/\/swapfile/d' /etc/fstab
        swapoff /swapfile || true
        rm -f /swapfile
        echo -e "[${GREEN} OK ${RESET}] Swap file removed."
    else
        echo "No swap file found. Skipping."
    fi
}

remove_cronjobs() {
    echo "Removing OpenPanel cronjobs..."
    rm -f /etc/cron.d/openpanel
    echo -e "[${GREEN} OK ${RESET}] Cronjobs removed."
}

remove_logrotate_configs() {
    echo "Removing logrotate configurations..."
    rm -f /etc/logrotate.d/openpanel
    rm -f /etc/logrotate.d/syslog
    echo -e "[${GREEN} OK ${RESET}] Logrotate configurations removed."
}

restore_fstab() {
    echo "Restoring /etc/fstab to remove quotas..."
    sed -i 's/,usrquota,grpquota//' /etc/fstab
    mount -o remount / || true
    quotaoff -a || true
    echo -e "[${GREEN} OK ${RESET}] Quotas removed from /etc/fstab."
}

remove_packages() {
    echo "Removing installed packages..."
    if command -v apt-get > /dev/null 2>&1; then
        apt-get purge -y docker.io sqlite3 python3.12 quota quotatool || true
        apt-get autoremove -y || true
    elif command -v dnf > /dev/null 2>&1; then
        dnf remove -y docker-ce sqlite python3.12 quota quotatool || true
    elif command -v yum > /dev/null 2>&1; then
        yum remove -y docker-ce sqlite python3.12 quota quotatool || true
    fi
    echo -e "[${GREEN} OK ${RESET}] Packages removed."
}

final_message() {
    echo ""
    echo -e "${GREEN}OpenPanel has been successfully uninstalled.${RESET}"
    echo "If you encounter any issues, please visit:"
    echo "  - Discord: https://discord.openpanel.com/"
    echo "  - Forums: https://community.openpanel.org/"
    echo "  - GitHub: https://github.com/stefanpejcic/OpenPanel/"
    echo ""
}

# ======================================================================
# Main Program

print_header
stop_and_remove_services
remove_docker_containers
remove_files_and_directories
remove_firewall_rules
remove_swap
remove_cronjobs
remove_logrotate_configs
restore_fstab
remove_packages
final_message
