#!/bin/bash

# Update from OpenPanel 0.2.8 to 0.2.9

# new version
NEW_PANEL_VERSION="0.2.9"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
RESET='\033[0m'

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


# block updates for older versions!
required_version="0.2.1"

if [[ "$CURRENT_PANEL_VERSION" < "$required_version" ]]; then
    # Version is less than 0.2.1, no update will be performed
    echo ""
    echo "NO UPDATES FOR VERSIONS =< 0.1.9 - PLEASE REINSTALL OPENPANEL"
    echo "Annoucement: https://community.openpanel.com/d/65-important-update-openpanel-version-021-announcement"
    echo ""
    exit 0
else
    # Version is 0.2.1 or newer
    echo "Starting update.."
fi




# Check if the --debug flag is provided
for arg in "$@"
do
    if [ "$arg" == "--debug" ]; then
        DEBUG_MODE=1
        break
    fi
done






# HELPERS


# Display error and exit
radovan() {
    echo -e "${RED}Error: $2${RESET}" >&2
    exit $1
}


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

    # ftp service
    update_configuration_files

    # update docker openpanel iamge
    download_new_panel

    # update admin from github
    download_new_admin

    # 0.2.9 only
    fix_for_inotifytools

    # update opencli
    opencli_update

    # ping us
    verify_license

    # kill existing ftp server for 0.2.9 only
    kill_existing_ftp
    
    # openpanel/openpanel should be downloaded now!
    docker_compose_up_with_newer_images

    # if user created a post-update script, run it now
    run_custom_postupdate_script

    # yay! we made it
    celebrate

    # show to user what was done and how to restore previous version if needed!
    post_install_message

    # must be at the end, otherwise breaks install process from gui
    restart_admin_panel_if_needed

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



# print fullwidth line
print_space_and_line() {
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
}



# END helper functions


fix_for_inotifytools() {
 # temporary for 0.2.9

 
        # Detect the package manager and install inotifywait-tools
        if command -v apt-get &> /dev/null; then
            sudo apt-get install -y -qq inotify-tools > /dev/null 2>&1
        elif command -v yum &> /dev/null; then
            sudo yum install -y -q inotify-tools > /dev/null 2>&1
        elif command -v dnf &> /dev/null; then
            sudo dnf install -y -q inotify-tools > /dev/null 2>&1
        else
            echo "Error: No compatible package manager found. Please install inotify-tools manually for DNS zone reload to work."
        fi

        service watcher restart
}



update_configuration_files() {
    #######cd /etc/openpanel && git pull

    echo "Downloading files for FTP modules.."
    
    mkdir -p /etc/openpanel/ftp
    wget -O /etc/openpanel/ftp/vsftpd.conf https://gist.githubusercontent.com/stefanpejcic/b0777f7ef39628703f2ccb41f3ab0358/raw/51f535a8758e3c7e714b71daaf120bb001029176/vsftpd.conf

    wget -O /etc/openpanel/ftp/start_vsftpd.sh https://gist.githubusercontent.com/stefanpejcic/e4c45848777d12330199dc78e6a6030f/raw/32a3fbcb40cbe5193ab74410052b1c412700814b/start_vsftpd.sh
    chmod +x /etc/openpanel/ftp/start_vsftpd.sh

    touch /etc/openpanel/ftp/all.users

    wget -O /etc/openpanel/ftp/Dockerfile https://gist.githubusercontent.com/stefanpejcic/53718bdbee13b6312223f00627badc98/raw/012af234ba3dee1527e5d0d91a4ba80b53f627a8/Dockerfile


DOCKER_COMPOSE_FILE="/root/docker-compose.yml"

# Check if the service 'ftp_env_generator' exists in the file
if ! grep -q "ftp_env_generator" "$DOCKER_COMPOSE_FILE"; then
  FTP_CODE="
# FTP
  ftp_env_generator:
    image: alpine:latest
    container_name: ftp_env_generator
    volumes:
      - /etc/openpanel/ftp/:/etc/openpanel/ftp/
      - /usr/local/admin/scripts/ftp/users:/usr/local/admin/scripts/ftp/users
    entrypoint: /bin/sh -c \"/usr/local/admin/scripts/ftp/users\"
    restart: \"no\"  # Do not restart, we just want it to run once

  openadmin_ftp:
    build:
      context: /etc/openpanel/ftp/
    container_name: openadmin_ftp
    restart: always
    ports:
      - \"21:21\"
      - \"21000-21010:21000-21010\"
    volumes:
      - /home/:/home/
      - /etc/openpanel/ftp/vsftpd.conf:/etc/vsftpd/vsftpd.conf
      - /etc/openpanel/ftp/start_vsftpd.sh:/bin/start_vsftpd.sh
      - /etc/openpanel/ftp/vsftpd.chroot_list:/etc/vsftpd.chroot_list
      - /etc/openpanel/users/:/etc/openpanel/ftp/users/
      # uncomment for ssl # - /etc/letsencrypt:/etc/letsencrypt:ro
    depends_on:
      - ftp_env_generator
    env_file:
      - /etc/openpanel/ftp/all.users
    # uncomment the following lines for SSL and replace ftp.YOUR_DOMAIN_HERE.com with your domain
    # environment:
      # - ADDRESS=ftp.YOUR_DOMAIN_HERE.com
      # - TLS_CERT=\"/etc/letsencrypt/live/ftp.YOUR_DOMAIN_HERE.com/fullchain.pem\"
      # - TLS_KEY=\"/etc/letsencrypt/live/ftp.YOUR_DOMAIN_HERE.com/privkey.pem\"
    mem_limit: 0.5g
    cpus: 0.5
"

  sed -i "/# make the mysql data persistent/i$FTP_CODE" "$DOCKER_COMPOSE_FILE"

  echo "ftp_env_generator and openadmin_ftp services have been added to the docker-compose file."
else
  echo "ftp_env_generator service already exists in the docker-compose file."
fi





    
}





kill_existing_ftp() {
    # Check if the Docker container named 'openadmin_ftp' is running
    if docker ps -a --format '{{.Names}}' | grep -q '^openadmin_ftp$'; then
        echo "Stopping and removing the existing 'openadmin_ftp' container..."
        
        docker stop openadmin_ftp
        docker rm openadmin_ftp
        echo "'openadmin_ftp' container stopped and removed. Starting with new command:"

        echo ""

        echo "Generating list of all accounts with: cd /root && docker compose up -d ftp_env_generator"
        cd /root && docker compose up -d ftp_env_generator

        echo "Starting new FTP server with commandL: cd /root && docker compose up -d openadmin_ftp"
        cd /root && docker compose up -d openadmin_ftp

        echo "Checking and opening FTP ports on firewall.."

            function open_port_csf() {
                local port=$1
                local csf_conf="/etc/csf/csf.conf"
                
                # Check if port is already open
                port_opened=$(grep "TCP_IN = .*${port}" "$csf_conf")
                if [ -z "$port_opened" ]; then
                    # Open port
                    sed -i "s/TCP_IN = \"\(.*\)\"/TCP_IN = \"\1,${port}\"/" "$csf_conf"
                    echo -e "Port ${port} is now open."
                    ports_opened=1
                else
                    echo -e "Port ${port} is already open."
                fi
            }


          open_port_csf 21 > /dev/null 2>&1 #ftp
    	  open_port_csf 21000:21010 > /dev/null 2>&1 #passive ftp

    	  ufw allow 21/tcp > /dev/null 2>&1 #ftp
          ufw allow 21000:21010/tcp > /dev/null 2>&1 #passive ftp


        echo "Finished setting FTP server."




        
    fi
}



restart_admin_panel_if_needed() {
    service admin restart
}

    

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
    echo "Updating OpenCLI commands from https://storage.googleapis.com/openpanel/${NEW_PANEL_VERSION}/get.openpanel.co/downloads/${NEW_PANEL_VERSION}/opencli/opencli-main.tar.gz"
    echo ""
    mkdir -p ${TEMP_DIR}opencli
    wget -O ${TEMP_DIR}opencli.tar.gz "https://storage.googleapis.com/openpanel/${NEW_PANEL_VERSION}/get.openpanel.co/downloads/${NEW_PANEL_VERSION}/opencli/opencli-main.tar.gz"
    cd ${TEMP_DIR} && tar -xzf opencli.tar.gz -C ${TEMP_DIR}opencli
    rm -rf /usr/local/admin/scripts/
    
    cp -r ${TEMP_DIR}opencli/ /usr/local/admin/scripts/
    rm ${TEMP_DIR}opencli.tar.gz 
    rm -rf ${TEMP_DIR}opencli

    cp /usr/local/admin/scripts/opencli /usr/local/bin/opencli
    chmod +x /usr/local/bin/opencli
    chmod +x -R /usr/local/admin/scripts/
    opencli commands
    echo "# opencli aliases
    ALIASES_FILE=\"/usr/local/admin/scripts/aliases.txt\"
    generate_autocomplete() {
        awk '{print \$NF}' \"\$ALIASES_FILE\"
    }
    complete -W \"\$(generate_autocomplete)\" opencli" >> ~/.bashrc
    
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

    mkdir -p $OPENADMIN_DIR
    echo "Updating OpenAdmin from https://github.com/stefanpejcic/openadmin"
    echo ""
    cd $OPENADMIN_DIR
    #stash is used for demo
    git stash
    git pull
    git stash pop

    # need for csf
    chmod +x ${OPENADMIN_DIR}modules/security/csf.pl
    
    cd $OPENADMIN_DIR
    pip3 install --force-reinstall zope.event || pip3 install --force-reinstall zope.event --break-system-packages

    mv /etc/openpanel/openadmin/config/terms /etc/openpanel/openadmin/config/terms_accepted_on_update

}



download_new_panel() {

    mkdir -p $OPENPANEL_DIR
    echo "Downloading latest OpenPanel image from https://hub.docker.com/r/openpanel/openpanel"
    echo ""
    docker pull openpanel/openpanel
}


docker_compose_up_with_newer_images(){

  echo "Restarting OpenPanel docker container.."
  echo ""
  docker stop openpanel &&  docker rm openpanel
  echo ""
  cd /root  
  docker compose up -d --no-deps --build openpanel

  #cp version file
  mkdir -p /usr/local/panel/  > /dev/null 2>&1
  docker cp openpanel:/usr/local/panel/version /usr/local/panel/version
}



verify_license() {
    # Get server ipv4
    current_ip=$(curl -s --max-time 10 https://ip.openpanel.co || wget -qO- --timeout=10 https://ip.openpanel.co)
    echo "Checking OpenPanel license for IP address: $current_ip" 
    echo ""
    server_hostname=$(hostname)
    license_data='{"hostname": "'"$server_hostname"'", "public_ip": "'"$current_ip"'"}'
    response=$(curl -s --max-time 10 -X POST -H "Content-Type: application/json" -d "$license_data" https://api.openpanel.co/license-check)
    #echo "Response: $response"
}




celebrate() {

    print_space_and_line

    echo ""
    echo -e "${GREEN}OpenPanel successfully updated to ${NEW_PANEL_VERSION}.${RESET}"
    echo ""

    # remove the unread notification that there is new update
    sed -i 's/UNREAD New OpenPanel update is available/READ New OpenPanel update is available/' $LOG_FILE

    # add notification that update was successful
    write_notification "OpenPanel successfully updated!" "OpenPanel successfully updated from $CURRENT_PANEL_VERSION to $NEW_PANEL_VERSION"
}



post_install_message() {

    print_space_and_line

    # Instructions for seeking help
    echo -e "\nIf you experience any problems or need further assistance, please do not hesitate to reach out on our community forums or join our Discord server for 
support:"
    echo ""
    echo "👉 Forums: https://community.openpanel.com/"
    echo ""
    echo "👉 Discord: https://discord.openpanel.com/"
    echo ""
    echo "Our community and support team are ready to help you!"
}



# main execution of the script
main
