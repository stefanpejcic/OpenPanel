#!/bin/bash
################################################################################
# OpenPanel Installer ✌️
# https://openpanel.com/install
#
# Supported OS:            Ubuntu, Debian, AlmaLinux, RockyLinux, CentOS
# Supported Architecture:  x86_64(AMD64), AArch64(ARM64)
#
# Usage:                   bash <(curl -sSL https://openpanel.org)
# Author:                  Stefan Pejcic <stefan@pejcic.rs>
# Created:                 11.07.2023
# Last Modified:           15.04.2026
################################################################################
# shellcheck disable=SC2015

GREEN='\033[0;32m'; YELLOW='\033[0;33m'; RED='\033[0;31m'; RESET='\033[0m'
export TERM=xterm-256color
export DEBIAN_FRONTEND=noninteractive

# defaults
PANEL_VERSION=""
CUSTOM_VERSION=false
ADMIN_PORT=2087
USER_PORT=2083
SKIP_APT_UPDATE=false
SKIP_DNS_SERVER=false
SKIP_FIREWALL=false
REPAIR=false
SET_HOSTNAME_NOW=false
USE_SELFSIGNED=false
SETUP_SWAP_ANYWAY=false
CORAZA=true
IMUNIFY_AV=false
SWAP_FILE=1
SEND_EMAIL_AFTER_INSTALL=false
SET_PREMIUM=false
SET_ADMIN_USERNAME=false
SET_ADMIN_PASSWORD=false
LICENSE="Community"
post_install_path=""
new_hostname=""
separate_panel_domain=""
custom_username=""
custom_password=""
EMAIL=""
license_key=""

readonly DEFAULT_PANEL_VERSION="1.7.53"
readonly DOCKER_COMPOSE_VERSION="v2.40.2"
readonly ETC_DIR="/etc/openpanel/"
readonly LOG_FILE="openpanel_install.log"
readonly SERVICES_DIR="/etc/systemd/system/"
readonly CONFIG_FILE="${ETC_DIR}openpanel/conf/openpanel.config"

exec > >(tee -a "$LOG_FILE") 2>&1
echo "" > /root/openpanel_restart_needed

ok()   { echo -e "[${GREEN} OK ${RESET}] $*"; }
warn() { echo -e "[${YELLOW}  ! ${RESET}] $*"; }
fail() { echo -e "[${RED}  X ${RESET}] $*"; }

die() {
    echo -e "${RED}INSTALLATION FAILED${RESET} - Please retry with '--repair' flag\nError: $2" >&2
    exit 1
}

run() {
    local ts; ts=$(date +'%Y-%m-%d %H:%M:%S')
    echo "[$ts] COMMAND: $*" >> "$LOG_FILE"
    "$@" >/dev/null 2>&1
}

pkg_installed() {
    case "$PACKAGE_MANAGER" in
        yum|dnf) "$PACKAGE_MANAGER" list installed "$1" &>/dev/null ;;
        apt-get) dpkg -l 2>/dev/null | grep -q "^ii[[:space:]]*${1}" ;;
    esac
}

line() { printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -; }

show_help() {
    cat <<EOF
Available options:
  --key=<key>                 License key for OpenPanel Enterprise edition.
  --domain=<domain>           Domain for OpenAdmin and OpenPanel.
  --panel-domain=<domain>     Separate domain just for OpenPanel UI.
  --username=<username>       Admin username (random if not provided).
  --password=<password>       Admin password (random if not provided).
  --version=<version>         Custom OpenPanel version to install.
  --email=<email>             Email to receive admin credentials.
  --admin-port=<port>         Port for OpenAdmin (default: 2087).
  --user-port=<port>          Port for OpenPanel (default: 2083).
  --imunifyav                 Install and set up ImunifyAV.
  --skip-requirements         Skip requirements check.
  --skip-panel-check          Skip check for existing panels.
  --skip-apt-update           Skip package manager update.
  --skip-firewall             Skip Sentinel Firewall installation.
  --skip-dns-server           Skip DNS (Bind9) setup.
  --no-waf                    Disable CorazaWAF / OWASP ruleset.
  --post_install=<path>       Post-install script path or URL.
  --swap=<1-10>               Swap size in GB.
  --selfsigned                Use a self-signed SSL certificate.
  --repair | --retry          Retry and overwrite existing installation.
  -h, --help                  Show this help message.
EOF
}

validate_port() {
    local name=$1 val=$2
    if [[ "$val" =~ ^[0-9]+$ ]] && (( val >= 1000 && val <= 30000 )); then
        echo "$val"
    else
        echo "Error: $name must be between 1000 and 30000" >&2; exit 1
    fi
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --key=*)           SET_PREMIUM=true;         license_key="${1#*=}" ;;
            --domain=*)        SET_HOSTNAME_NOW=true;    new_hostname="${1#*=}" ;;
            --panel-domain=*)  SET_HOSTNAME_NOW=true;    separate_panel_domain="${1#*=}" ;;
            --username=*)      SET_ADMIN_USERNAME=true;  custom_username="${1#*=}" ;;
            --password=*)      SET_ADMIN_PASSWORD=true;  custom_password="${1#*=}" ;;
            --post_install=*)  post_install_path="${1#*=}" ;;
            --version=*)       CUSTOM_VERSION=true;      PANEL_VERSION="${1#*=}" ;;
            --swap=*)          SETUP_SWAP_ANYWAY=true;   SWAP_FILE="${1#*=}" ;;
            --email=*)         SEND_EMAIL_AFTER_INSTALL=true; EMAIL="${1#*=}" ;;
            --admin-port=*)    ADMIN_PORT=$(validate_port "admin-port" "${1#*=}") ;;
            --user-port=*)     USER_PORT=$(validate_port "user-port"  "${1#*=}") ;;
            --skip-requirements) SKIP_REQUIREMENTS=true ;;
            --skip-panel-check)  SKIP_PANEL_CHECK=true ;;
            --skip-apt-update)   SKIP_APT_UPDATE=true ;;
            --skip-dns-server)   SKIP_DNS_SERVER=true ;;
            --skip-firewall)     SKIP_FIREWALL=true ;;
            --imunifyav)         IMUNIFY_AV=true ;;
            --no-waf)            CORAZA=false ;;
            --selfsigned)        USE_SELFSIGNED=true ;;
            --repair|--retry)    REPAIR=true; SKIP_PANEL_CHECK=true; SKIP_APT_UPDATE=true ;;
            -h|--help)           show_help; exit 0 ;;
            *) echo "Unknown option: $1"; show_help; exit 1 ;;
        esac
        shift
    done
}

check_requirements() {
    [[ -n "${SKIP_REQUIREMENTS:-}" ]] && return

    [[ "$(id -u)" == "0" ]] || die 1 "You must be root to run this script."
    [[ "$(uname)" != "Darwin" ]] || die 1 "macOS is not supported."

	if [[ -f /.dockerenv || -f /run/.containerenv ]] || { tr '\0' '\n' < /proc/1/environ | grep -qi '^container='; }; then
        die 1 "Running inside a container is not supported."
    fi

    local ram_mb disk_mb
    ram_mb=$(( $(grep MemTotal /proc/meminfo | awk '{print $2}') / 1024 ))
    disk_mb=$(( $(df / --output=avail | tail -1) / 1024 ))
    (( ram_mb  >= 1024 )) || die 1 "At least 1 GB RAM required (detected: ${ram_mb} MB)."
    (( disk_mb >= 5120 )) || die 1 "At least 5 GB free disk space required (detected: ${disk_mb} MB)."
}

detect_installed_panels() {
    [[ -n "${SKIP_PANEL_CHECK:-}" ]] && return

    declare -A panels=([/usr/local/admin/]="OpenPanel" [/usr/local/cpanel/whostmgr]="cPanel WHM" [/opt/psa/version]="Plesk" [/usr/local/psa/version]="Plesk" [/usr/local/CyberPanel]="CyberPanel" [/usr/local/directadmin]="DirectAdmin" [/usr/local/mgr5]="ispmanager" [/usr/local/cwpsrv]="CentOS Web Panel (CWP)" [/usr/local/vesta]="VestaCP" [/usr/local/hestia]="HestiaCP" [/usr/local/httpd]="Apache WebServer" [/usr/local/apache2]="Apache WebServer" [/usr/sbin/httpd]="Apache WebServer" [/sbin/httpd]="Apache WebServer" [/usr/lib/nginx]="Nginx WebServer")
    for path in "${!panels[@]}"; do
        [[ -e "$path" ]] || continue
        local name="${panels[$path]}"
        if [[ "$name" == "OpenPanel" ]]; then
            die 1 "OpenPanel is already installed. To update, run: opencli update --force"
        elif [[ "$name" == *WebServer* ]]; then
            die 1 "$name is already installed. OpenPanel requires a clean server with no web servers."
        else
            die 1 "$name is installed. OpenPanel requires a clean server with no control panels."
        fi
    done

    ok "No conflicting panels or web servers found."
}

detect_os_and_package_manager() {
    [[ -f /etc/os-release ]] || die 1 "Cannot detect OS: /etc/os-release not found."
	# shellcheck disable=SC1091
    . /etc/os-release
    OS_ID="${ID,,}"
    OS_VERSION_ID="${VERSION_ID:-}"
    OS_CODENAME="${VERSION_CODENAME:-}"
    OS_NAME="${NAME:-}"
    export OS_ID OS_VERSION_ID OS_CODENAME OS_NAME	
	
    case "$OS_ID" in
        ubuntu|debian)               PACKAGE_MANAGER="apt-get" ;;
        fedora|rocky|almalinux|alma) PACKAGE_MANAGER="dnf" ;;
        centos)                      PACKAGE_MANAGER="yum" ;;
        *) die 1 "Unsupported OS: $OS_ID" ;;
    esac
    case "$(uname -m)" in
        x86_64|amd64)  architecture="x86_64" ;;
        aarch64|arm64) architecture="aarch64" ;;
        *)             architecture="$(uname -m)" ;;
    esac
}

get_server_ipv4() {
    local ip
    ip=$(curl -s --max-time 3 -4 "https://ip.openpanel.com" || true)
    [[ -z "$ip" ]] && ip=$(ip -4 addr show scope global | grep -oP '(?<=inet\s)\d+(\.\d+){3}' | head -n1)
    [[ "$ip" =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}$ ]] || warn "Could not determine a valid public IPv4 address."
    SERVER_IPV4_ADDRESS="$ip"
}

set_panel_version() {
    if [[ "$CUSTOM_VERSION" == false ]]; then
        local response
        response=$(curl -4 -s "https://api.openpanel.com/statistics/" || true)
        PANEL_VERSION=$(echo "$response" | grep -oP '"latest_version":"\K[^"]+' || true)
        [[ "$PANEL_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || PANEL_VERSION="$DEFAULT_PANEL_VERSION"
    fi
}

print_header() {
    line
    echo -e "   ____                         _____                      _  "
    echo -e "  / __ \                       |  __ \                    | | "
    echo -e " | |  | | _ __    ___  _ __    | |__) | __ _  _ __    ___ | | "
    echo -e " | |  | || '_ \  / _ \| '_ \   |  ___/ / _\" || '_ \ / _  \| | "
    echo -e " | |__| || |_) ||  __/| | | |  | |    | (_| || | | ||  __/| | "
    echo -e "  \____/ | .__/  \___||_| |_|  |_|     \__,_||_| |_| \___||_| "
    echo -e "         | |                                                  "
    echo -e "         |_|                                   version: ${GREEN}$PANEL_VERSION${RESET} "
    line
    echo -e "  OS: ${GREEN}${OS_NAME^^} ${OS_VERSION_ID}${RESET}  |  Arch: ${GREEN}${architecture^^}${RESET}  |  IP: ${GREEN}${SERVER_IPV4_ADDRESS}${RESET}  |  Pkg: ${GREEN}${PACKAGE_MANAGER^^}${RESET}"
    line
}

update_package_manager() {
    [[ "$SKIP_APT_UPDATE" == true ]] && return
    echo "Updating $PACKAGE_MANAGER..."
    run $PACKAGE_MANAGER update -y && ok "all packages updated." || warn "Could not update all packages."
}

pkg_install_with_retry() {
    local pkg=$1
    pkg_installed "$pkg" && { echo -e "${GREEN}$pkg already installed, skipping.${RESET}"; return; }

    echo -e "Installing ${GREEN}${pkg}${RESET}..."
    $PACKAGE_MANAGER install -y "$pkg" >/dev/null 2>&1 && return

    case "$pkg" in
        docker.io)    $PACKAGE_MANAGER install -y docker-ce >/dev/null 2>&1 && return ;;
        linux-image-amd64) $PACKAGE_MANAGER install -y linux-image >/dev/null 2>&1 && return ;;
        dbus-user-session) $PACKAGE_MANAGER install -y dbus >/dev/null 2>&1 && return ;;
        uidmap)       $PACKAGE_MANAGER install -y shadow-utils >/dev/null 2>&1 && return ;;
        quota|quotatool) warn "Could not install $pkg — you may need to install it manually."; return ;;
    esac

    local attempt=1 max=10 delay=30
    until $PACKAGE_MANAGER install -y "$pkg" >/dev/null 2>&1; do
        (( attempt++ ))
        (( attempt > max )) && die 1 "Failed to install $pkg after $max attempts."
        echo "Retry $attempt/$max for $pkg in ${delay}s..."
        sleep $delay
    done
}

build_quotatool_from_source() {
    quotatool -V >/dev/null 2>&1 && return
    echo "Building quotatool from source..."
    local pm=$PACKAGE_MANAGER
    if [[ "$pm" == "dnf" ]]; then
        run dnf groupinstall "Development Tools" -y
        run dnf install -y git gcc make autoconf automake
    else
        run yum groupinstall "Development Tools" -y
        run yum install -y git gcc make autoconf automake
    fi
    git clone https://github.com/ekenberg/quotatool.git /tmp/quotatool >/dev/null 2>&1
    ( cd /tmp/quotatool && ./configure && make && make install ) >/dev/null 2>&1
    rm -rf /tmp/quotatool
}

check_kernel_compat() {
    local major="${OS_VERSION_ID%%.*}"
	if [[ ("$OS_ID" == "almalinux" || "$OS_ID" == "rockylinux") && "$major" -ge 10 ]]; then
        cat <<MSG
WARNING: ${OS_ID} ${major}+ detected. Docker may not work on this kernel.
See: https://github.com/stefanpejcic/OpenPanel/issues/745
To fix, run:
  sudo update-alternatives --set iptables /usr/sbin/iptables-legacy
  sudo dnf install -y kernel kernel-core kernel-modules
  sudo grubby --set-default /boot/vmlinuz-\$(rpm -q --qf '%{VERSION}-%{RELEASE}.%{ARCH}\n' kernel | tail -1)
  sudo reboot
Then re-run this installer.
MSG
    fi
}

install_packages() {
    echo "Installing required packages..."
    local packages=()
    case "$PACKAGE_MANAGER" in
        apt-get)
            local kernel_pkg="linux-image-amd64"
            [[ "$OS_ID" == "ubuntu" ]] && kernel_pkg="linux-generic"
            [[ -f /etc/needrestart/needrestart.conf ]] && run sed -i "s/#\$nrconf{restart} = 'i';/\$nrconf{restart} = 'a';/g" /etc/needrestart/needrestart.conf
            run $PACKAGE_MANAGER -qq install -y apt-transport-https ca-certificates
            echo 'APT::Acquire::Retries "3";' > /etc/apt/apt.conf.d/80-retries
            run update-ca-certificates
            packages=(curl cron git gnupg dbus-user-session systemd dbus systemd-container quota quotatool uidmap docker.io "$kernel_pkg" default-mysql-client jc jq sqlite3)
            ;;
        yum)
            check_kernel_compat
            build_quotatool_from_source
            packages=(curl cron git gnupg dbus-user-session systemd dbus systemd-container quota uidmap docker.io linux-generic default-mysql-client jc jq sqlite3)
            ;;
        dnf)
            check_kernel_compat
            build_quotatool_from_source
            run dnf install -y yum-utils epel-release perl python3-pip python3-devel gcc
            run yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo -y
            run dnf config-manager --add-repo=https://download.docker.com/linux/centos/docker-ce.repo
            if [[ -f /etc/fedora-release ]]; then
                packages=(git wget gnupg dbus-user-session systemd dbus systemd-container quota uidmap docker docker-compose mysql docker-compose-plugin sqlite sqlite-devel perl-Math-BigInt)
            else
                packages=(git ncurses wget gnupg systemd dbus systemd-container quota shadow-utils docker-ce docker-ce-cli mariadb containerd.io docker-compose-plugin sqlite sqlite-devel perl-Math-BigInt)
            fi
            ;;
    esac

    for pkg in "${packages[@]}"; do
        pkg_install_with_retry "$pkg"
    done
}

install_python() {
    if [[ "$OS_ID" == "debian" && "$OS_CODENAME" == "trixie" ]]; then
        echo "Building Python 3.12 from source for Debian 13..."
        run $PACKAGE_MANAGER install -y build-essential curl ca-certificates xz-utils libssl-dev zlib1g-dev libbz2-dev libreadline-dev libsqlite3-dev libffi-dev libncursesw5-dev libgdbm-dev liblzma-dev uuid-dev

        local latest
        latest=$(curl -fsSL https://www.python.org/ftp/python/ | grep -Po 'href="3\.12\.\d+/' | grep -Po '3\.12\.\d+' | sort -V | tail -1 || echo "3.12.7")
        local prefix="/opt/python/${latest}"

        run bash -lc "cd /usr/local/src && curl -fsSL -o Python-${latest}.tgz https://www.python.org/ftp/python/${latest}/Python-${latest}.tgz && tar -xzf Python-${latest}.tgz && cd Python-${latest} && ./configure --prefix='${prefix}' --enable-optimizations --with-ensurepip=install && make -j$(nproc) && make altinstall"

        PYTHON_BIN="${prefix}/bin/python3.12"
        [[ -x "$PYTHON_BIN" ]] || die 1 "Python build failed."
        $PYTHON_BIN -m ensurepip -U || true
        $PYTHON_BIN -m pip install --upgrade pip setuptools wheel || true
        return
    fi

    if command -v python3.12 &>/dev/null; then
        PYTHON_BIN="python3.12"
    else
        echo "Installing Python 3.12..."	
        case "$OS_ID" in
            ubuntu)
                run $PACKAGE_MANAGER install -y software-properties-common
                run add-apt-repository -y ppa:deadsnakes/ppa
                run $PACKAGE_MANAGER update -y
                run $PACKAGE_MANAGER install -y python3.12 python3.12-venv || true
                ;;
            debian)
                run install -d -m 0755 /etc/apt/keyrings
                run bash -lc 'curl -fsSL --ipv4 https://pascalroeleven.nl/deb-pascalroeleven.gpg | tee /etc/apt/keyrings/deb-pascalroeleven.gpg >/dev/null' || true
                run bash -lc "echo 'Types: deb
URIs: http://deb.pascalroeleven.nl/python3.12
Suites: ${OS_CODENAME}-backports
Components: main
Signed-By: /etc/apt/keyrings/deb-pascalroeleven.gpg' > /etc/apt/sources.list.d/pascalroeleven.sources" || true
                run $PACKAGE_MANAGER update -y
                run $PACKAGE_MANAGER install -y python3.12 python3.12-venv || true
                ;;
            almalinux|alma|rocky|centos)
                run $PACKAGE_MANAGER update -y
                command -v dnf &>/dev/null && {
                    run dnf install -y epel-release || true
                    run dnf config-manager --set-enabled crb 2>/dev/null || run dnf config-manager --set-enabled powertools || true
                }
                run $PACKAGE_MANAGER install -y python3.12 || true
                ;;
        esac
        command -v python3.12 &>/dev/null && PYTHON_BIN="python3.12" || PYTHON_BIN="python3"
    fi

	[[ "$PACKAGE_MANAGER" == "apt-get" ]] && run "$PACKAGE_MANAGER" install -y python3-venv python3.12-venv || :

    $PYTHON_BIN --version &>/dev/null && ok "Python is available." || die 1 "Python installation failed."
    export PYTHON_BIN
}

download_config() {
    [[ "$REPAIR" == true ]] && rm -rf "$ETC_DIR"
    echo "Cloning OpenPanel configuration to ${ETC_DIR}..."
    run timeout 300s git clone https://github.com/stefanpejcic/openpanel-configuration "$ETC_DIR" || die 1 "Failed to clone openpanel-configuration from GitHub."
    [[ -f "$CONFIG_FILE" ]] && ok "configuration files downloaded." || die 1 "Config file ${CONFIG_FILE} is missing after clone."	
}

setup_opencli() {
    echo "Setting up opencli..."
    [[ "$REPAIR" == true ]] && rm -rf /usr/local/opencli /usr/local/bin/opencli
    run timeout 300s git clone https://github.com/stefanpejcic/opencli.git /usr/local/opencli || die 1 "Failed to clone opencli from GitHub."
    chmod +x -R /usr/local/opencli
    ln -sf /usr/local/opencli/opencli /usr/local/bin/opencli
    export PATH="/usr/bin:$PATH"
    [[ -x /usr/local/bin/opencli ]] && ok "opencli commands are executable." || die 1 "opencli setup failed."
}

install_openadmin() {
    echo "Setting up OpenAdmin..."
    local dir="/usr/local/admin/"
    [[ "$REPAIR" == true ]] && rm -rf "$dir"
    mkdir -p "$dir"

    local branch="110"
    [[ "$architecture" == "aarch64" ]] && branch="armcpu"

    run timeout 300s git clone -b "$branch" --single-branch https://github.com/stefanpejcic/openadmin "$dir" || die 1 "Failed to clone openadmin from GitHub."

    if [[ "$ADMIN_PORT" != 2087 ]]; then
        sed -i "/# START HOSTNAME DOMAIN #/,/# END HOSTNAME DOMAIN #/ s/\(reverse_proxy localhost:\)[0-9]\+/\1$ADMIN_PORT/" "${ETC_DIR}caddy/Caddyfile"
        sed -i "/redir @openadmin/s/:[0-9]\+/:$ADMIN_PORT/g" "${ETC_DIR}caddy/redirects.conf"
        sed -i "/# openadmin/,/# roundcube/ s/:[0-9]\+/:$ADMIN_PORT/g" "${ETC_DIR}nginx/vhosts/openpanel_proxy.conf"
    fi

    cd "$dir" || die 1 "Failed to open $dir"
    ${PYTHON_BIN:-python3} -m venv "${dir}venv"
	# shellcheck disable=SC1090
    source "${dir}venv/bin/activate"
    pip install --default-timeout=300 --force-reinstall --ignore-installed -r requirements.txt >/dev/null 2>&1 || pip install --default-timeout=300 --force-reinstall --ignore-installed -r requirements.txt --break-system-packages >/dev/null 2>&1

    [[ "$OS_ID" == "debian" ]] && run apt install -y python3-yaml || true

    for f in "${ETC_DIR}openadmin/secret.key" "${ETC_DIR}openpanel/secret.key"; do
        [[ -f "$f" ]] || { openssl rand -hex 32 > "$f"; chmod 600 "$f"; }
    done

    cp "${ETC_DIR}openadmin/service/openadmin.service" "${SERVICES_DIR}admin.service"
    run systemctl daemon-reload
    run systemctl enable --now admin
	sleep 5 # allow admin to start
    ss -tuln | grep -q ":$ADMIN_PORT" && ok "OpenAdmin is running on port $ADMIN_PORT." || die 1 "OpenAdmin is not listening on port $ADMIN_PORT."
}

setup_docker_compose() {
    echo "Setting up Docker Compose..."
    local dc_dir="${DOCKER_CONFIG:-/root/.docker}/cli-plugins"
    mkdir -p "$dc_dir"

    local arch_label="x86_64"
    [[ "$architecture" == "aarch64" ]] && arch_label="aarch64"
    local dc_url="https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-linux-${arch_label}"

    curl -4 -SL "$dc_url" -o "$dc_dir/docker-compose" >/dev/null 2>&1
    chmod +x "$dc_dir/docker-compose"
    run curl -4 -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/bin/docker-compose
    chmod +x /usr/bin/docker-compose
    ln -sf /usr/bin/docker-compose /usr/local/bin/docker-compose

    local rc_file="/root/.bashrc"
    [[ -f /root/.zshrc ]] && rc_file="/root/.zshrc"
    if ! grep -q "docker() {" "$rc_file"; then
        cat >> "$rc_file" <<'EOF'
docker() {
    if [[ $1 == "compose" ]]; then
        /usr/local/bin/docker-compose "${@:2}"
    else
        command docker "$@"
    fi
}
EOF
    fi
	# shellcheck disable=SC1090
    source ~/.bashrc || true

    systemctl start docker

    local test_output
    test_output=$(timeout 10 docker run --rm alpine echo "Hello from Alpine!" 2>/dev/null || true)
    [[ "$test_output" == "Hello from Alpine!" ]] && ok "Docker alpine container ran successfully." || die 1 "Docker test failed. Check Docker Hub connectivity and Docker installation."

    local mysql_cnf="/etc/my.cnf"
    local root_pw; root_pw=$(openssl rand -base64 -hex 9)

    cd /root || die 1 "No read access to /root"
    rm -f "$mysql_cnf" .env
    cp "${ETC_DIR}docker/compose/docker-compose.yml" /root/docker-compose.yml
    cp "${ETC_DIR}docker/compose/.env"               /root/.env
    cp "${ETC_DIR}mysql/initialize/1.1/plans.sql"    /root/initialize.sql 2>/dev/null || true
    chmod +x "${ETC_DIR}mysql/scripts/dump.sh" "${ETC_DIR}openlitespeed/start.sh"

    sed -i "s/^VERSION=.*$/VERSION=\"$PANEL_VERSION\"/" /root/.env

    [[ "$USER_PORT" != 2083 ]] && {
        sed -i "s/^PORT=\"[^\"]*\"/PORT=\"$USER_PORT\"/" /root/.env
        sed -i "/redir @openpanel/s/:[0-9]\+/:$USER_PORT/g" "${ETC_DIR}caddy/redirects.conf"
        sed -i "/# openpanel/,/# openadmin/ s/:[0-9]\+/:$USER_PORT/g" "${ETC_DIR}nginx/vhosts/openpanel_proxy.conf"
    }

    sed -i "s/MYSQL_ROOT_PASSWORD=.*/MYSQL_ROOT_PASSWORD=${root_pw}/" /root/.env
    ln -s "${ETC_DIR}mysql/host_my.cnf" "$mysql_cnf"
    sed -i "s/password = .*/password = ${root_pw}/" "${ETC_DIR}mysql/host_my.cnf"
    sed -i "s/password = .*/password = ${root_pw}/" "${ETC_DIR}mysql/container_my.cnf"

    [[ "$OS_ID" == "almalinux" ]] && sed -i 's/mysql\/mysql-server/mysql/g' /root/docker-compose.yml
    if [[ "$OS_ID" == "debian" ]]; then
        run apt install -y apparmor apparmor-utils
        [[ "$OS_VERSION_ID" == "13" ]] && grep -q "skip-ssl" "$mysql_cnf" || echo "skip-ssl = true" >> "$mysql_cnf"
    fi

    [[ "$REPAIR" == true ]] && {
		run docker compose down
		run docker --context default volume rm root_openadmin_mysql
	}

	run docker compose up -d openpanel_mysql

    local cid; cid=$(docker compose ps -a -q openpanel_mysql)
    [[ -n "$cid" ]] || die 1 "MySQL container not found after docker compose up."
    docker --context default ps -q --no-trunc | grep -q "$cid" && ok "MySQL service started." || die 1 "MySQL container is not running."

    ln -sf / /hostfs 2>/dev/null || true
}

setup_bind() {
    [[ "$SKIP_DNS_SERVER" == true ]] && {
		echo "Skipping BIND (--skip-dns-server)."
		return
	}
    echo "Setting up BIND DNS..."
    install -d -m 755 /etc/bind
    cp -r "${ETC_DIR}bind9/"* /etc/bind/

    if [[ "$OS_ID" == "ubuntu" || "$OS_ID" == "debian" ]]; then
		local resolved_conf="/etc/systemd/resolved.conf"
	    if [ -f "$resolved_conf" ]; then
	        grep -q "^DNSStubListener=no" "$resolved_conf" || echo "DNSStubListener=no" >> "$resolved_conf"
	        systemctl restart systemd-resolved
	    fi
	fi

    local rndc_key="/etc/bind/rndc.key"
    if [[ ! -f "$rndc_key" ]]; then
        run timeout 90 docker --context default run --rm -v /etc/bind/:/etc/bind/ --entrypoint=/bin/sh ubuntu/bind9:latest -c 'rndc-confgen -a -A hmac-sha256 -b 256 -c /etc/bind/rndc.key'
        [[ -f "$rndc_key" ]] && ok "rndc.key generated." || warn "Could not generate rndc.key — DNS zone reloads may not work."
    fi

    find /etc/bind/ -type d -print0 | xargs -0 chmod 755
    find /etc/bind/ -type f -print0 | xargs -0 chmod 644
}

setup_firewall() {
    if [[ "$SKIP_FIREWALL" == true ]]; then
        echo "Skipping firewall (--skip-firewall)."
        sed -i 's/,csf//g' "${ETC_DIR}openadmin/config/notifications.ini"
        if command -v jq &>/dev/null; then
            jq 'map(select(.real_name != "csf" and .real_name != "lfd"))' /etc/openpanel/openadmin/config/services.json > /tmp/services.tmp.json && mv /tmp/services.tmp.json /etc/openpanel/openadmin/config/services.json
        fi
        return
    fi

    echo "Installing Sentinel Firewall..."

    wget --timeout=3 --tries=3 --inet4-only https://raw.githubusercontent.com/sentinelfirewall/sentinel/main/csf.tgz >/dev/null 2>&1
    tar -xzf csf.tgz; rm csf.tgz
    ( cd csf && sh install.sh >/dev/null 2>&1 )
    rm -rf csf

    if [[ "$PACKAGE_MANAGER" == "dnf" ]]; then
        run dnf install -y wget curl yum-utils policycoreutils-python-utils libwww-perl
        [[ -f /etc/fedora-release ]] && run yum --allowerasing install perl -y
    else
        run apt-get install -y perl libwww-perl libgd-dev libgd-perl libgd-graph-perl
    fi

    timeout 300s git clone https://github.com/stefanpejcic/csfpost-docker.sh /tmp/csfpost >/dev/null 2>&1
    mv /tmp/csfpost/csfpost.sh /usr/local/csf/bin/csfpost.sh
    chmod +x /usr/local/csf/bin/csfpost.sh
    rm -rf /tmp/csfpost

    sed -i -e 's/TESTING = "1"/TESTING = "0"/' -e 's/RESTRICT_SYSLOG = "0"/RESTRICT_SYSLOG = "3"/' -e 's/ETH_DEVICE_SKIP = ""/ETH_DEVICE_SKIP = "docker0"/' -e 's/DOCKER = "0"/DOCKER = "1"/' /etc/csf/csf.conf
    cp "${ETC_DIR}csf/csf.blocklists" /etc/csf/csf.blocklists

    local email; email=$(grep -E "^e-mail=" "$CONFIG_FILE" | cut -d= -f2 || true)
    [[ -n "$email" ]] && sed -i "s/LF_ALERT_TO = \"\"/LF_ALERT_TO = \"$email\"/" /etc/csf/csf.conf

    open_csf_port() {
        local type=$1 port=$2
        local conf="/etc/csf/csf.conf"
        for dir in "$type" "${type/4/6}"; do
            grep -q "${dir} = .*${port}" "$conf" || sed -i "s/${dir} = \"\(.*\)\"/${dir} = \"\1,${port}\"/" "$conf"
        done
    }

    local ssh_port; ssh_port=$(grep -Po "(?<=Port[ =])\d+" /etc/ssh/sshd_config 2>/dev/null || echo 22)
    for p in 3306 465 "$USER_PORT" "$ADMIN_PORT"; do open_csf_port TCP_OUT "$p"; done
    for p in 22 53 80 443 "$USER_PORT" "$ADMIN_PORT" "32768:60999" 21 "21000:21010" "$ssh_port"; do
        open_csf_port TCP_IN "$p"
    done

    run csf -r
    run systemctl enable csf
    run systemctl restart csf
    run systemctl restart docker

    install -m 755 /dev/null /usr/sbin/sendmail
    command -v csf &>/dev/null && ok "Sentinel Firewall installed." || fail "Sentinel Firewall not installed properly."
}

configure_caddy_extras() {
    sed -i "s/example\.net/$SERVER_IPV4_ADDRESS/g" "${ETC_DIR}caddy/redirects.conf" 2>/dev/null || true
    grep -qE '^127\.0\.0\.1\s+localhost' /etc/hosts || echo "127.0.0.1 localhost" >> /etc/hosts
}

set_hostname() {
    [[ "$SET_HOSTNAME_NOW" != true ]] && return

    if [[ -n "$separate_panel_domain" && "$separate_panel_domain" =~ ^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
        if [[ "$separate_panel_domain" != "$new_hostname" ]]; then
            cat >> "${ETC_DIR}caddy/Caddyfile" <<EOF

# START USERPANEL DOMAIN #
$separate_panel_domain {
    reverse_proxy localhost:2083
}
http://$separate_panel_domain {
    reverse_proxy localhost:2083
}
# END USERPANEL DOMAIN #
EOF
            touch "${ETC_DIR}caddy/domains/${separate_panel_domain}.conf"
        fi
    fi

    if [[ -n "$new_hostname" ]]; then
        if [[ "$new_hostname" =~ ^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
            sed -i "s/example\.net/$new_hostname/g" "${ETC_DIR}caddy/Caddyfile"
        else
            echo "Hostname '$new_hostname' is not a valid FQDN. OpenAdmin will use IP: $SERVER_IPV4_ADDRESS"
        fi
    fi
}

generate_ssl() {
    [[ "$SET_HOSTNAME_NOW" != true ]] && return
    local hostname
    hostname=$(awk '/# START HOSTNAME DOMAIN #/{f=1;next}/# END HOSTNAME DOMAIN #/{f=0}f{print $1;exit}' "${ETC_DIR}caddy/Caddyfile")
    [[ -z "$hostname" || "$hostname" == "example.net" ]] && return

    if [[ "$USE_SELFSIGNED" == true ]]; then
        local ssl_dir="${ETC_DIR}caddy/ssl/custom/$hostname"
        mkdir -p "$ssl_dir"
        openssl genrsa -out "$ssl_dir/$hostname.key" 2048
        openssl req -new -x509 -key "$ssl_dir/$hostname.key" -out "$ssl_dir/$hostname.crt" -days 365 -subj "/CN=$hostname"
        run systemctl restart admin
        echo "Self-signed certificate created. Remove $ssl_dir to allow Let's Encrypt later."
        return
    fi

    cd /root && run docker --context default compose up -d caddy
    for attempt in {1..5}; do
        echo "SSL attempt $attempt for $hostname..."
        if curl -4 -sf -o /dev/null "https://$hostname"; then
            ok "SSL ready. OpenAdmin is now using HTTPS."; run systemctl restart admin; return
        fi
        run docker restart caddy; sleep 5
    done
    warn "SSL generation failed after 5 attempts. OpenAdmin will use HTTP."
}

configure_waf() {
    touch "${ETC_DIR}caddy/coraza_rules.conf"
	opencli waf "$([[ "$CORAZA" == true ]] && printf '%s' enable || printf '%s' disable)" > /dev/null 2>&1
}

setup_redis() { install -d -m 777 /tmp/redis; }

enable_disk_quotas() {
    echo "Enabling disk quotas..."
    local fstab="/etc/fstab"
    if ! grep -E '^\S+\s+/\s+' "$fstab" | grep -q "usrquota"; then
        sed -i -E '/\s+\/\s+/s/(\S+)(\s+\/\s+\S+\s+\S+)(\s+[0-9]+\s+[0-9]+)$/\1\2,usrquota,grpquota\3/' "$fstab"
    fi
    run systemctl daemon-reload
    run quotaoff -a
    run mount -o remount,usrquota,grpquota /
    run quotacheck -cum / -f
    run quotaon -a
    repquota -u / > "/tmp/repquota" 2>/dev/null && ok "Disk quotas enabled." || fail "Quota check failed."
}

set_docker_cpu_limits() {
    mkdir -p /etc/systemd/system/user@.service.d
    echo -e "[Service]\nDelegate=cpu cpuset io memory pids" > /etc/systemd/system/user@.service.d/delegate.conf
    run systemctl daemon-reload
}

configure_premium() {
    [[ "$SET_PREMIUM" != true ]] && return
    LICENSE="Enterprise"
    opencli config update key "$license_key"
    run systemctl restart admin
    timeout 60 opencli email-server install
}

configure_imunifyav() {
    [[ "$IMUNIFY_AV" == true ]] && run opencli imunify install && run opencli imunify start
}

configure_ssh() {
    ln -sf "${ETC_DIR}ssh/admin_welcome.sh" /etc/profile.d/welcome.sh
    chmod +x /etc/profile.d/welcome.sh
    [[ -f /etc/ssh/sshd_config ]] || return
    sed -i "s/[#]LoginGraceTime [[:digit:]]m/LoginGraceTime 1m/" /etc/ssh/sshd_config
	[[ -f /etc/pam.d/sshd ]] && sed -i '/pam_motd\.so/s/^/#/' /etc/pam.d/sshd

    if [[ "$PACKAGE_MANAGER" == "apt-get" ]]; then
        grep -q "^DebianBanner no" /etc/ssh/sshd_config || echo "DebianBanner no" >> /etc/ssh/sshd_config
    fi
    run systemctl restart sshd 2>/dev/null || run systemctl restart ssh
    ok "SSH configured."
}

setup_cron() {
    install -m 600 -o root -g root "${ETC_DIR}cron" /etc/cron.d/openpanel
    [[ "$PACKAGE_MANAGER" =~ ^(dnf|yum)$ ]] && { run restorecon -R /etc/cron.d; run systemctl restart crond; }
    ok "Cron configured."
}

setup_logrotate() { opencli server-logrotate; }

setup_log_dirs() {
    local log_dir="/var/log/openpanel"
    install -d -m 755 "$log_dir" "$log_dir/user" "$log_dir/admin"
}

setup_swap() {
    [[ -n "$(swapon -s)" ]] && { echo "Swap already exists, skipping."; return; }

    create_swap() {
        fallocate -l "${SWAP_FILE}G" /swapfile
        chmod 600 /swapfile; mkswap /swapfile; swapon /swapfile
        echo "/swapfile none swap sw 0 0" >> /etc/fstab
        ok "Created ${SWAP_FILE}G swap file."
    }

    if [[ "$SETUP_SWAP_ANYWAY" == true ]]; then
        [[ "$SWAP_FILE" =~ ^[0-9]+$ && "$SWAP_FILE" -ge 1 && "$SWAP_FILE" -le 10 ]] || { warn "Invalid swap size '$SWAP_FILE'. Using 1 GB default."; SWAP_FILE=1; }
        create_swap
    else
        local ram_gb; ram_gb=$(awk '/MemTotal/{printf "%.1f", $2/1024/1024}' /proc/meminfo)
        (( $(awk "BEGIN {print ($ram_gb < 8)}") )) && create_swap || echo "RAM is ${ram_gb}GB — skipping swap creation."
    fi
}

hetzner_fix() {
    [[ -f /etc/hetzner-build ]] || return
    echo "Hetzner detected — adding Cloudflare DNS resolvers..."
    mv /etc/resolv.conf /etc/resolv.conf.bak
    printf "nameserver 1.1.1.1\nnameserver 1.0.0.1\n" > /etc/resolv.conf
    command -v docker &>/dev/null || run $PACKAGE_MANAGER install -y docker-cli
}

clean_cache() { run $PACKAGE_MANAGER clean all 2>/dev/null || true; }

verify_license() {
    curl -4 -s -X POST -H "Content-Type: application/json" -d "{\"hostname\":\"$(hostname)\",\"public_ip\":\"${SERVER_IPV4_ADDRESS}\"}" https://api.openpanel.com/license/index.php >/dev/null 2>&1 || true
}

start_user_panel() {
    nohup sh -c "cd /root && docker compose up -d openpanel" </dev/null >nohup.out 2>nohup.err &
}

create_admin_account() {
    if [[ "$SET_ADMIN_USERNAME" == true ]]; then
        new_username="$custom_username"
    else
		# shellcheck disable=SC1091,SC2154
        wget --inet4-only --timeout=3 --tries=2 -q -O /tmp/generate.sh https://raw.githubusercontent.com/stefanpejcic/random-username-generator/refs/heads/main/generator.sh 2>/dev/null && source /tmp/generate.sh && new_username="$random_name" || new_username="admin"
    fi

    if [[ "$SET_ADMIN_PASSWORD" == true && "$custom_password" =~ ^[A-Za-z0-9]{5,30}$ ]]; then
        new_password="$custom_password"
    else
        [[ "$SET_ADMIN_PASSWORD" == true ]] && warn "Provided password is invalid — generating a secure one."
        new_password=$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 16)
    fi

    sqlite3 "${ETC_DIR}openadmin/users.db" "CREATE TABLE IF NOT EXISTS user (id INTEGER PRIMARY KEY, username TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL, role TEXT NOT NULL DEFAULT 'user', is_active BOOLEAN DEFAULT 1 NOT NULL);" 2>/dev/null || true
    opencli admin new "$new_username" "$new_password" --super >/dev/null 2>&1 || true

    local count; count=$(sqlite3 "${ETC_DIR}openadmin/users.db" "SELECT COUNT(*) FROM user WHERE username = '$new_username';" 2>/dev/null || echo 0)

	if [[ "$count" -eq 0 ]]; then
        warn "opencli failed — inserting user manually..."
        local hash; hash=$(/usr/local/admin/venv/bin/python3 /usr/local/admin/core/users/hash "$new_password")
        sqlite3 "${ETC_DIR}openadmin/users.db" "INSERT INTO user (username, password_hash, role) VALUES ('$new_username', '$hash', 'admin');" || warn "Manual user creation also failed."
    fi

    display_logins
    send_email_if_configured
}

display_logins() {
    exec > /dev/tty 2>&1
    echo ""
	printf "${GREEN}OpenPanel %s %s installed successfully ${RESET}in %dm %ds\n" "$LICENSE" "$PANEL_VERSION" "$minutes" "$seconds"
    line
    opencli admin
    echo -e "  Username: ${GREEN}${new_username}${RESET}"
    echo -e "  Password: ${GREEN}${new_password}${RESET}"
    line
    exec > >(tee -a "$LOG_FILE") 2>&1
}

send_email_if_configured() {
    [[ "$SEND_EMAIL_AFTER_INSTALL" != true ]] && return
    [[ "$EMAIL" =~ ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]] || { warn "Invalid email '$EMAIL'. Skipping notification."; return; }

    opencli config update email "$EMAIL"
    local token; token=$(tr -dc 'a-zA-Z0-9' < /dev/urandom | head -c 64)
    sed -i "s|^mail_security_token=.*|mail_security_token=$token|" "$CONFIG_FILE"

    local protocol="http" domain="127.0.0.1"
    if [[ "$SET_HOSTNAME_NOW" == true && "$new_hostname" =~ ^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
        local ssl_dir="${ETC_DIR}caddy/ssl"
        if [[ -f "$ssl_dir/acme-v02.api.letsencrypt.org-directory/$new_hostname/$new_hostname.key" || -f "$ssl_dir/custom/$new_hostname/$new_hostname.key" ]]; then
            protocol="https"; domain="$new_hostname"
        fi
    fi

    curl -4 -k -X POST "$protocol://$domain:$ADMIN_PORT/send_email" -F "transient=$token" -F "recipient=$EMAIL" -F "subject=OpenPanel successfully installed" -F "body=OpenAdmin URL: http://$(hostname):$ADMIN_PORT/ | username: $new_username | password: $new_password" --max-time 15 >/dev/null 2>&1 || warn "Failed to send email notification."
}

run_post_install() {
    [[ -z "$post_install_path" ]] && return
    echo "Running post-install script: $post_install_path"
    if [[ "$post_install_path" =~ ^https?:// ]]; then
        local tmp; tmp=$(mktemp)
        wget -q -O "$tmp" "$post_install_path" || { warn "Failed to download post-install script."; return; }
        chmod +x "$tmp"; bash "$tmp"; rm -f "$tmp"
    else
        bash "$post_install_path"
    fi
}

support_message() {
    line
    cat <<MSG
🎉 Thank you for choosing OpenPanel!

  Getting started:  https://openpanel.com/docs/admin/intro/#post-install-steps
  Report issues:    https://github.com/stefanpejcic/OpenPanel/issues/new/choose
  Discord:          https://discord.openpanel.com/
  Community forums: https://community.openpanel.org/
MSG
    line
}

setup_progress_bar() {
    local url="https://raw.githubusercontent.com/pollev/bash_progress_bar/master/progress_bar.sh"
    if command -v curl &>/dev/null; then
        curl -4 --max-time 5 -s "$url" -o progress_bar.sh >/dev/null 2>&1
	elif command -v wget &>/dev/null; then
        wget --timeout=5 --tries=3 --inet4-only "$url" -O progress_bar.sh >/dev/null 2>&1
    else
        echo "Neither wget nor curl found."; exit 1
    fi
    [[ -f progress_bar.sh ]] || { echo "Failed to download progress_bar.sh"; exit 1; }
	# shellcheck disable=SC1091
    source progress_bar.sh
}

STEPS=(
    install_python
    update_package_manager
    install_packages
    hetzner_fix
    download_config
    enable_disk_quotas
    setup_bind
    install_openadmin
    setup_opencli
    setup_redis
    setup_docker_compose
    set_docker_cpu_limits
    configure_waf
    configure_caddy_extras
    set_hostname
    generate_ssl
    setup_firewall
	configure_premium
    setup_cron
    setup_logrotate
    configure_ssh
    setup_log_dirs
    configure_imunifyav
    setup_swap
    clean_cache
    verify_license
    start_user_panel
)

run_installation() {
    enable_trapping
    setup_scroll_area
    local total=${#STEPS[@]} current=0
    for step in "${STEPS[@]}"; do
        $step
        (( current++ ))
        draw_progress_bar $(( current * 100 / total ))
    done
    destroy_scroll_area
}

# Main
(
flock -n 200 || { echo "Another install is already running."; exit 1; }
detect_os_and_package_manager
parse_args "$@"
[[ -r /root && -w /root ]] || { echo "No read/write access to /root."; exit 1; }
get_server_ipv4
set_panel_version
detect_os_and_package_manager
print_header
check_requirements
detect_installed_panels
echo -e "Starting OpenPanel installation process..."
start=$(date +%s)
setup_progress_bar
run_installation
duration=$(($(date +%s) - start))
minutes=$((duration / 60))
seconds=$((duration % 60))
support_message
create_admin_account
run_post_install
) 200>/root/openpanel_install.lock

rm -f /root/openpanel_install.lock
