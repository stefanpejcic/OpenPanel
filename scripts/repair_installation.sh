#!/bin/bash
# OpenPanel Installation Repair Tool
# This script diagnoses and fixes common installation issues

GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
RESET='\033[0m'

echo -e "${YELLOW}OpenPanel Installation Repair Tool${RESET}"
echo "This script will diagnose and attempt to fix common installation issues."

# Check if running as root
if [ "$(id -u)" != "0" ]; then
    echo -e "${RED}Error: This script must be run as root${RESET}"
    exit 1
fi

# Function to check for common issues
check_system() {
    echo "Checking system status..."
    
    # Check for Docker
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}Docker is not installed. This is required for OpenPanel.${RESET}"
        echo "Attempting to install Docker..."
        if command -v apt-get &> /dev/null; then
            apt-get update && apt-get install -y docker.io
        elif command -v dnf &> /dev/null; then
            dnf install -y docker
        elif command -v yum &> /dev/null; then
            yum install -y docker
        else
            echo -e "${RED}Could not determine package manager. Please install Docker manually.${RESET}"
        fi
    else
        echo -e "${GREEN}Docker is installed.${RESET}"
    fi
    
    # Check Docker service
    if systemctl is-active --quiet docker; then
        echo -e "${GREEN}Docker service is running.${RESET}"
    else
        echo -e "${RED}Docker service is not running.${RESET}"
        echo "Starting Docker service..."
        systemctl start docker
        systemctl enable docker
    fi
    
    # Check for OpenAdmin installation
    if [ -d "/usr/local/admin" ]; then
        echo -e "${GREEN}OpenAdmin directory exists.${RESET}"
    else
        echo -e "${RED}OpenAdmin directory not found. OpenPanel may not be properly installed.${RESET}"
    fi
    
    # Check for config files
    if [ -d "/etc/openpanel" ]; then
        echo -e "${GREEN}OpenPanel configuration directory exists.${RESET}"
    else
        echo -e "${RED}OpenPanel configuration directory not found.${RESET}"
        echo "Creating configuration directory structure..."
        mkdir -p /etc/openpanel
    fi
    
    # Check for opencli
    if command -v opencli &> /dev/null; then
        echo -e "${GREEN}OpenCLI is installed.${RESET}"
    else
        echo -e "${RED}OpenCLI is not installed or not in PATH.${RESET}"
        echo "Checking for OpenCLI directory..."
        if [ -d "/usr/local/opencli" ]; then
            echo "Creating symlink for OpenCLI..."
            chmod +x -R /usr/local/opencli
            ln -sf /usr/local/opencli/opencli /usr/local/bin/opencli
        else
            echo -e "${YELLOW}OpenCLI directory not found. Cannot fix automatically.${RESET}"
        fi
    fi
}

# Function to fix common issues
fix_common_issues() {
    echo "Attempting to fix common issues..."
    
    # Fix unbound variable issues in install.sh
    if [ -f "/OpenPanel/install.sh" ]; then
        echo "Checking install.sh script for issues..."
        if grep -q "UNINSTALL=" "/OpenPanel/install.sh"; then
            echo -e "${GREEN}UNINSTALL variable is defined in install.sh${RESET}"
        else
            echo "Adding UNINSTALL variable initialization to install.sh..."
            sed -i '/^# Defaults for environment variables/a UNINSTALL=false' "/OpenPanel/install.sh"
        fi
    fi
    
    # Restart key services
    echo "Restarting key services..."
    if systemctl is-active --quiet admin; then
        echo "Restarting OpenAdmin service..."
        systemctl restart admin
    fi
}

# Function to check containers
check_containers() {
    echo "Checking Docker containers..."
    
    # Check MySQL container
    if docker ps | grep -q "openpanel_mysql"; then
        echo -e "${GREEN}MySQL container is running.${RESET}"
    else
        echo -e "${RED}MySQL container is not running.${RESET}"
        echo "Attempting to start MySQL container..."
        cd /root && docker compose up -d openpanel_mysql
    fi
    
    # Check for other essential containers
    essential_containers=("bind9" "caddy")
    for container in "${essential_containers[@]}"; do
        if docker ps | grep -q "$container"; then
            echo -e "${GREEN}$container container is running.${RESET}"
        else
            echo -e "${YELLOW}$container container is not running.${RESET}"
        fi
    done
}

# Run diagnostics
check_system
fix_common_issues
check_containers

# Final message
echo -e "\n${GREEN}Repair process completed.${RESET}"
echo "If issues persist, please run the installation script with repair flag:"
echo "bash install.sh --repair"
echo "Or contact OpenPanel support for assistance."
exit 0
