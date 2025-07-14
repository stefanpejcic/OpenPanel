#!/bin/bash
################################################################################
# Script Name: domain.sh
# Description: View and set domain/ip for accessing panels.
# Usage: opencli domain [set <domain_name> | ip] [--debug]
# Author: Stefan Pejcic
# Created: 09.02.2025
# Last Modified: 13.07.2025
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

# Configuration
readonly CADDY_FILE="/etc/openpanel/caddy/Caddyfile"
readonly DOMAINS_DIR="/etc/openpanel/caddy/domains/"
readonly MAILSERVER_ENV="/usr/local/mail/openmail/mailserver.env"
readonly ROUNDCUBE_COMPOSE="/usr/local/mail/openmail/compose.yml"
readonly REDIRECTS_CONF="/etc/openpanel/caddy/redirects.conf"
readonly DEFAULT_DOMAIN="example.net"

# Global variables
DEBUG=false
current_domain=""
new_hostname=""
server_ip=""

# IP check servers
readonly IP_SERVERS=(
    "https://ip.openpanel.com"
    "https://ipv4.openpanel.com"
    "https://ifconfig.me"
)

# Utility functions
log_debug() {
    if [[ "$DEBUG" == true ]]; then
        echo "$1"
    fi
}

log_error() {
    echo "Error: $1" >&2
}

log_info() {
    echo "$1"
}

# IPv4 validation
is_valid_ipv4() {
    local ip="$1"
    
    # Check if it's a valid IPv4 format
    [[ $ip =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}$ ]] || return 1
    
    # Check if it's not a private IP
    [[ $ip =~ ^10\. ]] && return 1
    [[ $ip =~ ^172\.(1[6-9]|2[0-9]|3[0-1])\. ]] && return 1
    [[ $ip =~ ^192\.168\. ]] && return 1
    
    return 0
}

# Get server's public IPv4 address
get_server_ipv4() {
    local current_ip=""
    
    # Try each IP service
    for service in "${IP_SERVERS[@]}"; do
        # Try curl first
        current_ip=$(curl --silent --max-time 2 -4 "$service" 2>/dev/null)
        if is_valid_ipv4 "$current_ip"; then
            echo "$current_ip"
            return 0
        fi
        
        # Try wget as fallback
        current_ip=$(wget --timeout=2 -qO- "$service" 2>/dev/null)
        if is_valid_ipv4 "$current_ip"; then
            echo "$current_ip"
            return 0
        fi
    done
    
    # Fallback to local IP detection
    current_ip=$(ip addr | grep 'inet ' | grep global | head -n1 | awk '{print $2}' | cut -f1 -d/)
    
    if is_valid_ipv4 "$current_ip"; then
        echo "$current_ip"
        return 0
    fi
    
    log_error "Invalid or private IPv4 address. Please contact support."
    return 1
}

# Domain validation
is_valid_domain() {
    local domain="$1"
    [[ "$domain" =~ ^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$ ]]
}

# Get current domain from Caddyfile
get_current_domain() {
    if [[ ! -f "$CADDY_FILE" ]]; then
        log_error "Caddyfile not found at $CADDY_FILE"
        return 1
    fi
    
    grep -oP '^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}' "$CADDY_FILE" | head -n 1
}

# Get current domain or IP
get_current_domain_or_ip() {
    local domain
    domain=$(get_current_domain)
    
    if [[ "$domain" == "$DEFAULT_DOMAIN" ]]; then
        get_server_ipv4
    else
        echo "$domain"
    fi
}

# Update Caddyfile with new domain
update_caddyfile() {
    log_debug "Updating Caddyfile with new domain: $new_hostname"
    
    if [[ ! -f "$CADDY_FILE" ]]; then
        log_error "Caddyfile not found at $CADDY_FILE"
        return 1
    fi
    
    sed -i "s/$current_domain/$new_hostname/g" "$CADDY_FILE"
}

# Create or move domain configuration file
create_mv_file() {
    log_debug "Managing domain configuration file"
    
    mkdir -p "$DOMAINS_DIR"
    
    if [[ "$new_hostname" == "$DEFAULT_DOMAIN" ]]; then
        # Remove existing domain file when switching to IP
        [[ -f "${DOMAINS_DIR}${current_domain}.conf" ]] && rm "${DOMAINS_DIR}${current_domain}.conf"
    else
        # Move or create domain file
        if [[ -f "${DOMAINS_DIR}${current_domain}.conf" ]]; then
            mv "${DOMAINS_DIR}${current_domain}.conf" "${DOMAINS_DIR}${new_hostname}.conf" 2>/dev/null
        fi
        touch "${DOMAINS_DIR}${new_hostname}.conf"
    fi
}

# Update redirect configuration
update_redirects() {
    local redirect_url
    
    if [[ "$new_hostname" == "$DEFAULT_DOMAIN" ]]; then
        redirect_url="http://$server_ip"
        log_debug "Updating redirects to use IP address"
    else
        redirect_url="https://$new_hostname"
        log_debug "Updating redirects to use custom domain"
    fi
    
    if [[ ! -f "$REDIRECTS_CONF" ]]; then
        log_error "Redirects configuration not found at $REDIRECTS_CONF"
        return 1
    fi
    
    sed -i -E "s|(redir @openpanel )https?://[^:]+:|\1${redirect_url}:|; s|(redir @openadmin )https?://[^:]+:|\1${redirect_url}:|" "$REDIRECTS_CONF"
}

# Configure mailserver
configure_mailserver() {
    [[ ! -f "$MAILSERVER_ENV" ]] && return 0
    
    if [[ "$new_hostname" == "$DEFAULT_DOMAIN" ]]; then
        log_debug "Configuring mailserver for plain-text authentication with IP"
        
        sed -i '/^SSL_TYPE=/c\SSL_TYPE=' "$MAILSERVER_ENV"
        sed -i '/^SSL_CERT_PATH=/d' "$MAILSERVER_ENV"
        sed -i '/^SSL_KEY_PATH=/d' "$MAILSERVER_ENV"
    else
        log_debug "Configuring mailserver for TLS/SSL authentication with domain"
        
        local cert_path="/etc/letsencrypt/live/${new_hostname}/${new_hostname}.crt"
        local key_path="/etc/letsencrypt/live/${new_hostname}/${new_hostname}.key"
        
        sed -i "/^SSL_TYPE=/c\SSL_TYPE=manual" "$MAILSERVER_ENV"
        
        # Update or add SSL paths
        if grep -q '^SSL_CERT_PATH=' "$MAILSERVER_ENV"; then
            sed -i "s|^SSL_CERT_PATH=.*|SSL_CERT_PATH=$cert_path|" "$MAILSERVER_ENV"
        else
            echo "SSL_CERT_PATH=$cert_path" >> "$MAILSERVER_ENV"
        fi
        
        if grep -q '^SSL_KEY_PATH=' "$MAILSERVER_ENV"; then
            sed -i "s|^SSL_KEY_PATH=.*|SSL_KEY_PATH=$key_path|" "$MAILSERVER_ENV"
        else
            echo "SSL_KEY_PATH=$key_path" >> "$MAILSERVER_ENV"
        fi
    fi
    
    log_debug "Restarting mailserver to apply new configuration"
    nohup sh -c "cd /usr/local/mail/openmail/ && docker --context default restart openadmin_mailserver" </dev/null >nohup.out 2>nohup.err &
}

# Configure Roundcube
configure_roundcube() {
    [[ ! -f "$ROUNDCUBE_COMPOSE" ]] && return 0
    
    if [[ "$new_hostname" == "$DEFAULT_DOMAIN" ]]; then
        log_debug "Configuring Roundcube for plain-text authentication"
        
        sed -i 's|ROUNDCUBEMAIL_DEFAULT_HOST=.*|ROUNDCUBEMAIL_DEFAULT_HOST=openadmin_mailserver|' "$ROUNDCUBE_COMPOSE"
        sed -i 's|ROUNDCUBEMAIL_DEFAULT_PORT=.*|ROUNDCUBEMAIL_DEFAULT_PORT=|' "$ROUNDCUBE_COMPOSE"
        sed -i 's|ROUNDCUBEMAIL_SMTP_SERVER=.*|ROUNDCUBEMAIL_SMTP_SERVER=openadmin_mailserver|' "$ROUNDCUBE_COMPOSE"
        sed -i 's|ROUNDCUBEMAIL_SMTP_PORT=.*|ROUNDCUBEMAIL_SMTP_PORT=|' "$ROUNDCUBE_COMPOSE"
    else
        log_debug "Configuring Roundcube for TLS/SSL authentication"
        
        sed -i "s|ROUNDCUBEMAIL_DEFAULT_HOST=.*|ROUNDCUBEMAIL_DEFAULT_HOST=ssl://$new_hostname|" "$ROUNDCUBE_COMPOSE"
        sed -i "s|ROUNDCUBEMAIL_DEFAULT_PORT=.*|ROUNDCUBEMAIL_DEFAULT_PORT=993|" "$ROUNDCUBE_COMPOSE"
        sed -i "s|ROUNDCUBEMAIL_SMTP_SERVER=.*|ROUNDCUBEMAIL_SMTP_SERVER=ssl://$new_hostname|" "$ROUNDCUBE_COMPOSE"
        sed -i "s|ROUNDCUBEMAIL_SMTP_PORT=.*|ROUNDCUBEMAIL_SMTP_PORT=465|" "$ROUNDCUBE_COMPOSE"
    fi
    
    log_debug "Restarting Roundcube to apply new configuration"
    nohup sh -c "cd /usr/local/mail/openmail/ && docker --context default restart roundcube" </dev/null >nohup.out 2>nohup.err &
}

# Restart services
restart_services() {
    local no_restart=false
    
    # Check for --no-restart flag
    for arg in "$@"; do
        if [[ "$arg" == "--no-restart" ]]; then
            no_restart=true
            break
        fi
    done
    
    if [[ "$no_restart" == true ]]; then
        log_debug "Skipping service restart due to --no-restart flag"
        return 0
    fi
    
    log_debug "Restarting OpenPanel and Caddy containers"
    
    cd /root || return 1
    
    # Restart OpenPanel
    docker --context default compose restart openpanel >/dev/null 2>&1
    
    # Restart Caddy
    docker --context default compose down caddy >/dev/null 2>&1
    docker --context default compose up -d caddy >/dev/null 2>&1
}

# Display success message
show_success_message() {
    local display_hostname="$new_hostname"
    
    if [[ "$new_hostname" == "$DEFAULT_DOMAIN" ]]; then
        display_hostname="$server_ip"
    fi
    
    log_info "$display_hostname is now set for accessing the OpenPanel and OpenAdmin interfaces."
}

# Main domain update function
update_domain() {
    current_domain=$(get_current_domain) || return 1
    server_ip=$(get_server_ipv4) || return 1
    
    log_debug "Current domain: $current_domain"
    log_debug "Server IP: $server_ip"
    log_debug "New hostname: $new_hostname"
    
    # Execute all update steps
    update_caddyfile || return 1
    create_mv_file || return 1
    configure_mailserver || return 1
    configure_roundcube || return 1
    update_redirects || return 1
    restart_services "$@" || return 1
    show_success_message
}

# Display usage information
show_usage() {
    echo "Usage:"
    echo ""
    echo "opencli domain                        - displays current URL"
    echo "opencli domain set example.net        - set domain name for access"
    echo "opencli domain ip                     - set IP for access"
    echo ""
    echo "Options:"
    echo "  --debug                             - enable verbose output"
    echo "  --no-restart                        - skip service restart"
}

# Parse command line arguments
parse_arguments() {
    # Check for debug flag
    for arg in "$@"; do
        if [[ "$arg" == "--debug" ]]; then
            DEBUG=true
            log_debug "'--debug' flag provided: displaying verbose information."
            break
        fi
    done
}

# Main function
main() {
    parse_arguments "$@"
    
    case "${1:-}" in
        "")
            get_current_domain_or_ip
            ;;
        "set")
            if [[ -z "$2" ]]; then
                log_error "Domain name required for 'set' command"
                show_usage
                exit 1
            fi
            
            if ! is_valid_domain "$2"; then
                log_error "Invalid domain format. Please provide a valid domain."
                show_usage
                exit 1
            fi
            
            new_hostname="$2"
            log_debug "Setting domain to: $new_hostname"
            update_domain "$@"
            ;;
        "ip")
            new_hostname="$DEFAULT_DOMAIN"
            log_debug "Setting IP address for accessing panels"
            update_domain "$@"
            ;;
        *)
            show_usage
            exit 1
            ;;
    esac
}

# Script entry point
main "$@"
exit $?
