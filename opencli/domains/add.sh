#!/bin/bash
################################################################################
# Script Name: domains/add.sh
# Description: Add a domain name for user.
# Usage: opencli domains-add <DOMAIN_NAME> <USERNAME> [--docroot DOCUMENT_ROOT] [--php_version N.N] [--skip_caddy --skip_vhost --skip_containers --skip_dns] --debug
# Author: Stefan Pejcic
# Created: 20.08.2024
# Last Modified: 10.03.2026
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


# ======================================================================
# Constants

readonly PANEL_CONFIG_FILE='/etc/openpanel/openpanel/conf/openpanel.config'
readonly CADDYFILE='/etc/openpanel/caddy/Caddyfile'
readonly ZONE_FILE_DIR='/etc/bind/zones/'
readonly NAMED_CONF_LOCAL='/etc/bind/named.conf.local'
readonly BIND_CONFIG_FILE='/etc/bind/named.conf.options'

# ======================================================================
# Argument Parsing

[[ "$#" -lt 2 ]] && { echo "Usage: opencli domains-add <DOMAIN_NAME> <USERNAME> [--debug]"; exit 1; }

domain_name="$1"
user="$2"
context="$2"

# Validate domain name early
if ! [[ "$domain_name" =~ ^(xn--[a-z0-9-]+\.[a-z0-9-]+|[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})$ ]]; then
    echo "FATAL ERROR: Invalid domain name: $domain_name"
    exit 1
fi

# Defaults
debug_mode=false
docroot=""
php_version=""
onion_domain=false
is_subdomain=false
apex_domain=""
SKIP_VHOST_CREATE=false
SKIP_CADDY_CREATE=false
SKIP_DNS_ZONE=false
SKIP_STARTING_CONTAINERS=false
USE_PARENT_DNS_ZONE=false
REMOTE_SERVER=""
VARNISH=false
ws=""
current_ip=""
IPV4="yes"
dedicated_ip=""
context_uid=""
user_id=""
hs_ed25519_public_key=""
hs_ed25519_secret_key=""
hs_public_key=""
hs_secret_key=""

shift 2
while [[ $# -gt 0 ]]; do
    case "$1" in
        --debug)           debug_mode=true ; shift ;;
        --skip_caddy)      SKIP_CADDY_CREATE=true ; shift ;;
        --skip_vhost)      SKIP_VHOST_CREATE=true ; shift ;;
        --skip_dns)        SKIP_DNS_ZONE=true ; shift ;;
        --skip_containers) SKIP_STARTING_CONTAINERS=true ; shift ;;
        --docroot)
            [[ -z "${2:-}" ]] && { echo "FATAL ERROR: Missing value for --docroot"; exit 1; }
            docroot="$2"; shift 2 ;;
        --php_version)
            [[ -z "${2:-}" ]] && { echo "FATAL ERROR: Missing value for --php_version"; exit 1; }
            php_version="$2"
            [[ ! "$php_version" =~ ^[0-9]+\.[0-9]+$ ]] && { echo "FATAL ERROR: Invalid PHP version format '$php_version'. Expected format: N.N (e.g., 8.2)"; exit 1; }
            shift 2 ;;
        --hs_ed25519_public_key)
            [[ -z "${2:-}" ]] && { echo "FATAL ERROR: Missing value for --hs_ed25519_public_key"; exit 1; }
            hs_ed25519_public_key="$2"; shift 2 ;;
        --hs_ed25519_secret_key)
            [[ -z "${2:-}" ]] && { echo "FATAL ERROR: Missing value for --hs_ed25519_secret_key"; exit 1; }
            hs_ed25519_secret_key="$2"; shift 2 ;;
        *) shift ;;
    esac
done

# ======================================================================
# Helpers

log() { [[ "$debug_mode" == true ]] && echo "$1"; }

die() { echo "FATAL ERROR: $*"; exit 1; }

get_config_value() {
    grep -E "^\s*${1}=" "$PANEL_CONFIG_FILE" | sed -E "s/^\s*${1}=//" | tr -d '[:space:]'
}

run_bg() { nohup sh -c "$*" </dev/null >nohup.out 2>nohup.err & }

# ======================================================================
# Validations

verify_docroot() {
    if [[ -n "$docroot" && ! "$docroot" =~ ^/var/www/html/ ]]; then
        die "Invalid docroot. It must start with /var/www/html/"
    fi
    if [[ -z "$docroot" ]]; then
        docroot="/var/www/html/$domain_name"
        log "No document root specified, using $docroot"
    fi
}

compare_with_forbidden_domains_list() {
    local config_file='/etc/openpanel/openpanel/conf/domain_restriction.txt'
    [[ ! -f "forbidden_domains.txt" ]] && return
    log "Checking domain against forbidden_domains list"
    local forbidden_domains
    mapfile -t forbidden_domains < forbidden_domains.txt
    [[ " ${forbidden_domains[*]} " =~ " ${domain_name} " ]] && { echo "ERROR: $domain_name is a forbidden domain."; exit 1; }
}

compare_with_system_domains() {
    log "Checking domain against system domains list"
    if grep -q -E "^\s*$domain_name\s*\{" "$CADDYFILE" 2>/dev/null; then
        echo "ERROR: $domain_name is already configured."
        exit 1
    fi
}

check_domain_exists() {
    log "Checking if domain already exists on the server"
    if ! opencli domains-whoowns "$domain_name" | grep -q "not found in the database."; then
        echo "ERROR: Domain $domain_name already exists."
        exit 1
    fi
}

get_docker_context() {
    IFS=',' read -r user_id context <<< "$(mysql -se "SELECT CONCAT(id, ',', server) FROM users WHERE username='${user}';")"
    [[ -z "$user_id" || -z "$context" ]] && die "Missing user ID or context for user $user."
}

detect_apex_or_onion() {
    if [[ "$domain_name" =~ ^[a-zA-Z0-9]{16}\.onion$ ]]; then
        onion_domain=true
        log ".onion address - Tor will be configured.."
        verify_onion_files
        return
    fi

    local domain_lower tld_file update_tlds suffixes matched_suffix max_match_len suffix apex_domain_part sld
    domain_lower=$(echo "$domain_name" | tr '[:upper:]' '[:lower:]')
    tld_file="/etc/openpanel/openpanel/conf/public_suffix_list.dat"
    update_tlds=false

    if [[ ! -f "$tld_file" ]] || [[ -n "$(find "$tld_file" -mtime +6 2>/dev/null)" ]]; then
        update_tlds=true
    fi

    if [[ "$update_tlds" == true ]]; then
        mkdir -p "$(dirname "$tld_file")"
        wget --timeout=5 --tries=3 -q --inet4-only -O "$tld_file" "https://publicsuffix.org/list/public_suffix_list.dat" \
            || log "Failed to download TLD list from IANA"
    fi

    suffixes=$(grep -v '^//' "$tld_file" | grep -v '^$')
    matched_suffix=""
    max_match_len=0

    while read -r suffix; do
        if [[ ".$domain_lower" == *".$suffix" ]]; then
            local suffix_len=${#suffix}
            if (( suffix_len > max_match_len )); then
                matched_suffix="$suffix"
                max_match_len=$suffix_len
            fi
        fi
    done <<< "$suffixes"

    [[ "$domain_lower" == "$matched_suffix" ]] && { echo "ERROR: '$domain_lower' is a public suffix and cannot be used as a domain."; exit 1; }
    [[ -z "$matched_suffix" ]] && { echo "ERROR: Invalid domain or unrecognized TLD for '$domain_name'"; exit 1; }

    log "Detected public suffix (TLD): .$matched_suffix"
    local apex_part="${domain_lower%.$matched_suffix}"
    local sld_part="${apex_part##*.}"
    apex_domain="${sld_part}.${matched_suffix}"

    if [[ "$domain_lower" != "$apex_domain" ]]; then
        is_subdomain=true
        log "Domain '$domain_lower' is a subdomain of '$apex_domain'."
    fi
}

check_if_apex_exists_in_server() {
    [[ "$is_subdomain" != true ]] && return
    local whoowns_output existing_user
    whoowns_output=$(opencli domains-whoowns "$apex_domain")
    existing_user=$(echo "$whoowns_output" | awk -F "Owner of '$apex_domain': " '{print $2}')
    [[ -z "$existing_user" ]] && { echo "Apex domain: $apex_domain does not exist on this server."; return; }

    if [[ "$existing_user" == "$user" ]]; then
        log "User $existing_user already owns the apex domain $apex_domain - adding subdomain.."
        USE_PARENT_DNS_ZONE=true
    else
        local allow_subdomain_sharing
        allow_subdomain_sharing=$(get_config_value 'permit_subdomain_sharing')
        if [[ "$allow_subdomain_sharing" == "yes" ]]; then
            log "WARNING: Another user owns the apex domain: $apex_domain - adding subdomain as a separate addon on this account."
        else
            echo "Another user owns the domain: $apex_domain - can't add subdomain: $domain_name"
            exit 1
        fi
    fi
}

# ======================================================================
# TOR

verify_onion_files() {
    $onion_domain || return
    [[ -z "$hs_ed25519_public_key" || -z "$hs_ed25519_secret_key" ]] && \
        die "Both --hs_ed25519_public_key and --hs_ed25519_secret_key are required for .onion domains."

    for key in hs_ed25519_public_key hs_ed25519_secret_key; do
        [[ ! "${!key}" =~ ^/var/www/html/ ]] && die "--$key must be inside your /var/www/html/ directory."
    done

    hs_public_key="/home/$context/docker-data/volumes/${context}_html_data/_data/${hs_ed25519_public_key#/var/www/html/}"
    hs_secret_key="/home/$context/docker-data/volumes/${context}_html_data/_data/${hs_ed25519_secret_key#/var/www/html/}"
    [[ ! -f "$hs_public_key" || ! -f "$hs_secret_key" ]] && die "hs_ed25519_public_key or hs_ed25519_secret_key do not exist!"
}

setup_tor_for_user() {
    local tor_dir="/home/$context/tor"
    local folder_name

    if [[ ! -d "$tor_dir/hidden_service" || ! -f "$tor_dir/torrc" ]]; then
        folder_name="hidden_service"
    else
        local highest_folder=0
        for folder in "$tor_dir/hidden_service"/*/; do
            local folder_num
            folder_num=$(basename "$folder")
            [[ "$folder_num" =~ ^[0-9]+$ ]] && (( folder_num > highest_folder )) && highest_folder=$folder_num
        done
        folder_name="hidden_service/$((highest_folder + 1))"
    fi

    mkdir -p "$tor_dir/$folder_name/authorized_clients"
    cp "$hs_public_key" "$tor_dir/$folder_name/hs_ed25519_public_key"
    cp "$hs_secret_key" "$tor_dir/$folder_name/hs_ed25519_secret_key"
    chown "$context_uid:$context_uid" "/home/$context/tor"
    chmod 0600 "/home/$context/tor/torrc"

    local proxy_ws
    proxy_ws=$( [[ "$VARNISH" == true ]] && echo "varnish" || echo "$ws" )

    printf 'HiddenServiceDir /var/lib/tor/%s/\nHiddenServicePort 80 %s:80\n\n' "$folder_name" "$proxy_ws" >> "$tor_dir/torrc"
    log ".onion files are saved in $folder_name directory."
}

start_tor_for_user() {
    if docker --context "$context" ps -q -f name=tor | grep -q .; then
        log "Tor service is already running, restarting to apply new service configuration"
        run_bg "cd /home/$context/ && docker --context $context restart tor"
    else
        log "Starting Tor service.."
        run_bg "cd /home/$context/ && docker --context $context compose up -d tor"
    fi
}

# ======================================================================
# Server / Network

get_server_ipv4_or_ipv6() {
    local script_path="/usr/local/opencli/ip_servers.sh"
    log "Checking IPv4 address for the account"

    local IP_SERVER_1 IP_SERVER_2 IP_SERVER_3
    if [[ -f "$script_path" ]]; then
        source "$script_path"
    else
        IP_SERVER_1=IP_SERVER_2=IP_SERVER_3="https://ip.openpanel.com"
    fi

    get_ip() {
        local ip_version=$1; shift
        if [[ "$ip_version" == "-4" ]]; then
            curl --silent --max-time 2 "$ip_version" "$1" \
                || wget --timeout=2 --tries=1 -qO- "$2" \
                || curl --silent --max-time 2 "$ip_version" "$3"
        else
            curl --silent --max-time 2 "$ip_version" "$1" \
                || curl --silent --max-time 2 "$ip_version" "$3"
        fi
    }

    current_ip=$(get_ip "-4" "$IP_SERVER_1" "$IP_SERVER_2" "$IP_SERVER_3")
    [[ -z "$current_ip" ]] && { log "Fetching IPv4 from local hostname..."; current_ip=$(ip addr | awk '/inet / && /global/{print $2; exit}' | cut -d/ -f1); }

    IPV4="yes"

    if [[ -z "$current_ip" ]]; then
        IPV4="no"
        log "No IPv4 found. Checking IPv6 address..."
        current_ip=$(get_ip "-6" "$IP_SERVER_1" "$IP_SERVER_2" "$IP_SERVER_3")
        [[ -z "$current_ip" ]] && { log "Fetching IPv6 from local hostname..."; current_ip=$(ip addr | awk '/inet6 / && /global/{print $2; exit}' | cut -d/ -f1); }
    fi

    [[ -z "$current_ip" ]] && die "Unable to determine IP address (IPv4 or IPv6)."

    local json_file="/etc/openpanel/openpanel/core/users/$user/ip.json"
    if [[ -f "$json_file" ]]; then
        dedicated_ip=$(jq -r '.ip' "$json_file")
        if [[ -n "$dedicated_ip" ]]; then
            REMOTE_SERVER=$(docker context ls | grep -q "$dedicated_ip" && echo "yes" || echo "no")
            [[ "$REMOTE_SERVER" == "yes" ]] && log "IP address is assigned to a node (slave) server." || log "User has a dedicated IP address $dedicated_ip"
        else
            REMOTE_SERVER="no"
        fi
    fi
}

# ======================================================================
# Webserver / VHost

make_folder() {
    log "Creating document root directory $docroot"
    local stripped_docroot="${docroot#/var/www/html/}"
    context_uid=$(awk -F: -v user="$context" '$1 == user {print $3}' /hostfs/etc/passwd)
    [[ -z "$context_uid" ]] && { log "Warning: failed detecting user id, permissions issue!"; return; }

    local dirs=(
        "/home/$context/docker-data/volumes/${context}_html_data/_data/$stripped_docroot"
        "/home/$context/docker-data/volumes/${context}_webserver_data/_data/"
        "/home/$context/docker-data/volumes/${context}_html_data"
        "/home/$context/docker-data/volumes/${context}_html_data/_data"
    )
    mkdir -p "${dirs[@]}"
    timeout 3s chown -R "$context_uid:$context_uid" "${dirs[@]}"
    timeout 1s chmod -R g+w "${dirs[@]}"
}

check_and_create_default_nginx_conf() {
    [[ -e "/home/$context/nginx.conf" ]] && return
    log "Creating default vhost file for Nginx: /etc/nginx/nginx.conf"
    cat > "/home/$context/nginx.conf" <<'EOF'
user  nginx;
worker_processes  auto;
pid        /var/run/nginx.pid;

events {
    worker_connections  1024;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;
    sendfile        on;
    keepalive_timeout  65;
    include /etc/nginx/conf.d/*.conf;
}
EOF
}

get_webserver_for_user() {
    log "Checking webserver configuration"
    local output
    output=$(opencli webserver-get_webserver_for_user "$user")
    ws=$(echo "$output" | grep -Eo 'nginx|openresty|apache|openlitespeed|litespeed' | head -n1)
    [[ "$ws" == "nginx" ]] && check_and_create_default_nginx_conf
}

get_varnish_for_user() {
    VARNISH=false
    grep -qE "^PROXY_HTTP_PORT=" "/home/$context/.env" && VARNISH=true
}

vhost_files_create() {
    local vhost_in_docker_file="/home/$context/docker-data/volumes/${context}_webserver_data/_data/${domain_name}.conf"
    local vhost_docker_template="/etc/openpanel/nginx/vhosts/1.1/docker_${ws}_domain.conf"
    get_varnish_for_user

    if ! $SKIP_STARTING_CONTAINERS; then
        local services="$ws"
        [[ "$VARNISH" == true ]] && services="$services varnish"
        log "Starting $services containers.."
        run_bg "docker --context $context compose -f /home/$context/docker-compose.yml up -d $services"
    else
        log "Skipping starting ${ws} container."
    fi

    log "Creating ${domain_name}.conf"
    cp "$vhost_docker_template" "$vhost_in_docker_file"

    chown "$context_uid:$context_uid" "/home/$context/docker-data/volumes/${context}_webserver_data/"
    chown -R "$context_uid:$context_uid" "/home/$context/docker-data/volumes/${context}_webserver_data/_data/"

    sed -i \
        -e "s|<DOMAIN_NAME>|$domain_name|g" \
        -e "s|<USER>|$user|g" \
        -e "s|<PHP>|$php_version|g" \
        -e "s|<DOCUMENT_ROOT>|$docroot|g" \
        "$vhost_in_docker_file"

    ! $SKIP_STARTING_CONTAINERS && run_bg "cd /home/$context/ && docker --context $context restart $ws"
}

# ======================================================================
# Caddy

create_domain_file() {
    local logs_dir="/var/log/caddy/domlogs/${domain_name}"
    local waf_dir="/var/log/caddy/coraza_waf"
    mkdir -p "$logs_dir" && touch "$logs_dir/access.log"
    mkdir -p "$waf_dir" && touch "$waf_dir/${domain_name}.log"

    local env_file="/home/${context}/.env"
    [[ ! -f "$env_file" ]] && { echo "Warning: .env file not found!"; return 1; }
    source "$env_file"

    local non_ssl_port ssl_port ip_format_for_nginx
    non_ssl_port="${HTTP_PORT##*:}"
    ssl_port="${HTTPS_PORT##*:}"
    ip_format_for_nginx=$( [[ "$IPV4" == "yes" ]] && echo "$current_ip" || echo "[$current_ip]" )

    mkdir -p /etc/openpanel/caddy/domains/
    local domains_file="/etc/openpanel/caddy/domains/$domain_name.conf"
    local conf_template="/etc/openpanel/caddy/templates/domain.conf"

    # Build domain config from template
    local domain_conf
    if [[ "$REMOTE_SERVER" == "yes" ]]; then
        domain_conf=$(sed \
            -e "s|<DOMAIN_NAME>|$domain_name|g" \
            -e "s|127.0.0.1:<SSL_PORT>|$current_ip:$ssl_port|g" \
            -e "s|127.0.0.1:<NON_SSL_PORT>|$current_ip:$non_ssl_port|g" \
            "$conf_template")
    else
        domain_conf=$(sed \
            -e "s|<DOMAIN_NAME>|$domain_name|g" \
            -e "s|<SSL_PORT>|$ssl_port|g" \
            -e "s|<NON_SSL_PORT>|$non_ssl_port|g" \
            "$conf_template")
    fi
    echo "$domain_conf" > "$domains_file"

    # Dedicated IP binding
    if [[ -n "$dedicated_ip" ]]; then
        _update_bind_in_block "$domains_file" "http://$domain_name, http://*.$domain_name {" "$dedicated_ip"
        _update_bind_in_block "$domains_file" "https://$domain_name, https://*.$domain_name {" "$dedicated_ip"
    fi

    # WAF setting
    local waf_value
    if grep -qi "waf" "$PANEL_CONFIG_FILE" 2>/dev/null; then
        waf_value="On"; log "WAF module is enabled, setting SecRuleEngine On"
    else
        waf_value="Off"; log "WAF module is disabled, setting SecRuleEngine Off"
    fi
    sed -i "s/SecRuleEngine .*/SecRuleEngine $waf_value/" "$domains_file"

    # Varnish toggle
    if [[ "$VARNISH" == true ]]; then
        log "Enabling Varnish cache for the domain.."
        sed -i '/# Handle HTTPS traffic (port 443)/,+6 s/^/#/' "$domains_file"
        sed -i '/# Terminate TLS and pass to Varnish/,+3 s/^#//' "$domains_file"
    fi

    # Reload Caddy
    if docker --context default ps -q -f name=caddy | grep -q .; then
        log "Caddy is running, reloading with new domain configuration"
        if docker --context default restart caddy >/dev/null 2>&1; then
            log "Domain successfully added and Caddy reloaded."
        else
            log "Failed to add domain configuration, changes reverted."
        fi
    else
        log "Caddy is not running, starting in background.."
        run_bg "cd /root && docker --context default compose up -d caddy"
    fi
}

_escape_sed() { printf '%s' "$1" | sed -e 's/[\/&^$.*[]/\\&/g'; }

_update_bind_in_block() {
    local conf=$1 block_header=$2 ip=$3
    local esc_header
    esc_header=$(_escape_sed "$block_header")
    if sed -n "/^$esc_header/{n;/^[[:space:]]*bind /p}" "$conf" | grep -q "bind "; then
        sed -i "/^$esc_header/{n;s/^[[:space:]]*bind .*/    bind $ip/}" "$conf"
    else
        sed -i "/^$esc_header/a\    bind $ip" "$conf"
    fi
}

# ======================================================================
# DNS

create_zone_file() {
    mkdir -p "$ZONE_FILE_DIR"

    if $USE_PARENT_DNS_ZONE; then
        log "Adding records to existing DNS zone for apex domain: $ZONE_FILE_DIR$apex_domain.zone"
        local spf_prefix
        spf_prefix=$( [[ "$IPV4" == "yes" ]] && echo "ip4" || echo "ip6" )
        echo "$domain_name    14400     IN      A       $current_ip" >> "$ZONE_FILE_DIR$apex_domain.zone"
        echo "$domain_name    14400     IN      TXT       'v=spf1 ${spf_prefix}:${current_ip} +a +mx ~all'" >> "$ZONE_FILE_DIR$apex_domain.zone"
        return
    fi

    local zone_template_path
    if [[ "$IPV4" == "yes" ]]; then
        zone_template_path='/etc/openpanel/bind9/zone_template.txt'
        log "Creating DNS zone file with A records: $ZONE_FILE_DIR$domain_name.zone"
    else
        zone_template_path='/etc/openpanel/bind9/zone_template_ipv6.txt'
        log "Creating DNS zone file with AAAA records: $ZONE_FILE_DIR$domain_name.zone"
    fi

    local ns1 ns2 ns3 ns4 rpemail timestamp zone_template zone_content
    ns1=$(get_config_value 'ns1'); ns1=${ns1:-ns1.openpanel.org}
    ns2=$(get_config_value 'ns2'); ns2=${ns2:-ns2.openpanel.org}
    ns3=$(get_config_value 'ns3')
    ns4=$(get_config_value 'ns4')
    rpemail=$(get_config_value 'email'); rpemail=${rpemail:-root.${domain_name}}
    rpemail="${rpemail//@/.}"
    timestamp=$(date +"%Y%m%d")
    zone_template=$(<"$zone_template_path")

    local sed_args=(
        -e "s|{domain}|$domain_name|g"
        -e "s|{ns1}|$ns1|g"
        -e "s|{ns2}|$ns2|g"
        -e "s|{rpemail}|$rpemail|g"
        -e "s|{server_ip}|$current_ip|g"
        -e "s|YYYYMMDD|$timestamp|g"
    )
    [[ -n "$ns3" ]] && sed_args+=( -e "s|{ns3}|$ns3|g" -e "s|{ns4}|$ns4|g" )

    zone_content=$(echo "$zone_template" | sed "${sed_args[@]}")
    echo "$zone_content" > "$ZONE_FILE_DIR$domain_name.zone"
}

update_named_conf() {
    if $USE_PARENT_DNS_ZONE; then
        grep -q "zone \"$apex_domain\"" "$NAMED_CONF_LOCAL" && return
        echo "zone \"$apex_domain\" IN { type master; file \"$ZONE_FILE_DIR$domain_name.zone\"; };" >> "$NAMED_CONF_LOCAL"
    else
        log "Adding the newly created zone file to the DNS server"
        if grep -q "zone \"$domain_name\"" "$NAMED_CONF_LOCAL"; then
            log "Domain '$domain_name' already exists in $NAMED_CONF_LOCAL"; return
        fi
        echo "zone \"$domain_name\" IN { type master; file \"$ZONE_FILE_DIR$domain_name.zone\"; };" >> "$NAMED_CONF_LOCAL"
    fi
}

reload_bind_after_slaves() {
    if docker --context default ps -q -f name=openpanel_dns | grep -q .; then
        log "DNS service is running, adding the zone"
        docker --context default exec openpanel_dns rndc reconfig >/dev/null 2>&1
    else
        log "DNS is enabled but the DNS service is not yet started, starting now.."
        run_bg "cd /root && docker --context default compose up -d bind9"
    fi
}

notify_slave() {
    $USE_PARENT_DNS_ZONE && return
    echo "Notifying Slave DNS server ($SLAVE_IP) to create a new zone for domain $domain_name"
    timeout 5 ssh -q -o LogLevel=ERROR -o ConnectTimeout=5 -T root@"$SLAVE_IP" <<EOF >/dev/null 2>&1
if ! grep -q "$domain_name.zone" /etc/bind/named.conf.local; then
    echo "zone \"$domain_name\" { type slave; masters { $MASTER_IP; }; file \"/etc/bind/zones/$domain_name.zone\"; };" >> /etc/bind/named.conf.local
    touch /etc/bind/zones/$domain_name.zone
fi
EOF
    timeout 5 ssh -q -o LogLevel=ERROR -o ConnectTimeout=5 -T root@"$SLAVE_IP" <<'EOF' >/dev/null 2>&1
if docker --context default ps -q -f name=openpanel_dns | grep -q .; then
    docker --context default exec openpanel_dns rndc reconfig >/dev/null 2>&1
else
    nohup sh -c "cd /root && docker --context default compose up -d bind9" </dev/null >nohup.out 2>nohup.err &
fi
EOF
}

get_slave_dns_option() {
    [[ ! -f "$BIND_CONFIG_FILE" ]] && return
    if ! grep -qP '^(?!\s*//).*allow-transfer\s+\{[^}]*\}' "$BIND_CONFIG_FILE" \
        || ! grep -qP '^(?!\s*//).*also-notify\s+\{[^}]*\}' "$BIND_CONFIG_FILE"; then
        return
    fi

    local ALLOW_TRANSFER ALSO_NOTIFY
    ALLOW_TRANSFER=$(grep -oP '^(?!\s*//).*allow-transfer\s+\{\s*\K[^}]*' "$BIND_CONFIG_FILE" | tr -d '[:space:]')
    ALSO_NOTIFY=$(grep -oP '^(?!\s*//).*also-notify\s+\{\s*\K[^}]*' "$BIND_CONFIG_FILE" | tr -d '[:space:]')

    local -a ALLOW_TRANSFER_IPS ALSO_NOTIFY_IPS
    IFS=';' read -r -a ALLOW_TRANSFER_IPS <<< "$ALLOW_TRANSFER"
    IFS=';' read -r -a ALSO_NOTIFY_IPS <<< "$ALSO_NOTIFY"

    for ip in "${ALLOW_TRANSFER_IPS[@]}"; do
        [[ -z "$ip" ]] && continue
        if [[ " ${ALSO_NOTIFY_IPS[*]} " == *" $ip "* ]]; then
            SLAVE_IP=$ip; MASTER_IP=$current_ip
            notify_slave
        fi
    done
}

dns_stuff() {
    local enabled_modules_line
    enabled_modules_line=$(grep '^enabled_modules=' "$PANEL_CONFIG_FILE")
    if [[ "$enabled_modules_line" == *"dns"* ]]; then
    # TODO: check for dynamicdns
        create_zone_file
        get_slave_dns_option
        update_named_conf
        reload_bind_after_slaves
    else
        log "DNS module is disabled - skipping creating DNS records"
    fi
}

# ======================================================================
# PHP / Mail

get_php_version() {
    [[ -z "$php_version" ]] && php_version=$(opencli php-default "$user" | grep -oP '\d+\.\d+')
}

start_default_php_fpm_service() {
    local enabled_modules_line
    enabled_modules_line=$(grep '^enabled_modules=' "$PANEL_CONFIG_FILE")
    if [[ "$enabled_modules_line" == *"php"* ]]; then
        log "Starting container for the PHP version ${php_version}"
        run_bg "docker --context $context compose -f /home/$context/docker-compose.yml up -d php-fpm-${php_version}"
    else
        log "'php' module is disabled, skip starting container for the PHP version ${php_version}"
    fi
}

create_mail_mountpoint() {
    local key_value
    key_value=$(grep "^key=" "$PANEL_CONFIG_FILE" | cut -d'=' -f2-)
    [[ -z "$key_value" ]] && return

    local compose_file="/usr/local/mail/openmail/compose.yml"
    [[ ! -f "$compose_file" ]] && return

    local store_in volume_to_add
    store_in=$(grep -E '^email_storage_location=' /etc/openpanel/openadmin/config/admin.ini | cut -d'=' -f2- | xargs)

    if [[ "$store_in" == /* ]]; then
        log "Using $store_in for email storage"
        mkdir -p "$store_in"
        volume_to_add="      - $store_in:/var/mail/"
    else
        local domain_dir="/home/$context/mail/$domain_name/"
        log "Creating directory $domain_dir for emails"
        mkdir -p "$domain_dir"
        volume_to_add="      - $domain_dir:/var/mail/$domain_name/"
    fi

    log "Configuring mailserver in background.."
    if ! grep -qF "$volume_to_add" "$compose_file"; then
        sed -i "/^  mailserver:/,/^  roundcube:/ {/^    volumes:/a\\
$volume_to_add
        }" "$compose_file"
        run_bg "cd /usr/local/mail/openmail/ && docker-compose up -d --force-recreate mailserver"
    fi
}

# ======================================================================
# FTP / Domain DB

check_and_fix_FTP_permissions() {
    local real_path="/home/${user}/docker-data/volumes/${user}_html_data/_data/"
    local relative_path="${docroot##/var/www/html/}"
    local new_directory="${real_path}${relative_path}"

    if opencli ftp-list "$user" | tail -n +2 | cut -d'|' -f2 | sed 's/^ *//;s/ *$//' | grep -Fxq "$docroot"; then
        chown -R "$user:$user" "$new_directory"
        local dir
        for dir in "/home/$user" "/home/$user/docker-data" "/home/$user/docker-data/volumes" \
                   "/home/$user/docker-data/volumes/${user}_html_data" \
                   "/home/$user/docker-data/volumes/${user}_html_data/_data"; do
            chmod +rx "$dir"
        done
    fi
}

add_domain() {
    local user_id="$1"
    local domain_name="$2"
    log "Adding $domain_name to the domains database"

    mysql -e "INSERT INTO domains (user_id, docroot, php_version, domain_url) VALUES ('$user_id', '$docroot', '$php_version', '$domain_name');"

    local result
    result=$(mysql -N -e "SELECT COUNT(*) FROM domains WHERE user_id = '$user_id' AND docroot = '$docroot' AND domain_url = '$domain_name';")

    if [[ "$result" -ne 1 ]]; then
        log "Adding domain $domain_name failed! Contact administrator to check if the mysql database is running."
        echo "Failed to add domain $domain_name for user $user (id:$user_id)."
        return 1
    fi

    make_folder
    get_server_ipv4_or_ipv6

    if ! $SKIP_VHOST_CREATE; then
        get_webserver_for_user
        vhost_files_create
    else
        log "Skipping VirtualHost file creation due to '--skip_vhost' flag."
    fi

    if $onion_domain; then
        setup_tor_for_user
        start_tor_for_user
    else
        $SKIP_CADDY_CREATE \
            && log "Skipping Reverse Proxy file creation due to '--skip_caddy' flag." \
            || create_domain_file

        $SKIP_DNS_ZONE \
            && log "Skipping DNS zone file creation due to '--skip_dns' flag." \
            || dns_stuff
    fi

    if ! $SKIP_STARTING_CONTAINERS && [[ "$ws" != *litespeed* ]]; then
        start_default_php_fpm_service
    fi

    ! $onion_domain && create_mail_mountpoint

    nohup opencli sentinel --action=domains_create --title="Domain added" --message="Domain name: '$domain_name' has been added to OpenPanel user: '$user'." >/dev/null 2>&1 &
    disown

    echo "Domain $domain_name added successfully"
}

# ======================================================================
# Main

detect_apex_or_onion
check_domain_exists
compare_with_forbidden_domains_list
compare_with_system_domains
check_if_apex_exists_in_server
get_docker_context
get_php_version
verify_docroot
add_domain "$user_id" "$domain_name"
check_and_fix_FTP_permissions
