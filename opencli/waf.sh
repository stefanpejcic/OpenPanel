#!/bin/bash
################################################################################
# Script Name: waf.sh
# Description: Manage CorazaWAF
# Usage: opencli waf <setting> 
# Author: Stefan Pejcic
# Created: 22.05.2025
# Last Modified: 03.02.2026
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
# Helpers

usage() {
    echo "Usage: opencli waf <command> [options]"
    echo ""
    echo "Commands:"
    echo "  status                                       Check if CorazaWAF is enabled for new domains and users."
    echo "  enable                                       Use Caddy image with CorazaWAF, enable module and use WAF on new domains."
    echo "  disable                                      Use official Caddy docker image, disable module and dont use WAF on new domains."    
    echo "  domain                                       Check if CorazaWAF is enabled for a domain."
    echo "  domain DOMAIN_NAME enable                    Enable CorazaWAF for a domain."
    echo "  domain DOMAIN_NAME disable                   Disable CorazaWAF for a domain."
    echo "  tags                                         Display all tags from enabled sets."
    echo "  ids                                          Display all rule IDs from enabled sets."
    echo "  update                                       Update OWASP CRS."
    echo "  stats <country|agent|hourly|ip|request|path> Display top requests by country, ip, path, etc."
    echo ""
    echo "Examples:"
    echo "  opencli waf status"
    echo "  opencli waf enable"
    echo "  opencli waf disable"
    echo "  opencli waf domain pcx3.com"
    echo "  opencli waf domain pcx3.com enable"
    echo "  opencli waf domain pcx3.com disable"
    echo "  opencli waf stats ip"
    echo "  opencli waf stats hourly"
    exit 1
}

check_domain() {
    local domain="$1"
    local file="/etc/openpanel/caddy/domains/${domain}.conf"
    
    if [[ ! -f "$file" ]]; then
        echo "Domain not found!"
        exit 1
    fi

    if grep -iq '^[[:space:]]*SecRuleEngine[[:space:]]\+On' "$file"; then
        echo "SecRuleEngine is set to On for domain $domain"
    elif grep -iq '^[[:space:]]*SecRuleEngine[[:space:]]\+Off' "$file"; then
        echo "SecRuleEngine is set to Off for domain $domain"
    else
        echo "SecRuleEngine is not set for domain $domain"
    fi
}

check_coraza_status() {
  local env_file="/root/.env"
  local custom_image='CADDY_IMAGE="openpanel/caddy-coraza"'
  local openpanel_config="/etc/openpanel/openpanel/conf/openpanel.config"

  local image_enabled=false
  local waf_enabled=false

  # Check image
  if grep -q "^$custom_image" "$env_file" 2>/dev/null; then
    image_enabled=true
  fi

  # Check waf module
  if grep -qi "waf" "$openpanel_config" 2>/dev/null; then
    waf_enabled=true
  fi

  # Report
  if $image_enabled && $waf_enabled; then
    echo "CorazaWAF is ENABLED"
  elif $image_enabled; then
    echo "CorazaWAF is DISABLED: 'waf' module is not enabled in OpenAdmin > Settings > Modules.)"
  elif $waf_enabled; then
    echo "CorazaWAF is DISABLED: 'waf' module is enabled, but Caddy is not using $custom_image"
  else
    echo "CorazaWAF is DISABLED"
  fi
}

reload_caddy_now() {
    docker --context=default exec caddy caddy reload --config /etc/caddy/Caddyfile > /dev/null 2>&1
}

set_coraza_waf_for_domain() {
    local domain="$1"
    local action="$2"   # "enable" or "disable"
    local file="/etc/openpanel/caddy/domains/${domain}.conf"
    local value

    if [[ ! -f "$file" ]]; then
        echo "Domain not found!"
        exit 1
    fi

    case "$action" in
        enable) value="On" ;;
        disable) value="Off" ;;
        *)
            echo "Invalid action: $action. Use 'enable' or 'disable'."
            exit 1
            ;;
    esac

    if sed -i "s/SecRuleEngine .*/SecRuleEngine $value/" "$file"; then
        reload_caddy_now
        echo "SecRuleEngine $value is now set for domain $domain"
    else
        echo "Failed setting SecRuleEngine $value - please contact Administrator."
        exit 1
    fi
}

get_stats_from_file() {
    local log_file="/var/log/caddy/coraza_audit.log"
    local mode="$1"
    
    if [ -f "$log_file" ]; then
        case "$mode" in
            country)
                grep -i 'cf-ipcountry:' "$log_file" | cut -d':' -f2 | tr -d ' ' | sort | uniq -c | sort -nr
                ;;
            agent)
                grep -i 'user-agent:' "$log_file" | cut -d':' -f2- | sort | uniq -c | sort -nr | head
                ;;
            hourly)
                grep -oP '\[\K[0-9]{4}/[0-9]{2}/[0-9]{2} [0-9]{2}' "$log_file" | sort | uniq -c
                ;;
            ip)
                grep -oP '\[.*\] \S+ \K[\d.]+' "$log_file" | sort | uniq -c | sort -nr | head
                ;;
            request)
                grep -oP '^(GET|POST|HEAD|PUT|DELETE|OPTIONS) .+' "$log_file" | sort | uniq -c | sort -nr | head
                ;;
            path)
                grep -oP 'GET\s+\K\S+' "$log_file" | sort | uniq -c | sort -nr | head
                ;;
            *)
                usage
                ;;
        esac
    fi
}

list_all_tags() {
    grep -oP "tag:\s*['\"]\K[^'\"]+" /etc/openpanel/caddy/coreruleset/rules/*.conf | grep -v "OWASP_CRS" | sort -u
}

list_all_ids() {
    grep -oP "id:\K[0-9]+" /etc/openpanel/caddy/coreruleset/rules/*.conf | sort -u
}

get_count_from_file() {
    local log_file="/var/log/caddy/coraza_audit.log"
    local record_count
    if [ -f "$log_file" ]; then
        record_count=$(grep -cE '^--.*-B--$' "$log_file")
        echo "$record_count"
    fi
}

update_owasp_rules() {
  cd /etc/openpanel/caddy/coreruleset/ || { echo "Failed to enter modsec directory: /etc/openpanel/caddy/coreruleset/"; return 1; }
  
  for f in rules/*.conf.disabled; do
    original="${f%.disabled}"
    echo "- Excluding disabled ruleset $original"
    git update-index --assume-unchanged "$original"
  done

  echo "Updating OWASP CRS.."
  
  if git pull --quiet; then
    echo "Update successful."
  else
    echo "Update failed."
    return 1
  fi
}



enable_coraza_waf() {

    # 1. download rules
    echo "Downloading Coraza rules.."
    wget --timeout=15 --tries=3 --inet4-only https://raw.githubusercontent.com/corazawaf/coraza/v3/dev/coraza.conf-recommended -O /etc/openpanel/caddy/coraza_rules.conf
    echo "Downloading OWASP CRS.."
    git clone https://github.com/coreruleset/coreruleset /etc/openpanel/caddy/coreruleset/
    
    # 2. enable module
    echo "Enabling WAF module.."
    sed -i '/^enabled_modules=/ { /waf/! s/$/,waf/ }' /etc/openpanel/openpanel/conf/openpanel.config

    # 3. set image to openpanel/caddy-coraza
    echo "Setting docker image 'openpanel/caddy-coraza'.."
    sed -i 's|^CADDY_IMAGE=".*"|CADDY_IMAGE="openpanel/caddy-coraza"|' /root/.env

    # 4. enable WAF for ALL user domains
    sed -i 's/SecRuleEngine Off/SecRuleEngine On/g' /etc/openpanel/caddy/domains/*.conf

    # 5. restart caddy
    echo "Restarting Web Server to use the new image with CorazaWAF.."
    cd /root || {
        echo "Failed to cd to /root - are you root?" >&2
        exit 1
    }
    docker --context=default compose down caddy && docker --context=default compose up -d caddy

    # 6. check status
    check_coraza_status
}

disable_coraza_waf() {
    # 1. check if used and ask for confirmation
    echo "Checking if CorazaWAF is used by any user domains.."
    local conf_files
    conf_files=$(grep -rl "SecRuleEngine On" /etc/openpanel/caddy/domains/*.conf 2>/dev/null)

    if [[ -n "$conf_files" ]]; then
        echo "WARNING: WAF is still active on some domains:"
        for file in $conf_files; do
            filename=$(basename "$file" .conf)
            echo " - $filename"
        done

        read -r -p "Do you really want to disable WAF protection on these domains and stop using Coraza WAF on the server? [y/N]: " confirm
        case "$confirm" in
            [yY][eE][sS]|[yY])
                echo "Disabling Coraza WAF..."
                ;;
            *)
                echo "Aborting."
                return 1
                ;;
        esac
    fi

    # 2. disable module
    echo "Disabling WAF module..."
    sed -i 's/waf,//g' /etc/openpanel/openpanel/conf/openpanel.config

    # 3. disable WAF for ALL user domains
    sed -i 's/SecRuleEngine On/SecRuleEngine Off/g' /etc/openpanel/caddy/domains/*.conf

    # 4. reload caddy to apply
    reload_caddy_now

    # 5. check status
    check_coraza_status
}



# ======================================================================
# Main
case "$1" in
    "status")
        check_coraza_status
        ;;
    "domain")
        if [[ -z "$2" ]]; then
            echo "Domain name is required."
            usage
        fi
        case "$3" in
                enable|disable)
                    set_coraza_waf_for_domain "$2" "$3"
                    ;;
            "")
                check_domain "$2"
                ;;
            *)
                echo "Invalid action for domain: $3"
                usage
                ;;
        esac
        ;;
    "enable")
        enable_coraza_waf
        exit 0
        ;;
    "disable")
        disable_coraza_waf
        exit 0
        ;;
    "update")
        if [[ -z "$2" ]]; then    
            update_owasp_rules
        else
            case "$2" in
                "log")
                    cd /etc/openpanel/caddy/coreruleset/ && git log --oneline
                    ;;
                *)
                    echo "Invalid action, available: opencli waf-update and opencli waf-update log"
                    exit 1
                    ;;
            esac
        fi      
        ;;        
    "stats")
        get_stats_from_file "$2"
        ;;
    "count")
        get_count_from_file
        ;;
    "tags")
        list_all_tags
        ;;
    "ids")
        list_all_ids
        ;;
    "help")
        usage
        ;;
    *)
        echo "Invalid option: $1"
        usage
        ;;
esac

exit 0
