#!/bin/bash
################################################################################
# Script Name: domains/add.sh
# Description: Add a domain name for user.
# Usage: opencli domains-add <DOMAIN_NAME> <USERNAME> [--docroot DOCUMENT_ROOT] [--php_version N.N] [--skip_caddy --skip_vhost --skip_containers --skip_dns] --debug
# Author: Stefan Pejcic
# Created: 20.08.2024
# Last Modified: 25.06.2026
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

# Constants
readonly PANEL_CONFIG_FILE='/etc/openpanel/openpanel/conf/openpanel.config'
readonly CADDYFILE='/etc/openpanel/caddy/Caddyfile'
readonly NAMED_CONF_LOCAL='/etc/bind/named.conf.local'
readonly BIND_CONFIG_FILE='/etc/bind/named.conf.options'
readonly ZONE_FILE_DIR='/etc/bind/zones/'
readonly FORBIDDEN_DOMAINS_FILE='/etc/openpanel/openpanel/conf/domain_restriction.txt'
readonly TLD_FILE='/etc/openpanel/openpanel/conf/public_suffix_list.dat'

# Globals
domain_name=""
user=""
context=""
context_uid=""
user_id=""
docroot=""
php_version=""
current_ip=""
dedicated_ip=""
apex_domain=""
is_subdomain=false
onion_domain=false
IPV4="yes"
VARNISH=false
WEB_SERVER=""
REMOTE_SERVER="no"
USE_PARENT_DNS_ZONE=false
ENTERPRISE=""

hs_ed25519_public_key=""
hs_ed25519_secret_key=""

debug_mode=false
SKIP_VHOST_CREATE=false
SKIP_CADDY_CREATE=false
SKIP_DNS_ZONE=false
SKIP_STARTING_CONTAINERS=false

# ======================================================================
# Logging / error helpers

log()  { [[ "$debug_mode" == true ]] && echo "$*" >&2 || true; }
die()  { echo "FATAL ERROR: $*" >&2; exit 1; } #TODO: hard_cleanup like on user/add.sh
err()  { echo "ERROR: $*" >&2; exit 1; }
warn() { echo "WARNING: $*" >&2; }

# ======================================================================
# Arg parsing
parse_args() {
	[[ "$#" -lt 2 ]] && { echo "Usage: opencli domains-add <DOMAIN_NAME> <USERNAME> [--docroot /var/www/html/...] [--php_version N.N] [--debug]"; exit 1; }

    domain_name="$1"
    user="$2"
    shift 2

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --debug)               debug_mode=true; shift ;;
            --skip_caddy)          SKIP_CADDY_CREATE=true; shift ;;
            --skip_vhost)          SKIP_VHOST_CREATE=true; shift ;;
            --skip_dns)            SKIP_DNS_ZONE=true; shift ;;
            --skip_containers)     SKIP_STARTING_CONTAINERS=true; shift ;;
            --docroot)
                [[ -z "${2:-}" ]] && die "Missing value for --docroot"
                docroot="$2"; shift 2 ;;
            --php_version)
                [[ -z "${2:-}" ]] && die "Missing value for --php_version"
                php_version="$2"
                [[ ! "$php_version" =~ ^[0-9]+\.[0-9]+$ ]] && \
                    die "Invalid PHP version format '$php_version'. Expected: N.N (e.g. 8.2)"
                shift 2 ;;
            --hs_ed25519_public_key)
                [[ -z "${2:-}" ]] && die "Missing value for --hs_ed25519_public_key"
                hs_ed25519_public_key="$2"; shift 2 ;;
            --hs_ed25519_secret_key)
                [[ -z "${2:-}" ]] && die "Missing value for --hs_ed25519_secret_key"
                hs_ed25519_secret_key="$2"; shift 2 ;;
            *) shift ;;
        esac
    done
}

# ======================================================================
# helpers

get_config_value() {
    grep -E "^\s*${1}=" "$PANEL_CONFIG_FILE" \
        | sed -E "s/^\s*${1}=//" \
        | tr -d '[:space:]' \
        | head -n1
}

read_config_file() {
    while IFS='=' read -r key value; do
        [[ "$key" =~ ^[[:space:]]*# ]] && continue   # skip comments
        [[ -z "$key" ]] && continue                  # skip blank lines
        case "$key" in
            key) ENTERPRISE="$value" ;;
            ns1) ns1="$value" ;;
            ns2) ns2="$value" ;;
            ns3) ns3="$value" ;;
            ns4) ns4="$value" ;;
			rpemail) rpemail="$value" ;;
			permit_subdomain_sharing) permit_subdomain_sharing="$value" ;;
            enabled_modules) enabled_modules="$value" ;;
        esac
    done < "$PANEL_CONFIG_FILE"
}

is_module_enabled() {
    local module="$1"
    [[ ",$enabled_modules," == *",$module,"* ]]	
}

validate_domain_format() {
    [[ "$domain_name" =~ ^(xn--[a-z0-9-]+\.[a-z0-9-]+|[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})$ ]] || die "Invalid domain name: $domain_name"
}

validate_docroot() {
    if [[ -n "$docroot" ]]; then
        [[ "$docroot" =~ ^/var/www/html/ ]] || die "Invalid docroot. Must start with /var/www/html/"
    else
        docroot="/var/www/html/$domain_name"
        log "No docroot specified, using $docroot"
    fi
}

sql_escape() {
    local val="$1"
    val="${val//\\/\\\\}"
    val="${val//\'/\\\'}"
    printf '%s' "$val"
}

check_forbidden_domains() {
    [[ -f "$FORBIDDEN_DOMAINS_FILE" ]] || return 0
    log "Checking domain against forbidden list"
    if grep -qxF "$domain_name" "$FORBIDDEN_DOMAINS_FILE"; then
        err "$domain_name is a forbidden domain."
    fi
}

check_system_domains() {
    log "Checking domain against system domains"
    if grep -qE "^\s*$domain_name\s*\{" "$CADDYFILE" 2>/dev/null; then
        err "$domain_name is already configured in Caddy."
    fi
}

check_domain_not_taken() {
    log "Checking if domain already exists in database"
    if ! opencli domains-whoowns "$domain_name" | grep -q "not found in the database."; then
        err "Domain $domain_name already exists."
    fi
}

refresh_tld_list() {
    mkdir -p "$(dirname "$TLD_FILE")"
    if wget --timeout=5 --tries=3 -q --inet4-only -O "${TLD_FILE}.tmp" "https://publicsuffix.org/list/public_suffix_list.dat"; then
        mv "${TLD_FILE}.tmp" "$TLD_FILE"
    else
        log "Failed to refresh TLD list; using existing copy"
        rm -f "${TLD_FILE}.tmp"
    fi
}

detect_apex_or_onion() {
    if [[ "$domain_name" =~ ^[a-zA-Z0-9]{16}\.onion$ ]]; then
        onion_domain=true
        log ".onion address — Tor will be configured"
        return
    fi

    local domain_lower
    domain_lower=$(echo "$domain_name" | tr '[:upper:]' '[:lower:]')

    # Refresh TLD list if missing or older than 7 days
    if [[ ! -f "$TLD_FILE" ]] || [[ -n "$(find "$TLD_FILE" -mtime +6 2>/dev/null)" ]]; then
        refresh_tld_list
    fi

    local matched_suffix="" max_match_len=0 suffix
    while IFS= read -r suffix; do
        local suffix_pattern=".$suffix"
        if [[ ".$domain_lower" == *"$suffix_pattern" ]]; then
            local suffix_len=${#suffix}
            if (( suffix_len > max_match_len )); then
                matched_suffix="$suffix"
                max_match_len=$suffix_len
            fi
        fi
    done < <(grep -v '^//' "$TLD_FILE" | grep -v '^$')

    [[ "$domain_lower" == "$matched_suffix" ]] && err "'$domain_lower' is a public suffix and cannot be used as a domain."

    [[ -z "$matched_suffix" ]] && err "Invalid domain or unrecognized TLD for '$domain_name'"

    log "Detected public suffix: .$matched_suffix"

    local registrable="${domain_lower%.$matched_suffix}"
    local sld="${registrable##*.}"
    apex_domain="${sld}.${matched_suffix}"

    if [[ "$domain_lower" != "$apex_domain" ]]; then
        is_subdomain=true
        log "Domain '$domain_lower' is a subdomain of '$apex_domain'"
    fi
}

check_apex_ownership() {
    [[ "$is_subdomain" == true ]] || return 0

    local whoowns_output existing_user
    whoowns_output=$(opencli domains-whoowns "$apex_domain")
    existing_user=$(echo "$whoowns_output" | awk -F "Owner of '$apex_domain': " '{print $2}')

    if [[ -z "$existing_user" ]]; then
        log "Apex domain $apex_domain not found on this server"
        return 0
    fi

    if [[ "$existing_user" == "$user" ]]; then
        log "User $user already owns apex $apex_domain — adding subdomain"
        USE_PARENT_DNS_ZONE=true
    else
        if [[ "$permit_subdomain_sharing" == "yes" ]]; then
            warn "Another user owns $apex_domain — adding as separate addon domain"
        else
            err "Another user owns $apex_domain — cannot add subdomain $domain_name"
        fi
    fi
}

get_docker_context() {
    log "Fetching user context from database"
    local row
    row=$(mysql -sN -e "SELECT id, server FROM users WHERE username='$(sql_escape "$user")' LIMIT 1;" 2>/dev/null) || die "Database query failed"
    read -r user_id context <<< "$row"
    [[ -n "$user_id" && -n "$context" ]] || die "Missing user ID or context for user $user"
}

check_domains_limit() {
    if [[ "$USE_PARENT_DNS_ZONE" == true ]]; then
        log "Adding subdomain of existing domain — does not count toward limit"
        return 0
    fi

    local limit
    limit=$(mysql -sN -e "
        SELECT p.domains_limit
        FROM users u
        JOIN plans p ON u.plan_id = p.id
        WHERE u.username='$(sql_escape "$user")'
        LIMIT 1;" 2>/dev/null)

    # 0 or empty = unlimited
    [[ -z "$limit" || "$limit" -eq 0 ]] && return 0

    local current_count
    current_count=$(mysql -sN -e "
        SELECT COUNT(*)
        FROM domains d
        WHERE d.user_id = '$user_id'
        AND NOT EXISTS (
            SELECT 1 FROM domains d2
            WHERE d2.user_id = '$user_id'
            AND d.domain_url LIKE CONCAT('%.', d2.domain_url)
        );" 2>/dev/null)

    log "Domains limit: $limit, current apex domains: $current_count"

    if (( current_count >= limit )); then
        err "Domain limit reached. Plan allows $limit domain(s), user currently has $current_count."
    fi
}

get_user_env_value() {
	# TODO: merge with source we do later
    grep "^${1}=" "/home/$context/.env" | head -n1 | awk -F '=' '{print $2}' | tr -d '[:space:]' | sed 's/^"\(.*\)"$/\1/'
}

get_webserver_for_user() {
    log "Detecting web server"
    local raw_ws
    raw_ws=$(get_user_env_value WEB_SERVER)
    WEB_SERVER=$(echo "$raw_ws" | grep -Eo 'nginx|openresty|apache|openlitespeed|litespeed' | head -n1)
    [[ "$WEB_SERVER" == "nginx" ]] && ensure_nginx_conf
}

get_varnish_for_user() {
    VARNISH=false
    grep -qE "^PROXY_HTTP_PORT=" "/home/$context/.env" && VARNISH=true || true
}

get_php_version() {
    if [[ -z "$php_version" ]]; then
        php_version=$(opencli php-default "$user" | grep -oP '\d+\.\d+')
    fi
}

get_server_ip() {
    local ip_server="https://ip.openpanel.com"

    current_ip=$(curl --silent --max-time 2 -4 "$ip_server" 2>/dev/null || wget --timeout=2 --tries=1 -qO- "$ip_server" 2>/dev/null || true)

    if [[ -z "$current_ip" ]]; then
        log "ip.openpanel.com unreachable, trying ifconfig.me"
        current_ip=$(curl --silent --max-time 2 -4 https://ifconfig.me 2>/dev/null || wget --timeout=2 --tries=1 -qO- https://ifconfig.me 2>/dev/null || true)
    fi

    if [[ -z "$current_ip" ]]; then
        log "Fetching IPv4 from local interface"
        current_ip=$(ip addr show | awk '/inet / && /global/{print $2; exit}' | cut -d/ -f1)
    fi

    if [[ -n "$current_ip" ]]; then
        IPV4="yes"
    else
        IPV4="no"
        log "No IPv4 found, trying IPv6"
        current_ip=$(curl --silent --max-time 2 -6 "$ip_server" 2>/dev/null || true)
        [[ -z "$current_ip" ]] && current_ip=$(ip addr show | awk '/inet6 / && /global/{print $2; exit}' | cut -d/ -f1)
    fi

    [[ -n "$current_ip" ]] || die "Unable to determine server IP address"

    # Check for dedicated IP
    local json_file="/etc/openpanel/openpanel/core/users/$user/ip.json"
    if [[ -f "$json_file" ]]; then
        dedicated_ip=$(jq -r '.ip // empty' "$json_file")
        if [[ -n "$dedicated_ip" ]]; then
            if docker context ls | grep -q "$dedicated_ip"; then
                REMOTE_SERVER="yes"
                log "IP belongs to a node (slave) server"
            else
                REMOTE_SERVER="no"
                log "User has dedicated IP: $dedicated_ip"
            fi
        fi
    fi
}

make_docroot_dirs() {
    log "Creating docroot directories"
    context_uid=$(stat -c '%u' "/home/$context" 2>/dev/null) || true
    if [[ -z "$context_uid" ]]; then
        warn "Could not detect UID for $context — skipping chown"
        return
    fi

    local stripped="${docroot#/var/www/html/}"
    local dirs=(
        "/home/$context/docker-data/volumes/${context}_html_data/_data/$stripped"
        "/home/$context/docker-data/volumes/${context}_webserver_data/_data/"
        "/home/$context/docker-data/volumes/${context}_html_data"
        "/home/$context/docker-data/volumes/${context}_html_data/_data"
    )

    mkdir -p "${dirs[@]}"
    timeout 3s chown -R "$context_uid:$context_uid" "${dirs[@]}" || true
    timeout 1s chmod -R g+w "${dirs[@]}" || true
}

# ======================================================================
# Web server vhost

ensure_nginx_conf() {
    [[ -e "/home/$context/nginx.conf" ]] && return
    log "Creating default nginx.conf"
    cat > "/home/$context/nginx.conf" <<'NGINXEOF'
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
NGINXEOF
}

fix_ftp_perms() {
    real_path="/home/${context}/docker-data/volumes/${context}_html_data/_data/"
    relative_path="${docroot##/var/www/html/}"
    new_directory="${real_path}${relative_path}"

    if opencli ftp-list "$user" | tail -n +2 | cut -d'|' -f2 | sed 's/^ *//;s/ *$//' | grep -Fxq "$docroot"; then
        chown -R "$user:$user" "$new_directory"
        for dir in \
            "/home/$context" \
            "/home/$context/docker-data" \
            "/home/$context/docker-data/volumes" \
            "/home/$context/docker-data/volumes/${context}_html_data" \
            "/home/$context/docker-data/volumes/${context}_html_data/_data"; do
            chmod +rx "$dir"
        done
    fi
}

create_vhost_file() {
    $SKIP_VHOST_CREATE && { log "Skipping vhost creation (--skip_vhost)"; return; }

    get_varnish_for_user

    if ! $SKIP_STARTING_CONTAINERS; then
        local services="$WEB_SERVER"
        [[ "$VARNISH" == true ]] && services="$services varnish"
	    local service_count=$(echo "$services" | wc -w)
	    local container_word=$([ "$service_count" -eq 1 ] && echo "container" || echo "containers")
	    log "Starting $service_count $container_word ($services)"	
        nohup sh -c "docker --context $context compose -f /home/$context/docker-compose.yml up -d $services" </dev/null >nohup.out 2>nohup.err &
		disown
    fi

    local vhost_file="/home/$context/docker-data/volumes/${context}_webserver_data/_data/${domain_name}.conf"
    local template="/etc/openpanel/nginx/vhosts/1.1/docker_${WEB_SERVER}_domain.conf"

    log "Creating vhost: $vhost_file"
    cp "$template" "$vhost_file"

    chown "$context_uid:$context_uid" "/home/$context/docker-data/volumes/${context}_webserver_data/" "/home/$context/docker-data/volumes/${context}_webserver_data/_data/" 2>/dev/null || true

	sed -i -e "s|<DOMAIN_NAME>|$domain_name|g" -e "s|<USER>|$user|g" -e "s|<PHP>|$php_version|g" -e "s|<DOCUMENT_ROOT>|$docroot|g" "$vhost_file"

    if ! $SKIP_STARTING_CONTAINERS; then
        nohup sh -c "cd /home/$context/ && docker --context $context restart $WEB_SERVER" </dev/null >nohup.out 2>nohup.err &
		disown
    fi
}

# ======================================================================
# Caddy domain config

escape_sed_regex() {
    printf '%s' "$1" | sed -e 's/[\/&^$.*[]/\\&/g'
}

update_bind_in_block() {
    local conf="$1" block_header="$2" ip="$3"
    local esc
    esc=$(escape_sed_regex "$block_header")

    if sed -n "/^$esc/{n;/^[[:space:]]*bind /p}" "$conf" | grep -q "bind "; then
        sed -i "/^$esc/{n;s/^[[:space:]]*bind .*/    bind $ip/}" "$conf"
    else
        sed -i "/^$esc/a\\    bind $ip" "$conf"
    fi
}

create_caddy_domain_file() {
    $SKIP_CADDY_CREATE && { log "Skipping Caddy config (--skip_caddy)"; return; }

    local logs_dir="/var/log/caddy/domlogs/$domain_name"
    local waf_dir="/var/log/caddy/coraza_waf"
    mkdir -p "$logs_dir" && touch "$logs_dir/access.log"
    mkdir -p "$waf_dir"  && touch "$waf_dir/${domain_name}.log"

    local env_file="/home/${context}/.env"
    [[ -f "$env_file" ]] || { warn ".env not found for $context"; return 1; }
    # shellcheck source=/dev/null
    source "$env_file"

    local non_ssl_port ssl_port
    non_ssl_port=$(echo "$HTTP_PORT"  | cut -d':' -f2)
    ssl_port=$(echo     "$HTTPS_PORT" | cut -d':' -f2)

    mkdir -p /etc/openpanel/caddy/domains/
    local domains_file="/etc/openpanel/caddy/domains/${domain_name}.conf"
    local conf_template="/etc/openpanel/caddy/templates/domain.conf"

	local host="127.0.0.1:"
	[[ "$REMOTE_SERVER" == "yes" ]] && host="$current_ip:"
	
	sed -e "s|<DOMAIN_NAME>|$domain_name|g" -e "s|127.0.0.1:||g" -e "s|<SSL_PORT>|${host}${ssl_port}|g" -e "s|<NON_SSL_PORT>|${host}${non_ssl_port}|g" "$conf_template" > "$domains_file"

    # Dedicated IP bind directives
    if [[ -n "$dedicated_ip" ]]; then
        update_bind_in_block "$domains_file" "http://$domain_name, http://*.$domain_name {" "$dedicated_ip"
        update_bind_in_block "$domains_file" "https://$domain_name, https://*.$domain_name {" "$dedicated_ip"
    fi

    # WAF rule engine
    local waf_value="Off"
    grep -qi "waf" "$PANEL_CONFIG_FILE" 2>/dev/null && waf_value="On"
    log "WAF SecRuleEngine: $waf_value"
    sed -i "s/SecRuleEngine .*/SecRuleEngine $waf_value/" "$domains_file"

    # Varnish routing
    if [[ "$VARNISH" == true ]]; then
        log "Enabling Varnish for domain"
        sed -i -e '/# Handle HTTPS traffic (port 443)/,+6 s/^/#/' -e '/# Terminate TLS and pass to Varnish/,+3 s/^#//' "$domains_file"
    fi

    # Reload / start Caddy
    if docker --context default ps -q -f name=caddy | grep -q .; then
        log "Reloading Caddy"
        docker --context=default exec caddy caddy reload --config /etc/caddy/Caddyfile >/dev/null 2>&1 && log "Caddy reloaded successfully" || log "Caddy reload failed — check config"
    else
        log "Caddy not running, starting in background"
        nohup sh -c "cd /root && docker --context default compose up -d caddy" </dev/null >nohup.out 2>nohup.err &
		disown
    fi
}

# ======================================================================
# DNS

create_zone_file() {
    mkdir -p "$ZONE_FILE_DIR"

    if $USE_PARENT_DNS_ZONE; then
        log "Appending records to apex zone: $ZONE_FILE_DIR${apex_domain}.zone"
        local spf_prefix="ip4"
        [[ "$IPV4" == "no" ]] && spf_prefix="ip6"
        {
            echo "$domain_name    14400     IN      A       $current_ip"
            echo "$domain_name    14400     IN      TXT     'v=spf1 ${spf_prefix}:${current_ip} +a +mx ~all'"
        } >> "$ZONE_FILE_DIR${apex_domain}.zone"
        return
    fi

    local zone_template_path
    if [[ "$IPV4" == "yes" ]]; then
        zone_template_path='/etc/openpanel/bind9/zone_template.txt'
        log "Creating zone with A records: $ZONE_FILE_DIR${domain_name}.zone"
    else
        zone_template_path='/etc/openpanel/bind9/zone_template_ipv6.txt'
        log "Creating zone with AAAA records: $ZONE_FILE_DIR${domain_name}.zone"
    fi

    local timestamp
    ns1=${ns1:-ns1.openpanel.org}
    ns2=${ns2:-ns2.openpanel.org}
    rpemail=${rpemail:-"root.${domain_name}"}
    rpemail="${rpemail//@/.}"
    timestamp=$(date +"%Y%m%d")

    local zone_template
    zone_template=$(<"$zone_template_path")

    local zone_content
    if [[ -z "$ns3" ]]; then
        zone_content=$(sed \
            -e "s|{domain}|$domain_name|g" \
            -e "s|{ns1}|$ns1|g" \
            -e "s|{ns2}|$ns2|g" \
            -e "s|{rpemail}|$rpemail|g" \
            -e "s|{server_ip}|$current_ip|g" \
            -e "s|YYYYMMDD|$timestamp|g" \
            <<< "$zone_template")
    else
        zone_content=$(sed \
            -e "s|{domain}|$domain_name|g" \
            -e "s|{ns1}|$ns1|g" \
            -e "s|{ns2}|$ns2|g" \
            -e "s|{ns3}|$ns3|g" \
            -e "s|{ns4}|$ns4|g" \
            -e "s|{rpemail}|$rpemail|g" \
            -e "s|{server_ip}|$current_ip|g" \
            -e "s|YYYYMMDD|$timestamp|g" \
            <<< "$zone_template")
    fi

    echo "$zone_content" > "$ZONE_FILE_DIR${domain_name}.zone"
}

update_named_conf() {
    if $USE_PARENT_DNS_ZONE; then
        grep -q "zone \"$apex_domain\"" "$NAMED_CONF_LOCAL" && return
        echo "zone \"$apex_domain\" IN { type master; file \"${ZONE_FILE_DIR}${domain_name}.zone\"; };" >> "$NAMED_CONF_LOCAL"
    else
        log "Adding zone to named.conf.local"
        if grep -q "zone \"$domain_name\"" "$NAMED_CONF_LOCAL"; then
            log "Zone '$domain_name' already in $NAMED_CONF_LOCAL"
            return
        fi
        echo "zone \"$domain_name\" IN { type master; file \"${ZONE_FILE_DIR}${domain_name}.zone\"; };" >> "$NAMED_CONF_LOCAL"
    fi
}

reload_bind() {
    if docker --context default ps -q -f name=openpanel_dns | grep -q .; then
        log "Reloading BIND"
        docker --context default exec openpanel_dns rndc reconfig >/dev/null 2>&1
    else
        log "BIND not running, starting in background"
        nohup sh -c "cd /root && docker --context default compose up -d bind9" </dev/null >nohup.out 2>nohup.err &
		disown
    fi
}

notify_slave() {
    $USE_PARENT_DNS_ZONE && return
    log "Notifying slave $SLAVE_IP for zone $domain_name"

    # Prefer rndc-based notification over root SSH
	# https://www.ibm.com/docs/en/aix/7.2.0?topic=security-transaction-signatures-bind-version-94
	if docker --context=default exec openpanel_dns rndc -s "$SLAVE_IP" -p 953 -k /etc/bind/rndc.key status 2>/dev/null | grep -q 'number of zones'; then
        log "Slave reachable via rndc — adding zone $domain_name"
        if docker --context=default exec openpanel_dns rndc -s "$SLAVE_IP" -p 953 -k /etc/bind/rndc.key addzone "$domain_name { type slave; masters { $MASTER_IP; }; file \"/etc/bind/zones/${domain_name}.zone\"; allow-notify { $MASTER_IP; }; };" 2>/dev/null; then
            log "Zone $domain_name successfully added to slave $SLAVE_IP"
            return
        else
            log "rndc addzone failed, falling back to SSH"
        fi
    else
        log "Slave not reachable via rndc, falling back to SSH"
    fi

    # SSH fallback for <1.7.61
    ssh -q \
		-o BatchMode=yes \
        -o LogLevel=ERROR \
        -o ConnectTimeout=5 \
        -o StrictHostKeyChecking=no \
        -o UserKnownHostsFile=/dev/null \
        "root@$SLAVE_IP" \
        "bash -s" <<SSHEOF
if ! grep -q "${domain_name}.zone" /etc/bind/named.conf.local; then
    cat >> /etc/bind/named.conf.local <<ZONE
zone "${domain_name}" { type slave; masters { $MASTER_IP; }; file "/etc/bind/zones/${domain_name}.zone"; allow-notify { $MASTER_IP; }; };
ZONE
    mkdir -p /etc/bind/zones/
    touch /etc/bind/zones/${domain_name}.zone
fi
if docker --context default ps -q -f name=openpanel_dns >/dev/null; then
    docker --context default exec openpanel_dns rndc reconfig
else
    cd /root && docker --context default compose up -d bind9
fi
SSHEOF
}

get_slave_dns_option() {
    # Both allow-transfer and also-notify must be uncommented
    grep -qP '^(?!\s*//).*allow-transfer\s+\{[^}]*\}' "$BIND_CONFIG_FILE" 2>/dev/null || return 0
    grep -qP '^(?!\s*//).*also-notify\s+\{[^}]*\}'    "$BIND_CONFIG_FILE" 2>/dev/null || return 0

    local ALLOW_TRANSFER ALSO_NOTIFY
    ALLOW_TRANSFER=$(grep -oP '^(?!\s*//).*allow-transfer\s+\{\s*\K[^}]*' "$BIND_CONFIG_FILE" | tr -d '[:space:]')
    ALSO_NOTIFY=$(grep -oP '^(?!\s*//).*also-notify\s+\{\s*\K[^}]*' "$BIND_CONFIG_FILE" | tr -d '[:space:]')

    local -a ALLOW_TRANSFER_IPS ALSO_NOTIFY_IPS
    IFS=';' read -r -a ALLOW_TRANSFER_IPS <<< "$ALLOW_TRANSFER"
    IFS=';' read -r -a ALSO_NOTIFY_IPS    <<< "$ALSO_NOTIFY"

    for ip in "${ALLOW_TRANSFER_IPS[@]}"; do
        [[ -z "$ip" ]] && continue
        if [[ " ${ALSO_NOTIFY_IPS[*]} " == *" $ip "* ]]; then
            SLAVE_IP="$ip"
            MASTER_IP="$current_ip"
            notify_slave
        fi
    done
}

run_dns_setup() {
    $SKIP_DNS_ZONE && { log "Skipping DNS zone creation (--skip_dns)"; return; }
    is_module_enabled "dns" || { log "DNS module disabled — skipping"; return; }

    create_zone_file
    get_slave_dns_option   # must happen before named.conf update on master
    update_named_conf
    reload_bind
}

# ======================================================================
# Tor / .onion

verify_onion_files() {
    [[ "$onion_domain" == true ]] || return 0
    [[ -n "$hs_ed25519_public_key" && -n "$hs_ed25519_secret_key" ]] || die "Both --hs_ed25519_public_key and --hs_ed25519_secret_key are required for .onion domains."

    for key in hs_ed25519_public_key hs_ed25519_secret_key; do
        [[ "${!key}" =~ ^/var/www/html/ ]] || die "--$key must be inside /var/www/html/"
    done

    hs_public_key="/home/$context/docker-data/volumes/${context}_html_data/_data/${hs_ed25519_public_key#/var/www/html/}"
    hs_secret_key="/home/$context/docker-data/volumes/${context}_html_data/_data/${hs_ed25519_secret_key#/var/www/html/}"

    [[ -f "$hs_public_key" && -f "$hs_secret_key" ]] || die "hs_ed25519_public_key or hs_ed25519_secret_key files not found"
}

setup_tor_for_user() {
    local tor_dir="/home/$context/tor"
    local folder_name

    if [[ ! -d "$tor_dir/hidden_service" || ! -f "$tor_dir/torrc" ]]; then
        folder_name="hidden_service"
    else
        local highest=0 folder folder_num next_num
        for folder in "$tor_dir/hidden_service"/*/; do
            folder_num=$(basename "$folder")
            [[ "$folder_num" =~ ^[0-9]+$ ]] && (( folder_num > highest )) && highest=$folder_num
        done
        next_num=$(( highest + 1 ))
        folder_name="hidden_service/$next_num"
    fi

    mkdir -p "$tor_dir/$folder_name/authorized_clients"
    cp "$hs_public_key" "$tor_dir/$folder_name/hs_ed25519_public_key"
    cp "$hs_secret_key" "$tor_dir/$folder_name/hs_ed25519_secret_key"

    local uid
    uid=$(stat -c '%u' "/home/$context" 2>/dev/null)
    [[ -n "$uid" ]] && chown "$uid:$uid" "$tor_dir" || true
    chmod 0600 "$tor_dir/torrc"

    get_varnish_for_user
    local proxy_ws
    [[ "$VARNISH" == true ]] && proxy_ws="varnish" || proxy_ws="$WEB_SERVER"

    printf 'HiddenServiceDir /var/lib/tor/%s/\nHiddenServicePort 80 %s:80\n\n' "$folder_name" "$proxy_ws" >> "$tor_dir/torrc"

    log ".onion files saved in $folder_name"
}

start_tor_for_user() {
    if docker --context "$context" ps -q -f name=tor | grep -q .; then
        log "Tor running — restarting to apply new config"
        nohup sh -c "cd /home/$context/ && docker --context $context restart tor" </dev/null >nohup.out 2>nohup.err &
		disown
    else
        log "Starting Tor"
        nohup sh -c "cd /home/$context/ && docker --context $context compose up -d tor" </dev/null >nohup.out 2>nohup.err &
		disown
    fi
}

# ======================================================================
# PHP / Mail / Enterprise

start_php_fpm() {
    $SKIP_STARTING_CONTAINERS && return
    [[ "$WEB_SERVER" == *litespeed* ]] && return
    is_module_enabled "php" || { log "PHP module disabled — skipping"; return; }
    log "Starting PHP-FPM $php_version"
    nohup sh -c "docker --context $context compose -f /home/$context/docker-compose.yml up -d php-fpm-${php_version}" </dev/null >nohup.out 2>nohup.err &
	disown
}

postfwd_setup() {
    [[ -n "$ENTERPRISE" ]] || return
    nohup opencli email-ratelimit --domain="$domain_name" >/dev/null 2>&1 &
    disown
}

generate_dkim() {
    [[ -n "$ENTERPRISE" ]] || return
    local compose_file="/usr/local/mail/openmail/compose.yml"
    [[ -f "$compose_file" ]] || return
    log "Generating DKIM for the domain"
    opencli email-setup config dkim domain "$domain_name" >/dev/null 2>&1

    if $USE_PARENT_DNS_ZONE; then
        zone_file="$ZONE_FILE_DIR${apex_domain}.zone"
	else
		zone_file="$ZONE_FILE_DIR${domain_name}.zone"
	fi

	dkim_file="/usr/local/mail/openmail/docker-data/dms/config/opendkim/keys/${domain_name}/mail.txt"
	if [[ -f "$dkim_file" ]] && ! grep -q "mail\._domainkey" "$zone_file"; then
	    printf '\n' >> "$zone_file"
	    sed -E "s/^mail\._domainkey[[:space:]]+IN/mail._domainkey\t14400\tIN/" "$dkim_file" >> "$zone_file"
	    printf '\n' >> "$zone_file"
	    log "DKIM was successfully generated and added in the local DNS zone for domain"
	else
	    log "DKIM was not configured: generation failed or the DNS zone already includes a mail._domainkey record"
	fi
}

create_mail_mountpoint() {
    [[ -n "$ENTERPRISE" ]] || return

    local compose_file="/usr/local/mail/openmail/compose.yml"
    [[ -f "$compose_file" ]] || return

    local store_in
    store_in=$(grep -E '^email_storage_location=' /etc/openpanel/openadmin/config/admin.ini 2>/dev/null | cut -d'=' -f2- | xargs)

    local volume_to_add
    if [[ "$store_in" == /* ]]; then
        log "Using $store_in for email storage"
        mkdir -p "$store_in"
        volume_to_add="      - $store_in:/var/mail/"
    else
        local domain_dir="/home/$context/mail/$domain_name/"
        log "Creating $domain_dir for emails"
        mkdir -p "$domain_dir"
        volume_to_add="      - $domain_dir:/var/mail/$domain_name/"
    fi

    if grep -qF "$volume_to_add" "$compose_file"; then
        return
    fi

    sed -i "/^  mailserver:/,/^  roundcube:/ {
        /^    volumes:/a\\
$volume_to_add
    }" "$compose_file"

    log "Reconfiguring mailserver in background"
    nohup sh -c "cd /usr/local/mail/openmail/ && docker-compose up -d --force-recreate mailserver" </dev/null >nohup.out 2>nohup.err &
	disown
}

# ======================================================================
# Database insert

insert_domain_to_db() {
    log "Inserting domain $domain_name into database"

    local safe_docroot safe_php safe_domain
    safe_docroot=$(sql_escape "$docroot")
    safe_php=$(sql_escape "$php_version")
    safe_domain=$(sql_escape "$domain_name")
    [[ "$user_id" =~ ^[0-9]+$ ]] || die "Invalid user_id: $user_id"

    mysql -e "INSERT INTO domains (user_id, docroot, php_version, domain_url) VALUES ('$user_id', '$safe_docroot', '$safe_php', '$safe_domain');" 2>/dev/null || die "Database insert failed for $domain_name"

    local count
    count=$(mysql -sN -e "SELECT COUNT(*) FROM domains WHERE user_id='$user_id' AND docroot='$safe_docroot' AND domain_url='$safe_domain';" 2>/dev/null)
    [[ "$count" -eq 1 ]] || die "Insert verification failed for $domain_name"
}

run_parallel_async() {
    local pids=()

    _bg() {
        local label="$1"; shift
        ( "$@" 2>&1 | while IFS= read -r line; do log "[$label] $line"; done ) &
        pids+=($!)
    }

    if [[ "$onion_domain" == true ]]; then
        _bg "tor-setup"  setup_tor_for_user
        _bg "tor-start"  start_tor_for_user
    else
        _bg "caddy"  create_caddy_domain_file
        _bg "dns"    run_dns_setup
    fi

    _bg "vhost"   create_vhost_file
	_bg "FTP permissions"   fix_ftp_perms
    _bg "php-fpm" start_php_fpm

    if [[ "$onion_domain" == false ]]; then
        _bg "mail"     create_mail_mountpoint
		_bg "dkim"     generate_dkim
        _bg "postfwd"  postfwd_setup
    fi

    local failed=0
    for pid in "${pids[@]}"; do
        wait "$pid" || (( failed++ )) || true
    done

    (( failed > 0 )) && warn "$failed background task(s) reported errors — check logs"
}

notify_sentinel() {
    nohup opencli sentinel --action=domains_create --title="Domain added" --message="Domain name: '$domain_name' has been added to OpenPanel user: '$user'." >/dev/null 2>&1 &
    disown
}

# Main
main() {
    parse_args "$@"

    validate_domain_format
    detect_apex_or_onion
	read_config_file
    check_domain_not_taken
    check_forbidden_domains
    check_system_domains
    check_apex_ownership
    get_docker_context
	check_domains_limit
    get_php_version
    validate_docroot
    verify_onion_files

    # Blocking: need IP + web server before we can write files
    get_server_ip
    get_webserver_for_user

    # DB insert (fast, blocking — everything else depends on it succeeding)
    insert_domain_to_db

    # Filesystem (fast, blocking — vhost/caddy/dns need this)
    make_docroot_dirs

    # Fire remaining tasks in parallel
    run_parallel_async

    notify_sentinel
    echo "Domain $domain_name added successfully"
}

main "$@"
