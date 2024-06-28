#!/bin/bash
################################################################################
# Script Name: INSTALL.sh
# Description: Install the latest version of OpenPanel
# Usage: cd /home && (curl -sSL https://get.openpanel.co || wget -O - https://get.openpanel.co) | bash
# Author: Stefan Pejcic
# Created: 11.07.2023
# Last Modified: 24.06.2024
# Company: openpanel.co
# Copyright (c) OPENPANEL
# 
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
# 
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
# 
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.
################################################################################


# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
RESET='\033[0m'

# Defaults
CUSTOM_VERSION=false
INSTALL_TIMEOUT=600 # 10 min
DEBUG=false
SKIP_APT_UPDATE=false
SKIP_IMAGES=false
REPAIR=false
LOCALES=true
NO_SSH=false
INSTALL_FTP=false
INSTALL_MAIL=false
OVERLAY=false
IPSETS=true
SET_HOSTNAME_NOW=false
SETUP_SWAP_ANYWAY=false
SWAP_FILE="1"
SELFHOSTED_SCREENSHOTS=false
SEND_EMAIL_AFTER_INSTALL=false
SET_PREMIUM=false

# Paths
ETC_DIR="/etc/openpanel/"
LOG_FILE="openpanel_install.log"
LOCK_FILE="/root/openpanel.lock"
OPENPANEL_DIR="/usr/local/panel/"
OPENPADMIN_DIR="/usr/local/admin/"
OPENCLI_DIR="/usr/local/admin/scripts/"
OPENPANEL_ERR_DIR="/var/log/openpanel/"
SERVICES_DIR="/etc/systemd/system/"
TEMP_DIR="/tmp/"

# Domains
SCREENSHOTS_API_URL="http://screenshots-api.openpanel.co/screenshot"

# Redirect output to the log file
exec > >(tee -a "$LOG_FILE") 2>&1






#####################################################################
#                                                                   #
# START helper functions                                            #
#                                                                   #
#####################################################################



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
    echo -e "         |_|                                   version: $version "

    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
}



install_started_message(){
    echo -e ""
    echo -e "\nStarting the installation of OpenPanel. This process will take approximately 3-5 minutes."
    echo -e "During this time, we will:"
    echo -e "- Install necessary services and tools."
    echo -e "- Create an admin account for you."
    echo -e "- Set up the firewall for enhanced security."
    echo -e "- Install needed Docker images."
    echo -e "- Set up basic hosting plans so you can start right away."
    echo -e "\nThank you for your patience. We're setting everything up for your seamless OpenPanel experience!\n"
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
    echo -e ""
}



# Display error and exit
radovan() {
    echo -e "${RED}Error: $2${RESET}" >&2
    exit $1
}


# print the command and its output if debug, else run and echo to /dev/null
debug_log() {
    if [ "$DEBUG" = true ]; then
        echo "Running: $@"
        "$@"
    else
        "$@" > /dev/null 2>&1
    fi
}


# Get server ipv4 from ip.openpanel.co
current_ip=$(curl -s https://ip.openpanel.co || wget -qO- https://ip.openpanel.co)

# If site is not available, get the ipv4 from the hostname -I
if [ -z "$current_ip" ]; then
   # current_ip=$(hostname -I | awk '{print $1}')
    # ip addr command is more reliable then hostname - to avoid getting private ip
    current_ip=$(ip addr|grep 'inet '|grep global|head -n1|awk '{print $2}'|cut -f1 -d/)
fi


if [ "$CUSTOM_VERSION" = false ]; then
    # Fetch the latest version
    version=$(curl -s https://get.openpanel.co/version)
    if [[ $version =~ [0-9]+\.[0-9]+\.[0-9]+ ]]; then
        version=$version
    else
        version="0.2.1"
    fi
fi


# print fullwidth line
print_space_and_line() {
    echo " "
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
    echo " "
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
#FUNKCIJE

download_skeleton_directory_from_github
install_openadmin
opencli_setup
add_file_watcher
configure_docker
docker_compose_up
#docker_compose_check_health
panel_customize

configure_nginx
download_and_import_docker_images
set_system_cronjob
setup_ufw


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



# END helper functions







#####################################################################
#                                                                   #
# START main functions                                              #
#                                                                   #
#####################################################################

check_requirements() {
    if [ -z "$SKIP_REQUIREMENTS" ]; then

        # https://github.com/stefanpejcic/openpanel/issues/63

        architecture=$(lscpu | grep Architecture | awk '{print $2}')

        if [ "$architecture" == "aarch64" ]; then
            echo -e "${RED}Error: ARM CPU is not supported!${RESET}" >&2
            exit 1
        fi

        # check if the current user is root
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
        allowed_versions=("39" "310" "311" "312" "38")
        if [[ ! " ${allowed_versions[@]} " =~ " ${current_python_version} " ]]; then
            echo -e "${RED}Error: Unsupported Python version $current_python_version. No corresponding branch available.${RESET}" >&2
            exit 1
        fi
    fi
}


parse_args() {
    show_help() {
        echo "Available options:"
        echo "  --key=<key_here>                Set the license key for OpenPanel Enterprise edition."
        echo "  --hostname=<hostname>           Set the hostname."
        echo "  --version=<version>             Set a custom OpenPanel version to be installed."
        echo "  --email=<stefan@example.net>    Set email address to receive email with admin credentials and future notifications."
        echo "  --skip-requirements             Skip the requirements check."
        echo "  --skip-panel-check              Skip checking if existing panels are installed."
        echo "  --skip-apt-update               Skip the APT update."
        echo "  --overlay2                      Enable overlay2 storage driver instead of device-mapper."
        echo "  --skip-firewall                 Skip UFW setup UFW - Only do this if you will set another Firewall manually!"
        echo "  --skip-images                   Skip installing openpanel/nginx and openpanel/apache docker images."
        echo "  --skip-blacklists               Do not set up IP sets and blacklists."
        echo "  --skip-ssl                      Skip SSL setup."
        echo "  --with_modsec                   Enable ModSecurity for Nginx."
        echo "  --ips                           Whiteliste IP addresses of OpenPanel Support Team."
        echo "  --no-ssh                        Disable port 22 and whitelist the IP address of user installing the panel."
        echo "  --enable-ftp                    Install FTP (experimental)."
        echo "  --enable-mail                   Install Mail (experimental)."
        echo "  --post_install=<path>           Specify the post install script path."
        echo "  --screenshots=<url>             Set the screenshots API URL."
        echo "  --swap=<2>                      Set space in GB to be allocated for SWAP."
        echo "  --debug                         Display debug information during installation."
        echo "  --repair                        Retry and overwrite everything."
        echo "  -h, --help                      Show this help message and exit."
    }







while [[ $# -gt 0 ]]; do
    case $1 in
        --key=*)
            SET_PREMIUM=true
            license_key="${1#*=}"
            ;;
        --hostname=*)
            SET_HOSTNAME_NOW=true
            new_hostname="${1#*=}"
            ;;
        --skip-requirements)
            SKIP_REQUIREMENTS=true
            ;;
        --skip-panel-check)
            SKIP_PANEL_CHECK=true
            ;;
        --skip-apt-update)
            SKIP_APT_UPDATE=true
            ;;
        --repair)
            REPAIR=true
            SKIP_PANEL_CHECK=true
            SKIP_REQUIREMENTS=true
            ;;
        --overlay2)
            OVERLAY=true
            ;;
        --skip-firewall)
            SKIP_FIREWALL=true
            ;;
        --skip-images)
            SKIP_IMAGES=true
            ;;
        --skip-blacklists)
            IPSETS=false
            ;;
        --skip-ssl)
            SKIP_SSL=true
            ;;
        --with_modsec)
            MODSEC=true
            ;;
        --debug)
            DEBUG=true
            ;;
        --ips)
            SUPPORT_IPS=true
            ;;
        --no-ssh)
            NO_SSH=true
            ;;
        --enable-ftp)
            INSTALL_FTP=true
            ;;
        --enable-mail)
            INSTALL_MAIL=true
            ;;
        --post_install=*)
            post_install_path="${1#*=}"
            ;;
        --screenshots=*)
            SCREENSHOTS_API_URL="${1#*=}"
            ;;
        --version=*)
            CUSTOM_VERSION=true
            version="${1#*=}"
            ;;
        --swap=*)
            SETUP_SWAP_ANYWAY=true
            SWAP="${1#*=}"
            ;;
        --email=*)
            SEND_EMAIL_AFTER_INSTALL=true
            EMAIL="${1#*=}"
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
    shift
done

}




detect_installed_panels() {
    if [ -z "$SKIP_PANEL_CHECK" ]; then
        # Define an associative array with key as the directory path and value as the error message
        declare -A paths=(
            ["/usr/local/panel"]="You already have OpenPanel installed. ${RESET}\nInstead, did you want to update? Run ${GREEN}'opencli update --force' to update OpenPanel."
            ["/usr/local/cpanel/whostmgr"]="cPanel WHM is installed. OpenPanel only supports servers without any hosting control panel installed."
            ["/opt/psa/version"]="Plesk is installed. OpenPanel only supports servers without any hosting control panel installed."
            ["/usr/local/psa/version"]="Plesk is installed. OpenPanel only supports servers without any hosting control panel installed."
            ["/usr/local/CyberPanel"]="CyberPanel is installed. OpenPanel only supports servers without any hosting control panel installed."
            ["/usr/local/directadmin"]="DirectAdmin is installed. OpenPanel only supports servers without any hosting control panel installed."
            ["/usr/local/cwpsrv"]="CentOS Web Panel (CWP) is installed. OpenPanel only supports servers without any hosting control panel installed."
            ["/usr/local/httpd"]="Apache WebServer is already installed. OpenPanel only supports servers without any webservers installed."
            ["/usr/local/apache2"]="Apache WebServer is already installed. OpenPanel only supports servers without any webservers installed."
            ["/usr/sbin/httpd"]="Apache WebServer is already installed. OpenPanel only supports servers without any webservers installed."
            ["/usr/lib/nginx"]="Nginx WebServer is already installed. OpenPanel only supports servers without any webservers installed."
        )

        for path in "${!paths[@]}"; do
            if [ -d "$path" ] || [ -e "$path" ]; then
                radovan 1 "${paths[$path]}"
            fi
        done

        echo -e "${GREEN}No currently installed hosting control panels or webservers found. Proceeding with the installation process.${RESET}"
    fi
}



detect_os_and_package_manager() {
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
                echo -e "${RED}INSTALL FAILED${RESET}"
                exit 1
                ;;
        esac
    else
        echo -e "${RED}Could not detect Linux distribution. Exiting..${RESET}"
        echo -e "${RED}INSTALL FAILED${RESET}"
        exit 1
    fi
}


download_and_import_docker_images() {
    echo "Downloading docker images in the background.."

    if [ "$SKIP_IMAGES" = false ]; then
        # See https://github.com/moby/moby/issues/16106#issuecomment-310781836 for pulling images in parallel
        # Nohup (no hang up) with trailing ampersand allows running the command in the background
        # The </dev/null bits redirects the outputs to files rather than strout/err
        nohup sh -c "echo openpanel/nginx:latest openpanel/apache:latest | xargs -P4 -n1 docker pull" </dev/null >nohup.out 2>nohup.err &
    fi
}


check_lock_file_age() {
    if [ "$REPAIR" = true ]; then
        rm "$LOCK_FILE"
        # and if lock file exists
        if [ -e "$LOCK_FILE" ]; then
            local current_time=$(date +%s)
            local file_time=$(stat -c %Y "$LOCK_FILE")
            local age=$((current_time - file_time))

            if [ "$age" -ge "$INSTALL_TIMEOUT" ]; then
                echo -e "${GREEN}Identified a prior interrupted OpenPanel installation; initiating a fresh installation attempt.${RESET}"
                rm "$LOCK_FILE"  # Remove the old lock file
            else
                echo -e "${RED}Detected another OpenPanel installation already running. Exiting.${RESET}"
                exit 1
            fi
        else
            # Create the lock file
            touch "$LOCK_FILE"
            echo "OpenPanel installation started at: $(date)"
        fi
    fi
}




configure_docker() {

    #########apt-get install docker.io -y
    
    docker_daemon_json_path="/etc/docker/daemon.json"
    mkdir -p $(dirname "$docker_daemon_json_path")

    if [ "$OVERLAY" = true ]; then
        echo "Setting 'overlay2' as the default storage driver for Docker.."
        cp ${ETC_DIR}docker/overlay2/daemon.json "$docker_daemon_json_path"
    else
        echo "Setting 'devicemapper' as the default storage driver for Docker.."
        cp ${ETC_DIR}docker/devicemapper/daemon.json "$docker_daemon_json_path"
    fi

    echo -e "Docker is configured."
    systemctl daemon-reload
    systemctl restart docker
}


docker_compose_check_health(){
    all_healthy=true
    containers=$(docker ps --format '{{.Names}}\t{{.Status}}')

    while IFS=$'\t' read -r name status; do
        if [[ "$status" != *"(healthy)"* ]]; then
            echo "Container $name is not healthy. Status: $status"
            all_healthy=false
        fi
    done <<< "$containers"

    # Proceed if all containers are healthy
    if $all_healthy; then
        echo "All containers are healthy. Proceeding with next installation steps..."
    else
        echo "Some containers are not healthy. Investigate and retry."
        exit 1
    fi

}






docker_compose_up(){
    echo "Setting Openpanel and MySQL docker containers.."
    echo ""
    # install docker compose
    DOCKER_CONFIG=${DOCKER_CONFIG:-$HOME/.docker}
    mkdir -p $DOCKER_CONFIG/cli-plugins
    curl -SL https://github.com/docker/compose/releases/download/v2.27.1/docker-compose-linux-x86_64 -o $DOCKER_CONFIG/cli-plugins/docker-compose  > /dev/null 2>&1
    chmod +x $DOCKER_CONFIG/cli-plugins/docker-compose
    chmod +x /usr/local/lib/docker/cli-plugins/docker-compose
    
    # TOFO: CHECK WITH 
    #docker compose version
    
    # mysql image needs this!
    wget -O  /root/initialize.sql https://gist.githubusercontent.com/stefanpejcic/8efe541c2b24b9cd6e1861e5ab7282f1/raw/140e69a5c8c7805bc9a6e6c4fa968f390a6d5c8c/structure.sql  > /dev/null 2>&1
    
    # compose doesnt alllow /
    cd /root
    
    # generate random password for mysql
    MYSQL_ROOT_PASSWORD=$(openssl rand -base64 -hex 9)
    echo "MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD" >> .env
    echo ""
    echo "MYSQL_ROOT_PASSWORD = $MYSQL_ROOT_PASSWORD"
    echo ""
    # save it to /etc/my.cnf
    ln -s /etc/openpanel/mysql/db.cnf /etc/my.cnf  > /dev/null 2>&1
    sed -i 's/password = .*/password = '"${MYSQL_ROOT_PASSWORD}"'/g' ${ETC_DIR}mysql/db.cnf  > /dev/null 2>&1
    
    wget -O  docker-compose.yml https://gist.githubusercontent.com/stefanpejcic/39c6a2f7d79c13d342601e2b0f14c6b8/raw/4b104d1adb986753f19827bcc37cf0fc39b66586/docker-compose.yml  > /dev/null 2>&1
    
    # start the stack
    docker compose up -d

}




clean_apt_cache(){
    # clear /var/cache/apt/archives/
    apt-get clean

    # TODO: cover https://github.com/debuerreotype/debuerreotype/issues/95
}



setup_ftp() {
        if [ "$INSTALL_FTP" = true ]; then
        echo "Installing experimental FTP service."
            curl -sSL https://raw.githubusercontent.com/stefanpejcic/OpenPanel-FTP/master/setup.sh | bash
        fi
}



setup_email() {
        if [ "$INSTALL_MAIL" = true ]; then
        echo "Installing experimental Email service."
            curl -sSL https://raw.githubusercontent.com/stefanpejcic/OpenMail/master/setup.sh | bash --dovecot
        fi
}


add_file_watcher(){
    bash <(curl -sSL https://raw.githubusercontent.com/stefanpejcic/file-watcher/main/install.sh)
}



setup_ufw() {
    if [ -z "$SKIP_FIREWALL" ]; then
        echo "Setting up the firewall.."
        debug_log wget -qO /usr/local/bin/ufw-docker https://github.com/chaifeng/ufw-docker/raw/master/ufw-docker  > /dev/null 2>&1 && 
        debug_log chmod +x /usr/local/bin/ufw-docker



        # block all docker ports so we can manually open only what is needed
        debug_log ufw-docker install
        debug_log ufw allow 80/tcp #http
        debug_log ufw allow 53  #dns
        debug_log ufw allow 443/tcp # https
        debug_log ufw allow 2083/tcp #openpanel
        debug_log ufw allow 2087/tcp #openadmin 
        

        if [ "$NO_SSH" = false ]; then

            # whitelist user running the script
            ip_of_user_running_the_script=$(w -h | grep -m1 -oP '\d+\.\d+\.\d+\.\d+')
            debug_log ufw allow from $ip_of_user_running_the_script

            # close port 22
            debug_log ufw allow 22  #ssh
        fi

        # set https://github.com/stefanpejcic/ipset-blacklist
        if [ "$IPSETS" = true ]; then
            if [ "$REPAIR" = true ]; then
                rm -rf ipset-blacklist-master
            fi
            if [ "$DEBUG" = true ]; then
                bash <(curl -sSL https://raw.githubusercontent.com/stefanpejcic/ipset-blacklist/master/setup.sh)
            else
                bash <(curl -sSL https://raw.githubusercontent.com/stefanpejcic/ipset-blacklist/master/setup.sh) > /dev/null 2>&1
            fi
        fi
        
        if [ "$SUPPORT_IPS" = true ]; then
            # Whitelisting our VPN ip addresses from https://ip.openpanel.co/ips/
            ip_list=$(curl -s https://ip.openpanel.co/ips/)
            ip_list=$(echo "$ip_list" | sed 's/<br \/>/\n/g')
        
        echo "Whitelisting IPs from https://ip.openpanel.co/ips/"

            while IFS= read -r ip; do
                ip=$(echo "$ip" | tr -d '[:space:]')
                debug_log ufw allow from $ip
            done <<< "$ip_list"
        fi

        debug_log ufw --force enable
        debug_log ufw reload
    fi
}

update_package_manager() {
    if [ "$SKIP_APT_UPDATE" = false ]; then
        echo "Updating package manager.."
        debug_log $PACKAGE_MANAGER update -y
    fi
}




install_packages() {

    echo "Installing required services.."

    # https://www.faqforge.com/linux/fixed-ubuntu-apt-get-upgrade-auto-restart-services/
    
    debug_log sed -i 's/#$nrconf{restart} = '"'"'i'"'"';/$nrconf{restart} = '"'"'a'"'"';/g' /etc/needrestart/needrestart.conf
    
    packages=("docker.io" "default-mysql-client" "nginx" "zip" "bind9" "unzip" "python3-pip" "pip" "gunicorn" "jc" "certbot" "python3-certbot-nginx" "sqlite3" "geoip-bin" "ufw")

    if [ "$PACKAGE_MANAGER" == "apt-get" ]; then
        #only once..
        debug_log $PACKAGE_MANAGER -qq install apt-transport-https ca-certificates -y
        
        echo "Updating certificates.."
        if [ "$DEBUG" = false ]; then
        update-ca-certificates > /dev/null 2>&1
        else
        update-ca-certificates
        fi

        echo -e "Installing services.."

        for package in "${packages[@]}"; do
            echo -e "Installing ${GREEN}$package${RESET}"
            debug_log $PACKAGE_MANAGER -qq install "$package" -y
        done   

        for package in "${packages[@]}"; do
            if is_package_installed "$package"; then
                echo -e "${GREEN}$package is already installed. Skipping.${RESET}"
            else
                debug_log $PACKAGE_MANAGER -qq install "$package"
                if [ $? -ne 0 ]; then
                    echo "Error: Installation of $package failed. Retrying.."
                    $PACKAGE_MANAGER -qq install "$package"
                    if [ $? -ne 0 ]; then
                    radovan 1 "ERROR: Installation failed. Please retry installation with '--retry' flag."
                        exit 1
                    fi
                fi
            fi
        done


    elif [ "$PACKAGE_MANAGER" == "yum" ]; then
        for package in "${packages[@]}"; do
            echo -e "Installing ${GREEN}$package${RESET}"
            $PACKAGE_MANAGER install "$package" -y
        done     
    elif [ "$PACKAGE_MANAGER" == "dnf" ]; then
        # MORA DRUGI ZA ALMU..
        packages=("python3-flask" "python3-pip" "docker-ce" "docker-compose" "docker-ce-cli" "mysql-client-core-8.0" "containerd.io" "docker-compose-plugin" "nginx" "zip" "unzip" "ufw" "certbot" "python3-certbot-nginx" "sqlite3" "geoip-bin")
        
        #utils must be added first, then install from that repo
        dnf install yum-utils  -y
        yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo

        #  needed for ufw and gunicorn
        dnf install epel-release

        # ovo za gunicorn
        dnf install python3-pip python3-devel gcc -y

        # bind radi ovako
        dnf install bind bind-utils -y

        for package in "${packages[@]}"; do
            echo -e "Installing  ${GREEN}$package${RESET}"
            $PACKAGE_MANAGER install "$package" -y
        done 
        #gunicorn mora preko pip na almi..
        pip3 install gunicorn flask
    else
        echo -e "${RED}Unsupported package manager: $PACKAGE_MANAGER${RESET}"
        return 1
    fi
}





configure_modsecurity() {

    # ModSecurity
    #
    # https://openpanel.co/docs/admin/settings/waf/#install-modsecurity
    #

    if [ "$MODSEC" ]; then
        echo "Installing ModSecurity and setting OWASP core ruleset.."
        debug_log opencli nginx-install_modsec
    fi
}



set_system_cronjob(){
    echo "Setting cronjobs.."
    mv ${ETC_DIR}cron /etc/cron.d/openpanel
    chown root:root /etc/cron.d/openpanel
    chmod 0600 /etc/cron.d/openpanel
}



cleanup() {
    echo "Cleaning up.."
    # https://www.faqforge.com/linux/fixed-ubuntu-apt-get-upgrade-auto-restart-services/
    sed -i 's/$nrconf{restart} = '"'"'a'"'"';/#$nrconf{restart} = '"'"'i'"'"';/g' /etc/needrestart/needrestart.conf
}





helper_function_for_nginx_on_aws_and_azure(){
    #
    # FIX FOR:
    #
    # https://stackoverflow.com/questions/3191509/nginx-error-99-cannot-assign-requested-address/13141104#13141104
    #

    # Check the status of nginx service and capture the output
    nginx_status=$(systemctl status nginx 2>&1)

    # Search for "Cannot assign requested address" in the output
    if echo "$nginx_status" | grep -q "Cannot assign requested address"; then

        # If found, append the required line to /etc/sysctl.conf
        echo "net.ipv4.ip_nonlocal_bind = 1" >> /etc/sysctl.conf
        
        # Reload the sysctl configuration
        sysctl -p /etc/sysctl.conf

        # Change the bind ip in default nginx config
        sed -i "s/IP_HERE/*/" /etc/nginx/sites-enabled/default

        debug_log "echo Configuration updated and applied."
    else
        debug_log "echo Nginx started normally."
    fi
}




set_custom_hostname(){
        if [ "$SET_HOSTNAME_NOW" = true ]; then
            # Check if the provided hostname is a valid FQDN
            if [[ $new_hostname =~ ^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
                # Check if PTR record is set to the provided hostname
                ptr=$(dig +short -x $current_ip)
                if [ "$ptr" != "$new_hostname." ]; then
                    echo "Warning: PTR record is not set to $new_hostname"
                fi
                
                # Check if A record for provided hostname points to server IP
                a_record_ip=$(dig +short $new_hostname)
                if [ "$a_record_ip" != "$current_ip" ]; then
                    echo "WARNING: A record for $new_hostname does not point to server IP: $current_ip"
                    echo "After pointing the domain run this command to set domain for panel: opencli config update force_domain $new_hostname"
                else
                    opencli config update force_domain "$new_hostname"
                fi

            else
                echo "Hostname provided: $new_hostname is not a valid FQDN, OpenPanel will use IP address $current_ip for access."
            fi

            # Set the provided hostname as the system hostname
            hostnamectl set-hostname $new_hostname
        fi
}            





opencli_setup(){
    echo "Downloading OpenCLI and adding to path.."
    echo ""
    mkdir -p /usr/local/admin/scripts

    wget -O ${TEMP_DIR}opencli.tar.gz "https://storage.googleapis.com/openpanel/0.2.1/get.openpanel.co/downloads/0.2.1/opencli/opencli-main.tar.gz" > /dev/null 2>&1 ||  radovan 1 "download failed for https://storage.googleapis.com/openpanel/0.2.1/get.openpanel.co/downloads/0.2.1/opencli/opencli-main.tar.gz"

    cd ${TEMP_DIR} && tar -xzf opencli.tar.gz -C ${TEMP_DIR}opencli
    cp -r ${TEMP_DIR}opencli/opencli-main/ /usr/local/admin/scripts/
    rm ${TEMP_DIR}opencli.tar.gz ${TEMP_DIR}opencli

    cp /usr/local/admin/scripts/opencli /usr/local/bin/opencli
    chmod +x /usr/local/bin/opencli
    chmod +x -R /usr/local/admin/scripts/
    #opencli commands
    echo "# opencli aliases
    ALIASES_FILE=\"/usr/local/admin/scripts/aliases.txt\"
    generate_autocomplete() {
        awk '{print \$NF}' \"\$ALIASES_FILE\"
    }
    complete -W \"\$(generate_autocomplete)\" opencli" >> ~/.bashrc
    
    source ~/.bashrc
}



configure_nginx() {

    # Nginx

    echo "Setting Nginx configuration.."

    # https://dev.openpanel.co/services/nginx
    rm /etc/nginx/nginx.conf && ln -s /etc/openpanel/nginx/nginx.conf /etc/nginx/nginx.conf

    # dir for domlogs
    mkdir -p /var/log/nginx/domlogs

    # 444 status for domains pointed to the IP but not added to nginx
    rm /etc/nginx/sites-available/default 
    rm /etc/nginx/sites-enabled/default
    ln -s /etc/openpanel/nginx/vhosts/default.conf /etc/nginx/sites-available/default
    ln -s /etc/openpanel/nginx/vhosts/default.conf /etc/nginx/sites-enabled/default

    # Replace IP_HERE with the value of $current_ip
    sed -i "s/listen 80;/listen $current_ip:80;/" /etc/nginx/sites-enabled/default

    # Setting pretty error pages for nginx, but need to add them inside containers also!
    mkdir /etc/nginx/snippets/  > /dev/null 2>&1

    ln -s /etc/openpanel/nginx/error_pages/snippets/error_pages.conf /etc/nginx/snippets/error_pages.conf
    ln -s /etc/openpanel/nginx/error_pages/snippets/error_pages_content.conf /etc/nginx/snippets/error_pages_content.conf

    service nginx restart
}



set_premium_features(){
 if [ "$SET_HOSTNAME_NOW" = true ]; then
    echo "Setting OpenPanel enterprise version license key $license_key"
    opencli config update key "$license_key"
 fi
}



set_email_address_and_email_admin_logins(){
        if [ "$SEND_EMAIL_AFTER_INSTALL" = true ]; then
            # Check if the provided email is valid
            if [[ $EMAIL =~ ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
                echo "Setting email address $EMAIL for notifications"
                opencli config update email "$EMAIL"
                # Send an email alert
                
                generate_random_token_one_time_only() {
                    local config_file="${OPENPANEL_DIR}conf/panel.config"
                    TOKEN_ONE_TIME="$(tr -dc 'a-zA-Z0-9' < /dev/urandom | head -c 64)"
                    local new_value="mail_security_token=$TOKEN_ONE_TIME"
                    sed -i "s|^mail_security_token=.*$|$new_value|" "${OPENPANEL_DIR}conf/panel.config"
                }

                
                email_notification() {
                  local title="$1"
                  local message="$2"
                  generate_random_token_one_time_only
                  TRANSIENT=$(awk -F'=' '/^mail_security_token/ {print $2}' "${OPENPANEL_DIR}conf/panel.config")
                                
                  SSL=$(awk -F'=' '/^ssl/ {print $2}' "${OPENPANEL_DIR}conf/panel.config")
                
                # Determine protocol based on SSL configuration
                if [ "$SSL" = "yes" ]; then
                  PROTOCOL="https"
                else
                  PROTOCOL="http"
                fi
                
                # Send email using appropriate protocol
                curl -k -X POST "$PROTOCOL://127.0.0.1:2087/send_email" -F "transient=$TRANSIENT" -F "recipient=$EMAIL" -F "subject=$title" -F "body=$message"
                
                }

                server_hostname=$(hostname)
                email_notification "OpenPanel successfully installed" "OpenAdmin URL: http://$server_hostname:2087/ | username: admin | password: $admin_password"
            else
                echo "Address provided: $EMAIL is not a valid email address. Admin login credentials and future notifications will not be sent."
            fi
        fi
}        





generate_and_set_ssl_for_panels() {
    if [ -z "$SKIP_SSL" ]; then
        echo "Checking if SSL can be generated for the server hostname.."
        debug_log opencli ssl-hostname
    fi
}

run_custom_postinstall_script() {
    if [ -n "$post_install_path" ]; then
        # run the custom script
        echo " "
        echo "Running post install script.."
        debug_log "https://dev.openpanel.co/customize.html#After-installation"
        debug_log bash $post_install_path
    fi
}

verify_license() {
    debug_log "echo Current time: $(date +%T)"
    server_hostname=$(hostname)
    license_data='{"hostname": "'"$server_hostname"'", "public_ip": "'"$current_ip"'"}'
    response=$(curl -s -X POST -H "Content-Type: application/json" -d "$license_data" https://api.openpanel.co/license-check)
    debug_log "echo Checking OpenPanel license for IP address: $current_ip"
    debug_log "echo Response: $response"
}


download_skeleton_directory_from_github(){
    echo "Downloading configuration files to ${ETC_DIR}"
    echo ""
    git clone https://github.com/stefanpejcic/openpanel-configuration ${ETC_DIR} > /dev/null 2>&1
}




send_install_log(){
    # Restore normal output to the terminal, so we dont save generated admin password in log file!
    exec > /dev/tty
    exec 2>&1
    opencli report --public >> "$LOG_FILE"
    curl -F "file=@/root/$LOG_FILE" http://support.openpanel.co/install_logs.php
    # Redirect again stdout and stderr to the log file
    exec > >(tee -a "$LOG_FILE")
    exec 2>&1
}


rm_helpers(){
    rm -rf $PROGRESS_BAR_FILE
}



setup_swap(){
    # Function to create swap file
    create_swap() {
        fallocate -l ${SWAP_FILE}G /swapfile
        chmod 600 /swapfile
        mkswap /swapfile
        swapon /swapfile
        echo "/swapfile   none    swap    sw    0   0" >> /etc/fstab
    }

    # Check if swap space already exists
    if [ -n "$(swapon -s)" ]; then
        echo "ERROR: Skipping creating swap space as there already exists a swap partition."
        return
    fi

    # Check if we should set up swap anyway
    if [ "$SETUP_SWAP_ANYWAY" = true ]; then
        create_swap
    else
        # Only create swap if RAM is less than 8GB
        memory_kb=$(grep 'MemTotal' /proc/meminfo | awk '{print $2}')
        memory_gb=$(awk "BEGIN {print $memory_kb/1024/1024}")

        if [ $(awk "BEGIN {print ($memory_gb < 8)}") -eq 1 ]; then
            create_swap
        else
            echo "Total available memory is ${memory_gb}GB, skipping creating swap file."
        fi
    fi
}




support_message() {
    echo ""
    echo "🎉 Welcome aboard and thank you for choosing OpenPanel! 🎉"
    echo ""
    echo "Your journey with OpenPanel has just begun, and we're here to help every step of the way."
    echo ""
    echo "To get started, check out our Getting Started guide:"
    echo "👉 https://openpanel.co/docs/admin/intro/#post-install-steps"
    echo ""
    echo "Need assistance or looking to learn more? We've got you covered:"
    echo ""
    echo "📚 Admin Docs: Dive into our comprehensive documentation for all things OpenPanel:"
    echo "👉 https://openpanel.co/docs/admin/intro/"
    echo ""
    echo "💬 Forums: Join our community forum to ask questions, share tips, and connect with fellow admins:"
    echo "👉 https://community.openpanel.co/"
    echo ""
    echo "🎮 Discord: For real-time chat and support, hop into our Discord server:"
    echo "👉 https://discord.openpanel.co/"
    echo ""
    echo "We're thrilled to have you with us. Let's make something amazing together! 🚀"
    echo ""
}

panel_customize(){
    if [ "$SCREENSHOTS_API_URL" == "local" ]; then
        echo "Setting the local API service for website screenshots.. (additional 1GB of disk space will be used for the self-hosted Playwright service)"
        debug_log playwright install
        debug_log playwright install-deps
        sed -i 's#screenshots=.*#screenshots=''#' "${ETC_DIR}openpanel/conf/openpanel.config" # must use '#' as delimiter
    else
        echo "Setting the remote API service '$SCREENSHOTS_API_URL' for website screenshots.."
        sed -i 's#screenshots=.*#screenshots='"$SCREENSHOTS_API_URL"'#' "${ETC_DIR}openpanel/conf/openpanel.config" # must use '#' as delimiter
    fi
}



install_openadmin(){

    # OpenAdmin
    #
    # https://openpanel.co/docs/admin/intro/
    #
    echo "Setting up Admin panel.."

    if [ "$REPAIR" = true ]; then
        rm -rf $OPENPADMIN_DIR
    fi
    
    mkdir -p $OPENPADMIN_DIR

    # Ubuntu 22
    if [ -f /etc/os-release ] && grep -q "Ubuntu 22" /etc/os-release; then   
        echo "Downloading files for Ubuntu22 and python version $current_python_version"
        git clone -b $current_python_version --single-branch https://github.com/stefanpejcic/openadmin $OPENPADMIN_DIR
        cd $OPENPADMIN_DIR
        debug_log pip install -r requirements.txt
    # Ubuntu 24
    elif [ -f /etc/os-release ] && grep -q "Ubuntu 24" /etc/os-release; then
        echo "Downloading files for Ubuntu24 and python version $current_python_version"
        git clone -b $current_python_version --single-branch https://github.com/stefanpejcic/openadmin $OPENPADMIN_DIR
        cd $OPENPADMIN_DIR
        debug_log pip install -r requirements.txt --break-system-packages
    # Debian12
    elif [ -f /etc/debian_version ]; then
        echo "Installing PIP and Git"
        apt-get install git pip -y > /dev/null 2>&1
        echo "Downloading files for Debian and python version $current_python_version"
        git clone -b debian-$current_python_version --single-branch https://github.com/stefanpejcic/openadmin $OPENPADMIN_DIR
        cd $OPENPADMIN_DIR
        debug_log pip install -r requirements.txt --break-system-packages
    # other
    else
        echo "Unsuported OS. Currently only Ubuntu22-24 and Debian11-12 are supported."
        echo 0
    fi


    
    cp -fr /usr/local/admin/service/admin.service ${SERVICES_DIR}admin.service  > /dev/null 2>&1
    
    systemctl daemon-reload  > /dev/null 2>&1
    service admin start  > /dev/null 2>&1

}


create_admin_and_show_logins_success_message() {

    #motd
    ln -s ${ETC_DIR}ssh/admin_welcome.sh /etc/profile.d/welcome.sh
    chmod +x /etc/profile.d/welcome.sh  

    #cp version file
    mkdir -p /usr/local/panel/  > /dev/null 2>&1
    docker cp openpanel:/usr/local/panel/version /usr/local/panel/version > /dev/null 2>&1
    
    echo -e "${GREEN}OpenPanel [$(cat /usr/local/panel/version)] installation complete.${RESET}"
    echo ""

    # Restore normal output to the terminal, so we dont save generated admin password in log file!
    exec > /dev/tty
    exec 2>&1

    # not saved in log!
    wget -O /tmp/generate.sh https://gist.githubusercontent.com/stefanpejcic/905b7880d342438e9a2d2ffed799c8c6/raw/a1cdd0d2f7b28f4e9c3198e14539c4ebb9249910/random_username_generator_docker.sh > /dev/null 2>&1
    source /tmp/generate.sh
    new_username=($random_name)
    new_password=$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 16)
    
    sqlite3 /etc/openpanel/openadmin/users.db "CREATE TABLE IF NOT EXISTS user (id INTEGER PRIMARY KEY, username TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL, role TEXT NOT NULL DEFAULT 'user', is_active BOOLEAN DEFAULT 1 NOT NULL);"  > /dev/null 2>&1 && 

    opencli admin new "$random_name" "$new_password"  > /dev/null 2>&1 && 

    opencli admin
    echo "Username: $random_name"
    echo "Password: $new_password"
    echo " "
    print_space_and_line
    
    # added in 0.2.0
    # email to user the new logins
    set_email_address_and_email_admin_logins

    # Redirect again stdout and stderr to the log file
    exec > >(tee -a "$LOG_FILE")
    exec 2>&1

}

# END main functions

















#####################################################################
#                                                                   #
# START main script execution                                       #
#                                                                   #
#####################################################################

print_header

parse_args "$@"

check_requirements

detect_installed_panels

check_lock_file_age

install_started_message

main

send_install_log

rm_helpers

print_space_and_line

support_message

print_space_and_line

create_admin_and_show_logins_success_message

run_custom_postinstall_script


# END main script execution







