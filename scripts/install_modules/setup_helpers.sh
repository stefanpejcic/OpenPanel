#!/bin/bash
# Helper functions for the OpenPanel installer

GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
RESET='\033[0m'

# Display error and exit
error_exit() {
    echo -e "${RED}INSTALLATION FAILED${RESET}"
    echo ""
    echo -e "Error: $1" >&2
    exit 1
}

# Debug logging with timestamp
debug_log() {
    local timestamp
    timestamp=$(date +'%Y-%m-%d %H:%M:%S')
    
    if [ "$DEBUG" = true ]; then
        # Show both on terminal and log file
        echo "[$timestamp] $*" | tee -a "$LOG_FILE"
        "$@" 2>&1 | tee -a "$LOG_FILE"
    else
        # No terminal output, only log file
        echo "[$timestamp] COMMAND: $*" >> "$LOG_FILE"
        "$@" > /dev/null 2>&1
    fi
}

# Progress bar functions
setup_progress_bar() {
    local PROGRESS_BAR_URL="https://raw.githubusercontent.com/pollev/bash_progress_bar/master/progress_bar.sh"
    local PROGRESS_BAR_FILE="progress_bar.sh"
    
    # Check if wget is available
    if command -v wget &> /dev/null; then
        wget "$PROGRESS_BAR_URL" -O "$PROGRESS_BAR_FILE" > /dev/null 2>&1
    # If wget is not available, check if curl is available
    elif command -v curl &> /dev/null; then
        curl -s "$PROGRESS_BAR_URL" -o "$PROGRESS_BAR_FILE" > /dev/null 2>&1
    else
        error_exit "Neither wget nor curl is available. Please install one of them to proceed."
    fi
    
    if [ ! -f "$PROGRESS_BAR_FILE" ]; then
        error_exit "Failed to download progress_bar.sh - Github is not reachable by your server: https://raw.githubusercontent.com"
    fi
    
    source "$PROGRESS_BAR_FILE"
}

# Print logo and header
print_header() {
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
    echo -e "   ____                         _____                      _  "
    echo -e "  / __ \                       |  __ \                    | | "
    echo -e " | |  | | _ __    ___  _ __    | |__) | __ _  _ __    ___ | | "
    echo -e " | |  | || '_ \  / _ \| '_ \   |  ___/ / _\" || '_ \ / _  \| | "
    echo -e " | |  | || |_) ||  __/| | | |  | |    | (_| || | | ||  __/| | "
    echo -e "  \____/ | .__/  \___||_| |_|  |_|     \__,_||_| |_| \___||_| "
    echo -e "         | |                                                  "
    echo -e "         |_|                                   version: ${GREEN}$PANEL_VERSION${RESET} "
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
}

# Check if running as root
check_root() {
    if [ "$(id -u)" != "0" ]; then
        error_exit "You must be root to execute this script."
    fi
}

# Print space and line
print_space_and_line() {
    echo " "
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
    echo " "
}

# Network detection uses multiple fallback mechanisms
current_ip=$(curl --silent --max-time 2 -4 $IP_SERVER_1 || \
             wget --timeout=2 -qO- $IP_SERVER_2 || \
             curl --silent --max-time 2 -4 $IP_SERVER_3)

# Managing services through standardized functions
start_service() {
    if [ "$1" == "on" ]; then
        echo "Starting $2..."
        service "$2" start
    fi
}
