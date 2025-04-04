#!/bin/bash
# Validate the Docker environment configuration for OpenPanel

set -eo pipefail

GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
RESET='\033[0m'

echo -e "${YELLOW}Validating OpenPanel environment...${RESET}"

# Check if Docker is installed and running
echo -n "Checking Docker installation: "
if command -v docker >/dev/null 2>&1; then
    if docker info >/dev/null 2>&1; then
        echo -e "${GREEN}OK${RESET}"
    else
        echo -e "${RED}Docker daemon is not running${RESET}"
        exit 1
    fi
else
    echo -e "${RED}Not installed${RESET}"
    exit 1
fi

# Check if Docker Compose is installed
echo -n "Checking Docker Compose installation: "
if command -v docker-compose >/dev/null 2>&1 || docker compose version >/dev/null 2>&1; then
    echo -e "${GREEN}OK${RESET}"
else
    echo -e "${RED}Not installed${RESET}"
    exit 1
fi

# Check for required ports availability
echo "Checking required ports availability:"
PORTS=(80 443 2083 2087 53 21)

for PORT in "${PORTS[@]}"; do
    echo -n "  Port $PORT: "
    if netstat -tuln | grep -q ":$PORT "; then
        echo -e "${RED}In use${RESET}"
    else
        echo -e "${GREEN}Available${RESET}"
    fi
done

# Check system resources
echo "Checking system resources:"

# Check CPU cores
CPU_CORES=$(grep -c ^processor /proc/cpuinfo)
echo -n "  CPU Cores: $CPU_CORES - "
if [ "$CPU_CORES" -lt 2 ]; then
    echo -e "${RED}Insufficient (minimum 2 recommended)${RESET}"
else
    echo -e "${GREEN}OK${RESET}"
fi

# Check memory
TOTAL_MEM=$(free -m | awk '/Mem:/ {print $2}')
echo -n "  Memory: ${TOTAL_MEM}MB - "
if [ "$TOTAL_MEM" -lt 2048 ]; then
    echo -e "${RED}Insufficient (minimum 2GB recommended)${RESET}"
else
    echo -e "${GREEN}OK${RESET}"
fi

# Check disk space
ROOT_DISK=$(df -h / | awk 'NR==2 {print $4}')
echo -n "  Disk Space: $ROOT_DISK available - "
ROOT_DISK_GB=$(df -BG / | awk 'NR==2 {print $4}' | sed 's/G//')
if [ "$ROOT_DISK_GB" -lt 20 ]; then
    echo -e "${RED}Insufficient (minimum 20GB recommended)${RESET}"
else
    echo -e "${GREEN}OK${RESET}"
fi

# Check for containerization/virtualization
echo -n "Checking virtualization: "
if grep -q -E '(vmx|svm)' /proc/cpuinfo; then
    echo -e "${GREEN}Virtualization extensions available${RESET}"
else
    echo -e "${YELLOW}No hardware virtualization detected${RESET}"
fi

# Check selinux status
echo -n "Checking SELinux status: "
if command -v getenforce >/dev/null 2>&1; then
    SELINUX=$(getenforce)
    if [ "$SELINUX" = "Enforcing" ]; then
        echo -e "${YELLOW}Enforcing - May need additional configuration${RESET}"
    else
        echo -e "${GREEN}$SELINUX${RESET}"
    fi
else
    echo -e "${GREEN}Not installed${RESET}"
fi

# Check if important Docker capabilities are available
echo -n "Checking Docker capabilities: "
if docker info | grep -q "AppArmor\|Seccomp"; then
    echo -e "${GREEN}Security profiles available${RESET}"
else
    echo -e "${YELLOW}Limited container security profiles${RESET}"
fi

echo -e "\n${GREEN}Environment validation complete!${RESET}"
exit 0
