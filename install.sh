#!/bin/bash
################################################################################
#
# OpenPanel Installer ✌️
# https://openpanel.com/install
#
# Supported OS:            Ubuntu, Debian, AlmaLinux, RockyLinux, CentOS
# Supported Architecture:  x86_64(AMD64), AArch64(ARM64)
#
# Usage:                   bash <(curl -sSL https://openpanel.org)
# Author:                  Stefan Pejcic <stefan@pejcic.rs>
# Created:                 11.07.2023
# Last Modified:           23.09.2025
#
################################################################################



# ======================================================================
# Constants
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
RESET='\033[0m'
export TERM=xterm-256color                                            # bug fix: tput: No value for $TERM and no -T specified
export DEBIAN_FRONTEND=noninteractive
# ======================================================================
# Defaults for environment variables
CUSTOM_VERSION=false                                                  # default version is latest
DEBUG=false                                                           # verbose output for debugging failed install
SKIP_APT_UPDATE=false                                                 # they are auto-pulled on account creation
SKIP_DNS_SERVER=false
REPAIR=false
LOCALES=true                                                          # only en
NO_SSH=false                                                          # deny port 22
SET_HOSTNAME_NOW=false                                                # must be a FQDN
SETUP_SWAP_ANYWAY=false                                               # setup swapfile regardless of server ram
CORAZA=true                                                           # install CorazaWAF, unless user provices --no-waf flag
IMUNIFY_AV=false                                                      # https://community.openpanel.org/d/193-dont-install-imunifyav-by-default
SWAP_FILE="1"                                                         # calculated based on ram
SEND_EMAIL_AFTER_INSTALL=false                                        # send admin logins to specified email
SET_PREMIUM=false                                                     # added in 0.2.1
SET_ADMIN_USERNAME=false                                              # random
SET_ADMIN_PASSWORD=false                                              # random
SCREENSHOTS_API_URL="http://screenshots-v2.openpanel.com/api/screenshot" # default since 0.5.9
DEV_MODE=false
post_install_path=""                                                  # not to run
# ======================================================================
# PATHs used throughout the script
ETC_DIR="/etc/openpanel/"                                             # https://github.com/stefanpejcic/openpanel-configuration
LOG_FILE="openpanel_install.log"                                      # install log                                      # install running
SERVICES_DIR="/etc/systemd/system/"                                   # used for admin and sentinel services
CONFIG_FILE="${ETC_DIR}openpanel/conf/openpanel.config"               # main config file for openpanel

exec > >(tee -a "$LOG_FILE") 2>&1

echo "" > /root/openpanel_restart_needed

# ======================================================================
# Helper functions that are not mandatory but still should not be modified

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
    if [ "$SKIP_FIREWALL" = false ]; then
    	echo -e "- Install necessary services and tools: CSF, Docker, MySQL, SQLite, Python3, PIP.. "
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
    if [ "$SKIP_FIREWALL" = false ]; then
    	echo -e "- Set up ConfigServer Firewall for enhanced security."
    fi

    echo -e "- Set up 2 hosting plans so you can start right away."

    echo -e "\nThank you for your patience. We're setting everything up for your seamless OpenPanel experience!\n"
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
    echo -e ""
}

radovan() {
    echo -e "${RED}INSTALLATION FAILED${RESET} - Please retry with '--repair' flag"
    echo ""
    echo -e "Error: $2" >&2
    exit 1
}

debug_log() {
    local timestamp
    timestamp=$(date +'%Y-%m-%d %H:%M:%S')

    if [ "$DEBUG" = true ]; then
        echo "[$timestamp] $message" | tee -a "$LOG_FILE"
        "$@" 2>&1 | tee -a "$LOG_FILE"
    else
    	# ❯❯❯
        echo "[$timestamp] COMMAND: $@" >> "$LOG_FILE"
        "$@" > /dev/null 2>&1
    fi
}

is_package_installed() {
    if [ "$DEBUG" = false ]; then
    $PACKAGE_MANAGER -qq list "$1" 2>/dev/null | grep -qE "^ii"
    else
    $PACKAGE_MANAGER -qq list "$1" | grep -qE "^ii"
    echo "Updating $PACKAGE_MANAGER package manager.."
    fi
}

get_server_ipv4() {
    local services=("https://ip.openpanel.com" "https://ipv4.openpanel.com" "https://ifconfig.me")
    local ip

    for url in "${services[@]}"; do
        ip=$(curl -s --max-time 2 -4 "$url" || wget -qO- --timeout=2 --inet4-only "$url")
        [ -n "$ip" ] && break
    done

    if [ -z "$ip" ]; then
        ip=$(ip -4 addr show scope global | grep -oP '(?<=inet\s)\d+(\.\d+){3}' | head -n1)
    fi

    if ! [[ "$ip" =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}$ ]] || \
         [[ "$ip" =~ ^10\.|^172\.(1[6-9]|2[0-9]|3[0-1])\.|^192\.168\. ]]; then
        echo "Invalid or private IPv4 address: $ip. OpenPanel requires a public IPv4 address to bind domains configuration files."
    fi
	current_ip=$ip
}

set_version_to_install() {
    if [ "$CUSTOM_VERSION" = false ]; then
        response=$(curl -4 -s "https://usage-api.openpanel.org/latest_version")

        if command -v jq &> /dev/null; then
            PANEL_VERSION=$(echo "$response" | jq -r '.latest_version')
        else
            PANEL_VERSION=$(echo "$response" | grep -oP '"latest_version":"\K[^"]+')
        fi

        [[ "$PANEL_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || PANEL_VERSION="1.6.1"
    fi
}


print_space_and_line() {
    echo " "
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
    echo " "
}


setup_progress_bar_script(){
	PROGRESS_BAR_URL="https://raw.githubusercontent.com/pollev/bash_progress_bar/master/progress_bar.sh"
	PROGRESS_BAR_FILE="progress_bar.sh"

	if command -v wget &> /dev/null; then
	    wget --timeout=5 --inet4-only "$PROGRESS_BAR_URL" -O "$PROGRESS_BAR_FILE" > /dev/null 2>&1
	    if [ $? -ne 0 ]; then
	        echo "ERROR: wget failed or timed out after 5 seconds while downloading from github"
	 	echo "repeat with --debug flag to see where errored."
	        exit 1
	    fi
	elif command -v curl -4 &> /dev/null; then # fallback for fedora
	    curl -4 --max-time 5 -s "$PROGRESS_BAR_URL" -o "$PROGRESS_BAR_FILE" > /dev/null 2>&1
	    if [ $? -ne 0 ]; then
	        echo "ERROR: curl failed or timed out after 5 seconds while downloading progress_bar.sh"
	        exit 1
	    fi
	else
	    echo "Neither wget nor curl is available. Please install one of them to proceed."
	    exit 1
	fi

	if [ ! -f "$PROGRESS_BAR_FILE" ]; then
	    echo "ERROR: Failed to download progress_bar.sh - Github may be unreachable from your server: $PROGRESS_BAR_URL"
	    exit 1
	fi
}




display_what_will_be_installed(){
    echo -e "[ OK ] DETECTED OPERATING SYSTEM: ${GREEN} ${NAME^^} $VERSION_ID ${RESET}"
    if [ -z "$SKIP_REQUIREMENTS" ]; then
		if [ "$architecture" == "x86_64" ] || [ "$architecture" == "aarch64" ]; then
	  		echo -e "[ OK ] CPU ARCHITECTURE:          ${GREEN} ${architecture^^} ${RESET}"
	   	else
	      	echo -e "[PASS] CPU ARCHITECTURE:          ${YELLOW} ${architecture^^} ${RESET}"
	 	fi
  	fi
 	echo -e "[ OK ] PACKAGE MANAGEMENT SYSTEM: ${GREEN} ${PACKAGE_MANAGER^^} ${RESET}"
 	echo -e "[ OK ] PUBLIC IPV4 ADDRESS:       ${GREEN} ${current_ip} ${RESET}"
  	echo ""
}




# ======================================================================
setup_progress_bar_script
source "$PROGRESS_BAR_FILE"               # Source the progress bar script

FUNCTIONS=(
detect_os_cpu_and_package_manager         # detect os and package manager
display_what_will_be_installed            # display os, version, ip
install_python
update_package_manager                    # update dnf/yum/apt-get
install_packages                          # install docker, csf, sqlite, etc.
download_skeleton_directory_from_github   # download configuration to /etc/openpanel/
edit_fstab                                # enable quotas
setup_bind                                # must run after -configuration
install_openadmin                         # set admin interface
opencli_setup                             # set terminal commands
extra_step_on_hetzner                     # run it here, then csf install does docker restart later
setup_redis_service                       # for redis container
create_rdnc                               # generate rdnc key for managing domains
panel_customize                           # customizations
docker_compose_up                         #
docker_cpu_limiting                       # https://docs.docker.com/engine/security/rootless/#limiting-resources
set_premium_features                      # must be after docker_compose_up
configure_coraza                          # download corazawaf coreruleset or change docker image
extra_step_for_caddy                      # so that webmail domain works without any setups!
enable_dev_mode                           # https://dev.openpanel.com/cli/config.html#dev-mode
set_custom_hostname                       # set hostname if provided
generate_and_set_ssl_for_panels           # if FQDN then lets setup https
setup_firewall_service                    # setup firewall
set_system_cronjob                        # setup crons, must be after csf
set_logrotate                             # setup logrotate, ignored on fedora
tweak_ssh                                 # basic ssh
log_dirs                                  # for almalinux
download_ui_image                         # pull openpanel-ui image
setup_imunifyav                           # setum imunifyav and enable autologin from openadmin
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

		check_requirement() {
		    local value=$1
		    local min=$2
		    local unit=$3
		    local message=$4

		    if [ "$value" -lt "$min" ]; then
		        echo -e "${RED}Error: ${message}. Detected: ${value}${unit}${RESET}" >&2
		        echo ""
		        echo "Requirements: https://openpanel.com/docs/admin/intro/#requirements"
		        exit 1
		    fi
		}

		check_condition() {
		    local condition=$1
		    local message=$2
		    local show_link=$3

		    if eval "$condition"; then
		        echo -e "${RED}Error: ${message}${RESET}" >&2
		        [ "$show_link" = true ] && echo -e "\nRequirements: https://openpanel.com/docs/admin/intro/#requirements"
		        exit 1
		    fi
		}

		# Check if running as root
		check_condition '[ "$(id -u)" != "0" ]' "you must be root to execute this script" false

		# Check OS
		check_condition '[ "$(uname)" = "Darwin" ]' "MacOS is not currently supported" true

		# Check if running inside a container
		check_condition '[[ -f /.dockerenv || $(grep -sq "docker\|lxc" /proc/1/cgroup) ]]' "running inside a container is not supported" false

		# Check RAM
		total_mb=$(( $(grep MemTotal /proc/meminfo | awk '{print $2}') / 1024 ))
		check_requirement "$total_mb" 1024 "MB" "at least 1GB of RAM is required"

		# Check available disk space on /
		available_mb=$(( $(df / --output=avail | tail -1) / 1024 ))
		check_requirement "$available_mb" 5120 "MB" "at least 5GB of free disk space is required on /"
    fi

}


parse_args() {

# ======================================================================
# Usage info
    show_help() {
        echo "Available options:"
        echo "  --key=<key_here>                Set the license key for OpenPanel Enterprise edition."
        echo "  --domain=<domain>               Set the server hostname and domain for accessing panel."
        echo "  --username=<username>           Set Admin username - random generated if not provided."
        echo "  --password=<password>           Set Admin Password - random generated if not provided."
        echo "  --version=<version>             Set a custom OpenPanel version to be installed."
        echo "  --email=<stefan@example.net>    Set email address to receive email with admin credentials and future notifications."
        echo "  --imunifyav                     Install and setup ImunifyAV."
        echo "  --skip-requirements             Skip the requirements check."
        echo "  --skip-panel-check              Skip checking if existing panels are installed."
        echo "  --skip-apt-update               Skip the APT update."
        echo "  --skip-firewall                 Skip installing CSF - Only do this if you will set another external firewall!"
        echo "  --no-waf                        Do not configure CorazaWAF with OWASP Coreruleset."
        echo "  --no-ssh                        Disable port 22 and whitelist the IP address of user installing the panel."
        echo "  --skip-dns-server               Skip setup for DNS (Bind9) server."
        echo "  --post_install=<path>           Specify the post install script path."
        echo "  --screenshots=<url>             Set the screenshots API URL."
        echo "  --swap=<2>                      Set space in GB to be allocated for SWAP."
        echo "  --debug                         Display debug information during installation."
        echo "  --enable-dev-mode               Enable dev_mode after installation."
        echo "  --repair OR --retry             Retry and overwrite everything."
        echo "  -h, --help                      Show this help message and exit."
    }






# ======================================================================
# Change defaults
while [[ $# -gt 0 ]]; do
    case $1 in
        --key=*|--domain=*|--username=*|--password=*|--post_install=*|--screenshots=*|--version=*|--swap=*|--email=*)
            opt="${1%%=*}"
            val="${1#*=}"
            case "$opt" in
                --key)         SET_PREMIUM=true;          license_key="$val" ;;
                --domain)      SET_HOSTNAME_NOW=true;     new_hostname="$val" ;;
                --username)    SET_ADMIN_USERNAME=true;   custom_username="$val" ;;
                --password)    SET_ADMIN_PASSWORD=true;   custom_password="$val" ;;
                --post_install) post_install_path="$val" ;;
                --screenshots) SCREENSHOTS_API_URL="$val" ;;
                --version)     CUSTOM_VERSION=true;       PANEL_VERSION="$val" ;;
                --swap)        SETUP_SWAP_ANYWAY=true;    SWAP_FILE="$val" ;;
                --email)       SEND_EMAIL_AFTER_INSTALL=true; EMAIL="$val" ;;
            esac
            ;;
        --skip-requirements)   SKIP_REQUIREMENTS=true ;;
        --skip-panel-check)    SKIP_PANEL_CHECK=true ;;
        --skip-apt-update)     SKIP_APT_UPDATE=true ;;
        --skip-dns-server)     SKIP_DNS_SERVER=true ;;
        --skip-firewall)       SKIP_FIREWALL=true ;;
        --imunifyav)      IMUNIFY_AV=true ;;
        --no-waf)              CORAZA=false ;;
        --debug)               DEBUG=true ;;
        --no-ssh)              NO_SSH=true ;;
        --repair|--retry)
            REPAIR=true
            SKIP_PANEL_CHECK=true
            SKIP_APT_UPDATE=true
            ;;
        --enable-dev-mode)     DEV_MODE=true ;;
        -h|--help)             show_help; exit 0 ;;
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
    if [ -n "$SKIP_PANEL_CHECK" ]; then return; fi

    declare -A panels=(
        ["/usr/local/admin/"]="OpenPanel"
        ["/usr/local/cpanel/whostmgr"]="cPanel WHM"
        ["/opt/psa/version"]="Plesk"
        ["/usr/local/psa/version"]="Plesk"
        ["/usr/local/CyberPanel"]="CyberPanel"
        ["/usr/local/directadmin"]="DirectAdmin"
        ["/usr/local/cwpsrv"]="CentOS Web Panel (CWP)"
        ["/usr/local/vesta"]="VestaCP"
        ["/usr/local/hestia"]="HestiaCP"
        ["/usr/local/httpd"]="Apache WebServer"
        ["/usr/local/apache2"]="Apache WebServer"
        ["/usr/sbin/httpd"]="Apache WebServer"
        ["/sbin/httpd"]="Apache WebServer"
        ["/usr/lib/nginx"]="Nginx WebServer"
    )

    for path in "${!panels[@]}"; do
        if [ -e "$path" ]; then
            if [ "${panels[$path]}" = "OpenPanel" ]; then
                radovan 1 "You already have OpenPanel installed. ${RESET}\nInstead, did you want to update? Run ${GREEN}'opencli update --force' to update OpenPanel."
            elif [[ "${panels[$path]}" == *"WebServer"* ]]; then
                radovan 1 "${panels[$path]} is already installed. OpenPanel only supports servers without any webservers installed."
            else
                radovan 1 "${panels[$path]} is installed. OpenPanel only supports servers without any hosting control panel installed."
            fi
        fi
    done

    echo -e "${GREEN}No currently installed hosting control panels or webservers found. Starting the installation process.${RESET}"
}



detect_os_cpu_and_package_manager() {
    if [ -f "/etc/os-release" ]; then
        . /etc/os-release

		case "$ID" in
		    ubuntu|debian)
		        PACKAGE_MANAGER="apt-get"
		        ;;
		    fedora|rocky|almalinux|alma)
		        PACKAGE_MANAGER="dnf"
		        ;;
		    centos)
		        PACKAGE_MANAGER="yum"
		        ;;
		    *)
		        echo -e "${RED}Unsupported Operating System: $ID. Exiting.${RESET}"
		        echo -e "${RED}INSTALL FAILED${RESET}"
		        exit 1
		        ;;
		esac

        # Locale-agnostic CPU architecture detection
        arch_raw="$(uname -m)"
        case "$arch_raw" in
            x86_64|amd64)
                architecture="x86_64"   # 64-bit x86
                ;;
            aarch64|arm64)
                architecture="aarch64"  # 64-bit ARM
                ;;
            *)
                architecture="$arch_raw" # Fallback: keep raw value
                ;;
        esac
    else
        echo -e "${RED}Could not detect Linux distribution from /etc/os-release${RESET}"
        echo -e "${RED}INSTALL FAILED${RESET}"
        exit 1
    fi
}


docker_compose_up(){
    echo "Setting docker-compose.."
    DOCKER_CONFIG=${DOCKER_CONFIG:-/root/.docker}
    mkdir -p $DOCKER_CONFIG/cli-plugins

    if [ "$architecture" == "aarch64" ]; then
		link="https://github.com/docker/compose/releases/download/v2.36.0/docker-compose-linux-aarch64"
  	else
   		link="https://github.com/docker/compose/releases/download/v2.36.0/docker-compose-linux-x86_64"
 	fi

	    curl -4 -SL $link -o $DOCKER_CONFIG/cli-plugins/docker-compose  > /dev/null 2>&1
	    debug_log chmod +x $DOCKER_CONFIG/cli-plugins/docker-compose
		debug_log curl -4 -L "https://github.com/docker/compose/releases/download/v2.30.3/docker-compose-$(uname -s)-$(uname -m)"  -o /usr/local/bin/docker-compose
		debug_log mv /usr/local/bin/docker-compose /usr/bin/docker-compose
  		ln -s /usr/bin/docker-compose /usr/local/bin/docker-compose
		debug_log chmod +x /usr/bin/docker-compose

		function_to_insert='docker() {
		  if [[ $1 == "compose" ]]; then
		    /usr/local/bin/docker-compose "${@:2}"
		  else
		    command docker "$@"
		  fi
		}'

		if [ -f "/root/.bashrc" ]; then
		    config_file="/root/.bashrc"
		elif [ -f "/root/.zshrc" ]; then
		    config_file="/root/.zshrc"
		else
		    radovan 1 "ERROR: Neither .bashrc nor .zshrc file found. Exiting."
		fi

		source ~/.bashrc                                               # for armcpu, or if user did sudo su to switch..

		if ! grep -q "docker() {" "$config_file"; then
		    echo "$function_to_insert" >> "$config_file"
		    debug_log "Function 'docker' has been added to $config_file."
		    source "$config_file"
		fi

 	systemctl start docker

	testing_docker=$(timeout 10 docker run --rm alpine echo "Hello from Alpine!")
	if [ "$testing_docker" != "Hello from Alpine!" ]; then              # https://community.openpanel.org/d/157-issue-with-installation-script-error-mysql-container-not-found
		radovan 1: "ERROR: Unable to run the Alpine Docker image! This suggests an issue with connecting to Docker Hub or with the Docker installation itself. To troubleshoot, try running the following command manually: 'docker run --rm alpine'."
	fi

    cp /etc/openpanel/mysql/initialize/1.1/plans.sql /root/initialize.sql  > /dev/null 2>&1

  	chmod +x /etc/openpanel/mysql/scripts/dump.sh                       # added in 1.2.5 for dumping dbs
  	chmod +x /etc/openpanel/openlitespeed/start.sh                      # added in 1.5.6 for openlitespeed

    cd /root || radovan 1 "ERROR: Failed to change directory to /root. OpenPanel needs to be installed by the root user and have write access to the /root directory." # compose doesnt alllow /

    rm -rf /etc/my.cnf .env > /dev/null 2>&1                            # on centos we get default my.cnf, and on repair we already have symlink and .env
    cp /etc/openpanel/docker/compose/new.yml /root/docker-compose.yml > /dev/null 2>&1
    cp /etc/openpanel/docker/compose/.env /root/.env > /dev/null 2>&1

    sed -i "s/^VERSION=.*$/VERSION=\"$PANEL_VERSION\"/" /root/.env

    MYSQL_ROOT_PASSWORD=$(openssl rand -base64 -hex 9)                  # generate random password for mysql
    sed -i 's/MYSQL_ROOT_PASSWORD=.*/MYSQL_ROOT_PASSWORD='"${MYSQL_ROOT_PASSWORD}"'/g' /root/.env  > /dev/null 2>&1
    echo "MYSQL_ROOT_PASSWORD = $MYSQL_ROOT_PASSWORD"
    ln -s /etc/openpanel/mysql/host_my.cnf /etc/my.cnf  > /dev/null 2>&1 # save it to /etc/my.cnf
    sed -i 's/password = .*/password = '"${MYSQL_ROOT_PASSWORD}"'/g' ${ETC_DIR}mysql/host_my.cnf  > /dev/null 2>&1
    sed -i 's/password = .*/password = '"${MYSQL_ROOT_PASSWORD}"'/g' ${ETC_DIR}mysql/container_my.cnf  > /dev/null 2>&1
    os_name=$(grep ^ID= /etc/os-release | cut -d'=' -f2 | tr -d '"')
    if [ "$os_name" == "almalinux" ]; then
        sed -i 's/mysql\/mysql-server/mysql/g' /root/docker-compose.yml   # fix for bug with mysql-server image on Almalinux 9.2
        echo "mysql/mysql-server docker image has known issues on AlmaLinux - editing docker compose to use the mysql:latest instead"
    elif [ "$os_name" == "debian" ]; then
    	echo "Setting AppArmor profiles for Debian"
   		apt install apparmor -y   > /dev/null 2>&1
    fi


    if [ "$REPAIR" = true ]; then
    	echo "Deleting all existing MySQL data in volume root_openadmin_mysql due to the '--repair' flag."
		cd /root && docker compose down > /dev/null 2>&1                            # in case mysql was running
        docker --context default volume rm root_openadmin_mysql > /dev/null 2>&1    # delete database
    fi

    if [ "$architecture" == "aarch64" ]; then
    	sed -i 's/mysql\/mysql-server/mariadb:10-focal/' docker-compose.yml
    fi

    cd /root && docker compose up -d openpanel_mysql > /dev/null 2>&1               # from 0.2.5 we only start mysql by default

    mysql_container=$(docker compose ps -q openpanel_mysql)
    if [ -z "$mysql_container" ]; then
	    radovan 1 "ERROR: MySQL container not found. Please ensure Docker Compose is set up correctly."
    fi

    if ! docker --context default ps -q --no-trunc | grep -q "$mysql_container"; then
        radovan 1 "ERROR: MySQL container is not running. Please retry installation with '--repair' flag."
    else
        echo -e "[${GREEN}OK${RESET}] MySQL service started successfully."
    fi

    ln -s / /hostfs > /dev/null 2>&1                                                 # needed from 1.0.0 for docker contexts to work both inside openpanel ui container and host os
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
   if [ -f /etc/ssh/sshd_config ]; then
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
	   systemctl restart sshd >/dev/null 2>&1 || systemctl restart ssh >/dev/null 2>&1
	   echo -e "[${GREEN} OK ${RESET}] SSH service is configured."
   fi
}


setup_firewall_service() {
    if [ -z "$SKIP_FIREWALL" ]; then
        echo "Installing ConfigServer Firewall & Security.."

        install_csf() {
            wget --inet4-only https://raw.githubusercontent.com/stefanpejcic/sentinelfw/main/csf.tgz > /dev/null 2>&1
            debug_log tar -xzf csf.tgz
            rm csf.tgz
            cd csf
            sh install.sh > /dev/null 2>&1
            cd ..
            rm -rf csf
            echo "Setting CSF auto-login from OpenAdmin interface.."
            if [ "$PACKAGE_MANAGER" == "dnf" ]; then
                debug_log dnf install -y wget curl yum-utils policycoreutils-python-utils libwww-perl
                # fixes bug when starting csf: Can't locate locale.pm in @INC (you may need to install the locale module)
                if [ -f /etc/fedora-release ]; then
                    debug_log yum --allowerasing install perl -y
                elif [ "$PACKAGE_MANAGER" == "apt-get" ]; then
                   debug_log apt-get install -y perl libwww-perl libgd-dev libgd-perl libgd-graph-perl
                fi
                timeout 60s git clone https://github.com/stefanpejcic/csfpost-docker.sh > /dev/null 2>&1
                mv csfpost-docker.sh/csfpost.sh /usr/local/csf/bin/csfpost.sh
                chmod +x /usr/local/csf/bin/csfpost.sh
                rm -rf csfpost-docker.sh
            fi
        }

        open_csf_port() {
            local type=$1   # TCP_IN or TCP_OUT
            local port=$2
            local csf_conf="/etc/csf/csf.conf"

            for dir in "$type" "${type/4/6}"; do
                if grep -q "${dir} = .*${port}" "$csf_conf"; then
                    echo "Port $port already open in $dir"
                else
                    sed -i "s/${dir} = \"\(.*\)\"/${dir} = \"\1,${port}\"/" "$csf_conf"
                    echo "Port $port opened in $dir"
                fi
            done
        }

        edit_csf_conf() {
            echo "Tweaking /etc/csf/csf.conf"
            sed -i 's/TESTING = "1"/TESTING = "0"/' /etc/csf/csf.conf
            sed -i 's/RESTRICT_SYSLOG = "0"/RESTRICT_SYSLOG = "3"/' /etc/csf/csf.conf
            sed -i 's/ETH_DEVICE_SKIP = ""/ETH_DEVICE_SKIP = "docker0"/' /etc/csf/csf.conf
            sed -i 's/DOCKER = "0"/DOCKER = "1"/' /etc/csf/csf.conf
            echo "Copying CSF blocklists" # https://github.com/stefanpejcic/OpenPanel/issues/573
            cp /etc/openpanel/csf/csf.blocklists /etc/csf/csf.blocklists
        }

        set_csf_email_address() {
            email_address=$(grep -E "^e-mail=" $CONFIG_FILE | cut -d "=" -f2)
            if [[ -n "$email_address" ]]; then
                sed -i "s/LF_ALERT_TO = \"\"/LF_ALERT_TO = \"$email_address\"/" /etc/csf/csf.conf
            fi
        }

        function sshd_port() {
            local file_path="/etc/ssh/sshd_config"
            local pattern='Port'
            local port=$(grep -Po "(?<=${pattern}[ =])\d+" "$file_path")
            echo "$port"
        }

        install_csf
        edit_csf_conf

        # OUT ports
        for p in 3306 465 2087; do
            open_csf_port TCP_OUT "$p"
        done

        # IN ports
        for p in 22 53 80 443 2083 2087 32768:60999 21 21000:21010 \
            $(sshd_port); do
            open_csf_port TCP_IN "$p"
        done

        set_csf_email_address
        csf -r    > /dev/null 2>&1
        echo "Restarting CSF service"
        systemctl enable csf
        systemctl restart csf                                                   # also restarts docker at csfpost.sh
		systemctl restart docker                                                # not sure why

        install -m 755 /dev/null /usr/sbin/sendmail                             # https://github.com/stefanpejcic/OpenPanel/issues/338


        if command -v csf > /dev/null 2>&1; then
            echo -e "[${GREEN} OK ${RESET}] ConfigServer Firewall is installed and configured."
        else
            echo -e "[${RED} X  ${RESET}] ConfigServer Firewall is not installed properly."
        fi
    fi
}

update_package_manager() {
    if [ "$SKIP_APT_UPDATE" = false ]; then
        echo "Updating $PACKAGE_MANAGER package manager.."
        debug_log $PACKAGE_MANAGER update -y
    fi
}


setup_imunifyav() {
    if [ "$IMUNIFY_AV" = true ]; then
        echo "Installing ImunifyAV"
        debug_log opencli imunify install
        debug_log opencli imunify start
    fi
}


create_rdnc() {
    if [ "$SKIP_DNS_SERVER" = false ]; then
	    echo "Setting remote name daemon control (rndc) for DNS.."
	    mkdir -p /etc/bind/
	    cp -r /etc/openpanel/bind9/* /etc/bind/

	    if [ -f /etc/os-release ] && grep -qE "Ubuntu|Debian" /etc/os-release; then            # Only on Ubuntu and Debian 12, systemd-resolved is installed
	        echo "DNSStubListener=no" >> /etc/systemd/resolved.conf
	        systemctl restart systemd-resolved
	    fi

	    RNDC_KEY_PATH="/etc/bind/rndc.key"
	    if [ -f "$RNDC_KEY_PATH" ]; then
	        echo "rndc.key already exists."
	        return 0
	    fi

	    echo "Generating rndc.key for DNS zone management."
	    debug_log timeout 90 docker --context default run --rm \
	        -v /etc/bind/:/etc/bind/ \
	        --entrypoint=/bin/sh \
	        ubuntu/bind9:latest \
	        -c 'rndc-confgen -a -A hmac-sha256 -b 256 -c /etc/bind/rndc.key'

	    if [ -f "$RNDC_KEY_PATH" ]; then                                                       # Check if rndc.key was successfully generated
			echo -e "[${GREEN} OK ${RESET}] rndc.key successfully generated."
	    else
	 		echo -e "[${YELLOW}  !  ${RESET}] Warning: Unable to generate rndc.key."
			echo "RNDC is required for managing the named service. Without it, you won’t be able to reload DNS zones."
			echo "That is OK if you don’t plan on using custom nameservers or DNS Clustering on this server."
	    fi

	    find /etc/bind/ -type d -print0 | xargs -0 chmod 755
	    find /etc/bind/ -type f -print0 | xargs -0 chmod 644
	else
		echo "Skipping rndc.key generation due to the '--skip-dns-server' flag."
 	fi
}


extra_step_on_hetzner() {
	if [ -f /etc/hetzner-build ]; then
	    echo "Hetzner provider detected, adding Google DNS resolvers..."
	    echo "info: https://github.com/stefanpejcic/OpenPanel/issues/471"
	    mv /etc/resolv.conf /etc/resolv.conf.bak
	    echo "nameserver 8.8.8.8" >> /etc/resolv.conf
	    echo "nameserver 8.8.4.4" >> /etc/resolv.conf
	fi
}


set_logrotate(){

	echo "Setting Logrotate for Caddy webserver.."
	opencli server-logrotate
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

    case "$PACKAGE_MANAGER" in
        apt-get)
            if [ -f /etc/os-release ] && grep -q "Ubuntu" /etc/os-release; then
                packages=("curl" "cron" "git" "gnupg" "dbus-user-session" "systemd" "dbus" "systemd-container" \
                          "quota" "quotatool" "uidmap" "docker.io" "linux-generic" "default-mysql-client" \
                          "jc" "jq" "sqlite3")
            else
                packages=("curl" "cron" "git" "gnupg" "dbus-user-session" "systemd" "dbus" "systemd-container" \
                          "quota" "quotatool" "uidmap" "docker.io" "linux-image-amd64" "default-mysql-client" \
                          "jc" "jq" "sqlite3")
            fi

            debug_log sed -i 's/#$nrconf{restart} = '"'"'i'"'"';/$nrconf{restart} = '"'"'a'"'"';/g' \
                      /etc/needrestart/needrestart.conf
            debug_log $PACKAGE_MANAGER -qq install apt-transport-https ca-certificates -y
            echo "APT::Acquire::Retries \"3\";" > /etc/apt/apt.conf.d/80-retries
            debug_log update-ca-certificates
            ;;

        yum)
            dnf config-manager --add-repo=https://download.docker.com/linux/centos/docker-ce.repo
            packages=("wget" "git" "gnupg" "dbus-user-session" "systemd" "dbus" "systemd-container" \
                      "quota" "quotatool" "uidmap" "docker-ce" "mysql" "pip" "jc" "sqlite" \
                      "perl-Math-BigInt")
            ;;

        dnf)
            dnf config-manager --add-repo=https://download.docker.com/linux/centos/docker-ce.repo

            if [ -f /etc/fedora-release ]; then
                packages=("git" "wget" "gnupg" "dbus-user-session" "systemd" "dbus" "systemd-container" \
                          "quota" "quotatool" "uidmap" "docker" "docker-compose" "mysql" \
                          "docker-compose-plugin" "sqlite" "sqlite-devel" "perl-Math-BigInt")
            else
                packages=("git" "ncurses" "wget" "gnupg" "systemd" "dbus" "systemd-container" \
                          "quota" "quotatool" "shadow-utils" "docker-ce" "docker-ce-cli" "mariadb" \
                          "containerd.io" "docker-compose-plugin" "sqlite" "sqlite-devel" "perl-Math-BigInt")
            fi

            debug_log dnf install -y yum-utils epel-release perl python3-pip python3-devel gcc
            debug_log yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo -y
            ;;
    esac

    echo -e "Installing $PACKAGE_MANAGER services.."
    for package in "${packages[@]}"; do
        echo -e "Installing ${GREEN}$package${RESET}"
        if ! is_package_installed "$package"; then
            debug_log $PACKAGE_MANAGER install -y "$package" || {
                echo "Error: Installation of $package failed. Retrying.."
            	if [[ "$package" == "docker.io" ]]; then
	                echo "Trying to install docker-ce instead..."
	                $PACKAGE_MANAGER install -y docker-ce || {
	                    radovan 1 "ERROR: Installation of both docker.io and docker-ce failed. Please install manually and then proceed with openpanel installation."
					}
				elif [[ "$package" == "quota" || "$package" == "quotatool" ]]; then
					$PACKAGE_MANAGER install -y "$package" || {
						echo "WARNING: Installation of '$package' failed. You may need to install it manually."
					}
				elif [[ "$package" == "linux-image-amd64" || "$package" == "linux-image" ]]; then
	                $PACKAGE_MANAGER install -y linux-image || {
				 		echo "WARNING: Installation of both linux-image-amd64 and linux-image failed. You may need to install it manually."
					}	
					$PACKAGE_MANAGER install -y "$package" || {
						echo "WARNING: Installation of '$package' failed. You may need to install it manually."
					}	 
				elif [[ "$package" == "dbus-user-session" || "$package" == "dbus" ]]; then
	                $PACKAGE_MANAGER install -y dbus || {
				 		echo "WARNING: Installation of both dbus-user-session and dbus failed. You may need to install it manually."
					}	
					$PACKAGE_MANAGER install -y "$package" || {
						echo "WARNING: Installation of '$package' failed. You may need to install it manually."
					}	
				elif [[ "$package" == "uidmap" || "$package" == "shadow-utils" ]]; then
	                $PACKAGE_MANAGER install -y shadow-utils || {
				 		echo "WARNING: Installation of both uidmap and shadow-utils failed. You may need to install it manually."
					}	
					$PACKAGE_MANAGER install -y "$package" || {
						echo "WARNING: Installation of '$package' failed. You may need to install it manually."
					}	
  
				else
					$PACKAGE_MANAGER install -y "$package" || {
						radovan 1 "ERROR: Installation failed. Please retry installation with '--repair' flag."
					}
				fi
             }
        else
            echo -e "${GREEN}$package is already installed. Skipping.${RESET}"
        fi
    done
}


docker_cpu_limiting() {

	# https://docs.docker.com/engine/security/rootless/#limiting-resources
	mkdir -p /etc/systemd/system/user@.service.d

	cat > /etc/systemd/system/user@.service.d/delegate.conf << EOF
[Service]
Delegate=cpu cpuset io memory pids
EOF

	debug_log systemctl daemon-reload
}


edit_fstab() {

echo "Setting quotas for disk limits of user files"
fstab_file="/etc/fstab"
root_entry=$(grep -E '^\S+\s+/.*' "$fstab_file")

if [[ "$root_entry" =~ "usrquota" && "$root_entry" =~ "grpquota" ]]; then
    echo "Success, usrquota and grpquota are already set for /"
else
    # Add usrquota and grpquota to the fstab entry (only for the root entry)
    sed -i -E '/\s+\/\s+/s/(\S+)(\s+\/\s+\S+\s+\S+)(\s+[0-9]+\s+[0-9]+)$/\1\2,usrquota,grpquota\3/' "$fstab_file"
fi
	systemctl daemon-reload >/dev/null 2>&1
	quotaoff -a >/dev/null 2>&1
	mount -o remount,usrquota,grpquota / >/dev/null 2>&1
	mount /dev/vda1 /mnt >/dev/null 2>&1
	cd /mnt >/dev/null 2>&1
	chmod 600 aquota.user aquota.group >/dev/null 2>&1
	quotacheck -cum / -f >/dev/null 2>&1
	quotaon -a >/dev/null 2>&1
	repquota / >/dev/null 2>&1
	quota -v >/dev/null 2>&1

    debug_log "Testing quotas.."
    repquota -u / > /etc/openpanel/openpanel/core/users/repquota 2>/dev/null
    if [ $? -eq 0 ]; then
        echo -e "[${GREEN} OK ${RESET}] Quotas are now enabled for users."
    else
        echo -e "[${RED} FAIL ${RESET}] Quota check failed."
    fi


}

set_system_cronjob() {
    echo "Setting cronjobs..."
    install -m 600 -o root -g root "${ETC_DIR}cron" /etc/cron.d/openpanel

    if [[ "$PACKAGE_MANAGER" == "dnf" || "$PACKAGE_MANAGER" == "yum" ]]; then                       # Handle SELinux if using dnf or yum
        restorecon -R /etc/cron.d >/dev/null 2>&1
        systemctl restart crond.service >/dev/null 2>&1
    fi

    if systemctl is-active --quiet crond 2>/dev/null || systemctl is-active --quiet cron 2>/dev/null; then
        echo -e "[${GREEN} OK ${RESET}] ."
		[ -f /etc/cron.d/openpanel ] && echo -e "[${GREEN} OK ${RESET}] Cron service is running and crontab configured."
    else
        echo -e "[${RED} FAIL ${RESET}] Cron service is NOT running!"
    fi

}


set_custom_hostname(){
        if [ "$SET_HOSTNAME_NOW" = true ]; then
		    if [[ $new_hostname =~ ^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
	            sed -i "s/example\.net/$new_hostname/g" /etc/openpanel/caddy/Caddyfile
	        else
                echo "Hostname provided: $new_hostname is not a valid domain name, OpenPanel will use IP address $current_ip for access."
            fi
        fi
}



opencli_setup(){
    echo "Downloading OpenCLI and adding to path.."
    cd /usr/local
	[ "$REPAIR" = true ] && rm -rf /usr/local/opencli
    timeout 60s git clone https://github.com/stefanpejcic/opencli.git

	if [ ! -d "/usr/local/opencli" ]; then
	 	radovan 1 "Failed to clone OpenCLI from Github - please retry install with '--retry --debug' flags."
	fi

    chmod +x -R /usr/local/opencli
    ln -s /usr/local/opencli/opencli /usr/local/bin/opencli
    echo "# opencli aliases
    ALIASES_FILE=\"/usr/local/opencli/aliases.txt\"
    generate_autocomplete() {
        awk '{print \$NF}' \"\$ALIASES_FILE\"
    }
    complete -W \"\$(generate_autocomplete)\" opencli" >> ~/.bashrc

    # Fix for: The command could not be located because '/usr/local/bin' is not included in the PATH environment variable.
    export PATH="/usr/bin:$PATH"

    source ~/.bashrc

    echo "Testing 'opencli' commands:"
    if [ -x "/usr/local/bin/opencli" ]; then
        echo -e "[${GREEN} OK ${RESET}] opencli commands are available."
    else
        radovan 1 "'opencli --version' command failed."
    fi
}



enable_dev_mode() {
	if [ "$DEV_MODE" = true ]; then
	    echo "Enabling dev_mode"
	    opencli config update dev_mode "on" > /dev/null 2>&1
	fi
}

set_premium_features(){
 if [ "$SET_PREMIUM" = true ]; then
    LICENSE="Enterprise"
    echo "Setting OpenPanel enterprise version license key $license_key"
    opencli config update key "$license_key"
    systemctl restart admin > /dev/null 2>&1
    echo "Setting mailserver.."
    timeout 60 opencli email-server install #added in 0.2.5 https://community.openpanel.org/d/91-email-support-for-openpanel-enterprise-edition
    echo "Enabling Roundcube webmail.."
    timeout 60 opencli email-webmail roundcube
 else
    LICENSE="Community"
 fi
}


log_dirs() {
    local log_dir="/var/log/openpanel"  # https://dev.openpanel.com/logs.html
    install -d -m 755 "$log_dir" "$log_dir/user" "$log_dir/admin"
}


set_email_address_and_email_admin_logins(){
        if [ "$SEND_EMAIL_AFTER_INSTALL" = true ]; then
            if [[ $EMAIL =~ ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
                echo "Setting email address $EMAIL for notifications"
                opencli config update email "$EMAIL"

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

                  PROTOCOL="http"
                  admin_domain="127.0.0.1"

                  if [ "$SET_HOSTNAME_NOW" = true ]; then
                  	if [[ $new_hostname =~ ^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
                  		if [ -f "/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/$new_hostname/$new_hostname.key" ]; then
                  			PROTOCOL="https"
                  			admin_domain="$new_hostname"
                  		fi
                  	fi
                  fi
                  curl -4 -k -X POST "$PROTOCOL://$admin_domain:2087/send_email" -F "transient=$TRANSIENT" -F "recipient=$EMAIL" -F "subject=$title" -F "body=$message" --max-time 15
                }

                server_hostname=$(hostname)
                email_notification "OpenPanel successfully installed" "OpenAdmin URL: http://$server_hostname:2087/ | username: $new_username  | password: $new_password"
            else
                echo "Address provided: $EMAIL is not a valid email address. Admin login credentials and future notifications will not be sent."
            fi
        fi
}



generate_and_set_ssl_for_panels() {
    if [ "$SET_HOSTNAME_NOW" = true ]; then
        echo "Checking if SSL can be generated for the server hostname.."
        CADDYFILE="/etc/openpanel/caddy/Caddyfile"
        HOSTNAME=$(awk '/# START HOSTNAME DOMAIN #/{flag=1; next} /# END HOSTNAME DOMAIN #/{flag=0} flag' "$CADDYFILE" | awk 'NF {print $1; exit}')

        if [[ -n "$HOSTNAME" && "$HOSTNAME" != "example.net" ]]; then
            cd /root && docker --context default compose up -d caddy               # start and generate ssl

            MAX_RETRIES=5
            SLEEP_SECONDS=5
            SUCCESS=0
            for ((i=1; i<=MAX_RETRIES; i++)); do
                debug_log echo "Attempt $i to generate SSL for $HOSTNAME..."
                if curl -4 -sf -o /dev/null "https://$HOSTNAME"; then
                    debug_log echo "SSL certificate is ready! OpenAdmin is now using HTTPS protocol."
                    SUCCESS=1
                    debug_log systemctl restart admin
                    break
                else
                    debug_log echo "SSL not ready yet, retrying in $SLEEP_SECONDS seconds..."
                    debug_log docker restart caddy
                    sleep $SLEEP_SECONDS
                fi
            done
            if [ $SUCCESS -ne 1 ]; then
                echo "Failed to generate SSL certificate after $MAX_RETRIES attempts. OpenAdmin fallback to using HTTP protocol."
            fi
        fi
    fi
}



download_ui_image() {
        echo "Pulling OpenPanel image in background (not starting the service).."
		nohup sh -c "cd /root && docker --context default compose pull openpanel" </dev/null >nohup.out 2>nohup.err &
  }


setup_redis_service() {
	install -d -m 777 /tmp/redis
}

run_custom_postinstall_script() {
    if [ -n "$post_install_path" ]; then
        echo "Running post install script.."
        debug_log bash $post_install_path
    fi
}


verify_license() {
    debug_log "echo Current time: $(date +%T)"
    server_hostname=$(hostname)
    license_data='{"hostname": "'"$server_hostname"'", "public_ip": "'"$current_ip"'"}'
    response=$(curl -4 -s -X POST -H "Content-Type: application/json" -d "$license_data" https://api.openpanel.com/license/index.php)
    debug_log "echo Checking OpenPanel license for IP address: $current_ip"
    debug_log "echo Response: $response"
}

download_skeleton_directory_from_github() {
    local repo_url="https://github.com/stefanpejcic/openpanel-configuration"

	[ "$REPAIR" = true ] && rm -rf "$ETC_DIR"

    echo "Downloading configuration files to ${ETC_DIR}..."
    timeout 60s git clone "$repo_url" "$ETC_DIR" >/dev/null 2>&1 || \
        radovan 1 "Failed to clone OpenPanel Configuration from GitHub - retry with '--retry --debug'."

    [ -f "$CONFIG_FILE" ] || radovan 1 "Main configuration file ${CONFIG_FILE} is missing."
    systemctl daemon-reload >/dev/null 2>&1
}


setup_bind(){
    if [ "$SKIP_DNS_SERVER" = true ]; then
        echo "Skipping BIND setup due to the '--skip-dns-server' flag."
        return
    fi

    echo "Setting up DNS service..."
	install -d -m 755 /etc/bind
	cp -r /etc/openpanel/bind9/* /etc/bind/

	if [ -f /etc/os-release ] && grep -Eq "Ubuntu|Debian" /etc/os-release; then
		grep -q "^DNSStubListener=no" /etc/systemd/resolved.conf || \
			echo "DNSStubListener=no" >> /etc/systemd/resolved.conf
		systemctl restart systemd-resolved # systemd-resolved is installed on debian based
	fi

	chmod -R 755 /etc/bind
}

send_install_log(){
    # Restore normal output to the terminal, so we dont save generated admin password in log file!
    exec > /dev/tty
    exec 2>&1
    opencli report --public >> "$LOG_FILE"
    curl -4 --max-time 15 -k -F "file=@/root/$LOG_FILE" https://support.openpanel.org/install_logs.php
    # Redirect again stdout and stderr to the log file
    exec > >(tee -a "$LOG_FILE")
    exec 2>&1
}


rm_helpers(){
    rm -rf $PROGRESS_BAR_FILE
}



setup_swap(){

    create_swap() {
        fallocate -l ${SWAP_FILE}G /swapfile > /dev/null 2>&1
        chmod 600 /swapfile
        mkswap /swapfile
        swapon /swapfile
        echo "/swapfile   none    swap    sw    0   0" >> /etc/fstab

 		echo -e "[${GREEN} OK ${RESET}] Created SWAP file of ${SWAP_FILE}G."
    }

    if [ -n "$(swapon -s)" ]; then
        echo "ERROR: Skipping creating swap space as there already exists a swap partition."
        return
    fi

    if [ "$SETUP_SWAP_ANYWAY" = true ]; then
        create_swap
    else
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
    local discord_invite_url="https://discord.openpanel.com/"
    local forums_link="https://community.openpanel.org/"
    local docs_link="https://openpanel.com/docs/admin/intro/"
    local docs_get_started_link="https://openpanel.com/docs/admin/intro/#post-install-steps"
    local github_link="https://github.com/stefanpejcic/OpenPanel/"
    local tickets_url="https://my.openpanel.com/submitticket.php?step=2&deptid=2"

    if [[ "$LICENSE" == "Enterprise" ]]; then
        echo ""
        echo "🎉 Welcome aboard and thank you for choosing OpenPanel Enterprise edition! 🎉"
        echo ""
        echo "Need assistance or looking to learn more? We've got you covered:"
        echo "  - Check the Admin Docs: $docs_link"
        echo "  - Open Support Ticket: $tickets_url"
        echo "  - Chat with us on Discord: $discord_invite_url"
        echo ""
    else
        echo ""
        echo "🎉 Welcome aboard and thank you for choosing OpenPanel! 🎉"
        echo ""
        echo "To get started, check out our Post Install Steps:"
        echo "👉 $docs_get_started_link"
        echo ""
        echo "Join our community and connect with us on:"
        echo "  - Github: $github_link"
        echo "  - Discord: $discord_invite_url"
        echo "  - Our community forums: $forums_link"
        echo ""
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


check_permissions_in_root_dir() {
	local root_dir="/root"
	if ! [ -r "$root_dir" ] || ! [ -w "$root_dir" ]; then
		radovan 1 "User $(whoami) does NOT have read and write permissions on $root_dir"
	fi
}


install_python() {
    OS=$(grep '^ID=' /etc/os-release | cut -d= -f2 | tr -d '"')
    CODENAME=$(grep '^VERSION_CODENAME=' /etc/os-release | cut -d= -f2 | tr -d '"')

    # Default
    local chosen_py="python3"

    # Helper to install venv package for apt-based systems (when using distro python)
    install_venv_pkg() {
        if [ "$PACKAGE_MANAGER" = "apt-get" ]; then
            debug_log $PACKAGE_MANAGER install -y python3-venv || true
        fi
    }

    if [ "$OS" = "debian" ] && [ "$CODENAME" = "trixie" ]; then
        # Debian 13 ships Python 3.13 (cgi removed). Build latest Python 3.12.x from source to keep compatibility.
        echo "Building latest Python 3.12.x from source for Debian 13 (trixie) … will take approximately 3-10 minutes."
        debug_log $PACKAGE_MANAGER update -y
        # Minimal deps (no tk/X11)
        debug_log $PACKAGE_MANAGER install -y build-essential curl ca-certificates xz-utils \
            libssl-dev zlib1g-dev libbz2-dev libreadline-dev libsqlite3-dev \
            libffi-dev libncursesw5-dev libgdbm-dev liblzma-dev uuid-dev

        # Ensure curl exists even on very minimal images
        if ! command -v curl >/dev/null 2>&1; then
            debug_log $PACKAGE_MANAGER install -y curl ca-certificates
        fi

        SRCDIR=/usr/local/src
        debug_log mkdir -p "$SRCDIR"

        # Detect latest 3.12.x from python.org; fallback to 3.12.7 if detection fails
        LATEST_312="$(curl -fsSL https://www.python.org/ftp/python/ \
            | grep -Po 'href=\"3\\.12\\.\\d+/' \
            | grep -Po '3\\.12\\.\\d+' \
            | sort -V | tail -1 || true)"
        if [ -z "$LATEST_312" ]; then
            LATEST_312="3.12.7"
        fi

        PREFIX="/opt/python/${LATEST_312}"
        TARBALL="Python-${LATEST_312}.tgz"
        URL="https://www.python.org/ftp/python/${LATEST_312}/${TARBALL}"

        debug_log bash -lc "cd '$SRCDIR' && curl -fsSL -o '${TARBALL}' '${URL}' && \
            rm -rf 'Python-${LATEST_312}' && \
            tar -xzf '${TARBALL}' && cd 'Python-${LATEST_312}' && \
            ./configure --prefix='${PREFIX}' --enable-optimizations --with-ensurepip=install && \
            make -j\"$(nproc)\" && make altinstall"

        chosen_py="${PREFIX}/bin/python3.12"
        [ -x "$chosen_py" ] || radovan 1 "Custom Python not found at $chosen_py (build may have failed)."
        # Ensure pip tooling is up to date for the custom build
        $chosen_py -m ensurepip -U || true
        $chosen_py -m pip install --upgrade pip setuptools wheel || true

    else
        # Legacy behavior for other distros – prefer 3.x from system repos
        if command -v python3.12 &>/dev/null; then
            chosen_py="python3.12"
            install_venv_pkg
        else
            echo "Installing distro Python and venv tooling …"
            if [ "$PACKAGE_MANAGER" = "apt-get" ]; then
                debug_log $PACKAGE_MANAGER install -y software-properties-common || true
            fi

            if [ "$OS" = "ubuntu" ]; then
                debug_log add-apt-repository -y ppa:deadsnakes/ppa
                debug_log $PACKAGE_MANAGER update -y
                debug_log $PACKAGE_MANAGER install -y python3.12 python3.12-venv || true
                if command -v python3.12 &>/dev/null; then
                    chosen_py="python3.12"
                else
                    chosen_py="python3"
                    install_venv_pkg
                fi

            elif [ "$OS" = "debian" ]; then
                # Older Debian – try backports for 3.12; fall back to system python3
                debug_log install -d -m 0755 /etc/apt/keyrings
                debug_log bash -lc 'curl -fsSL --ipv4 https://pascalroeleven.nl/deb-pascalroeleven.gpg | tee /etc/apt/keyrings/deb-pascalroeleven.gpg >/dev/null' || true
                debug_log bash -lc 'cat >> /etc/apt/sources.list.d/pascalroeleven.sources <<EOF
Types: deb
URIs: http://deb.pascalroeleven.nl/python3.12
Suites: '"${CODENAME}"'-backports
Components: main
Signed-By: /etc/apt/keyrings/deb-pascalroeleven.gpg
EOF' || true
                debug_log $PACKAGE_MANAGER update -y
                debug_log $PACKAGE_MANAGER install -y python3.12 python3.12-venv || true
                if command -v python3.12 &>/dev/null; then
                    chosen_py="python3.12"
                else
                    chosen_py="python3"
                    install_venv_pkg
                fi

            elif [ "$OS" = "almalinux" ] || [ "$OS" = "alma" ] || [ "$OS" = "rocky" ] || [ "$OS" = "centos" ]; then
                debug_log $PACKAGE_MANAGER update -y
                if command -v dnf >/dev/null 2>&1; then
                    debug_log dnf install -y epel-release || true
                    debug_log dnf config-manager --set-enabled crb || debug_log dnf config-manager --set-enabled powertools || true
                fi
                debug_log $PACKAGE_MANAGER install -y python3.12 || true
                if command -v python3.12 &>/dev/null; then
                    chosen_py="python3.12"
                else
                    chosen_py="python3"
                fi

            else
                # Fallback
                chosen_py="python3"
                install_venv_pkg
            fi
        fi
    fi

    # Sanity & export for downstream steps
    $chosen_py --version &>/dev/null || radovan 1 "Python installation failed for ${chosen_py}."
    export PYTHON_BIN="$chosen_py"
}



extra_step_for_caddy() {
	sed -i "s/example\.net/$current_ip/g" /etc/openpanel/caddy/redirects.conf > /dev/null 2>&1
}





configure_coraza() {
	if [ "$CORAZA" = true ]; then
		echo "Installing CorazaWAF and setting OWASP core ruleset.."
		debug_log mkdir -p /etc/openpanel/caddy/
		debug_log wget --inet4-only https://raw.githubusercontent.com/corazawaf/coraza/v3/dev/coraza.conf-recommended -O /etc/openpanel/caddy/coraza_rules.conf
  		[ "$REPAIR" = true ] && rm -rf /etc/openpanel/caddy/coreruleset/
		debug_log timeout 60s git clone https://github.com/coreruleset/coreruleset /etc/openpanel/caddy/coreruleset/
	else
 		echo "Disabling CorazaWAF: setting caddy:latest docker image instead of openpanel/caddy-coraza"
		sed -i 's|image: .*caddy.*|image: caddy:latest|' /root/docker-compose.yml
	fi
}


install_openadmin(){
    echo "Setting up OpenAdmin panel.."
    local openadmin_dir="/usr/local/admin/"

    [ "$REPAIR" = true ] && rm -rf "$openadmin_dir"

    mkdir -p $openadmin_dir
    debug_log echo "Downloading OpenAdmin files"

    local branch="110"
    [ "$architecture" = "aarch64" ] && branch="armcpu"

    timeout 60s git clone -b "$branch" --single-branch https://github.com/stefanpejcic/openadmin "$openadmin_dir" || {
        radovan 1 "Failed to clone OpenAdmin from Github - please retry install with '--retry --debug' flags."
    }

    cd "$openadmin_dir" || exit 1

	${PYTHON_BIN:-python3} -m venv ${openadmin_dir}venv
	source ${openadmin_dir}venv/bin/activate

    pip install --default-timeout=300 --force-reinstall --ignore-installed -r requirements.txt > /dev/null 2>&1 \
        || pip install --default-timeout=300 --force-reinstall --ignore-installed -r requirements.txt --break-system-packages > /dev/null 2>&1

     # on debian12 yaml is also needed to read conf files!
     if [ -f /etc/os-release ] && grep -q "Debian" /etc/os-release; then
     	apt install python3-yaml -y  > /dev/null 2>&1
     fi

    cp -fr /etc/openpanel/openadmin/service/openadmin.service ${SERVICES_DIR}admin.service  > /dev/null 2>&1
    cp -fr /etc/openpanel/openadmin/service/watcher.service ${SERVICES_DIR}watcher.service  > /dev/null 2>&1

    systemctl daemon-reload  > /dev/null 2>&1
    systemctl start admin > /dev/null 2>&1
    systemctl enable admin > /dev/null 2>&1

	if [ "$SKIP_DNS_SERVER" = false ]; then
	    chmod +x /etc/openpanel/services/watcher.sh
	    systemctl start watcher > /dev/null 2>&1
 	    systemctl enable watcher > /dev/null 2>&1

	else
	    echo "Skipping Watcher service setup due to the '--skip-dns-server' flag."
 	fi

	echo "Testing if OpenAdmin service is available on default port '2087':"
	if ss -tuln | grep ':2087' >/dev/null; then
		echo -e "[${GREEN} OK ${RESET}] OpenAdmin service is running."
	else
		radovan 1 "OpenAdmin service is NOT listening on port 2087."
	fi
}


create_admin_and_show_logins_success_message() {

    ln -s ${ETC_DIR}ssh/admin_welcome.sh /etc/profile.d/welcome.sh   #motd
    chmod +x /etc/profile.d/welcome.sh

    echo -e "${GREEN}OpenPanel ${LICENSE} $PANEL_VERSION installation complete.${RESET}"
    echo ""

    if [ "$SET_ADMIN_USERNAME" = true ]; then
       new_username="${custom_username}"
    else
       wget --inet4-only -O /tmp/generate.sh https://gist.githubusercontent.com/stefanpejcic/905b7880d342438e9a2d2ffed799c8c6/raw/a1cdd0d2f7b28f4e9c3198e14539c4ebb9249910/random_username_generator_docker.sh > /dev/null 2>&1

       if [ -f "/tmp/generate.sh" ]; then
	       source /tmp/generate.sh
	       new_username=($random_name)
       else
	       new_username="admin"
       fi

    fi

	if [ "$SET_ADMIN_PASSWORD" = true ]; then
	    if [[ "$custom_password" =~ ^[A-Za-z0-9]{5,16}$ ]]; then
	        new_password="${custom_password}"
	    else
	        echo "Warning: provided password is invalid (must be alphanumeric and 5–16 characters). Generating a secure password."
	        new_password=$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 16)
	    fi
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
    opencli admin new "$new_username" "$new_password" --super > /dev/null 2>&1 &&
	user_exists=$(sqlite3 /etc/openpanel/openadmin/users.db "SELECT COUNT(*) FROM user WHERE username = '$new_username';")
	if [ "$user_exists" -gt 0 ]; then
	    echo "User $new_username has been successfully added."
		display_admin_status_and_logins
	else
	    echo "WARNING: Admin user $new_username was not created using opencli. Trying to insert user manually to SQLite database.."
	    password_hash=$(/usr/local/admin/venv/bin/python3 /usr/local/admin/core/users/hash $new_password)
	    create_table_sql="CREATE TABLE IF NOT EXISTS user (id INTEGER PRIMARY KEY, username TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL, role TEXT NOT NULL DEFAULT 'user', is_active BOOLEAN DEFAULT 1 NOT NULL);"
	    insert_user_sql="INSERT INTO user (username, password_hash, role) VALUES ('$new_username', '$password_hash', 'admin');"
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

(
flock -n 200 || { echo "Error: Another instance of the install script is already running. Exiting."; exit 1; }
# shellcheck disable=SC2068
parse_args "$@"
check_permissions_in_root_dir
get_server_ipv4
set_version_to_install
print_header
check_requirements
detect_installed_panels
install_started_message
main
rm_helpers
print_space_and_line
support_message
print_space_and_line
send_install_log
create_admin_and_show_logins_success_message
run_custom_postinstall_script
)200>/root/openpanel_install.lock
