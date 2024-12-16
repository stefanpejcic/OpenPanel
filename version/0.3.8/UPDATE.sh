#!/bin/bash

# Update from OpenPanel 0.3.7 to 0.3.8

# new version
NEW_PANEL_VERSION="0.3.8"
PREVIOUS_VERSION="0.3.7"

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

all_success=true







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

update_blocker

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

# Dsiplay progress bar
FUNCTIONS=(

    #notify user we started
    print_header

    # update docker openpanel image
    download_new_panel

    # only for 0.3.8
    add_mysql_container_table
    add_mysql_homedir_table

    # update opencli
    opencli_update
    
    update_nginx_conf
    
    # ping us
    verify_license
    
    docker_compose_up_with_newer_images

    # yay! we made it
    celebrate


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
    
    # Create progress bar
    setup_scroll_area
    for func in "${FUNCTIONS[@]}"
    do
        # Execute each function
        $func
        update_progress
    done
    destroy_scroll_area
}

# for 0.3.8 only!
update_nginx_conf() {
    wget -O /etc/openpanel/nginx/suspended_user.html https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/nginx/suspended_user.html
    wget -O /etc/openpanel/nginx/suspended_website.html https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/nginx/suspended_website.html
    wget -O /etc/openpanel/nginx/vhosts/domain.conf https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/nginx/vhosts/domain.conf
    wget -O  /etc/openpanel/nginx/vhosts/domain.conf_with_modsec https://github.com/stefanpejcic/openpanel-configuration/blob/main/nginx/vhosts/domain.conf_with_modsec

    file="/root/docker-compose.yml"
    line_to_add="        - /etc/openpanel/nginx/vhosts/:/etc/openpanel/nginx/vhosts/ # for custom suspended pages from 0.3.8"
    reference_line="        - /etc/openpanel/nginx/vhosts/openpanel_proxy.conf:/etc/openpanel/nginx/vhosts/openpanel_proxy.conf"
    conf_dir="/etc/nginx/sites-available"
    new_config_block='
        # custom templates
        set $suspended_user 0;
        set $suspended_website 0;
    
        location /suspended_website.html {
            root /etc/openpanel/nginx/;
            internal;
        }
    
        location /suspended_user.html {
            root /etc/openpanel/nginx/;
            internal;
        }
    
        # container
        location / {
            if ($suspended_user) {
                rewrite ^ /suspended_user.html last;
            }
            if ($suspended_website) {
                rewrite ^ /suspended_website.html last;
            }
        }
    '
    
    for conf_file in "$conf_dir"/*.conf; do
        # Check if the file contains the "# custom templates" line
        if ! grep -Fq "# custom templates" "$conf_file"; then
            if grep -Fq "# container" "$conf_file" && grep -Fq "location / {" "$conf_file"; then
                sed -i '/# container/,/location \//c\'"$new_config_block" "$conf_file"
                echo "Configuration updated in $conf_file"
            else
                echo "WARNING: Either # container or location / { section not found in $conf_file"
            fi
        else
            echo "Skipping domain $conf_file"
        fi
    done

    if grep -Fxq "$reference_line" "$file" && ! grep -Fxq "$line_to_add" "$file"; then
      sed -i "/$reference_line/a \\$line_to_add" "$file"
      echo "/etc/openpanel/nginx/vhosts/ successfully added in docker compose file, reloading nginx.."
      cd /root && docker compose down nginx && docker compose up -d nginx
    fi
    
}

# print fullwidth line
print_space_and_line() {
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
}



# END helper functions

    

# START MAIN FUNCTIONS



# Function to write notification to log file
write_notification() {
  local title="$1"
  local message="$2"
  local current_message="$(date '+%Y-%m-%d %H:%M:%S') UNREAD $title MESSAGE: $message"

  echo "$current_message" >> "$LOG_FILE"
}



# logo
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
    # Check if the file /root/openpanel_run_after_update exists
    if [ -f "/root/openpanel_run_after_update" ]; then
        echo " "
        echo "Running post update script: '/root/openpanel_run_after_update'"
        echo "https://dev.openpanel.com/customize.html#After-update"
        bash /root/openpanel_run_after_update

    fi
}


download_new_admin() {
        touch /etc/openpanel/openadmin/config/admin.ini
        
  cd $OPENADMIN_DIR
  git pull

  restart_admin_panel_if_needed() {
      echo "Restarting OpenAdmin service.."
      service admin restart
  }
  
 restart_admin_panel_if_needed
   
}


add_mysql_container_table() {
  COLUMN_EXISTS=$(mysql -N -e "SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = 'panel' AND TABLE_NAME = 'users' AND COLUMN_NAME = 'container';")
  if [ "$COLUMN_EXISTS" -eq 0 ]; then
    echo "Column 'container' does not exist. Creating column..."
    timeout 5s mysql -e "ALTER TABLE \`users\` ADD COLUMN \`container\` VARCHAR(255) DEFAULT NULL;"
    if [ $? -eq 124 ]; then
      echo "Warning: Timeout occurred while creating the column 'container' - retrying.."
      cd /root && docker compose down openpanel_mysql && docker compose up -d openpanel_mysql
      timeout 5s mysql -e "ALTER TABLE \`users\` ADD COLUMN \`container\` VARCHAR(255) DEFAULT NULL;"
        if [ $? -eq 124 ]; then
          echo "ERROR: Unable to add 'container' column, please restart mysql and run manually the query: ALTER TABLE users ADD COLUMN container VARCHAR(255) DEFAULT NULL;"
          exit 1
        else
          echo "Column 'container' created after restarting mysql."
        fi
    else
      echo "Column 'container' created."
    fi
  else
    echo "Column 'container' already exists."
  fi
  echo ""
}



add_mysql_homedir_table() {
  COLUMN_EXISTS=$(mysql -N -e "SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = 'panel' AND TABLE_NAME = 'users' AND COLUMN_NAME = 'homedir';")
  if [ "$COLUMN_EXISTS" -eq 0 ]; then
    echo "Column 'homedir' does not exist. Creating column..."
    timeout 5s mysql -e "ALTER TABLE \`users\` ADD COLUMN \`homedir\` VARCHAR(255) DEFAULT NULL;"
    if [ $? -eq 124 ]; then
      echo "Warning: Timeout occurred while creating the column 'homedir' - retrying.."
      cd /root && docker compose down openpanel_mysql && docker compose up -d openpanel_mysql
      timeout 5s mysql -e "ALTER TABLE \`users\` ADD COLUMN \`homedir\` VARCHAR(255) DEFAULT NULL;"
        if [ $? -eq 124 ]; then
          echo "ERROR: Unable to add 'homedir' column, please restart mysql and run manually the query: ALTER TABLE users ADD COLUMN homedir VARCHAR(255) DEFAULT NULL;"
          exit 1
        else
          echo "Column 'homedir' created after restarting mysql."
        fi
    else
      echo "Column 'homedir' created."
    fi
  else
    echo "Column 'homedir' already exists."
  fi
  echo ""
}




download_new_panel() {

    mkdir -p $OPENPANEL_DIR
    echo "Downloading latest OpenPanel image from https://hub.docker.com/r/openpanel/openpanel"
    echo ""
    docker pull openpanel/openpanel
    task1_result=$?
    if [ $task1_result -ne 0 ]; then
        all_success=false
    fi


    
}


docker_compose_up_with_newer_images(){

  echo "Restarting OpenPanel docker container.."
  echo ""
  docker stop openpanel &&  docker rm openpanel
  echo ""
  cd /root  
  docker compose up -d --no-deps --build openpanel

  mkdir -p /usr/local/panel/  > /dev/null 2>&1
  docker cp openpanel:/usr/local/panel/version /usr/local/panel/version
}



verify_license() {
    # Get server ipv4
    current_ip=$(curl -s --max-time 10 https://ip.openpanel.com || wget -qO- --timeout=10 https://ip.openpanel.com)
    echo "Checking OpenPanel license for IP address: $current_ip" 
    echo ""
    server_hostname=$(hostname)
    license_data='{"hostname": "'"$server_hostname"'", "public_ip": "'"$current_ip"'"}'
    response=$(curl -s --max-time 10 -X POST -H "Content-Type: application/json" -d "$license_data" https://api.openpanel.co/license-check)
}


celebrate() {

    print_space_and_line

    # Final status check
    if [ "$all_success" = true ]; then
        echo ""
        echo -e "${GREEN}OpenPanel successfully updated to ${NEW_PANEL_VERSION}.${NC}"
        echo ""
    
        
        sed -i 's/UNREAD New OpenPanel update is available/READ New OpenPanel update is available/' $LOG_FILE # remove the unread notification that there is new update
        write_notification "OpenPanel successfully updated!" "OpenPanel successfully updated from $CURRENT_PANEL_VERSION to $NEW_PANEL_VERSION" # add notification that update was successful
        post_install_message # more text
        download_new_admin # this will restart from admin panel and break progress - so do it at the end!
        run_custom_postupdate_script  # if user created a post-update script, run it now
    else
        echo ""
        echo -e "${RED}OpenPanel update failed!${NC}"
        echo ""
        sed -i 's/UNREAD New OpenPanel update is available/READ New OpenPanel update is available/' $LOG_FILE
        # TODO: in main update script exclude this version from future updates!
        write_notification "OpenPanel update failed!" "Error updating OpenPanel from $CURRENT_PANEL_VERSION to $NEW_PANEL_VERSION - please retry manually."

    fi

}



post_install_message() {

    print_space_and_line

    # Instructions for seeking help
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
