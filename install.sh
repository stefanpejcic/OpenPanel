
#!/bin/bash
################################################################################
#
# Install the latest version of OpenPanel ✌️
# https://openpanel.com/install
#
# Supported OS:            Ubuntu, Debian, AlmaLinux, RockyLinux, CentOS
# Supported Python         3.8 3.9 3.10 3.11 3.12
#
# Usage:                   bash <(curl -sSL https://openpanel.org)
# Author:                  Stefan Pejcic <stefan@pejcic.rs>
# Created:                 11.07.2023
# Last Modified:           11.12.2024
#
################################################################################



# ======================================================================
# Constants
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
RESET='\033[0m'
export TERM=xterm-256color                                            # bug fix: tput: No value for $TERM and no -T specified

# ======================================================================
# Defaults for environment variables
CUSTOM_VERSION=false                                                  # default version is latest
INSTALL_TIMEOUT=600                                                   # after 10min, consider the install failed
DEBUG=false                                                           # verbose output for debugging failed install
SKIP_APT_UPDATE=false
SKIP_IMAGES=false                                                     # they are auto-pulled on account creation
REPAIR=false
LOCALES=true                                                          # only en
NO_SSH=false                                                          # deny port 22
IPSETS=true                                                           # currently only works with ufw
SET_HOSTNAME_NOW=false                                                # must be a FQDN
CUSTOM_GB_DOCKER=false                                                # space in gb, if not set fallback to 50% of available du
SETUP_SWAP_ANYWAY=false
SWAP_FILE="1"                                                         # calculated based on ram
SEND_EMAIL_AFTER_INSTALL=false 
SET_PREMIUM=false                                                     # added in 0.2.1
UFW_SETUP=false                                                       # previous default on <0.2.3
CSF_SETUP=true                                                        # default since >0.2.2
SET_ADMIN_USERNAME=false                                              # random
SET_ADMIN_PASSWORD=false                                              # random
SCREENSHOTS_API_URL="http://screenshots-api.openpanel.com/screenshot" # default since 0.2.1

# ======================================================================
# PATHs used throughout the script
ETC_DIR="/etc/openpanel/"                                             # https://github.com/stefanpejcic/openpanel-configuration
LOG_FILE="openpanel_install.log"                                      # install log
LOCK_FILE="/root/openpanel.lock"                                      # install running
OPENPANEL_DIR="/usr/local/panel"                                      # currently only used to store version
OPENPADMIN_DIR="/usr/local/admin/"                                    # https://github.com/stefanpejcic/openadmin/branches
OPENCLI_DIR="/usr/local/admin/scripts/"                               # https://dev.openpanel.com/cli/commands.html
OPENPANEL_ERR_DIR="/var/log/openpanel/"                               # https://dev.openpanel.com/logs.html
SERVICES_DIR="/etc/systemd/system/"                                   # used for admin, sentinel and floatingip services
CONFIG_FILE="${ETC_DIR}openpanel/conf/openpanel.config"               # main config file for openpanel

exec > >(tee -a "$LOG_FILE") 2>&1

# ======================================================================
# Helper functions that are not mandatory and should not be modified

setup_terminal() {
  TPUT_RESET=""
  TPUT_WHITE=""
  TPUT_BGRED=""
  TPUT_BGGREEN=""
  TPUT_BOLD=""
  TPUT_DIM=""

  test -t 2 || return 1

  if command -v tput > /dev/null 2>&1; then
    if num_colors=$(tput colors 2> /dev/null) && [ "${num_colors:-0}" -ge 8 ]; then
      TPUT_RESET="$(tput sgr 0)"
      TPUT_WHITE="$(tput setaf 7)"
      TPUT_BGRED="$(tput setab 1)"
      TPUT_BGGREEN="$(tput setab 2)"
      TPUT_BOLD="$(tput bold)"
      TPUT_DIM="$(tput dim)"
    fi
  fi

  echo "${TPUT_RESET}"

  return 0
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
    echo -e "         |_|                                   version: ${GREEN}$PANEL_VERSION${RESET} "
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
}



install_started_message(){
    echo -e "\nStarting the installation of OpenPanel. This process will take approximately 3-5 minutes."
    echo -e "During this time, we will:"
    if [ "$CSF_SETUP" = true ]; then
    	echo -e "- Install necessary services and tools: CSF, Docker, MySQL, SQLite, Python3, PIP.. "
    elif [ "$UFW_SETUP" = true ]; then
    	echo -e "- Install necessary services and tools: UFW, Docker, MySQL, SQLite, Python3, PIP.. "
    else
	echo -e "- Install necessary services and tools: Docker, MySQL, SQLite, Python3, PIP.. "
    fi
    if [ "$SET_ADMIN_USERNAME" = true ]; then
        if [ "$SET_ADMIN_PASSWORD" = true ]; then
		echo -e "- Create an admin account $custom_username with password $custom_password for you."
	else
		echo -e "- Create an admin account $custom_username with a strong random password for you."
  	fi
    else
	echo -e "- Create an admin account with random username and strong password for you."
    fi    
    if [ "$CSF_SETUP" = true ]; then
    	echo -e "- Set up ConfigServer Firewall for enhanced security."
    elif [ "$UFW_SETUP" = true ]; then
    	echo -e "- Set up Uncomplicated Firewall for enhanced security."
    fi
    echo -e "- Install needed Docker images."
    if [ "$SET_PREMIUM" = true ]; then
    	echo -e "- Set up 4 hosting plans (Nginx with MySQL, Nginx with MariaDB, Apache with MySQL, Apache with MariaDB) so you can start right away."
    else
    	echo -e "- Set up 2 hosting plans (Nginx and Apache) so you can start right away."
    fi
    echo -e "\nThank you for your patience. We're setting everything up for your seamless OpenPanel experience!\n"
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
    echo -e ""
}



# Display error and exit
radovan() {
    printf >&2 "%s\n\n" "${TPUT_BGRED}${TPUT_WHITE}${TPUT_BOLD} FAILED ${TPUT_RESET}"
    echo -e "Error: $2$" >&2
    exit 1
}


debug_log() {
    local timestamp
    timestamp=$(date +'%Y-%m-%d %H:%M:%S')

    if [ "$DEBUG" = true ]; then
        # Show both on terminal and log file
        echo "[$timestamp] $message" | tee -a "$LOG_FILE"
        "$@" 2>&1 | tee -a "$LOG_FILE"
    else
        # No terminal output, only log file
        echo "[$timestamp] COMMAND: $@" >> "$LOG_FILE"
        "$@" > /dev/null 2>&1
    fi
}

# Check if a package is already installed
is_package_installed() {
    if [ "$DEBUG" = false ]; then
    $PACKAGE_MANAGER -qq list "$1" 2>/dev/null | grep -qE "^ii"
    else
    $PACKAGE_MANAGER -qq list "$1" | grep -qE "^ii"
    echo "Updating $PACKAGE_MANAGER package manager.."
    fi
}


detect_filesystem(){
	FS_TYPE=$(df -T "/var" | awk 'NR==2 {print $2}')
}

get_server_ipv4(){
	# Get server ipv4
 
	# list of ip servers for checks
	IP_SERVER_1="https://ip.openpanel.com"
	IP_SERVER_2="https://ipv4.openpanel.com"
	IP_SERVER_3="https://ifconfig.me"

	current_ip=$(curl --silent --max-time 2 -4 $IP_SERVER_1 || \
                 wget --timeout=2 -qO- $IP_SERVER_2 || \
                 curl --silent --max-time 2 -4 $IP_SERVER_3)

	# If no site is available, get the ipv4 from the hostname -I
	if [ -z "$current_ip" ]; then
	    # ip addr command is more reliable then hostname - to avoid getting private ip
	    current_ip=$(ip addr|grep 'inet '|grep global|head -n1|awk '{print $2}'|cut -f1 -d/)
	fi

	    is_valid_ipv4() {
	        local ip=$1
	        # is it ip
	        [[ $ip =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}$ ]] && \
	        # is it private
	        ! [[ $ip =~ ^10\. ]] && \
	        ! [[ $ip =~ ^172\.(1[6-9]|2[0-9]|3[0-1])\. ]] && \
	        ! [[ $ip =~ ^192\.168\. ]]
	    }
	
	   
	if ! is_valid_ipv4 "$current_ip"; then
	        echo "Invalid or private IPv4 address: $current_ip. OpenPanel requires a public IPv4 address to bind Nginx configuration files."
	fi

}

set_version_to_install(){

	if [ "$CUSTOM_VERSION" = false ]; then
	    PANEL_VERSION=$(curl --silent --max-time 10 -4 https://openpanel.org/version)
	    if [[ $PANEL_VERSION =~ [0-9]+\.[0-9]+\.[0-9]+ ]]; then
	        PANEL_VERSION=$PANEL_VERSION
	    else
	        PANEL_VERSION="0.3.7"
	    fi
	fi
}


# prints fullwidth line
print_space_and_line() {
    echo " "
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
    echo " "
}


setup_progress_bar_script(){
	# Progress bar script
	PROGRESS_BAR_URL="https://raw.githubusercontent.com/pollev/bash_progress_bar/master/progress_bar.sh"
	PROGRESS_BAR_FILE="progress_bar.sh"
 
	# Check if wget is available
	if command -v wget &> /dev/null; then
	    wget "$PROGRESS_BAR_URL" -O "$PROGRESS_BAR_FILE" > /dev/null 2>&1
	# If wget is not available, check if curl is available *(fallback for fedora)
	elif command -v curl &> /dev/null; then
	    curl -s "$PROGRESS_BAR_URL" -o "$PROGRESS_BAR_FILE" > /dev/null 2>&1
	else
	    echo "Neither wget nor curl is available. Please install one of them to proceed."
	    exit 1
	fi
 
	if [ ! -f "$PROGRESS_BAR_FILE" ]; then
	    echo "ERROR: Failed to download progress_bar.sh - Github is not reachable by your server: https://raw.githubusercontent.com"
	    exit 1
	fi
}




display_what_will_be_installed(){
 	echo -e "[ OK ] DETECTED OPERATING SYSTEM: ${GREEN} ${NAME^^} $VERSION_ID ${RESET}"
    	if [ -z "$SKIP_REQUIREMENTS" ]; then
	if [ "$architecture" == "x86_64" ]; then
  	echo -e "[ OK ] CPU ARCHITECTURE:          ${GREEN} ${architecture^^} ${RESET}"
	elif [ "$architecture" == "aarch64" ]; then
  	echo -e "[PASS] CPU ARCHITECTURE:          ${YELLOW} ${architecture^^} ${RESET}"
   	else
      	echo -e "[PASS] CPU ARCHITECTURE:          ${YELLOW} ${architecture^^} ${RESET}"
 	fi
  	fi
 	echo -e "[ OK ] PACKAGE MANAGEMENT SYSTEM: ${GREEN} ${PACKAGE_MANAGER^^} ${RESET}"
 	echo -e "[ OK ] INSTALLED PYTHON VERSION:  ${GREEN} ${current_python_version} ${RESET}"
	if [ "$FS_TYPE" = "xfs" ]; then
 	echo -e "[ OK ] BACKING FILESYSTEM TYPE:   ${GREEN} ${FS_TYPE^^} ${RESET}"
	else 
  	echo -e "[PASS] BACKING FILESYSTEM TYPE:   ${YELLOW} ${FS_TYPE^^} ${RESET}"
 	fi
 	echo -e "[ OK ] PUBLIC IPV4 ADDRESS:       ${GREEN} ${current_ip} ${RESET}"
  	echo ""

}




# ======================================================================
# Core program logic
setup_progress_bar_script
source "$PROGRESS_BAR_FILE"               # Source the progress bar script

FUNCTIONS=(
detect_os_and_package_manager             # detect os and package manager
display_what_will_be_installed            # display os, version, ip
update_package_manager                    # update dnf/yum/apt-get
install_packages                          # install docker, csf/ufw, sqlite, etc.
download_skeleton_directory_from_github   # download configuration to /etc/openpanel/
setup_bind                                # must run after -configuration
install_openadmin                         # set admin interface
opencli_setup                             # set terminal commands
panel_customize                           # customizations
docker_compose_up                         # must be after configure_nginx
set_premium_features                      # must be after docker_compose_up
configure_modsecurity                     # TEMPORARY OFF FROM 0.2.5
set_custom_hostname                       # set hostname if provided
generate_and_set_ssl_for_panels           # if FQDN then lets setup https
setup_firewall_service                    # setup firewall
set_system_cronjob                        # setup crons, must be after csf
set_logrotate                             # setup logrotate, ignored on fedora
tweak_ssh                                 # basic ssh
setup_swap                                # swap space
clean_apt_and_dnf_cache                   # clear
verify_license                            # ping our server
)


TOTAL_STEPS=${#FUNCTIONS[@]}
CURRENT_STEP=0

update_progress() {
    CURRENT_STEP=$((CURRENT_STEP + 1))
    PERCENTAGE=$(($CURRENT_STEP * 100 / $TOTAL_STEPS))
    draw_progress_bar $PERCENTAGE
}

main() {
    enable_trapping                       # clean on CTRL+C
    setup_scroll_area                     # load progress bar
    for func in "${FUNCTIONS[@]}"
    do
        $func                             # Execute each function
        update_progress                   # update progress after each
    done
    destroy_scroll_area
}




check_requirements() {
    if [ -z "$SKIP_REQUIREMENTS" ]; then

        architecture=$(lscpu | grep Architecture | awk '{print $2}')

        if [ "$architecture" == "aarch64" ]; then
	    # https://github.com/stefanpejcic/openpanel/issues/63 
            echo -e "${RED}ERROR: ARM CPU architecture is not yet supported! Feature request: https://github.com/stefanpejcic/openpanel/issues/63 ${RESET}"
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
    fi
    
    # check if python version is supported
    current_python_version=$(python3 --version 2>&1 | cut -d " " -f 2 | cut -d "." -f 1,2 | tr -d '.')
    allowed_versions=("39" "310" "311" "312" "38")
    if [[ ! " ${allowed_versions[@]} " =~ " ${current_python_version} " ]]; then
       echo -e "${RED}Error: Unsupported Python version $current_python_version. No corresponding branch available.${RESET}" >&2
       exit 1
    fi
}


parse_args() {

# ======================================================================
# Usage info
    show_help() {
        echo "Available options:"
        echo "  --key=<key_here>                Set the license key for OpenPanel Enterprise edition."
        echo "  --hostname=<hostname>           Set the hostname - must be FQDN, example: server.example.net."
        echo "  --username=<username>           Set Admin username - random generated if not provided."
        echo "  --password=<password>           Set Admin Password - random generated if not provided."
        echo "  --version=<version>             Set a custom OpenPanel version to be installed."
        echo "  --email=<stefan@example.net>    Set email address to receive email with admin credentials and future notifications."
        echo "  --skip-requirements             Skip the requirements check."
        echo "  --skip-panel-check              Skip checking if existing panels are installed."
        echo "  --skip-apt-update               Skip the APT update."
        echo "  --skip-firewall                 Skip installing UFW or CSF - Only do this if you will set another external firewall!"
        echo "  --csf                           Install and setup ConfigServer Firewall  (default from >0.2.3)"
        echo "  --ufw                           Install and setup Uncomplicated Firewall (was default in <0.2.3)"
        echo "  --skip-images                   Skip installing openpanel/nginx and openpanel/apache docker images."
        echo "  --skip-blacklists               Do not set up IP sets and blacklists."
        echo "  --skip-ssl                      Skip SSL setup."
        echo "  --with_modsec                   Enable ModSecurity for Nginx."
        echo "  --no-ssh                        Disable port 22 and whitelist the IP address of user installing the panel."
        echo "  --post_install=<path>           Specify the post install script path."
        echo "  --screenshots=<url>             Set the screenshots API URL."
        echo "  --swap=<2>                      Set space in GB to be allocated for SWAP."
        echo "  --docker-space=<2>              Set space in GB to be allocated for Docker containers (default 50% of available storage)."
        echo "  --debug                         Display debug information during installation."
        echo "  --repair                        Retry and overwrite everything."
        echo "  -h, --help                      Show this help message and exit."
    }






# ======================================================================
# Change defaults
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
        --username=*)
            SET_ADMIN_USERNAME=true
            custom_username="${1#*=}"
            ;;
        --password=*)
            SET_ADMIN_PASSWORD=true
            custom_password="${1#*=}"
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
            #SKIP_REQUIREMENTS=true
            ;;
        --skip-firewall)
            SKIP_FIREWALL=true
            ;;
        --csf)
            UFW_SETUP=false
            CSF_SETUP=true
            ;;
        --ufw)
            UFW_SETUP=true
            CSF_SETUP=false
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
        --no-ssh)
            NO_SSH=true
            ;;
        --post_install=*)
            post_install_path="${1#*=}"
            ;;
        --screenshots=*)
            SCREENSHOTS_API_URL="${1#*=}"
            ;;
        --version=*)
            CUSTOM_VERSION=true
            PANEL_VERSION="${1#*=}"
            ;;
        --swap=*)
            SETUP_SWAP_ANYWAY=true
            SWAP="${1#*=}"
            ;;
        --docker-space=*)
            CUSTOM_GB_DOCKER=true
            SPACE_FOR_DOCKER_FILE="${1#*=}"
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
            ["$OPENPANEL_DIR"]="You already have OpenPanel installed. ${RESET}\nInstead, did you want to update? Run ${GREEN}'opencli update --force' to update OpenPanel."
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

        echo -e "${GREEN}No currently installed hosting control panels or webservers found. Starting the installation process.${RESET}"
    fi
}



detect_os_and_package_manager() {
    if [ -f "/etc/os-release" ]; then
        . /etc/os-release

        case $ID in
            ubuntu)
                PACKAGE_MANAGER="apt-get"
                py_enchoded_for_distro="$current_python_version"
                ;;
            debian)
                PACKAGE_MANAGER="apt-get"
                py_enchoded_for_distro="debian-$current_python_version"
                ;;
            fedora)
                PACKAGE_MANAGER="dnf"
                py_enchoded_for_distro="$current_python_version"
                ;;
            rocky)
                PACKAGE_MANAGER="dnf"
                py_enchoded_for_distro="$current_python_version"
                ;;
            centos)
                PACKAGE_MANAGER="yum"
                py_enchoded_for_distro="$current_python_version"
                ;;
            almalinux|alma)
                PACKAGE_MANAGER="dnf"
                py_enchoded_for_distro="$current_python_version"
                ;;
            *)
                echo -e "${RED}Unsupported Operating System: $ID. Exiting.${RESET}"
                echo -e "${RED}INSTALL FAILED${RESET}"
                exit 1
                ;;
        esac
	 
    else
        echo -e "${RED}Could not detect Linux distribution from /etc/os-release${RESET}"
        echo -e "${RED}INSTALL FAILED${RESET}"
        exit 1
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





docker_compose_up(){
    echo "Setting docker-compose.."
    # install docker compose on dnf
    DOCKER_CONFIG=${DOCKER_CONFIG:-$HOME/.docker}
    mkdir -p $DOCKER_CONFIG/cli-plugins
    curl -SL https://github.com/docker/compose/releases/download/v2.27.1/docker-compose-linux-x86_64 -o $DOCKER_CONFIG/cli-plugins/docker-compose  > /dev/null 2>&1
    debug_log chmod +x $DOCKER_CONFIG/cli-plugins/docker-compose

        architecture=$(lscpu | grep Architecture | awk '{print $2}')

	if [ "$architecture" == "aarch64" ]; then
		# need to download compose and add it as alias
		debug_log curl -L "https://github.com/docker/compose/releases/download/v2.30.3/docker-compose-$(uname -s)-$(uname -m)"  -o /usr/local/bin/docker-compose
		debug_log mv /usr/local/bin/docker-compose /usr/bin/docker-compose
		debug_log chmod +x /usr/bin/docker-compose   

		function_to_insert='docker() {
		  if [[ $1 == "compose" ]]; then
		    /usr/local/bin/docker-compose "${@:2}"
		  else
		    command docker "$@"
		  fi
		}'
		
		# Check which shell configuration file to edit
		if [ -f "$HOME/.bashrc" ]; then
		    config_file="$HOME/.bashrc"
		elif [ -f "$HOME/.zshrc" ]; then
		    config_file="$HOME/.zshrc"
		else
		    radovan 1 "ERROR: Neither .bashrc nor .zshrc file found. Exiting."
		fi
		
		# Check if the function already exists in the config file
		if grep -q "docker() {" "$config_file"; then
		    :
		else
		    # Add the function to the configuration file
		    echo "$function_to_insert" >> "$config_file"
		    debug_log "Function 'docker' has been added to $config_file."
		    source "$config_file"
		fi
  
        fi











 
    if [ "$SET_PREMIUM" = true ]; then
    	cp /etc/openpanel/mysql/initialize/0.4/mariadb_plans.sql /root/initialize.sql  > /dev/null 2>&1
    else
    	cp /etc/openpanel/mysql/initialize/0.4/mysql_plans.sql /root/initialize.sql  > /dev/null 2>&1
    fi

    # compose doesnt alllow /
    cd /root
    rm -rf /etc/my.cnf .env > /dev/null 2>&1 # on centos we get default my.cnf, and on repair we already have symlink and .env

    # generate random password for mysql
    MYSQL_ROOT_PASSWORD=$(openssl rand -base64 -hex 9)
    echo "MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD" > .env
    echo "MYSQL_ROOT_PASSWORD = $MYSQL_ROOT_PASSWORD"
        
    # save it to /etc/my.cnf
    ln -s /etc/openpanel/mysql/db.cnf /etc/my.cnf  > /dev/null 2>&1
    sed -i 's/password = .*/password = '"${MYSQL_ROOT_PASSWORD}"'/g' ${ETC_DIR}mysql/db.cnf  > /dev/null 2>&1
    
    cp /etc/openpanel/docker/compose/new-docker-compose.yml /root/docker-compose.yml > /dev/null 2>&1 # from 0.2.5  new-docker-compose.yml instead of docker-compose.yml
    
    # added in 0.2.9
    # fix for bug with mysql-server image on Almalinux 9.2
    os_name=$(grep ^ID= /etc/os-release | cut -d'=' -f2 | tr -d '"')
    if [ "$os_name" == "almalinux" ]; then
        sed -i 's/mysql\/mysql-server/mysql/g' /root/docker-compose.yml
        echo "mysql/mysql-server docker image has known issues on AlmaLinux - editing docker compose to use the mysql:latest instead"
    elif [ "$os_name" == "debian" ]; then
    	echo "Setting AppArmor profiles for Debian"
   	apt install apparmor -y   > /dev/null 2>&1
    fi


    if [ "$REPAIR" = true ]; then
    	echo "Deleting all existing MySQL data in volume root_openadmin_mysql due to the '--repair' flag."
	cd /root && docker compose down > /dev/null 2>&1          # in case mysql was running
        docker volume rm root_openadmin_mysql > /dev/null 2>&1    # delete database
    fi

   
    # from 0.2.5 we only start mysql by default
    cd /root && docker compose up -d openpanel_mysql > /dev/null 2>&1

    # check if compose started the mysql container, and if is currently running
    	mysql_container=$(docker compose ps -q openpanel_mysql)
	if [ -z `docker ps -q --no-trunc | grep "$mysql_container"` ]; then
        	echo "ERROR: MySQL container is not running. Please retry installation with '--retry' flag."
	else
		echo -e "[${GREEN} OK ${RESET}] MySQL service started successfuly"
	fi
}




clean_apt_and_dnf_cache(){

     if [ "$PACKAGE_MANAGER" == "dnf" ]; then
	    	dnf clean  > /dev/null > /dev/null 2>&1 
      elif [ "$PACKAGE_MANAGER" == "apt-get" ]; then
      		# clear /var/cache/apt/archives/   # TODO: cover https://github.com/debuerreotype/debuerreotype/issues/95
      		apt-get clean  > /dev/null > /dev/null 2>&1 
	fi
}

tweak_ssh(){
   echo "Tweaking SSH service.."

   sed -i "s/[#]LoginGraceTime [[:digit:]]m/LoginGraceTime 1m/g" /etc/ssh/sshd_config
   
   if [ "$PACKAGE_MANAGER" == "apt-get" ]; then
	   if [ -z "$(grep "^DebianBanner no" /etc/ssh/sshd_config)" ]; then
		   sed -i '/^[#]Banner .*/a DebianBanner no' /etc/ssh/sshd_config
		   if [ -z "$(grep "^DebianBanner no" /etc/ssh/sshd_config)" ]; then
			   echo '' >> /etc/ssh/sshd_config # fallback
			   echo 'DebianBanner no' >> /etc/ssh/sshd_config
		   fi
	   fi
   fi


	# ssh on debian, sshd on rhel
	if [ "$PACKAGE_MANAGER" == "dnf" ] || [ "$PACKAGE_MANAGER" == "yum" ]; then
	 	systemctl restart sshd  > /dev/null 2>&1
	else
		systemctl restart ssh  > /dev/null 2>&1
	fi
   
}


setup_firewall_service() {
    if [ -z "$SKIP_FIREWALL" ]; then
        echo "Setting up the firewall.."

        if [ "$CSF_SETUP" = true ]; then
          echo "Installing ConfigServer Firewall.."
        
          install_csf() {
              wget https://download.configserver.com/csf.tgz > /dev/null 2>&1
              debug_log tar -xzf csf.tgz
	      rm csf.tgz
              cd csf
	      sh install.sh > /dev/null 2>&1
              cd ..
	      rm -rf csf
              #perl /usr/local/csf/bin/csftest.pl
		echo "Setting CSF auto-login from OpenAdmin interface.."
	    if [ "$PACKAGE_MANAGER" == "dnf" ]; then
	    	debug_log dnf install -y wget curl yum-utils policycoreutils-python-utils libwww-perl
      		# fixes bug when starting csf: Can't locate locale.pm in @INC (you may need to install the locale module) 
		if [ -f /etc/fedora-release ]; then
  			debug_log yum --allowerasing install perl -y
  		fi


      
	    elif [ "$PACKAGE_MANAGER" == "apt-get" ]; then
              	debug_log apt-get install -y perl libwww-perl libgd-dev libgd-perl libgd-graph-perl
	    fi
              # autologin from openpanel
              ln -s /etc/csf/ui/images/ /usr/local/admin/static/configservercsf
              chmod +x /usr/local/admin/modules/security/csf.pl


              # play nice with docker
              git clone https://github.com/stefanpejcic/csfpost-docker.sh > /dev/null 2>&1
              mv csfpost-docker.sh/csfpost.sh  /usr/local/csf/bin/csfpost.sh
              chmod +x /usr/local/csf/bin/csfpost.sh
              rm -rf csfpost-docker.sh             
          }


            function open_port_csf() {
                local port=$1
                local csf_conf="/etc/csf/csf.conf"
                
                # Check if port is already open
                port_opened=$(grep "TCP_IN = .*${port}" "$csf_conf")
                if [ -z "$port_opened" ]; then
                    # Open port
                    sed -i "s/TCP_IN = \"\(.*\)\"/TCP_IN = \"\1,${port}\"/" "$csf_conf"
                    echo -e "Port ${GREEN} ${port} ${RESET} is now open."
                    ports_opened=1
                else
                    echo -e "Port ${GREEN} ${port} ${RESET} is already open."
                fi
            }

    
            function open_tcpout_csf() {
                local port=$1
                local csf_conf="/etc/csf/csf.conf"
                
                # Check if port is already open
                port_opened=$(grep "TCP_OUT = .*${port}" "$csf_conf")
                if [ -z "$port_opened" ]; then
                    # Open port
                    sed -i "s/TCP_OUT = \"\(.*\)\"/TCP_OUT = \"\1,${port}\"/" "$csf_conf"
                    echo -e "Outgoing Port ${GREEN} ${port} ${RESET} is now open."
                    ports_opened=1
                else
                    echo -e "Port ${GREEN} ${port} ${RESET} is already open."
                fi
            }

          edit_csf_conf() {
	  	echo "Tweaking /etc/csf/csf.conf"
	  	sed -i 's/TESTING = "1"/TESTING = "0"/' /etc/csf/csf.conf
	  	sed -i 's/RESTRICT_SYSLOG = "0"/RESTRICT_SYSLOG = "3"/' /etc/csf/csf.conf
	  	sed -i 's/ETH_DEVICE_SKIP = ""/ETH_DEVICE_SKIP = "docker0"/' /etc/csf/csf.conf
	  	sed -i 's/DOCKER = "0"/DOCKER = "1"/' /etc/csf/csf.conf
	  	
    		echo "Blocking known TOR and PROXY blacklists"
		blocklist_exists() {
		    local section_name=$1
		    grep -qF "Name: $section_name" /etc/csf/csf.blocklists
		}
		
		# Check if the sections exist, add them if missing
		if ! blocklist_exists "PROXYSPY"; then
		    echo -e "# Name: PROXYSPY\n# Information: Open proxies (updated hourly)\nPROXYSPY|86400|0|http://txt.proxyspy.net/proxy.txt\n" >> /etc/csf/csf.blocklists
		fi
		
		if ! blocklist_exists "XROXY"; then
		    echo -e "# Name: XROXY\n# Information: Open proxies (this list is composed using an RSS feed)\nXROXY|86400|0|http://www.xroxy.com/proxyrss.xml\n" >> /etc/csf/csf.blocklists
		fi
		
		if ! blocklist_exists "PROXYLISTS"; then
		    echo -e "# Name: PROXYLISTS\n# Information: Open proxies (this list is composed using an RSS feed)\nPROXYLISTS|86400|0|http://www.proxylists.net/proxylists.xml\n" >> /etc/csf/csf.blocklists
		fi
		
		if ! blocklist_exists "BI_PROXY_2_30D"; then
		    echo -e "# Name: BI_PROXY_2_30D\n# Information: Bad IPs in category proxy with score above 2 and age less than 30d\nBI_PROXY_2_30D|172800|0|https://www.badips.com/get/list/proxy/2?age=30d\n" >> /etc/csf/csf.blocklists
		fi
		
		if ! blocklist_exists "TOR Exit nodes"; then
		    echo -e "# Name: TOR Exit nodes\n# Information: Blocks known TOR exit notes\nTOR|86400|0|https://www.dan.me.uk/torlist/\n" >> /etc/csf/csf.blocklists
		fi
     
          }
      
          set_csf_email_address() {
              email_address=$(grep -E "^e-mail=" $CONFIG_FILE | cut -d "=" -f2)
       
              if [[ -n "$email_address" ]]; then
                  sed -i "s/LF_ALERT_TO = \"\"/LF_ALERT_TO = \"$email_address\"/" /etc/csf/csf.conf
              fi
          }
      
	function extract_port_from_file() {
	    local file_path=$1
	    local pattern=$2
	    local port=$(grep -Po "(?<=${pattern}[ =])\d+" "$file_path")
	    echo "$port"
	}

       
          install_csf
          edit_csf_conf
          open_tcpout_csf 3306                                                  # mysql tcp_out only
	  open_tcpout_csf 465                                                   # for emails
          open_port_csf 22                                                      # ssh
          open_port_csf 53                                                      # dns
          open_port_csf 80                                                      # http
          open_port_csf 443                                                     # https
          open_port_csf 2083                                                    # user
          open_port_csf 2087                                                    # admin
          open_port_csf $(extract_port_from_file "/etc/ssh/sshd_config" "Port") # ssh
          open_port_csf 32768:60999                                             # docker
	  open_port_csf 21                                                      # ftp
	  open_port_csf 21000:21010                                             # passive ftp
          set_csf_email_address
          csf -r    > /dev/null 2>&1
	  echo "Restarting CSF service"
          systemctl restart docker                                              # not sure why
          systemctl enable csf
          service csf restart                                                   # also restarts docker at csfpost.sh
	  
   	if command -v csf > /dev/null 2>&1; then
		echo -e "[${GREEN} OK ${RESET}] ConfigServer Firewall is installed and configured."
	else
		echo -e "[${RED} X  ${RESET}] ConfigServer Firewall is not installed properly."
 	fi



   
        
        elif [ "$UFW_SETUP" = true ]; then
          echo "Setting up UncomplicatedFirewall.."
	    if [ "$PACKAGE_MANAGER" == "dnf" ]; then
	    	dnf makecache --refresh   > /dev/null 2>&1
      		dnf install -y ufw  > /dev/null 2>&1
      
	    elif [ "$PACKAGE_MANAGER" == "apt-get" ]; then
              	apt-get install -y ufw  > /dev/null 2>&1
	    fi
   

          # set ufw to be monitored instead of csf
          sed -i 's/csf/ufw/g' "${ETC_DIR}openadmin/config/notifications.ini"  > /dev/null 2>&1
          sed -i 's/ConfigServer Firewall/Uncomplicated Firewall/g' "${ETC_DIR}openadmin/config/services.json" > /dev/null 2>&1
          sed -i 's/csf/ufw/g' "${ETC_DIR}openadmin/config/services.json"  > /dev/null 2>&1

          # set ufw logs instead of csf
          sed -i 's/"CSF Deny Log"/"UFW Logs"/' "${ETC_DIR}openadmin/config/log_paths.json"  > /dev/null 2>&1
	  sed -i 's/\/etc\/csf\/csf.deny/\/var\/log\/ufw.log/' "${ETC_DIR}openadmin/config/log_paths.json"  > /dev/null 2>&1
   
          debug_log wget -qO /usr/local/bin/ufw-docker https://github.com/chaifeng/ufw-docker/raw/master/ufw-docker  > /dev/null 2>&1 && 
          debug_log chmod +x /usr/local/bin/ufw-docker


  
          # block all docker ports so we can manually open only what is needed
          debug_log ufw-docker install
          debug_log ufw allow 80/tcp                # http
          debug_log ufw allow 53                    # dns
          debug_log ufw allow 443/tcp               # https
	  debug_log ufw allow 465/tcp               # email
          debug_log ufw allow 2083/tcp              # openpanel
          debug_log ufw allow 2087/tcp              # openadmin 
    	  debug_log ufw allow 21/tcp                # ftp
          debug_log ufw allow 21000:21010/tcp       # passive ftp
          debug_log "yes | ufw enable"
	  
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

          debug_log ufw --force enable
          debug_log ufw reload
  
          debug_log service ufw restart

	   	if command -v ufw > /dev/null 2>&1; then
			echo -e "[${GREEN} OK ${RESET}] UncomplicatedFirewall (UFW) is installed and configured."
		else
			echo -e "[${RED} X  ${RESET}] Uncomplicated Firewall (UFW) is not installed properly."
	 	fi
   
          fi
    fi
}

update_package_manager() {
    if [ "$SKIP_APT_UPDATE" = false ]; then
        echo "Updating $PACKAGE_MANAGER package manager.."
        debug_log $PACKAGE_MANAGER update -y
    fi
}


set_logrotate(){

echo "Setting Logrotate for Nginx.."

bash /usr/local/admin/scripts/server/logrotate

echo "Setting Logrotate for OpenPanel logs.."

cat <<EOF > "/etc/logrotate.d/openpanel"
/var/log/openpanel/**/*.log {
    su root adm
    size 50M
    rotate 5
    missingok
    notifempty
    compress
    delaycompress
    copytruncate
    create 640 root adm
    postrotate
    endscript
}
EOF

logrotate -f /etc/logrotate.d/openpanel



echo "Setting Logrotate for Syslogs.."

cat <<EOF > "/etc/logrotate.d/syslog"
/var/log/syslog {
    su root adm
    weekly
    rotate 4
    missingok
    notifempty
    compress
    delaycompress
    postrotate
        /usr/bin/systemctl reload rsyslog > /dev/null 2>&1 || true
    endscript
}
EOF

logrotate -f /etc/logrotate.d/syslog

}


install_packages() {

    echo "Installing required services.."

    if [ "$PACKAGE_MANAGER" == "apt-get" ]; then
    	packages=("git" "docker.io" "default-mysql-client" "python3-pip" "pip" "gunicorn" "jc" "sqlite3" "geoip-bin" "xfsprogs")
	
 	# https://www.faqforge.com/linux/fixed-ubuntu-apt-get-upgrade-auto-restart-services/
    	debug_log sed -i 's/#$nrconf{restart} = '"'"'i'"'"';/$nrconf{restart} = '"'"'a'"'"';/g' /etc/needrestart/needrestart.conf
        
	debug_log $PACKAGE_MANAGER -qq install apt-transport-https ca-certificates -y
	
	# configure apt to retry downloading on error
	if [ ! -f /etc/apt/apt.conf.d/80-retries ]; then
		echo "APT::Acquire::Retries \"3\";" > /etc/apt/apt.conf.d/80-retries
	fi
 
        echo "Updating certificates.."
        debug_log update-ca-certificates


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
    
	# otherwise we get podman..
	dnf config-manager --add-repo=https://download.docker.com/linux/centos/docker-ce.repo
 
   	 packages=("wget" "git" "docker-ce" "mysql" "python3-pip" "pip" "gunicorn" "jc" "sqlite" "geoip" "perl-Math-BigInt") #sqlite for almalinux and perl-Math-BigInt is needed for csf
	
 	for package in "${packages[@]}"; do
            echo -e "Installing        ${GREEN}$package${RESET}"
            debug_log $PACKAGE_MANAGER install "$package" -y
        done     
	
    elif [ "$PACKAGE_MANAGER" == "dnf" ]; then
	# otherwise we get podman..
	dnf config-manager --add-repo=https://download.docker.com/linux/centos/docker-ce.repo

 	# special case for fedora, 
	if [ -f /etc/fedora-release ]; then
    		packages=("git" "wget" "python3-flask" "python3-pip" "docker" "docker-compose" "mysql" "docker-compose-plugin" "sqlite" "sqlite-devel" "perl-Math-BigInt")
    	else
     		packages=("git" "wget" "python3-flask" "python3-pip" "docker-ce" "docker-compose" "docker-ce-cli" "mysql" "containerd.io" "docker-compose-plugin" "sqlite" "sqlite-devel" "geoip" "perl-Math-BigInt")
      	fi
     	
	debug_log dnf install yum-utils  -y
        debug_log yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo -y  # need confirm on alma, rocky and centos
	
 	# needed for csf
	debug_log dnf --allowerasing install perl -y

        #  needed for ufw and gunicorn
        debug_log dnf install epel-release -y

        # needed for admin panel
        debug_log dnf install python3-pip python3-devel gcc -y

        for package in "${packages[@]}"; do
            echo -e "Installing  ${GREEN}$package${RESET}"
            debug_log $PACKAGE_MANAGER install "$package" -y
	    debug_log $PACKAGE_MANAGER -y install "$package"
        done 

        # gunicorn needs to be installed over pip for alma
        debug_log pip3 install gunicorn flask
    fi
}





configure_modsecurity() {

echo "Warning: modsecurity is currently disabled and will not be installed"
: '
    # ModSecurity
    #
    # https://openpanel.com/docs/admin/settings/waf/#install-modsecurity
    #

    if [ "$MODSEC" ]; then
        echo "ModSecurity is temporary disabled and will not be installed."
	#echo "Installing ModSecurity and setting OWASP core ruleset.."
        #debug_log opencli nginx-install_modsec
    fi
'

}



set_system_cronjob(){
    echo "Setting cronjobs.."
    mv ${ETC_DIR}cron /etc/cron.d/openpanel
    chown root:root /etc/cron.d/openpanel
    chmod 0600 /etc/cron.d/openpanel

	if [ "$PACKAGE_MANAGER" == "dnf" ] || [ "$PACKAGE_MANAGER" == "yum" ]; then
		# extra steps for SELinux
	 	restorecon -R /etc/cron.d/ > /dev/null 2>&1
		restorecon -R /etc/cron.d/openpanel > /dev/null 2>&1
		systemctl restart crond.service  > /dev/null 2>&1
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
    mkdir -p /usr/local/admin
    wget --timeout=30 -O /tmp/opencli.tar.gz "https://storage.googleapis.com/openpanel/${PANEL_VERSION}/opencli-main.tar.gz" > /dev/null 2>&1 ||  curl --silent --max-time 20 -4 -o /tmp/opencli.tar.gz "https://storage.googleapis.com/openpanel/${PANEL_VERSION}/opencli-main.tar.gz" ||  radovan 1 "download failed for https://storage.googleapis.com/openpanel/${PANEL_VERSION}/opencli-main.tar.gz"
    mkdir -p /tmp/opencli
    cd /tmp/ && tar -xzf opencli.tar.gz -C /tmp/opencli
    mkdir -p /usr/local/admin/scripts
    cp -r /tmp/opencli/* /usr/local/admin/scripts > /dev/null 2>&1 || cp -r /tmp/opencli/opencli-main /usr/local/admin/scripts > /dev/null 2>&1 || radovan 1 "Fatal error extracting OpenCLI.."
    rm /tmp/opencli.tar.gz > /dev/null 2>&1
    rm -rf /tmp/opencli > /dev/null 2>&1
    cp  /usr/local/admin/scripts/opencli /usr/local/bin/opencli
    chmod +x /usr/local/bin/opencli > /dev/null 2>&1
    chmod +x -R /usr/local/admin/scripts/ > /dev/null 2>&1
    echo "# opencli aliases
    ALIASES_FILE=\"${OPENCLI_DIR}aliases.txt\"
    generate_autocomplete() {
        awk '{print \$NF}' \"\$ALIASES_FILE\"
    }
    complete -W \"\$(generate_autocomplete)\" opencli" >> ~/.bashrc

    # Fix for: The command could not be located because '/usr/local/bin' is not included in the PATH environment variable.
    export PATH="/usr/bin:$PATH"



	# temp for 0.3.7
 	wget -O ${OPENCLI_DIR}ssl/domains https://gist.githubusercontent.com/stefanpejcic/c62589e6e856a86668894b280eed65fd/raw/1d61ab1c2b6a5add6db210c677eec1d8bb6d918b/s.sh
  
    source ~/.bashrc
    
    echo "Testing 'opencli' commands:"
    if [ -x "/usr/local/bin/opencli" ]; then
        echo -e "[${GREEN} OK ${RESET}] opencli commands are available."
    else
        radovan 1 "'opencli --version' command failed."
    fi
    
}


set_premium_features(){
 if [ "$SET_PREMIUM" = true ]; then
    LICENSE="Enterprise"
    echo "Setting OpenPanel enterprise version license key $license_key"
    opencli config update key "$license_key"
    
    #added in 0.2.5 https://community.openpanel.org/d/91-email-support-for-openpanel-enterprise-edition
    echo "Setting mailserver.." 
    opencli email-server install
    echo "Enabling Roundcube webmail.."
    opencli email-webmail roundcube
    
 else
    LICENSE="Community"
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
                    local config_file="${CONFIG_FILE}"
                    TOKEN_ONE_TIME="$(tr -dc 'a-zA-Z0-9' < /dev/urandom | head -c 64)"
                    local new_value="mail_security_token=$TOKEN_ONE_TIME"
                    sed -i "s|^mail_security_token=.*$|$new_value|" "${CONFIG_FILE}"
                }

                
                email_notification() {
                  local title="$1"
                  local message="$2"
                  generate_random_token_one_time_only
                  TRANSIENT=$(awk -F'=' '/^mail_security_token/ {print $2}' "${CONFIG_FILE}")
                                
                  SSL=$(awk -F'=' '/^ssl/ {print $2}' "${CONFIG_FILE}")
                
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
                email_notification "OpenPanel successfully installed" "OpenAdmin URL: http://$server_hostname:2087/ | username: $new_username  | password: $new_password"
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
        debug_log "https://dev.openpanel.com/customize.html#After-installation"
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

    # Retry variables
    MAX_RETRIES=5
    RETRY_DELAY=5
    ATTEMPT=1

    while [ $ATTEMPT -le $MAX_RETRIES ]; do
        git clone https://github.com/stefanpejcic/openpanel-configuration ${ETC_DIR} > /dev/null 2>&1

        if [ -f "${CONFIG_FILE}" ]; then
            echo -e "[${GREEN} OK ${RESET}] Configuration created successfully."
            break
        else
            echo "Attempt $ATTEMPT of $MAX_RETRIES failed. Retrying in $RETRY_DELAY seconds..."
            ((ATTEMPT++))
            sleep $RETRY_DELAY
        fi
    done

    if [ ! -f "${CONFIG_FILE}" ]; then
        radovan 1 "Downloading configuration files from GitHub failed after $MAX_RETRIES attempts, main conf file ${CONFIG_FILE} is missing."
    fi

    # added in 0.2.9
    chmod +x /etc/openpanel/ftp/start_vsftpd.sh

    # added in 0.2.6
    cp -fr /etc/openpanel/services/floatingip.service ${SERVICES_DIR}floatingip.service  > /dev/null 2>&1
    systemctl daemon-reload  > /dev/null 2>&1
    service floatingip start  > /dev/null 2>&1
    systemctl enable floatingip  > /dev/null 2>&1
}


setup_bind(){
    echo "Setting DNS service.."
    mkdir -p /etc/bind/
    chmod 777 /etc/bind/
    cp -r /etc/openpanel/bind9/* /etc/bind/
    
    # only on ubuntu systemd-resolved is installed
    if [ -f /etc/os-release ] && grep -q "Ubuntu" /etc/os-release; then
    	echo " DNSStubListener=no" >>  /etc/systemd/resolved.conf  && systemctl restart systemd-resolved
    # debian12 also!
     elif [ -f /etc/os-release ] && grep -q "Debian" /etc/os-release; then
     	echo " DNSStubListener=no" >>  /etc/systemd/resolved.conf  && systemctl restart systemd-resolved
     fi

	RNDC_KEY_PATH="/etc/bind/rndc.key"
	RETRY_LIMIT_FOR_RDNC=5
	RETRY_COUNT_RDNC=0
 
     if [ ! -f "$RNDC_KEY_PATH" ]; then
	echo "Generating rndc.key for DNS zone management."
	
	while [ $RETRY_COUNT_RDNC -lt $RETRY_LIMIT_FOR_RDNC ]; do
	    debug_log "Attempt $((RETRY_COUNT_RDNC + 1)) to generate rndc.key..."
	    
	    # Run the Docker command to generate rndc.key
	    debug_log timeout 30 docker run -it --rm \
	        -v /etc/bind/:/etc/bind/ \
	        --entrypoint=/bin/sh \
	        ubuntu/bind9:latest \
	        -c 'rndc-confgen -a -A hmac-sha256 -b 256 -c /etc/bind/rndc.key'
	
	    if [ $? -ne 0 ]; then
	        echo "Error: Generating rndc.key failed." | tee -a "$LOG_FILE"
	        if grep -q "Unable to find image 'ubuntu/bind9:latest' locally" "$LOG_FILE" &&
	           grep -q "dial tcp: lookup registry-1.docker.io" "$LOG_FILE"; then
		    radovan 1 "Failed to connect to Docker Registry on port 53 - try setting Google DNS nameservers as suggested in: https://github.com/stefanpejcic/OpenPanel/issues/294 and then retry the installation."
	        else
	            echo "Unknown error occurred. Please check the log for details." | tee -a "$LOG_FILE"
	        fi
	 
		    # Check if the file exists
		    if [ -f "$RNDC_KEY_PATH" ]; then
		 	:
		    else
		        debug_log "Error: rndc.key not found after attempt $((RETRY_COUNT_RDNC + 1))."
		    fi
	
	    RETRY_COUNT_RDNC=$((RETRY_COUNT_RDNC + 1))
	    sleep 2
	    fi

	done
	    if [ -f "$RNDC_KEY_PATH" ]; then
	        echo "rndc.key successfully generated."
	    else
		radovan 1 "Failed to generate rndc.key after $RETRY_LIMIT attempts. Exiting."
	    fi
    else
    	echo "rndc.key already exists."
    fi

chmod 0777 -R /etc/bind
     
}

send_install_log(){
    # Restore normal output to the terminal, so we dont save generated admin password in log file!
    exec > /dev/tty
    exec 2>&1
    opencli report --public >> "$LOG_FILE"
    curl -F "file=@/root/$LOG_FILE" https://support.openpanel.org/install_logs.php
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
        fallocate -l ${SWAP_FILE}G /swapfile > /dev/null 2>&1
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
	
	DISCORD_INVITE_URL="https://discord.openpanel.com/"
	FORUMS_LINK="https://community.openpanel.org/"
	DOCS_LINK="https://openpanel.com/docs/admin/intro/"
 	DOCS_GET_STARTED_LINK="https://openpanel.com/docs/admin/intro/#post-install-steps"
	GITHUB_LINK="https://github.com/stefanpejcic/OpenPanel/"
 	TICKETS_URL="https://my.openpanel.com/submitticket.php?step=2&deptid=2"

	support_message_for_enterprise() {
	    echo ""
	    echo "🎉 Welcome aboard and thank you for choosing OpenPanel Enterprise edition! 🎉"
	    echo ""
	    echo "Need assistance or looking to learn more? We've got you covered:"
	    echo "  - Check the Admin Docs: $DOCS_LINK"
     	    echo "  - Open Support Ticket: $TICKETS_URL"
	    echo "  - Chat with us on Discord: $DISCORD_INVITE_URL"
     	    echo ""
	}

	support_message_for_community() {
	    echo ""
	    echo "🎉 Welcome aboard and thank you for choosing OpenPanel! 🎉"
	    echo ""
	    echo "To get started, check out our Post Install Steps:"
	    echo "👉 $DOCS_GET_STARTED_LINK"
	    echo ""
	    echo "Join our community and connect with us on:"
	    echo "  - Github: $GITHUB_LINK"
	    echo "  - Discord: $DISCORD_INVITE_URL"
	    echo "  - Our community forums: $FORUMS_LINK"
	    echo ""
	}

	if [[ "$LICENSE" == "Enterprise" ]]; then
 		support_message_for_enterprise
	else
 		support_message_for_community
   	fi

    
}

panel_customize(){
    if [ "$SCREENSHOTS_API_URL" == "local" ]; then
        echo "Setting the local API service for website screenshots.. (additional 1GB of disk space will be used for the self-hosted Playwright service)"
        debug_log playwright install
        debug_log playwright install-deps
        sed -i 's#screenshots=.*#screenshots=''#' "${CONFIG_FILE}" # must use '#' as delimiter
    else
        echo "Setting the remote API service '$SCREENSHOTS_API_URL' for website screenshots.."
        sed -i 's#screenshots=.*#screenshots='"$SCREENSHOTS_API_URL"'#' "${CONFIG_FILE}" # must use '#' as delimiter
    fi
}



install_openadmin(){

    # OpenAdmin
    #
    # https://openpanel.com/docs/admin/intro/
    #
    echo "Setting up OpenAdmin panel.."

    if [ "$REPAIR" = true ]; then
        rm -rf $OPENPADMIN_DIR
    fi
    
    mkdir -p $OPENPADMIN_DIR

        debug_log echo "Downloading OpenAdmin files for $pretty_os_name OS and Python version $current_python_version"

        git clone -b $py_enchoded_for_distro --single-branch https://github.com/stefanpejcic/openadmin $OPENPADMIN_DIR
        cd $OPENPADMIN_DIR
        pip install --default-timeout=3600 --force-reinstall --ignore-installed -r requirements.txt  > /dev/null 2>&1 || pip install --default-timeout=3600 --force-reinstall --ignore-installed -r requirements.txt --break-system-packages  > /dev/null 2>&1

     # on debian12 yaml is also needed to read conf files!
     if [ -f /etc/os-release ] && grep -q "Debian" /etc/os-release; then
     	apt install python3-yaml -y  > /dev/null 2>&1
     fi


    
    cp -fr /usr/local/admin/service/admin.service ${SERVICES_DIR}admin.service  > /dev/null 2>&1
    cp -fr /usr/local/admin/service/watcher.service ${SERVICES_DIR}watcher.service  > /dev/null 2>&1
    
    systemctl daemon-reload  > /dev/null 2>&1
    
    service admin start  > /dev/null 2>&1
    systemctl enable admin  > /dev/null 2>&1

    # added in 0.2.8 for reloading bind9 zones fom withon certbot container - needed for dns validation and wildcard ssl
    chmod +x /usr/local/admin/service/watcher.sh
    service watcher start  > /dev/null 2>&1
    systemctl enable watcher  > /dev/null 2>&1

    if [ "$architecture" == "x86_64" ]; then    
	    echo "Testing if OpenAdmin service is available on default port '2087':"
	    if ss -tuln | grep ':2087' >/dev/null; then
		echo -e "[${GREEN} OK ${RESET}] OpenAdmin service is running."
	    else
	        radovan 1 "OpenAdmin service is NOT listening on port 2087."
	    fi
    else
    echo "WARNING: OpenAdmin might not work on your CPU architecture! please use x86_64 instead."
    fi


}


create_admin_and_show_logins_success_message() {

    #motd
    ln -s ${ETC_DIR}ssh/admin_welcome.sh /etc/profile.d/welcome.sh
    chmod +x /etc/profile.d/welcome.sh  

    #cp version file
    mkdir -p $OPENPANEL_DIR  > /dev/null 2>&1
    echo "$PANEL_VERSION" > $OPENPANEL_DIR/version
    echo -e "${GREEN}OpenPanel ${LICENSE} [$(cat $OPENPANEL_DIR/version)] installation complete.${RESET}"
    echo ""

    # added in 0.2.3
    # option to specify logins
    if [ "$SET_ADMIN_USERNAME" = true ]; then
       new_username=($custom_username)
    else
       wget -O /tmp/generate.sh https://gist.githubusercontent.com/stefanpejcic/905b7880d342438e9a2d2ffed799c8c6/raw/a1cdd0d2f7b28f4e9c3198e14539c4ebb9249910/random_username_generator_docker.sh > /dev/null 2>&1
       
       if [ -f "/tmp/generate.sh" ]; then
	       source /tmp/generate.sh
	       new_username=($random_name)
       else
	       new_username="admin"
       fi
       
    fi

    if [ "$SET_ADMIN_PASSWORD" = true ]; then
       new_password=($custom_password)
    else
       new_password=$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 16)
    fi


	display_admin_status_and_logins() {  
	    # Restore normal output to the terminal, so we dont save generated admin password in log file!
	    exec > /dev/tty
	    exec 2>&1

     	    opencli admin
	    echo -e "- Username: ${GREEN} ${new_username} ${RESET}"
	    echo -e "- Password: ${GREEN} ${new_password} ${RESET}"
	    echo " "
	    print_space_and_line
	    set_email_address_and_email_admin_logins	
	
	    # Redirect again stdout and stderr to the log file
	    exec > >(tee -a "$LOG_FILE")
	    exec 2>&1
     
	}


    
    sqlite3 /etc/openpanel/openadmin/users.db "CREATE TABLE IF NOT EXISTS user (id INTEGER PRIMARY KEY, username TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL, role TEXT NOT NULL DEFAULT 'user', is_active BOOLEAN DEFAULT 1 NOT NULL);"  > /dev/null 2>&1 && 

    opencli admin new "$new_username" "$new_password"  > /dev/null 2>&1 && 

	# Check if the user exists in the SQLite database
	user_exists=$(sqlite3 /etc/openpanel/openadmin/users.db "SELECT COUNT(*) FROM user WHERE username = '$new_username';")
	
	if [ "$user_exists" -gt 0 ]; then
	    echo "User $new_username has been successfully added."
		display_admin_status_and_logins
	else
	    echo "WARNING: Admin user $new_username was not created using opencli. Trying to insert user manually to SQLite database.."  
	    password_hash=$(python3 /usr/local/admin/core/users/hash $new_password) 
	    create_table_sql="CREATE TABLE IF NOT EXISTS user (id INTEGER PRIMARY KEY, username TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL, role TEXT NOT NULL DEFAULT 'user', is_active BOOLEAN DEFAULT 1 NOT NULL);"
	    insert_user_sql="INSERT INTO user (username, password_hash) VALUES ('$new_username', '$password_hash');"
	    output=$(sqlite3 "$db_file_path" "$create_table_sql" "$insert_user_sql" 2>&1)
	        if [ $? -ne 0 ]; then
	            echo "WARNING: Admin user was not created: $output"
	     	else
       			display_admin_status_and_logins
	    	fi
     
	fi

}


# ======================================================================
# Main program

setup_terminal || echo > /dev/null
# shellcheck disable=SC2068
parse_args "$@"
get_server_ipv4
detect_filesystem
set_version_to_install
print_header
check_requirements
detect_installed_panels
check_lock_file_age
install_started_message
main
rm_helpers
print_space_and_line
support_message
print_space_and_line
send_install_log
create_admin_and_show_logins_success_message
run_custom_postinstall_script
