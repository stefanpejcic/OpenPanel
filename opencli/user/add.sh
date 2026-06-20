#!/bin/bash
################################################################################
# Script Name: user/add.sh
# Description: Create a new user with the provided plan_name.
# Usage: opencli user-add <USERNAME> <PASSWORD|generate> <EMAIL> "<PLAN_NAME>" [--send-email] [--debug]  [--webserver="<nginx|apache|openresty|openlitespeed|litespeed|varnish+nginx|varnish+apache|varnish+openresty|varnish+openlitespeed>"] [--sql=<mysql|mariadb>] [--RESELLER=<RESELLER_USERNAME>][--server=<IP_ADDRESS>]  [--key=<SSH_KEY_PATH>] [--no-sentinel]
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 01.10.2023
# Last Modified: 19.06.2026
# Company: openpanel.com
# Copyright (c) openpanel.com
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

readonly FORBIDDEN_USERNAMES_FILE="/etc/openpanel/openadmin/config/forbidden_usernames.txt"
readonly DB_CONFIG_FILE="/usr/local/opencli/db.sh"
readonly PANEL_CONFIG_FILE="/etc/openpanel/openpanel/conf/openpanel.config"
readonly LOCK_FILE="/var/lock/openpanel_user_add.lock"
readonly DOCKER_COMPOSE_VERSION="2.36.0"
readonly DOCKER_COMPOSE_ARM="https://github.com/docker/compose/releases/download/v${DOCKER_COMPOSE_VERSION}/docker-compose-linux-aarch64"
readonly DOCKER_COMPOSE_X86="https://github.com/docker/compose/releases/download/v${DOCKER_COMPOSE_VERSION}/docker-compose-linux-x86_64"
readonly ROOTLESS_SETUP_SCRIPT="/etc/openpanel/docker/dockerd-rootless-setuptool.sh"

if [[ "$#" -lt 4 || "$#" -gt 11 ]]; then
    echo "Usage: opencli user-add <username> <password|generate> <email> '<plan_name>' [--send-email] [--debug] [--reseller=<RESELLER_USER>] [--server=<IP_ADDRESS>] [--key=<KEY_PATH>]"
    echo
    echo "Required arguments:"
    echo "  <username>                 The username of the new user."
    echo "  <password|generate>        The password for the new user, or 'generate' to auto-generate a password."
    echo "  <email>                    The email address associated with the new user."
    echo "  <plan_name>                The plan to assign to the new user."
    echo
    echo "Optional flags:"
    printf "%-25s %-45s\n" "  --send-email" "Send a welcome email to the user."
    printf "%-25s %-45s\n" "  --debug" "Enable debug mode for additional output."
    echo
    exit 1
fi

# required
readonly USERNAME="${1,,}"
PASSWORD="$2"
readonly EMAIL="$3"
readonly PLAN_NAME="$4"
shift 4

# optional
DEBUG=false
SKIP_IMAGE_PULL=false
SEND_EMAIL=false
SENTINEL=true
RESELLER=""
NODE_IP=""
SSH_KEY=""
WEBSERVER=""
SQL_TYPE=""
HOSTNAME_LABEL=""

cleanup() {
    rm -f "$LOCK_FILE"
}

hard_cleanup() {
	[[ -n "$USERNAME" ]] || { echo "ERROR: USERNAME is empty"; return 1; }

	# kill processes
    killall -u "${USERNAME}" -9 >/dev/null 2>&1 || true

	# delete user on master
    if command -v deluser >/dev/null 2>&1; then
        deluser --remove-home "$USERNAME" # Debian
    elif command -v userdel >/dev/null 2>&1; then
        userdel -r "$USERNAME"            # RHEL
    else
        echo "ERROR: Neither deluser nor userdel found"
    fi

	# delete user on slave
     [[ -n "$NODE_IP" ]] && node_ssh "bash -s" << EOF
killall -u "${USERNAME}" -9 >/dev/null 2>&1 || true
if command -v deluser >/dev/null 2>&1; then
    deluser --remove-home "${USERNAME}"
elif command -v userdel >/dev/null 2>&1; then
    userdel -r "${USERNAME}"
fi
EOF

	# delete user files
    rm -rf /etc/openpanel/openpanel/core/users/"$USERNAME" > /dev/null 2>&1
	# delete docker context
    docker context rm "$USERNAME"  > /dev/null 2>&1
}

trap cleanup EXIT

read_config_file() {
    while IFS='=' read -r key value; do
        [[ "$key" =~ ^[[:space:]]*# ]] && continue   # skip comments
        [[ -z "$key" ]] && continue                  # skip blank lines
        case "$key" in
            key) ENTERPRISE="$value" ;;
            weakpass) weakpass="$value" ;;
            email_plaintext_passwords) email_plaintext_passwords="$value" ;;
        esac
    done < "$PANEL_CONFIG_FILE"
}

die()  { echo "[✘] ERROR: $*" >&2; exit 1; }

node_ssh() {
    # shellcheck disable=SC2086
    ssh -q -o LogLevel=ERROR -i ${SSH_KEY} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o BatchMode=yes "root@${NODE_IP}" "$@"
}

is_valid_ipv4() {
    local ip="$1"
    local re='^([0-9]{1,3}\.){3}[0-9]{1,3}$'
    [[ "$ip" =~ $re ]] || return 1
    IFS='.' read -r -a octets <<< "$ip"
    for o in "${octets[@]}"; do
        (( o >= 0 && o <= 255 )) || return 1
    done
}

load_default_node() {
    local default_node default_key

	: '
	[CLUSTERING]
	default_node="11.22.33.44"
	default_ssh_key_path="/root/some-key.rsa"
	'

    while IFS='=' read -r key value; do
        case "$key" in
            default_node) default_node="$value" ;;
            default_ssh_key_path) default_key="$value" ;;
        esac
    done < "/etc/openpanel/openadmin/config/admin.ini"

    [[ -z "$default_node" ]] && return 0

    if [[ -n "$default_key" ]]; then
        NODE_IP="$default_node"
        SSH_KEY="$default_key"
    fi
}


read_config_file
load_default_node         # we run it before parse args so it can be overwritten!

# Parse args
for arg in "$@"; do
    case "$arg" in
        --debug)          DEBUG=true ;;
        --send-email)     SEND_EMAIL=true ;;
        --skip-images)    SKIP_IMAGE_PULL=true ;;
        --no-sentinel)    SENTINEL=false ;;
        --reseller=*)     RESELLER="${arg#*=}" ;;
        --server=*)       NODE_IP="${arg#*=}" ;;
        --key=*)          SSH_KEY="${arg#*=}" ;;
        --sql=*)          SQL_TYPE="${arg#*=}" ;;
        --webserver=*)    WEBSERVER="${arg#*=}" ;;
        *)                echo "[!] Warning: unknown flag '$arg'" >&2 ;;
    esac
done

log() { [[ "$DEBUG" == true ]] && echo "$*"; }


# shellcheck disable=SC1091
. "$DB_CONFIG_FILE"

escape() {
    printf "%s" "$1" | sed "s/'/''/g"
}

db_query() {
    # shellcheck disable=SC2154
    mysql --defaults-extra-file="$config_file" -D "$mysql_database" -sN -e "$1"
}

update_reseller_account_count() {
    [[ -n "$RESELLER" ]] || return 0
    local limits_file="/etc/openpanel/openadmin/resellers/${RESELLER}.json"
    [[ -f "$limits_file" ]] || return 0

    local count
    count="$(db_query "SELECT COUNT(*) FROM users WHERE owner='${RESELLER//\'/\\\'}'")"
    jq --argjson n "${count:-0}" '.current_accounts = $n' "$limits_file" > "/tmp/reseller_${RESELLER}_update_$$.json" && mv "/tmp/reseller_${RESELLER}_update_$$.json" "$limits_file"
    log "Reseller $RESELLER account count updated to $count"
}

check_reseller_limits() {
    [[ -n "$RESELLER" ]] || return 0

    local db_file="/etc/openpanel/openadmin/users.db"

    RESELLER_ESC=$(escape "$RESELLER")
    local user_exists
    user_exists="$(sqlite3 "$db_file" "SELECT COUNT(*) FROM user WHERE username='${RESELLER//\'/\'\'}' AND role='reseller';")"
    (( user_exists >= 1 )) || die "User '$RESELLER' is not a reseller or is not allowed to create users."

    local limits_file="/etc/openpanel/openadmin/resellers/${RESELLER}.json"
    if [[ ! -f "$limits_file" ]]; then
        echo "[!] Warning: Reseller $RESELLER has no limits file; unlimited accounts allowed."
        return 0
    fi

    local current_accounts
    current_accounts="$(db_query "SELECT COUNT(*) FROM users WHERE owner='${RESELLER_ESC//\'/\\\'}'")"
    current_accounts="${current_accounts:-0}"

    jq --argjson n "$current_accounts" '.current_accounts = $n' "$limits_file" > "/tmp/reseller_${RESELLER}_$$.json" && mv "/tmp/reseller_${RESELLER}_$$.json" "$limits_file"

    local max_accounts
    max_accounts="$(jq -r '.max_accounts // "unlimited"' "$limits_file")"
    if [[ "$max_accounts" != "unlimited" && "$max_accounts" != "0" && "$current_accounts" -ge "$max_accounts" ]]; then
        die "Reseller '$RESELLER' has reached the maximum account limit ($max_accounts)."
    fi

    local allowed_plans
    allowed_plans="$(jq -r '.allowed_plans | join(",")' "$limits_file")"
    grep -wq "$PLAN_ID" <<< "$allowed_plans" || die "Plan ID '$PLAN_ID' is not assigned to reseller '$RESELLER'."
}

resolve_node() {
    if [[ -z "$NODE_IP" ]]; then HOSTNAME_LABEL="$(hostname)"; return 0; fi

    is_valid_ipv4 "$NODE_IP" || die "'$NODE_IP' is not a valid IPv4 address."
    [[ -n "$SSH_KEY" ]] || die "--key is required when --server is specified."
    [[ -f "$SSH_KEY" ]] || die "SSH key path '$SSH_KEY' does not exist."
    [[ "$(stat -c %a "$SSH_KEY")" == "600" ]] || { chmod 600 "$SSH_KEY"; log "Fixed SSH key permissions."; }

    HOSTNAME_LABEL="$(node_ssh hostname 2>/dev/null)"
    [[ -n "$HOSTNAME_LABEL" ]] || die "Cannot reach node $NODE_IP. Verify SSH access with: ssh -i ${SSH_KEY} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o BatchMode=yes root@$NODE_IP"

    log "Containers will be created on node: $NODE_IP ($HOSTNAME_LABEL)"
}

validate_password_in_lists() {
    # https://weakpass.com/wordlist
    # https://github.com/steveklabnik/password-cracker/blob/master/dictionary.txt

    [[ "$weakpass" == "no" ]] && return 0

    log "Checking password against weak-password dictionary"

    local dict="/tmp/weakpass_dictionary.txt"
    local url="https://github.com/steveklabnik/password-cracker/raw/master/dictionary.txt"

    if [[ ! -f "$dict" ]]; then
        log "Downloading weak-password dictionary (first time only)"
        wget -qO "$dict" "$url" || { echo "[!] Warning: Could not fetch dictionary; skipping check."; return 0; }
    elif [[ $(find "$dict" -mtime +7 -print -quit 2>/dev/null) ]]; then
        log "Refreshing weak-password dictionary (older than 7 days)"
        wget -qO "$dict" "$url" || echo "[!] Warning: Update failed; using cached dictionary."
    fi

    local lower="${PASSWORD,,}"

    if grep -qiF "^${lower}$" "$dict" 2>/dev/null; then
        die "Password is a common dictionary word. Use a stronger password or disable the check with: opencli config update weakpass no"
    fi
}

validate_username() {
    log "Validating username '$USERNAME'"

    (( ${#USERNAME} >= 3 && ${#USERNAME} <= 20 )) || die "Username must be 3–20 characters. See https://openpanel.com/docs/articles/accounts/forbidden-usernames/#openpanel"

    [[ "$USERNAME" =~ ^[a-zA-Z][a-zA-Z0-9]*$ ]] || die "Username must start with a letter and contain only letters and numbers."

    if [[ -f "$FORBIDDEN_USERNAMES_FILE" ]]; then
        while IFS= read -r forbidden; do
            [[ "${USERNAME,,}" == "${forbidden,,}" ]] && die "Username '$USERNAME' is not allowed. See https://openpanel.com/docs/articles/accounts/forbidden-usernames/#reserved-usernames"
        done < "$FORBIDDEN_USERNAMES_FILE"
    fi
}


validate_user_creation() {
    if [[ -z "$ENTERPRISE" ]]; then
        [[ -z "$RESELLER" ]] || die "Resellers require the Enterprise edition."
        log "Community edition: checking 3-user limit"
    else
        [[ -z "$RESELLER" ]] && log "Enterprise edition: unlimited users"
    fi

    local USERNAME_ESC result
	USERNAME_ESC=$(escape "$USERNAME")
    result="$(db_query "SELECT COUNT(*), SUM(username='${USERNAME_ESC}' OR username LIKE 'SUSPENDED\_%\_${USERNAME_ESC}') FROM users")"

    local user_count username_taken
    read -r user_count username_taken <<< "$result"

    [[ -z "$ENTERPRISE" && "$user_count" -gt 2 ]] && die "Community edition is limited to 3 accounts. Upgrade to Enterprise for unlimited users."

    [[ "${username_taken:-0}" -gt 0 ]] && die "Username '$USERNAME' is already taken."
}

bootstrap_node() {
    [[ -n "$NODE_IP" ]] || return 0

    log "Bootstrapping node $NODE_IP"

    node_ssh 'bash -s' << 'REMOTE'
set -e
if [[ ! -d /etc/openpanel/openpanel ]]; then
    grep -qxF 'PubkeyAuthentication yes' /etc/ssh/sshd_config || echo 'PubkeyAuthentication yes' >> /etc/ssh/sshd_config
    grep -qxF 'AuthorizedKeysFile .ssh/authorized_keys' /etc/ssh/sshd_config || echo 'AuthorizedKeysFile .ssh/authorized_keys' >> /etc/ssh/sshd_config
    service ssh restart >/dev/null 2>&1 || true

    if command -v apt-get &>/dev/null; then
        DEBIAN_FRONTEND=noninteractive apt-get -yq update >/dev/null 2>&1
        DEBIAN_FRONTEND=noninteractive apt-get -yq install systemd-container uidmap >/dev/null 2>&1
    elif command -v dnf &>/dev/null; then
        dnf install -y systemd-container uidmap >/dev/null 2>&1
    elif command -v yum &>/dev/null; then
        yum install -y systemd-container uidmap >/dev/null 2>&1
    else
        echo "ERROR: Cannot install prerequisites." >&2; exit 1
    fi

    mkdir -p /etc/systemd/system/user@.service.d
    cat > /etc/systemd/system/user@.service.d/delegate.conf << 'EOF'
[Service]
Delegate=cpu cpuset io memory pids
EOF
    systemctl daemon-reload
fi
REMOTE
	# TODO: dont scp if already exists!
	scp -i "${SSH_KEY}" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o BatchMode=yes -r /etc/openpanel "root@${NODE_IP}:/etc/openpanel"
}

setup_sshfs_mount() {
    [[ -n "$NODE_IP" ]] || return 0

    if ! command -v sshfs &>/dev/null; then
        local pm
        if command -v apt-get &>/dev/null; then pm="apt-get install -y";
        elif command -v dnf &>/dev/null;     then pm="dnf install -y";
        elif command -v yum &>/dev/null;     then pm="yum install -y";
        else die "Cannot install sshfs; no supported package manager found."; fi
        # shellcheck disable=SC2086
        $pm sshfs
    fi

    sshfs -o "IdentityFile=${SSH_KEY},StrictHostKeyChecking=no" "root@${NODE_IP}:/home/${USERNAME}" "/home/${USERNAME}" || die "sshfs mount failed for $NODE_IP:/home/$USERNAME"
}

load_plan() {
    log "Looking up plan '$PLAN_NAME'"
	local PLAN_NAME_ESC
    PLAN_NAME_ESC=$(escape "$PLAN_NAME")
    PLAN_ID="$(db_query "SELECT id FROM plans WHERE name = '${PLAN_NAME_ESC//\'/\\\'}'")"
    [[ -n "$PLAN_ID" ]] || die "Plan '$PLAN_NAME' not found."
}

autostart_services() {
	local AUTOSTART_FILE="/etc/openpanel/docker/compose/1.0/autostart.services"

    [[ "$SKIP_IMAGE_PULL" == true ]] && { log "Skipping image pull (--skip-images)."; return 0; }

    local env_file="/home/${USERNAME}/.env"
    [[ -f "$env_file" ]] || { echo "[!] Warning: $env_file not found; skipping image pull."; return 1; }

	# https://community.openpanel.org/d/239-no-such-container-openlitespeed
    get_env() { grep -E "^$1=" "$env_file" | cut -d= -f2- | tr -d '\r"'\'; }
    local sql_type ws_type php_version php_svc
    sql_type="$(get_env MYSQL_TYPE)"
    ws_type="$(get_env WEB_SERVER)"
	php_version="$(get_env DEFAULT_PHP_VERSION)"
    varnish="$(get_env PROXY_HTTP_PORT)"
    php_svc="php-fpm-${php_version}"

    [[ -f "$AUTOSTART_FILE" ]] || { echo "[!] Warning: $AUTOSTART_FILE not found; skipping image pull."; return 1; }

    local autostart
    mapfile -t autostart < <(grep -v '^\s*#' "$AUTOSTART_FILE" | grep -v '^\s*$')

    local images=()
    local ols_ws=false
    [[ "$ws_type" =~ ^(openlitespeed|litespeed)$ ]] && ols_ws=true
    [[ "$ols_ws" == false ]] && [[ "$php_version" =~ ^[0-9]+\.[0-9]+$ ]] && [[ " ${autostart[*]} " == *" $php_svc "* ]] && images+=("$php_svc")
    [[ "$sql_type" =~ ^(mysql|mariadb)$ ]] && [[ " ${autostart[*]} " == *" $sql_type "* ]] && images+=("$sql_type")
	[[ -n "$varnish" ]] && [[ " ${autostart[*]} " == *" varnish "* ]] && images+=("varnish")
    [[ "$ws_type" =~ ^(nginx|apache|openresty|openlitespeed|litespeed)$ ]] && [[ " ${autostart[*]} " == *" $ws_type "* ]] && images+=("$ws_type")
    local sql_ws_types="mysql mariadb nginx apache openresty openlitespeed litespeed"
    for svc in "${autostart[@]}"; do
        [[ " $sql_ws_types " == *" $svc "* ]] && continue                 # only webserver and sql type for user, ignore others
        [[ "$ols_ws" == true ]] && [[ "$svc" == php-fpm-* ]] && continue  # OLS/LS manages PHP internally, no need to start php-fpm services
        [[ "$svc" == php-fpm-* || "$svc" == varnish ]] && continue        # only current default php version and varnish if enabled
        images+=("$svc")
    done
    [[ ${#images[@]} -eq 0 ]] && { echo "[!] Warning: No autostart services match user config."; return 1; }
	log "Starting services in background: ${images[*]}"
	nohup bash -c "e=0; ok=0; while [[ \$e -lt 60 ]]; do if docker --context=${USERNAME} info >/dev/null 2>&1; then ((ok++)); [[ \$ok -ge 3 ]] && break; else ok=0; fi; sleep 3; ((e+=3)); done; cd /home/${USERNAME}/; for svc in ${images[*]}; do docker --context=${USERNAME} compose up -d \$svc || true; done" </dev/null >"/tmp/autostart_${USERNAME}.log" 2>&1 &
	disown
}

setup_ssh_key_for_user() {
    [[ -n "$NODE_IP" ]] || return 0
    log "Configuring SSH key for user '$USERNAME'"

    local pub_key
    pub_key="$(ssh-keygen -y -f "$SSH_KEY")" || die "Cannot read public key from $SSH_KEY"

    node_ssh "bash -s" << EOF
mkdir -p /home/${USERNAME}/.ssh
touch /home/${USERNAME}/.ssh/authorized_keys
chown "${USERNAME}" -R /home/${USERNAME}/.ssh
grep -qF '${pub_key}' /home/${USERNAME}/.ssh/authorized_keys || echo '${pub_key}' >> /home/${USERNAME}/.ssh/authorized_keys
EOF

    mkdir -p ~/.ssh/cm_socket
    chmod 700 ~/.ssh
    local key_copy="${HOME}/.ssh/${NODE_IP}"
    cp "$SSH_KEY" "$key_copy" && chmod 600 "$key_copy"

    if ! grep -qF "Host ${USERNAME}" ~/.ssh/config 2>/dev/null; then

		mkdir -p /tmp/ssh_cm
		chmod 700 /tmp/ssh_cm
		# TODO: recreate on reboot!
        cat >> ~/.ssh/config << EOF

Host ${USERNAME}
    HostName ${NODE_IP}
    User ${USERNAME}
    IdentityFile ~/.ssh/${NODE_IP}
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
    ControlPath /tmp/ssh_cm/%r@%h:%p
    ControlMaster auto
    ControlPersist 5m
    TCPKeepAlive yes
    ServerAliveInterval 15
    ServerAliveCountMax 3
EOF
    fi

    ssh "${USERNAME}" exit &>/dev/null || die "Failed to establish SSH connection for user '$USERNAME'."
    log "SSH connection established for $USERNAME"
}

create_linux_user_local() {
    log "Creating local Linux user '$USERNAME'"
    useradd -m -d "/home/${USERNAME}" "$USERNAME" || die "Failed to create Linux user '$USERNAME' on master."
    USER_ID="$(id -u "$USERNAME")"
}

create_linux_user_remote() {
    [[ -n "$NODE_IP" ]] || return 0
    log "Creating Linux user '$USERNAME' on node $NODE_IP"
    local id_flag=""
    [[ -n "${USER_ID:-}" ]] && id_flag="-u $USER_ID"

    # shellcheck disable=SC2086
    node_ssh "useradd -m -s /bin/bash -d /home/${USERNAME} ${id_flag} ${USERNAME}" || { hard_cleanup; die "Failed to create Linux user '$USERNAME' on node $NODE_IP"; }
    USER_ID="$(node_ssh "id -u ${USERNAME}")"
}

ensure_docker_on_node() {
    [[ -n "$NODE_IP" ]] || return 0

    if ! node_ssh "command -v docker >/dev/null 2>&1"; then
        log "Installing Docker on $NODE_IP"
        node_ssh "bash -s" << 'REMOTE'
set -e
apt-get update -qq
apt-get install -y docker.io
systemctl enable --now docker
REMOTE
    fi

    log "Adding '$USERNAME' to docker group on $NODE_IP"
    node_ssh "bash -s" << EOF
if ! id -nG "${USERNAME}" 2>/dev/null | grep -qw docker; then
    usermod -aG docker "${USERNAME}"
fi
EOF
}

setup_docker_compose() {
    log "Setting up Docker Compose for '$USERNAME'"
    local arch link system_file
    arch="$(uname -m)"
    case "$arch" in
        aarch64) link="$DOCKER_COMPOSE_ARM" ;;
        *)       link="$DOCKER_COMPOSE_X86" ;;
    esac
    system_file="/etc/openpanel/docker/docker-compose-linux-${arch}"
    [[ -f "$system_file" ]] || curl -fsSL "$link" -o "$system_file"
    chmod +x "$system_file"
    mkdir -p "/home/${USERNAME}/.docker/cli-plugins"
    ln -sf "$system_file" "/home/${USERNAME}/.docker/cli-plugins/docker-compose"
}

setup_docker_rootless() {
    log "Configuring rootless Docker for '$USERNAME'"

    local home_dir="/home/${USERNAME}"
    local apparmor_file="/etc/apparmor.d/home.${USERNAME}.bin.rootlesskit"

    _build_apparmor_profile() {
        sed "s/USERNAME/${USERNAME}/g" << 'EOF'
abi <abi/4.0>,
include <tunables/global>

"/home/USERNAME/bin/rootlesskit" flags=(unconfined) {
  userns,
  include if exists <local/home.USERNAME.bin.rootlesskit>
}
EOF
    }

    mkdir -p "${home_dir}/docker-data" "${home_dir}/.config/docker" "${home_dir}/.docker/run" "${home_dir}/bin" "/etc/apparmor.d/"
    cp /etc/openpanel/docker/daemon/rootless.json "${home_dir}/.config/docker/daemon.json"
    sed -i "s/USERNAME/${USERNAME}/g" "${home_dir}/.config/docker/daemon.json"
    chmod 755 -R "${home_dir}"
    [[ -f "${home_dir}/.bashrc" ]] && sed -i "1i export PATH=/home/${USERNAME}/bin:\$PATH" "${home_dir}/.bashrc"

    if [[ -n "$NODE_IP" ]]; then
        node_ssh "bash -s" << REMOTE
set -e
$(_build_apparmor_profile | cat > "/etc/apparmor.d/home.${USERNAME}.bin.rootlesskit")
systemctl restart apparmor.service >/dev/null 2>&1 || true
loginctl enable-linger "${USERNAME}" >/dev/null 2>&1
mkdir -p /home/${USERNAME}/.docker/run
chmod 700 /home/${USERNAME}/.docker/run
chmod 755 -R /home/${USERNAME}/
chown -R ${USERNAME}:${USERNAME} /home/${USERNAME}/
REMOTE

        node_ssh "bash -s" << REMOTE
su - "${USERNAME}" -c 'bash -l -c "
    mkdir -p ~/bin
    curl -fsSL https://get.docker.com/rootless -o ~/bin/dockerd-rootless-setuptool.sh
    chmod +x ~/bin/dockerd-rootless-setuptool.sh
    ~/bin/dockerd-rootless-setuptool.sh install >/dev/null 2>&1

    grep -qF XDG_RUNTIME_DIR ~/.bashrc || echo \"export XDG_RUNTIME_DIR=/home/${USERNAME}/.docker/run\" >> ~/.bashrc
    grep -qF DBUS_SESSION_BUS_ADDRESS ~/.bashrc || echo \"export DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/\$(id -u)/bus\" >> ~/.bashrc

    mkdir -p ~/.config/systemd/user/
    cat > ~/.config/systemd/user/docker.service << EOF
[Unit]
Description=Docker Application Container Engine (Rootless)
After=network.target
[Service]
Environment=PATH=/home/${USERNAME}/bin:\$PATH
Environment=DOCKER_HOST=unix:///home/${USERNAME}/.docker/run/docker.sock
ExecStart=/home/${USERNAME}/bin/dockerd-rootless.sh -H unix:///home/${USERNAME}/.docker/run/docker.sock
Restart=on-failure
StartLimitBurst=3
StartLimitInterval=60s
[Install]
WantedBy=default.target
EOF
"'
REMOTE

        node_ssh "machinectl shell ${USERNAME}@ /bin/bash -c '
            systemctl --user daemon-reload >/dev/null 2>&1
            systemctl --user enable docker >/dev/null 2>&1
            systemctl --user start docker >/dev/null 2>&1
        ' 2>/dev/null" || true

    else
        _build_apparmor_profile > "$apparmor_file"
        systemctl restart apparmor.service >/dev/null 2>&1 || true
        loginctl enable-linger "$USERNAME" >/dev/null 2>&1

      	if [ ! -f "$ROOTLESS_SETUP_SCRIPT" ]; then
			curl -sSL https://get.docker.com/rootless -o "$ROOTLESS_SETUP_SCRIPT"
   			chmod +x "$ROOTLESS_SETUP_SCRIPT"
		fi

        mkdir -p "${home_dir}/.docker/run" "${home_dir}/bin"
        chmod 700 "${home_dir}/.docker/run"
        chown -R "${USERNAME}:${USERNAME}" "${home_dir}"
        ln -sf "$ROOTLESS_SETUP_SCRIPT" "${home_dir}/bin/dockerd-rootless-setuptool.sh"

        machinectl shell "${USERNAME}@" /bin/bash -c "
            source ~/.bashrc 2>/dev/null || true
            /home/${USERNAME}/bin/dockerd-rootless-setuptool.sh install >/dev/null 2>&1

            grep -qF XDG_RUNTIME_DIR ~/.bashrc || echo 'export XDG_RUNTIME_DIR=/home/${USERNAME}/.docker/run' >> ~/.bashrc
            grep -qF 'export PATH=/home/${USERNAME}/bin' ~/.bashrc || echo 'export PATH=/home/${USERNAME}/bin:\$PATH' >> ~/.bashrc
            grep -qF DOCKER_HOST ~/.bashrc || echo 'export DOCKER_HOST=unix:///home/${USERNAME}/.docker/run/docker.sock' >> ~/.bashrc

            source ~/.bashrc 2>/dev/null || true
            mkdir -p ~/.config/systemd/user/
            cat > ~/.config/systemd/user/docker.service << 'SVCEOF'
[Unit]
Description=Docker Application Container Engine (Rootless)
After=network.target
[Service]
Environment=PATH=/home/${USERNAME}/bin:$PATH
Environment=DOCKER_HOST=unix://%t/docker.sock
ExecStart=/home/${USERNAME}/bin/dockerd-rootless.sh
Restart=on-failure
StartLimitBurst=3
StartLimitInterval=60s
[Install]
WantedBy=default.target
SVCEOF
            systemctl --user daemon-reload >/dev/null 2>&1
            systemctl --user restart docker >/dev/null 2>&1
        " 2>/dev/null || true
    fi
}

get_docker_service_errors() {
    docker_service_errors=$(timeout 5 machinectl shell "${USERNAME}@" /bin/bash -c 'systemctl --user status docker --no-pager 2>&1' 2>&1) #systemctl --user status docker --no-pager 2>&1 | grep -i "error\|failed\|fatal"
}

test_docker_service() {
    # dockerd-rootless-setuptool.sh executed?
	if [ ! -f "/home/${USERNAME}/bin/dockerd-rootless-setuptool.sh" ]; then
	    hard_cleanup
	    die "Installer script '${ROOTLESS_SETUP_SCRIPT}' exists but installation appears incomplete."
	fi

    # dockerd-rootless-setuptool.sh finished and created dockerd-rootless.sh?
	if [ ! -e "/home/${USERNAME}/bin/dockerd-rootless.sh" ]; then
	    hard_cleanup
	    die "Installer script '${ROOTLESS_SETUP_SCRIPT}' failed."
	fi

    # wait for the docker socket to be available (docker started and initialized)
	local elapsed=0 max_time=30
    while [[ $elapsed -lt $max_time ]]; do
        if [[ -S "/hostfs/run/user/${USER_ID}/docker.sock" ]]; then
            log "Docker service started (socket: /run/user/${USER_ID}/docker.sock)"
            break
        fi
        sleep 1
        (( elapsed++ )) || true
    done

	# is docker socket available?
	if [ ! -S "/hostfs/run/user/${USER_ID}/docker.sock" ]; then
	    get_docker_service_errors
	    hard_cleanup
	    die "Docker service did not start after $max_time seconds!"
	fi

    # context created and compose plugin loaded?
	docker_compose_output=$(timeout 3 docker --context="$USERNAME" compose version 2>&1)
	if echo "$docker_compose_output" | grep -q "Docker Compose version"; then
	    log "Docker context is working and compose plugin is loaded."
	else
		hard_cleanup
		die "Docker Compose is not working in context '$USERNAME'."
	fi

	# is docker service ready?
	docker_info_output=$(timeout 5 docker --context="$USERNAME" info 2>&1)
	if echo "$docker_info_output" | grep -q "Server Version"; then
	    log "Docker service is responding and working correctly."
	else
	    log "docker info output: $docker_info_output"
		get_docker_service_errors
	    hard_cleanup
	    die "Docker service started (socket created) but is not running: $docker_service_errors"
	fi
}

find_available_ports_bg() { find_available_ports > /tmp/ports_$$; }

find_available_ports() {
    local last_user min_port
    last_user="$(db_query "SELECT server FROM users ORDER BY id DESC LIMIT 1" 2>/dev/null || true)"
    min_port=32768

    if [[ -n "$last_user" ]]; then
        local env_file="/home/${last_user}/.env"
        if [[ -f "$env_file" ]]; then
            local highest_port
            highest_port="$(grep -E '^[A-Z_]+_PORT=' "$env_file" | grep -oE '[0-9]+' | sort -rn | head -1)"
            [[ -n "$highest_port" ]] && min_port="$highest_port"
        fi
    fi

    local ports=()
    for (( i = 1; i <= 7; i++ )); do
        ports+=("$(( min_port + i ))")
    done
    echo "${ports[@]}"
}

configure_environment() {
    log "Configuring environment for '$USERNAME'"

    local port_1 port_2 port_3 port_4 port_5 port_6 port_7
    if [[ -n "$NODE_IP" ]]; then
        port_1="${NODE_IP}:${P1}:80"
        port_2="${NODE_IP}:${P2}:3306"
        port_3="${NODE_IP}:${P3}:5432"
        port_4="${NODE_IP}:${P4}:80"
        port_5="${NODE_IP}:${P5}:80"
        port_6="${NODE_IP}:${P6}:443"
        port_7="${NODE_IP}:${P7}:80"
    else
        port_1="${P1}:80"
        port_2="${P2}:3306"
        port_3="${P3}:5432"
        port_4="${P4}:80"
        port_5="127.0.0.1:${P5}:80"
        port_6="127.0.0.1:${P6}:443"
        port_7="127.0.0.1:${P7}:80"
    fi

    local root_password_for_services
    root_password_for_services="$(openssl rand -base64 16 | tr -dc 'a-zA-Z0-9' | head -c 24)"

    local home_dir="/home/${USERNAME}"
    cp /etc/openpanel/docker/compose/1.0/docker-compose.yml "${home_dir}/docker-compose.yml" || die "docker-compose.yml template missing from /etc/openpanel/"

    sed -i 's/\r$//' /etc/openpanel/docker/compose/1.0/.env
    cp /etc/openpanel/docker/compose/1.0/.env "${home_dir}/.env" || die ".env template missing from /etc/openpanel/"

    sed -i \
        -e "s|USERNAME=\"[^\"]*\"|USERNAME=\"${USERNAME}\"|g" \
        -e "s|USER_ID=\"[^\"]*\"|USER_ID=\"${USER_ID}\"|g" \
        -e "s|CONTEXT=\"[^\"]*\"|CONTEXT=\"${USERNAME}\"|g" \
        -e "s|^HTTP_PORT=\"[^\"]*\"|HTTP_PORT=\"${port_5}\"|g" \
        -e "s|HTTPS_PORT=\"[^\"]*\"|HTTPS_PORT=\"${port_6}\"|g" \
        -e "s|PGADMIN_PORT=\"[^\"]*\"|PGADMIN_PORT=\"${port_1}\"|g" \
        -e "s|POSTGRES_PORT=\"[^\"]*\"|POSTGRES_PORT=\"127.0.0.1:${port_3}\"|g" \
        -e "s|PMA_PORT=\"[^\"]*\"|PMA_PORT=\"${port_4}\"|g" \
        -e "s|{PMA_PORT}|${P4}|g" \
		-e "s|rootpassword|${root_password_for_services}|g" \
        -e "s|MYSQL_PORT=\"[^\"]*\"|MYSQL_PORT=\"127.0.0.1:${port_2}\"|g" \
        -e "s|PROXY_HTTP_PORT=\"[^\"]*\"|PROXY_HTTP_PORT=\"${port_7}\"|g" \
        "${home_dir}/.env"

    if [[ "$EMAIL" =~ ^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$ ]]; then
        sed -i "s|PGADMIN_MAIL=[^\"]*|PGADMIN_MAIL=${EMAIL}|g" "${home_dir}/.env"
    fi

    # Webserver
    if [[ -n "$WEBSERVER" ]]; then
        if [[ "$WEBSERVER" =~ ^varnish\+([a-zA-Z]+)$ ]]; then
            local ws="${BASH_REMATCH[1]}"
            log "Enabling Varnish + $ws"
            sed -i \
                -e "s|WEB_SERVER=\"[^\"]*\"|WEB_SERVER=\"${ws}\"|g" \
                -e "s|^#\(PROXY_HTTP_PORT=.*\)|\1|" \
                "${home_dir}/.env"
        elif [[ "$WEBSERVER" =~ ^(nginx|apache|openresty|openlitespeed|litespeed)$ ]]; then
            log "Setting webserver: $WEBSERVER"
            sed -i "s|WEB_SERVER=\"[^\"]*\"|WEB_SERVER=\"${WEBSERVER}\"|g" "${home_dir}/.env"
        else
            echo "[!] Warning: Invalid webserver '$WEBSERVER'. Using template default."
        fi
    fi

    # SQL
    if [[ -n "$SQL_TYPE" ]]; then
        if [[ "$SQL_TYPE" =~ ^(mysql|mariadb)$ ]]; then
            log "Setting SQL type: $SQL_TYPE"
            sed -i "s|MYSQL_TYPE=\"[^\"]*\"|MYSQL_TYPE=\"${SQL_TYPE}\"|g" "${home_dir}/.env"
        else
            echo "[!] Warning: Invalid SQL type '$SQL_TYPE'. Using template default."
        fi
    fi

    [[ -f "${home_dir}/.env" ]] || { hard_cleanup; die "Failed to create .env file."; }

    # Socket dirs and config files
    mkdir -p "${home_dir}/sockets/"{mysqld,postgres,redis,valkey,memcached}
    cp /etc/openpanel/mysql/user.cnf             "${home_dir}/custom.cnf"
    cp /etc/openpanel/postgres/postgresql.conf   "${home_dir}/postgre_custom.conf"
    cp /etc/openpanel/nginx/user-nginx.conf      "${home_dir}/nginx.conf"
    cp /etc/openpanel/openresty/nginx.conf       "${home_dir}/openresty.conf"
    cp /etc/openpanel/openlitespeed/httpd_config.conf "${home_dir}/openlitespeed.conf"
    cp /etc/openpanel/apache/httpd.conf          "${home_dir}/httpd.conf"
    cp /etc/openpanel/varnish/default.vcl        "${home_dir}/default.vcl"
    cp /etc/openpanel/ofelia/users.ini           "${home_dir}/crons.ini"   2>/dev/null || true
    cp /etc/openpanel/backups/backup.env         "${home_dir}/backup.env"  2>/dev/null || true
    cp -r /etc/openpanel/php/ini                 "${home_dir}/php.ini"

    chown -R "${USERNAME}:${USERNAME}" "${home_dir}/sockets"
    chmod 777 "${home_dir}/sockets/"

    printf '[client]\nuser=root\npassword=%s\n' "$root_password_for_services" > "${home_dir}/my.cnf"
    [[ -f "${home_dir}/my.cnf" ]] || echo "[!] Warning: Failed to create my.cnf."

    # shellcheck disable=SC2034
    MYSQL_ROOT_PASSWORD="$root_password_for_services"   # exported for pull_images
}

copy_skeleton_files() {
    log "Creating configuration files for the newly created user"
    cp -r /etc/openpanel/skeleton/ "/etc/openpanel/openpanel/core/users/${USERNAME}/" 2>/dev/null || true

    if [[ -n "$NODE_IP" ]]; then
        echo "{ \"ip\": \"${NODE_IP}\" }" > "/etc/openpanel/openpanel/core/users/${USERNAME}/ip.json"
    fi
}

create_docker_context() {
    local host
    if [[ -n "$NODE_IP" ]]; then
        host="ssh://${USERNAME}"
    else
        host="unix:///hostfs/run/user/${USER_ID}/docker.sock"
    fi
	docker context create "$USERNAME" --docker "host=${host}" --description "$USERNAME" >/dev/null 2>&1 || true
}

save_user_to_database() {
    log "Saving user '$USERNAME' to database"

    local escaped_hash
    escaped_hash="${HASHED_PASSWORD//\'/\'\'}"

    local query
    if [[ -n "$RESELLER" ]]; then
        query="INSERT INTO users (username, password, owner, email, plan_id, server) VALUES ('${USERNAME}', '${escaped_hash}', '${RESELLER//\'/\'\'}', '${EMAIL//\'/\'\'}', '${PLAN_ID}', '${USERNAME}');"
    else
        query="INSERT INTO users (username, password, email, plan_id, server) VALUES ('${USERNAME}', '${escaped_hash}', '${EMAIL//\'/\'\'}', '${PLAN_ID}', '${USERNAME}');"
    fi

    db_query "$query" || { hard_cleanup; die "Database insert failed."; }
    echo "[✔] Successfully added user $USERNAME with password: $PASSWORD"
}

send_welcome_email() {
    [[ "$SEND_EMAIL" == true ]] || return 0

    [[ "$EMAIL" =~ ^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$ ]] || { echo "[!] Warning: $EMAIL is not a valid address; skipping email."; return 0; }

    local domain protocol admin_port port login_url email_pw

    domain="$(opencli domain)"
    port="$(opencli port)"

    if [[ -f "/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/${domain}/${domain}.key" || -f "/etc/openpanel/caddy/ssl/custom/${domain}/${domain}.key" ]]; then
        protocol="https"
    else
        protocol="http"
    fi

    admin_port="$(awk '/# START HOSTNAME DOMAIN #/{flag=1;next}/# END HOSTNAME DOMAIN #/{flag=0}flag' /etc/openpanel/caddy/Caddyfile | grep -oP 'localhost:\K[0-9]+' | head -1)"
    login_url="${protocol:-http}://${domain}:${port}/login"

    local token
    token="$(tr -dc 'a-zA-Z0-9' < /dev/urandom | head -c 64)"
    sed -i "s|^mail_security_token=.*|mail_security_token=${token}|" "${PANEL_CONFIG_FILE}"

    email_pw="$PASSWORD"
    [[ "$email_plaintext_passwords" == "no" ]] && email_pw="********"

    curl -4 -fsSk -X POST "${protocol:-http}://${domain}:${admin_port}/send_email" -F "transient=${token}" -F "recipient=${EMAIL}" -F "subject=New OpenPanel account information" -F "body=OpenPanel URL: ${login_url} | username: ${USERNAME} | password: ${email_pw}" --max-time 15 || echo "[!] Warning: Failed to send welcome email."
}

generate_password_hash() {
    if [[ "$PASSWORD" == "generate" ]]; then
        PASSWORD="$(openssl rand -base64 12)"
        log "Generated password: $PASSWORD"
    fi

    local python3
    if [[ -x /usr/local/admin/venv/bin/python3 ]]; then
        python3=/usr/local/admin/venv/bin/python3
    elif command -v python3 &>/dev/null; then
        python3=python3
    else
        hard_cleanup; die "No Python 3 interpreter found. Install Python 3 or check the venv."
    fi

    # stdin to avoid it appearing in the process list!
    HASHED_PASSWORD="$(echo "$PASSWORD" \
        | "$python3" -c "
import sys
from werkzeug.security import generate_password_hash
pw = sys.stdin.readline().rstrip('\n')
print(generate_password_hash(pw))
")"

}

create_user_volume() {
    local vol_path="/home/${USERNAME}/docker-data/volumes/${USERNAME}_html_data/_data/"
    mkdir -p "$vol_path"
    chmod -R g+w "$vol_path"
    ln -sfn "$vol_path" "/home/${USERNAME}/files" 2>/dev/null || true
	chown -R "$USERNAME":"$USERNAME" "/home/${USERNAME}/docker-data/volumes"
}

notify_sentinel() {
    [[ "$SENTINEL" == true ]] || return 0
    nohup opencli sentinel --action=user_create --title="User account '${USERNAME}' created" --message="Account '${USERNAME}' created; email: ${EMAIL}; plan: ${PLAN_NAME}" >/dev/null 2>&1 &
    disown
}


# MAIN
(
flock -n 200 || { echo "[✘] Error: A user creation process is already running."; echo "Please wait for it to complete before starting a new one. Exiting."; exit 1; }

########################################################################
# 1. validate data and fetch plan info
validate_user_creation
validate_username
validate_password_in_lists
resolve_node
load_plan
check_reseller_limits

########################################################################
# 2. create system user
create_linux_user_local
# if slave server: create remote user, mount homedir, install docker
create_linux_user_remote
bootstrap_node
setup_sshfs_mount
ensure_docker_on_node
setup_ssh_key_for_user

########################################################################
# 3. setup docker rootless for user
setup_docker_rootless &
PID_ROOTLESS_INSTALL=$!

########################################################################
# 4. do background stuff while docker install is running

find_available_ports_bg() { find_available_ports > /tmp/ports_$$; }
find_available_ports_bg &
PID_PORTS=$!

copy_skeleton_files &

nohup sh -c "cd /root && docker compose up -d openpanel" </dev/null >/dev/null 2>&1 &
disown

generate_password_hash
create_user_volume &

wait $PID_PORTS
read -r P1 P2 P3 P4 P5 P6 P7 < /tmp/ports_$$
rm -f /tmp/ports_$$
configure_environment

setup_docker_compose
create_docker_context

########################################################################
# 6. validate docker service is started for user (socket exists), compose command and context are working
wait $PID_ROOTLESS_INSTALL
test_docker_service
autostart_services

########################################################################
# 8. save and notify
save_user_to_database

# needs to run AFTER saving user to database
nohup opencli plan-apply "$PLAN_ID" "$USERNAME" >/dev/null 2>&1 &
disown

# this needs to run AFTER plan-apply to show updated du info
nohup opencli user-quota >/dev/null 2>&1 &
disown

send_welcome_email &
notify_sentinel
update_reseller_account_count # must run after saving user to db

nohup chown -R "${USERNAME}:${USERNAME}" "/home/${USERNAME}/" 2>/dev/null &
disown

exit 0
) 200>"$LOCK_FILE"
