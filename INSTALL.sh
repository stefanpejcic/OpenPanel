#!/bin/bash
################################################################################
# Script Name: INSTALL.sh
# Description: Install the latest version of OpenPanel
# Usage: cd /home && (curl -sSL https://get.openpanel.co || wget -O - https://get.openpanel.co) | bash
# Author: Stefan Pejcic
# Created: 11.07.2023
# Last Modified: 30.05.2024
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
INSTALL_TIMEOUT=1800 # 30 min
DEBUG=false
SKIP_APT_UPDATE=false
SKIP_IMAGES=false
REPAIR=false
LOCALES=true
NO_SSH=false
INSTALL_FTP=false
INSTALL_MAIL=false
OVERLAY=false
IPSETS=false
SET_HOSTNAME_NOW=false

# Paths
LOG_FILE="openpanel_install.log"
LOCK_FILE="/root/openpanel.lock"
OPENPANEL_DIR="/usr/local/panel/"
OPENPADMIN_DIR="/usr/local/admin/"
OPENCLI_DIR="/usr/local/admin/scripts/"
OPENPANEL_ERR_DIR="/var/log/openpanel/"
SERVICES_DIR="/etc/systemd/system/"
TEMP_DIR="/tmp/"

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
    echo -e "         |_|                                                  "

    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
}


install_started_message(){
    echo -e ""
    echo -e "\nStarting the installation of OpenPanel. This process will take approximately 5-10 minutes."
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

# Check if a package is already installed
is_package_installed() {
    if [ "$DEBUG" = false ]; then
    $PACKAGE_MANAGER -qq list "$1" 2>/dev/null | grep -qE "^ii"
    else
    $PACKAGE_MANAGER -qq list "$1" | grep -qE "^ii"
    echo "Updating package manager.."
    fi
}

# Get server ipv4 from ip.openpanel.co
current_ip=$(curl -s https://ip.openpanel.co || wget -qO- https://ip.openpanel.co)

# If site is not available, get the ipv4 from the hostname -I
if [ -z "$current_ip" ]; then
    current_ip=$(hostname -I | awk '{print $1}')
fi



if [ "$CUSTOM_VERSION" = false ]; then
    # Fetch the latest version
    version=$(curl -s https://update.openpanel.co/)
    if [[ $version =~ [0-9]+\.[0-9]+\.[0-9]+ ]]; then
        version=$version
    else
        version="0.1.8"
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
    detect_os_and_package_manager
    update_package_manager
    install_packages
    setup_openpanel
    setup_openadmin
    configure_docker

    configure_nginx
    configure_modsecurity

    run_mysql_docker_container
    setup_opencli
    download_and_import_docker_images
    setup_ufw
    setup_ftp
    setup_email
    install_all_locales
    helper_function_for_nginx_on_aws_and_azure

    configure_mysql

    start_services
    set_system_cronjob
    cleanup
    set_custom_hostname
    generate_and_set_ssl_for_panels
    clean_apt_cache
    verify_license
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
        allowed_versions=("39" "310" "311" "312" "38")
        if [[ ! " ${allowed_versions[@]} " =~ " ${current_python_version} " ]]; then
            echo -e "${RED}Error: Unsupported Python version $current_python_version. No corresponding branch available.${RESET}" >&2
            exit 1
        fi
    fi
}

parse_args() {
    for arg in "$@"; do
        case $arg in
            --hostname=*)
                # Extract domain after "--hostname="
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
                # Extract path after "--post_install="
                post_install_path="${1#*=}"
                ;;
            --version=*)
                # Extract path after "--version="
                CUSTOM_VERSION=true
                version="${1#*=}"
                ;;
            *)
                echo "Unknown option: $arg"
                exit 1
                ;;
        esac
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


install_all_locales() {
            
        # OpenPanel translations
        #
        # https://dev.openpanel.co/localization.html
        #

        echo "Installing FR, DE, TR locales."

        # FR
        cd ${OPENPANEL_DIR} && pybabel init -i messages.pot -d translations -l fr
        debug_log "wget -O ${OPENPANEL_DIR}translations/fr/LC_MESSAGES/messages.po https://raw.githubusercontent.com/stefanpejcic/openpanel-translations/main/fr-fr/messages.pot"

        # DE
        cd ${OPENPANEL_DIR} && pybabel init -i messages.pot -d translations -l de
        debug_log "wget -O ${OPENPANEL_DIR}translations/de/LC_MESSAGES/messages.po https://raw.githubusercontent.com/stefanpejcic/openpanel-translations/main/de-de/messages.pot"

        # TR
        cd ${OPENPANEL_DIR} && pybabel init -i messages.pot -d translations -l tr
        debug_log "wget -O ${OPENPANEL_DIR}translations/tr/LC_MESSAGES/messages.po https://raw.githubusercontent.com/stefanpejcic/openpanel-translations/main/tr-tr/messages.pot"

        pybabel compile -d translations
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
        
        debug_log "Whitelisting IPs from https://ip.openpanel.co/ips/"

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
    
    packages=("python3-flask" "python3-pip" "docker.io" "mysql-client-core-8.0" "docker-compose" "nginx" "zip" "unzip" "bind9" "gunicorn" "ufw" "jc" "certbot" "python3-certbot-nginx" "sqlite3" "geoip-bin")

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


configure_nginx() {

    # Nginx

    echo "Setting Nginx configuration.."

    # https://dev.openpanel.co/services/nginx
    debug_log cp -fr ${OPENPANEL_DIR}conf/nginx.conf /etc/nginx/nginx.conf

    # dir for domlogs
    debug_log mkdir -p /var/log/nginx/domlogs

    # 444 status for domains pointed to the IP but not added to nginx
    debug_log cp -fr ${OPENPANEL_DIR}conf/default.nginx.conf /etc/nginx/sites-enabled/default

    # Replace IP_HERE with the value of $current_ip
    debug_log sed -i "s/IP_HERE/$current_ip/" /etc/nginx/sites-enabled/default

    # Setting pretty error pages for nginx, but need to add them inside containers also!
    debug_log mkdir -p /srv/http/default
    debug_log git clone https://github.com/denysvitali/nginx-error-pages /srv/http/default   > /dev/null 2>&1

    debug_log mkdir /etc/nginx/snippets/  > /dev/null 2>&1

    debug_log ln -s /srv/http/default/snippets/error_pages.conf /etc/nginx/snippets/error_pages.conf
    debug_log ln -s /srv/http/default/snippets/error_pages_content.conf /etc/nginx/snippets/error_pages_content.conf
}


run_mysql_docker_container() {

    # MySQL

    # FIX for Ubuntu24: docker: failed to register layer: devicemapper: Error running deviceSuspend dm_task_run failed.
    #docker pull --platform linux/amd64 mysql

    # set random password
    MYSQL_ROOT_PASSWORD=$(openssl rand -base64 -hex 9)

    if [ "$REPAIR" = true ]; then
        echo "RAPAIR: Removing existing mysql database."
        docker stop openpanel_mysql
        docker rm openpanel_mysql
        docker volume rm openpanel_mysql_data
    fi

    # run the container
    docker run -d -p 3306:3306 --name openpanel_mysql \
        -e MYSQL_ROOT_PASSWORD="$MYSQL_ROOT_PASSWORD" \
        -e MYSQL_DATABASE=panel \
        -e MYSQL_USER=panel \
        -e MYSQL_PASSWORD="$MYSQL_ROOT_PASSWORD" \
        -v openpanel_mysql_data:/var/lib/mysql \
        --memory="1g" --cpus="1" \
        --restart=always \
        --oom-kill-disable \
        mysql/mysql-server > /dev/null 2>&1
    

    if docker ps -a --format '{{.Names}}' | grep -q "openpanel_mysql"; then

        # show password
        echo "Generated MySQL password: $MYSQL_ROOT_PASSWORD"

        if [ "$REPAIR" = true ]; then
            rm /etc/my.cnf
        fi
        
        ln -s /usr/local/admin/db.cnf /etc/my.cnf
        
        # Update configuration files with new password
        sed -i "s/\"mysql_password\": \".*\"/\"mysql_password\": \"${MYSQL_ROOT_PASSWORD}\"/g" /usr/local/admin/config.json
        sed -i "s/\"mysql_user\": \".*\"/\"mysql_user\": \"panel\"/g" /usr/local/admin/config.json
        sed -i "s/password = \".*\"/password = \"${MYSQL_ROOT_PASSWORD}\"/g" /usr/local/admin/db.cnf
        sed -i "s/user = \".*\"/user = \"panel\"/g" /usr/local/admin/db.cnf

    else
        radovan 1 "Installation failed! Unable to start docker container for MySQL."
    fi
}

configure_mysql() {

    # Check if the Docker container exists
    if docker ps -a --format '{{.Names}}' | grep -q "openpanel_mysql"; then

        # Fix for: ERROR 2013 (HY000): Lost connection to MySQL server at 'reading initial communication packet', system error: 2
        
        # Function to check if MySQL is running
        mysql_is_running() {
            if mysqladmin --defaults-extra-file="/usr/local/admin/db.cnf" ping &> /dev/null; then
                return 0 # MySQL is running
            else
                return 1 # MySQL is not running
            fi
        }

        # Wait for MySQL to start
        wait_for_mysql() {
            retries=5
            while [ $retries -gt 0 ]; do
                if mysql_is_running; then
                    return 0 # MySQL is running
                else
                    echo "Waiting for MySQL to start..."
                    sleep 5
                    retries=$((retries - 1))
                fi
            done
            return 1 # MySQL did not start after retries
        }

        # Wait for MySQL to start
        wait_for_mysql

        # Create database
        mysql --defaults-extra-file="/usr/local/admin/db.cnf" -e "CREATE DATABASE IF NOT EXISTS panel;"
        #mysql --defaults-extra-file="/usr/local/admin/db.cnf" -e "GRANT PROCESS ON *.* TO 'panel'@'%';"
        mysql --defaults-extra-file="/usr/local/admin/db.cnf" -D "panel" < ${OPENPANEL_DIR}DATABASE.sql

        # Check if SQL file was imported successfully
        if mysql --defaults-extra-file="/usr/local/admin/db.cnf" -D "panel" -e "SELECT 1 FROM plans LIMIT 1;" &> /dev/null; then
            echo -e "${GREEN}Database is ready.${RESET}"
        else
            echo "SQL file import failed or database is not ready."
            radovan 1 "Installation failed!"
        fi

    else
        echo "Docker container 'openpanel_mysql' does not exist. Please make sure the container is running."
        radovan 1 "Installation failed! "
    fi
}


configure_docker() {


    docker_daemon_json_path="/etc/docker/daemon.json"
    debug_log mkdir -p $(dirname "$docker_daemon_json_path")

    # Docker
if [ "$OVERLAY" = true ]; then
    debug_log "Setting default storage driver for Docker from to 'overlay2'.."


    ### to be removed in 0.1.8
    daemon_json_content='{
      "storage-driver": "overlay2",
      "log-driver": "local",
      "log-opts": {
         "max-size": "5m"
         }
      }'
    echo "$daemon_json_content" > "$docker_daemon_json_path"
    ###

else
    debug_log "Changing default storage driver for Docker from 'overlay2' to 'devicemapper'.."


    ### to be removed in 0.1.8
    daemon_json_content='{
      "storage-driver": "devicemapper",
      "log-driver": "local",
      "log-opts": {
         "max-size": "5m"
         }
      }'
    echo "$daemon_json_content" > "$docker_daemon_json_path"
    ###

fi


    # to be used in 0.1.8+
    #cp /usr/local/panel/conf/docker.daemon.json $docker_daemon_json_path


    echo -e "${GREEN}Docker is configured.${RESET}"
    debug_log systemctl daemon-reload
    systemctl restart docker
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

setup_openpanel() {

    if [ "$REPAIR" = true ]; then
        rm -rf $OPENPANEL_DIR
    fi

    # OpenPanel
    #
    # https://openpanel.co/docs/panel/intro/
    #
    echo "Setting up User panel.."

    # check python version again, in case if --skip-requirements or --repair flags are used
    current_python_version=$(python3 --version 2>&1 | cut -d " " -f 2 | cut -d "." -f 1,2 | tr -d '.')
    
    mkdir -p $OPENPANEL_DIR
    mkdir -p ${OPENPANEL_DIR}users

    # Clone the git branch for that python version
    wget -O ${TEMP_DIR}openpanel.tar.gz "https://storage.googleapis.com/openpanel/$version/get.openpanel.co/downloads/$version/openpanel/$current_python_version/compressed.tar.gz" > /dev/null 2>&1 || radovan 1 "wget failed for https://storage.googleapis.com/openpanel/$version/get.openpanel.co/downloads/$version/openpanel/$current_python_version/compressed.tar.gz"
    cd ${TEMP_DIR} && tar -xzf openpanel.tar.gz -C $OPENPANEL_DIR
    rm ${TEMP_DIR}openpanel.tar.gz

    export PYTHONPATH=$OPENPANEL_DIR:$PYTHONPATH

    cd $OPENPANEL_DIR
    cp -fr services/panel.service ${SERVICES_DIR}panel.service

    echo "Installing PIP requirements for User panel.."

    # FIX FOR: https://peps.python.org/pep-0668/
    if [[ ($current_python_version == "311" || $current_python_version == "312") ]]; then
        debug_log "Installing PIP requirements for OpenPanel with break-system-packages..."
        debug_log pip install -r requirements.txt --break-system-packages
    else
        debug_log "Installing PIP requirements for OpenPanel without break-system-packages..."
        debug_log pip install -r requirements.txt
    fi


    echo "Setting the API service for website screenshots.."
    debug_log playwright install
    debug_log playwright install-deps

    mv ${OPENPANEL_DIR}icons/ ${OPENPANEL_DIR}static/images/icons
}

setup_openadmin() {

    # OpenAdmin
    #
    # https://openpanel.co/docs/admin/intro/
    #
    echo "Setting up Admin panel.."

    if [ "$REPAIR" = true ]; then
        rm -rf $OPENPADMIN_DIR
    fi


    mkdir -p $OPENPADMIN_DIR

    # Clone the branch for that python version
    wget -O ${TEMP_DIR}openadmin.tar.gz "https://storage.googleapis.com/openpanel/$version/get.openpanel.co/downloads/$version/openadmin/$current_python_version/compressed.tar.gz" > /dev/null 2>&1 || radovan 1 "wget failed for https://storage.googleapis.com/openpanel/$version/get.openpanel.co/downloads/$version/openadmin/$current_python_version/compressed.tar.gz"
    debug_log cd ${TEMP_DIR}
    debug_log tar -xzf openadmin.tar.gz -C $OPENPADMIN_DIR
    debug_log unzip ${OPENPADMIN_DIR}static/dist.zip -d ${OPENPADMIN_DIR}static/dist/
    debug_log rm ${TEMP_DIR}openadmin.tar.gz ${OPENPADMIN_DIR}static/dist.zip

    debug_log cd $OPENPADMIN_DIR

    cp -fr service/admin.service ${SERVICES_DIR}admin.service

    # Fix for: ModuleNotFoundError: No module named 'pyarmor_runtime_000000'
    wget -O ${OPENPADMIN_DIR}service/service.config.py https://gist.githubusercontent.com/stefanpejcic/37805c6781dc3beb1730fec82ee5ae34/raw/d7e8a6c1608c265aed89e97dcecea518b222ac86/service.config.py > /dev/null 2>&1


    echo "Installing PIP requirements for Admin panel.."
    # FIX FOR: https://peps.python.org/pep-0668/
    if [[ ($current_python_version == "311" || $current_python_version == "312") ]]; then
        debug_log "Installing PIP requirements for OpenAdmin with break-system-packages..."
        debug_log pip install -r requirements.txt --break-system-packages
    else
        debug_log "Installing PIP requirements for OpenAdmin without break-system-packages..."
        debug_log pip install -r requirements.txt
    fi

    echo "Creating Admin user.."

    touch ${OPENPADMIN_DIR}users.db

    export PYTHONPATH=$OPENPADMIN_DIR:$PYTHONPATH

    admin_password=$(openssl rand -base64 12 | tr -d '=+/')
    password_hash=$(python3 ${OPENPADMIN_DIR}core/users/hash $admin_password)

    debug_log sqlite3 ${OPENPADMIN_DIR}users.db "CREATE TABLE IF NOT EXISTS user (id INTEGER PRIMARY KEY, username TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL, role TEXT NOT NULL DEFAULT 'user', is_active BOOLEAN DEFAULT 1 NOT NULL);" "INSERT INTO user (username, password_hash, role) VALUES ('admin', \"$password_hash\", 'admin');"

    # added in 0.1.9
    cp helpers/welcome.sh /etc/profile.d/welcome.sh
    chmod +x /etc/profile.d/welcome.sh  

}



set_system_cronjob(){

    echo "Setting cronjobs.."

    mv /usr/local/panel/conf/cron /etc/cron.d/openpanel
    chown root:root /etc/cron.d/openpanel
    chmod 0600 /etc/cron.d/openpanel
}

setup_opencli() {

    # OpenCLI
    #
    # https://dev.openpanel.co/cli/
    #
    echo "Setting OpenPanel CLI scripts.."

    if [ "$REPAIR" = true ]; then
        rm -rf $OPENCLI_DIR
    fi

    debug_log mkdir -p $OPENCLI_DIR

    wget -O ${TEMP_DIR}opencli.tar.gz "https://storage.googleapis.com/openpanel/$version/get.openpanel.co/downloads/$version/opencli/compressed.tar.gz" > /dev/null 2>&1 || radovan 1 "wget failed for https://storage.googleapis.com/openpanel/$version/get.openpanel.co/downloads/$version/opencli/compressed.tar.gz"

    debug_log cd ${TEMP_DIR} && tar -xzf opencli.tar.gz -C $OPENCLI_DIR
    debug_log rm ${TEMP_DIR}opencli.tar.gz
    debug_log bash ${OPENCLI_DIR}install.sh
    debug_log source ~/.bashrc
    echo -e "${GREEN}opencli commands are now available.${RESET}"
    debug_log chmod +x -R ${OPENCLI_DIR}


   # added in 0.1.8
    mkdir -p /etc/openpanel/openadmin/config/
    wget -O /etc/openpanel/openadmin/config/forbidden_usernames.txt  https://gist.githubusercontent.com/stefanpejcic/f08e6841fbf953b7aff108a8c154e9eb/raw/9ac3efabbde48faf95435221e7e09e28b46340ae/forbidden_usernames.ttxt > /dev/null 2>&1

    echo "Creating directories for logs.."
    debug_log mkdir -p ${OPENPANEL_ERR_DIR}admin ${OPENPANEL_ERR_DIR}user
    debug_log -e "${GREEN}OpenCLI has been successfully installed.${RESET}"

}

cleanup() {
    echo "Cleaning up.."
    # https://www.faqforge.com/linux/fixed-ubuntu-apt-get-upgrade-auto-restart-services/
    sed -i 's/$nrconf{restart} = '"'"'a'"'"';/#$nrconf{restart} = '"'"'i'"'"';/g' /etc/needrestart/needrestart.conf
    rm -rf ${OPENPANEL_DIR}INSTALL.sh ${OPENPANEL_DIR}DATABASE.sql ${OPENPANEL_DIR}requirements.txt ${OPENPANEL_DIR}conf/nginx.conf ${OPENPANEL_DIR}conf/mysqld.cnf ${OPENPANEL_DIR}conf/named.conf.options ${OPENPANEL_DIR}conf/default.nginx.conf ${OPENPANEL_DIR}conf/.gitkeep ${OPENPANEL_DIR}templates/dist.zip ${OPENPANEL_DIR}install
    rm -rf ${OPENPANEL_DIR}services
}


start_services() {
    echo "Starting services.."
    systemctl restart panel admin nginx ufw
    systemctl enable panel admin nginx ufw
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

download_and_import_docker_images() {
    echo "Downloading docker images in the background.."

    if [ "$SKIP_IMAGES" = false ]; then
        # See https://github.com/moby/moby/issues/16106#issuecomment-310781836 for pulling images in parallel
        # Nohup (no hang up) with trailing ampersand allows running the command in the background
        # The </dev/null bits redirects the outputs to files rather than strout/err
        nohup sh -c "echo openpanel/nginx:latest openpanel/apache:latest | xargs -P10 -n1 docker pull" </dev/null >nohup.out 2>nohup.err &
        # stopped working on 0.1.8 :(
        # opencli docker-update_images &
        # pid1=$!
        # disown $pid1
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



success_message() {

    echo -e "${GREEN}OpenPanel installation complete.${RESET}"
    echo ""

    # Restore normal output to the terminal, so we dont save generated admin password in log file!
    exec > /dev/tty
    exec 2>&1

    opencli admin
    echo "Username: admin"
    echo "Password: $admin_password"
    echo " "
    print_space_and_line

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

success_message

run_custom_postinstall_script


# END main script execution
