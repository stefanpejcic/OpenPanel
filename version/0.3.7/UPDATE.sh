#!/bin/bash

# Update from OpenPanel 0.3.6 to 0.3.7

# new version
NEW_PANEL_VERSION="0.3.7"
PREVIOUS_VERSION="0.3.6"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# Directories
OPENADMIN_DIR="/usr/local/admin/"
OPENCLI_DIR="/usr/local/admin/scripts/"
OPENPANEL_LOG_DIR="/var/log/openpanel/"
SERVICES_DIR="/etc/systemd/system/"
TEMP_DIR="/tmp/"

OPENPANEL_DIR="/usr/local/panel/"
CURRENT_PANEL_VERSION=$(< ${OPENPANEL_DIR}/version)

LOG_FILE="${OPENPANEL_LOG_DIR}admin/notifications.log"
DEBUG_MODE=0

update_blocker() {

    echo -e "${RED}==================================================${NC}"
    echo -e "${RED}                                                  ${NC}"
    echo -e "${RED}         THIS UPDATE IS NOT YET RELEASED           ${NC}"
    echo -e "${RED}                                                  ${NC}"
    echo -e "${RED}======== DO NOT RUN ON PRODUCTION SERVERS ========${NC}"
    echo -e "${RED}                                                  ${NC}"
    echo -e "${RED}==================================================${NC}"

    exit 1
}

#update_blocker

echo "Starting update.."

# Check if the --debug flag is provided
for arg in "$@"
do
    if [ "$arg" == "--debug" ]; then
        DEBUG_MODE=1
        break
    fi
done

# HELPERS

# Progress bar script
PROGRESS_BAR_URL="https://raw.githubusercontent.com/pollev/bash_progress_bar/master/progress_bar.sh"
PROGRESS_BAR_FILE="progress_bar.sh"
wget "$PROGRESS_BAR_URL" -O "$PROGRESS_BAR_FILE" > /dev/null 2>&1

if [ ! -f "$PROGRESS_BAR_FILE" ]; then
    echo "Failed to download progress_bar.sh"
    exit 1
fi

# Source the progress bar script
source "$PROGRESS_BAR_FILE"

# Display progress bar
FUNCTIONS=(

    # notify user we started
    print_header

    # update docker openpanel image
    download_new_panel

    # update opencli
    opencli_update

    # ping us
    verify_license

    docker_compose_up_with_newer_images

    # yay! we made it
    celebrate

    post_install_message

    # admin panel
    download_new_admin

    # if user created a post-update script, run it now
    run_custom_postupdate_script

)

TOTAL_STEPS=${#FUNCTIONS[@]}
CURRENT_STEP=0

update_progress() {
    CURRENT_STEP=$((CURRENT_STEP + 1))
    PERCENTAGE=$(($CURRENT_STEP * 100 / $TOTAL_STEPS))
    draw_progress_bar $PERCENTAGE
}

main() {
    # Make sure that the progress bar is cleaned up when user presses ctrl+c
    enable_trapping
    setup_scroll_area
    for func in "${FUNCTIONS[@]}"
    do
        $func
        update_progress
    done
    destroy_scroll_area
}

# print fullwidth line
print_space_and_line() {
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
}

# END helper functions

# START MAIN FUNCTIONS

write_notification() {
  local title="$1"
  local message="$2"
  local current_message="$(date '+%Y-%m-%d %H:%M:%S') UNREAD $title MESSAGE: $message"
  echo "$current_message" >> "$LOG_FILE"
}

print_header() {
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
    echo -e "   ____                         _____                      _  "
    echo -e "  / __ \                       |  __ \                    | | "
    echo -e " | |  | | _ __    ___  _ __    | |__) | __ _  _ __    ___ | | "
    echo -e " | |  | || '_ \  / _ \| '_ \   |  ___/ / _\" || '_ \ / _  \| | "
    echo -e " | |__| || |_) ||  __/| | | |  | |    | (_| || | | ||  __/| | "
    echo -e "  \____/ | .__/  \___||_| |_|  |_|     \__,_||_| |_| \___||_| "
    echo -e "         | |                                                  "
    echo -e "         |_|                                                  "
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
    echo -e "Starting update to OpenPanel version $NEW_PANEL_VERSION"
    echo -e ""
    echo -e "Changelog: https://openpanel.com/docs/changelog/$NEW_PANEL_VERSION"        
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
    echo -e ""
}

opencli_update(){
    echo "Updating OpenCLI commands from https://storage.googleapis.com/openpanel/${NEW_PANEL_VERSION}/opencli-main.tar.gz"
    echo ""
    mkdir -p ${TEMP_DIR}opencli
    wget -O ${TEMP_DIR}opencli.tar.gz "https://storage.googleapis.com/openpanel/${NEW_PANEL_VERSION}/opencli-main.tar.gz"  > /dev/null 2>&1
    cd ${TEMP_DIR} && tar -xzf opencli.tar.gz -C ${TEMP_DIR}opencli
    rm -rf /usr/local/admin/scripts/
    cp -r ${TEMP_DIR}opencli/ /usr/local/admin/scripts/
    rm ${TEMP_DIR}opencli.tar.gz 
    rm -rf ${TEMP_DIR}opencli
    cp /usr/local/admin/scripts/opencli /usr/local/bin/opencli
    chmod +x /usr/local/bin/opencli
    chmod +x -R /usr/local/admin/scripts/
    opencli commands
    source ~/.bashrc
}

run_custom_postupdate_script() {
    echo "Checking if post-update script is provided.."
    echo ""
    if [ -f "/root/openpanel_run_after_update" ]; then
        echo " "
        echo "Running post update script: '/root/openpanel_run_after_update'"
        echo "https://dev.openpanel.com/customize.html#After-update"
        bash /root/openpanel_run_after_update
    fi
}

download_new_admin() {
  cd $OPENADMIN_DIR
  git pull
  restart_admin_panel_if_needed() {
      echo "Restarting OpenAdmin service.."
      service admin restart
  }
  restart_admin_panel_if_needed
}

download_new_panel() {
    mkdir -p $OPENPANEL_DIR
    if [ "$(uname -m)" = "aarch64" ]; then
        echo "ARM (AArch64) detected. Pulling ARM-compatible OpenPanel image..."
        docker pull openpanel/openpanel-arm || docker pull openpanel/openpanel
    else
        echo "Downloading latest OpenPanel image from https://hub.docker.com/r/openpanel/openpanel"
        docker pull openpanel/openpanel
    fi
}

docker_compose_up_with_newer_images(){
  echo "Restarting OpenPanel docker container.."
  echo ""
  docker stop openpanel && docker rm openpanel
  echo ""
  cd /root  
  docker compose up -d --no-deps --build openpanel
  mkdir -p /usr/local/panel/ > /dev/null 2>&1
  docker cp openpanel:/usr/local/panel/version /usr/local/panel/version
}

verify_license() {
    current_ip=$(curl -s --max-time 10 https://ip.openpanel.com || wget -qO- --timeout=10 https://ip.openpanel.com)
    echo "Checking OpenPanel license for IP address: $current_ip" 
    echo ""
    server_hostname=$(hostname)
    license_data='{"hostname": "'"$server_hostname"'", "public_ip": "'"$current_ip"'"}'
    response=$(curl -s --max-time 10 -X POST -H "Content-Type: application/json" -d "$license_data" https://api.openpanel.co/license-check)
}

celebrate() {
    print_space_and_line
    echo ""
    echo -e "${GREEN}OpenPanel successfully updated to ${NEW_PANEL_VERSION}.${NC}"
    echo ""
    sed -i 's/UNREAD New OpenPanel update is available/READ New OpenPanel update is available/' $LOG_FILE
    write_notification "OpenPanel successfully updated!" "OpenPanel successfully updated from $CURRENT_PANEL_VERSION to $NEW_PANEL_VERSION"
}

post_install_message() {
    print_space_and_line
    echo -e "\nIf you experience any problems or need further assistance, please do not hesitate to reach out on our community forums or join our Discord server for support:"
    echo ""
    echo "👉 Forums: https://community.openpanel.org/"
    echo ""
    echo "👉 Discord: https://discord.openpanel.com/"
    echo ""
    echo "Our community and support team are ready to help you!"
}

# main execution of the script
main

# Placeholder: Add update logic for version 0.3.7 here.
