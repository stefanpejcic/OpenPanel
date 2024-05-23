#!/bin/bash

# Update from OpenPanel 0.1.4 to 0.1.5

# new version
NEW_PANEL_VERSION="0.1.5"


# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
RESET='\033[0m'

# Directories
OPENPANEL_DIR="/usr/local/panel/"
OPENADMIN_DIR="/usr/local/admin/"
OPENCLI_DIR="/usr/local/admin/scripts/"
OPENPANEL_ERR_DIR="/var/log/openpanel/"
SERVICES_DIR="/etc/systemd/system/"


# Update source
REPO_URL="https://get.openpanel.co/downloads/"
CURRENT_PANEL_VERSION=$(< ${OPENPANEL_DIR}/version)

# temporary directories
TEMP_DIR="/tmp/"
OLD_OPENPANEL_DIR="/usr/local/panel-${CURRENT_PANEL_VERSION}/"
OLD_OPENADMIN_DIR="/usr/local/admin-${CURRENT_PANEL_VERSION}/"

LOG_FILE="${OPENADMIN_DIR}logs/notifications.log"
DEBUG_MODE=0


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

    # check license first, if not licenses enterprise version we will abort
    verify_license

    # check requirements: OS, root, not docker..
    check_requirements

    # check if user has additional webservers to avoid conflicts!
    detect_installed_panels

    # python version again - to be used in following functions
    check_python_version

    # check ubuntu and apt-get
    detect_os_and_package_manager

    # apt-update
    update_package_manager

    # move the existing panel folder and use it as fallback if update fails
    mv_panel_to_tmp_dir

    # install new user panel from git
    download_new_panel

    # copy configuration files that user modified and other custom data
    move_openpanel_data

    # install pip requirements for user panel
    pip_install_for_panel

    # move the existing admin folder and use it as fallback if update fails
    mv_admin_to_tmp_dir

    # install new admin panel from git
    download_new_admin

    # copy configuration files that user modified and other custom data
    move_openadmin_data

    # install pip requirements for admin panel
    pip_install_for_admin

    # install new opencli version from github
    setup_opencli_new_version

    # reload admin and panel services
    reload_services

    # check if services started, if not restore the old version
    check_if_panel_services_started

    # set cronjobs
    set_system_cronjob

    # delete temp files and (maybe) old panel versison
    cleanup

    # since openpanel <0.1.4 also run /usr/local/admin/scripts/INSTALL.sh post update, lets remove it!
    remove_opencli_install_script_from_cli

    # if user created a post-update script, run it now
    run_custom_postupdate_script

    # send us server info to collect new version data and usage
    report_status_after_update

    # yay! we made it
    celebrate

    # show to user what was done and how to restore previous version if needed!
    post_install_message

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



















# START MAIN FUNCTIONS



mv_panel_to_tmp_dir() {
    mv $OPENPANEL_DIR $OLD_OPENPANEL_DIR
}

mv_admin_to_tmp_dir() {
    mv $OPENADMIN_DIR $OLD_OPENADMIN_DIR
}








pip_install_for_panel(){
    echo "Installing PIP requirements for OpenPanel.."
    cd $OPENPANEL_DIR
    pip install -r requirements.txt  > /dev/null 2>&1
}



pip_install_for_admin(){
    echo "Installing PIP requirements for OpenAdmin.."
    cd $OPENADMIN_DIR
    pip install -r requirements.txt  > /dev/null 2>&1
}





check_requirements() {
    # https://github.com/stefanpejcic/openpanel/issues/63

    architecture=$(lscpu | grep Architecture | awk '{print $2}')

    if [ "$architecture" == "aarch64" ]; then
        echo -e "${RED}Error: ARM CPU is not supported!${RESET}" >&2
        exit 1
    fi

    # check if the current user is not root
    if [ "$(id -u)" != "0" ]; then
        echo -e "${RED}Error: you must be root to execute this script.${RESET}" >&2
        exit 1
    # check if OS is MacOS
    elif [ "$(uname)" = "Darwin" ]; then
        echo -e "${RED}Error: MacOS is not currently supported.${RESET}" >&2
        exit 1
    # check if running inside a container
    elif [[ -f /.dockerenv || $(grep -sq 'docker\|lxc' /proc/1/cgroup) ]]; then
        echo -e "${RED}Error: running openpanel inside a container is not supported.${RESET}" >&2
        exit 1
    fi
    # check if python version is supported
    current_python_version=$(python3 --version 2>&1 | cut -d " " -f 2 | cut -d "." -f 1,2 | tr -d '.')
    allowed_versions=("39" "310" "311" "38")
    if [[ ! " ${allowed_versions[@]} " =~ " ${current_python_version} " ]]; then
        echo -e "${RED}Error: Unsupported Python version $current_python_version. No corresponding branch available.${RESET}" >&2
        exit 1
    fi
}



detect_installed_panels() {
    if [ -z "$SKIP_PANEL_CHECK" ]; then
        # Define an associative array with key as the directory path and value as the error message
        declare -A paths=(
            ["/usr/local/cpanel/whostmgr"]="cPanel WHM is installed. OpenPanel only supports servers without any hosting control panel installed."
            ["/opt/psa/version"]="Plesk is installed. OpenPanel only supports servers without any hosting control panel installed."
            ["/usr/local/psa/version"]="Plesk is installed. OpenPanel only supports servers without any hosting control panel installed."
            ["/usr/local/CyberPanel"]="CyberPanel is installed. OpenPanel only supports servers without any hosting control panel installed."
            ["/usr/local/directadmin"]="DirectAdmin is installed. OpenPanel only supports servers without any hosting control panel installed."
            ["/usr/local/cwpsrv"]="CentOS Web Panel (CWP) is installed. OpenPanel only supports servers without any hosting control panel installed."
        )

        for path in "${!paths[@]}"; do
            if [ -d "$path" ] || [ -e "$path" ]; then
                radovan 1 "${paths[$path]}"
            fi
        done

        echo -e "No additional webservers found that might block the update process, proceeding with the update process."
    fi
}




check_python_version(){
    echo "Checking Python version to use for OpenPanel.."
    current_python_version=$(python3 --version 2>&1 | cut -d " " -f 2 | cut -d "." -f 1,2 | tr -d '.')
}



# Function to write notification to log file
write_notification() {
  local title="$1"
  local message="$2"
  local current_message="$(date '+%Y-%m-%d %H:%M:%S') UNREAD $title MESSAGE: $message"

  echo "$current_message" >> "$LOG_FILE"
}


cleanup(){
    echo "Cleaning up.."

    # empty error and access logs
    rm -rf ${OPENPANEL_ERR_DIR}
    mkdir -p ${OPENPANEL_ERR_DIR}user ${OPENPANEL_ERR_DIR}admin

    # install data
    rm -rf ${OPENPANEL_DIR}INSTALL.sh ${OPENPANEL_DIR}DATABASE.sql ${OPENPANEL_DIR}requirements.txt ${OPENPANEL_DIR}conf/nginx.conf ${OPENPANEL_DIR}conf/mysqld.cnf ${OPENPANEL_DIR}conf/named.conf.options ${OPENPANEL_DIR}conf/default.nginx.conf ${OPENPANEL_DIR}conf/.gitkeep ${OPENPANEL_DIR}templates/dist.zip ${OPENPANEL_DIR}install
    rm -rf ${OPENPANEL_DIR}services "${OPENPANEL_DIR}conf/GeoLite2-City_20231219/" "${OPENPANEL_DIR}conf/GeoLite2-Country_20231219/"
    rm ${OPENPANEL_DIR}conf/GeoLite2-City_20231219.tar.gz ${OPENPANEL_DIR}conf/GeoLite2-Country_20231219.tar.gz ${OPENPANEL_DIR}conf/goaccess.conf 
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

    echo -e "Starting update process.."
    echo -e ""
}



reload_services() {
    echo "Reloading services.."
    service panel restart
    #TODO: no restart, in case admin deliberately stopped the service.. 
    service admin reload 
}



check_if_panel_services_started() {
    echo "Checking if OpenAdmin and OpenPanel services started successfully.."

    admin_status=$(systemctl status admin 2>&1)
    panel_status=$(systemctl status panel 2>&1)

    if echo "$admin_status" | grep -q "failed" || echo "$panel_status" | grep -q "failed"; then
        echo "OpenAdmin or OpenPanel service failed to start after the update! Reverting back to the previous version: $CURRENT_PANEL_VERSION"

        # delete new version files for both services
        find ${OPENADMIN_DIR} -maxdepth 0 ! -name "${CURRENT_PANEL_VERSION}" -exec rm -rf {} +
        find ${OPENPANEL_DIR} -maxdepth 0 ! -name "${CURRENT_PANEL_VERSION}" -exec rm -rf {} +

        # restore the previous version for both services
        mv ${OPENADMIN_DIR}${CURRENT_PANEL_VERSION} $OPENADMIN_DIR && rmdir ${OPENADMIN_DIR}${CURRENT_PANEL_VERSION}
        mv ${OPENPANEL_DIR}${CURRENT_PANEL_VERSION} $OPENPANEL_DIR && rmdir ${OPENPANEL_DIR}${CURRENT_PANEL_VERSION}

        # write the update failed notification
        write_notification "OpenPanel update failed!" "OpenPanel update failed, please contact support."

        echo -e "${RED}UPDATE FAILED${RESET}"
        exit $1
    else
        echo "OpenAdmin and Panel services started normally."
    fi
}



run_custom_postupdate_script() {

    echo "Checking if post-update script is provided.."

    # Check if the file /root/openpanel_run_after_update exists
    if [ -f "/root/openpanel_run_after_update" ]; then
        # run the custom script
        echo " "
        echo "Running post install script: '/root/openpanel_run_after_update'"
        echo "https://dev.openpanel.co/customize.html#After-update"
        bash /root/openpanel_run_after_update

    fi
}



update_services() {
    echo "Updating services.."
    # nothing for 0.1.5
}


download_new_admin() {

    mkdir -p $OPENADMIN_DIR

    echo "Downloading OpenAdmin version ${NEW_PANEL_VERSION}"

    wget -O ${TEMP_DIR}openadmin.tar.gz "${REPO_URL}${NEW_PANEL_VERSION}/openadmin/${current_python_version}/compressed.tar.gz" > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        echo "Error downloading OpenAdmin from ${REPO_URL}. Exiting."

        # write the update failed notification
        #write_notification "OpenPanel update failed!" "OpenPanel update failed, error downloading OpenAdmin from ${REPO_URL}."

        check_if_panel_services_started

        exit 1
    fi

    #echo "wget -O ${TEMP_DIR}openadmin.tar.gz \"${REPO_URL}${NEW_PANEL_VERSION}/openadmin/${current_python_version}/compressed.tar.gz\""

    cd ${TEMP_DIR} && tar -xzf openadmin.tar.gz -C $OPENADMIN_DIR

    unzip ${OPENADMIN_DIR}static/dist.zip -d ${OPENADMIN_DIR}static/dist/

    rm ${TEMP_DIR}openadmin.tar.gz ${OPENADMIN_DIR}static/dist.zip
}

download_new_panel() {

    mkdir -p $OPENPANEL_DIR

    echo "Downloading OpenPanel version ${NEW_PANEL_VERSION}"

    wget -O ${TEMP_DIR}openpanel.tar.gz "${REPO_URL}${NEW_PANEL_VERSION}/openpanel/${current_python_version}/compressed.tar.gz" > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        echo "Error downloading OpenPanel from ${REPO_URL}. Exiting."

        # write the update failed notification
        #write_notification "OpenPanel update failed!" "OpenPanel update failed, error downloading OpenPanel from ${REPO_URL}."

        check_if_panel_services_started

        exit 1
    fi

    #echo "wget -O ${TEMP_DIR}openpanel.tar.gz \"${REPO_URL}${NEW_PANEL_VERSION}/openpanel/${current_python_version}/compressed.tar.gz\""

    cd ${TEMP_DIR} && tar -xzf openpanel.tar.gz -C $OPENPANEL_DIR

    rm ${TEMP_DIR}openpanel.tar.gz
}




move_openpanel_data() {

    echo "Restoring custom OpenPanel configuration files from previous version: ${CURRENT_PANEL_VERSION}"

    # server conf
    cp -r ${OLD_OPENPANEL_DIR}conf/* ${OPENPANEL_DIR}conf/

    # users data
    cp -r ${OLD_OPENPANEL_DIR}core/* ${OPENPANEL_DIR}core/
    #cp -r ${OLD_OPENPANEL_DIR}core/stats/* ${OPENPANEL_DIR}core/stats/

}

move_openadmin_data() {

    echo "Restoring custom OpenAdmin configuration files from previous version: ${CURRENT_PANEL_VERSION}"

    # admin users
    cp ${OLD_OPENADMIN_DIR}users.db $OPENADMIN_DIR

    # mysql logins
    cp ${OLD_OPENADMIN_DIR}db.cnf $OPENADMIN_DIR
    #cp ${OLD_OPENADMIN_DIR}config.py $OPENADMIN_DIR
    cp ${OLD_OPENADMIN_DIR}config.json $OPENADMIN_DIR

    # logs
    cp ${OLD_OPENADMIN_DIR}logs/notifications.log ${OPENADMIN_DIR}logs/notifications.log

}


set_system_cronjob(){

    echo "Setting cronjobs.."

    mv /usr/local/panel/conf/cron /etc/cron.d/openpanel
}



setup_opencli_new_version() {

    mkdir $OPENCLI_DIR

    echo "Downloading OpenCLI.."

    wget -O ${TEMP_DIR}opencli.tar.gz "${REPO_URL}${NEW_PANEL_VERSION}/opencli/compressed.tar.gz" > /dev/null 2>&1

    cd ${TEMP_DIR} && tar -xzf opencli.tar.gz -C $OPENCLI_DIR

    rm ${TEMP_DIR}opencli.tar.gz

    chmod +x -R $OPENCLI_DIR

    echo "Creating a list of opencli commands for autocomplete"

    opencli commands > /dev/null 2>&1
}




detect_os_and_package_manager() {

    echo "Checking OS distribution and package manager to use for services.."

    if [ -f "/etc/os-release" ]; then
        . /etc/os-release
        case "$ID" in
            "debian"|"ubuntu")
                PACKAGE_MANAGER="apt-get"
                ;;
            "centos"|"cloudlinux"|"rhel"|"fedora"|"almalinux")
                PACKAGE_MANAGER="yum"
                if [ "$(command -v dnf)" ]; then
                    PACKAGE_MANAGER="dnf"
                fi
                ;;
            *)
                echo -e "${RED}Unsupported distribution: $ID. Exiting.${RESET}"
                echo -e "${RED}UPDATE FAILED${RESET}"

                # write the update failed notification
                write_notification "OpenPanel update failed!" "OpenPanel update failed, please contact support."

                exit 1
                ;;
        esac
    else
        echo -e "${RED}Could not detect Linux distribution. Exiting..${RESET}"
        echo -e "${RED}UPDATE FAILED${RESET}"

        # write the update failed notification
        write_notification "OpenPanel update failed!" "OpenPanel update failed, please contact support."

        exit 1
    fi
}



update_package_manager() {
        echo "Updating package manager.."
        $PACKAGE_MANAGER update -y
}



report_status_after_update() {
    opencli report --public > /dev/null 2>&1
    rm -rf ${OPENADMIN_DIR}/logs/reports/   
}



verify_license() {

    # Get server ipv4
    current_ip=$(curl -s https://ip.openpanel.co || wget -qO- https://ip.openpanel.co)

    echo "Checking OpenPanel license for IP address: $current_ip" 

    server_hostname=$(hostname)

    license_data='{"hostname": "'"$server_hostname"'", "public_ip": "'"$current_ip"'"}'

    response=$(curl -s -X POST -H "Content-Type: application/json" -d "$license_data" https://api.openpanel.co/license-check)

    #echo "Response: $response"
}


remove_opencli_install_script_from_cli() {
    rm -rf ${OPENADMIN_DIR}scripts/INSTALL.sh  
}

celebrate() {

    print_space_and_line

    echo ""
    echo -e "${GREEN}OpenPanel successfully updated to ${NEW_PANEL_VERSION}.${RESET}"
    echo ""

    print_space_and_line

    write_notification "OpenPanel successfully updated!" "OpenPanel successfully updated from $CURRENT_PANEL_VERSION to $NEW_PANEL_VERSION"
}


post_install_message() {

    # steps to revert the update
    echo "If you encounter any issues with this update, you can revert to the previous version using the following steps:"
    echo ""
    echo "1. Delete the new version files:"
    echo ""
    echo "   rm -rf /usr/local/admin /usr/local/panel"
    echo ""
    echo "2. Restore files from the previous verison (${CURRENT_PANEL_VERSION}):"
    echo ""
    echo "   mv /usr/local/admin-${CURRENT_PANEL_VERSION}/ /usr/local/admin ; mv /usr/local/panel-${CURRENT_PANEL_VERSION}/ /usr/local/panel"
    echo ""
    echo "3. Restart OpenPanel and OpenAdmin serviced:"
    echo ""
    echo "   service panel restart && service admin restart"
    echo ""
    echo "4. Check if OpenCLI is working and OpenAdmin is accessible:"
    echo ""
    echo "   opencli admin"
    echo ""

    print_space_and_line

    # Instructions for seeking help
    echo -e "\nIf you experience any problems or need further assistance, please do not hesitate to reach out on our community forums or join our Discord server for support:"
    echo ""
    echo "👉 Forums: https://community.openpanel.co/"
    echo ""
    echo "👉 Discord: https://discord.openpanel.co/"
    echo ""
    echo "Our community and support team are ready to help you!"
}


# main execution of the script
main

