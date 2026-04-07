#!/bin/bash
################################################################################
# Script Name: user/block_ip.sh
# Description: Block IP addresses from accessng user domains.
# Usage: opencli user-block_ip <username> [--list='ip_here another_ip' | --delete-all]
# Author: Stefan Pejcic
# Created: 13.03.2026
# Last Modified: 06.04.2026
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

# Usage
if [ $# -eq 0 ] || [ $# -gt 2 ]; then
    echo "Usage: opencli user-block_ip <username> [--list='IP_ADDRESS ANOTHER_IP' | --delete-all]"
    echo ""
    echo "Examples:" 
    echo "opencli user-block_ip <username>"
    echo "opencli user-block_ip <username> --list='11.22.33.44 124.64.23.0/24'"
    echo "opencli user-block_ip <username> --delete-all"   
    exit 1
fi

readonly CADDY_VHOST_DIR="/etc/openpanel/caddy/domains"
delete_all=false
list=""
show_ips=false

# Parse
USERNAME="$1"

if [ -z "$2" ]; then
    show_ips=true
else
    case "$2" in
        --list=*) list="${2#*=}" ;;
        --delete-all) delete_all=true ;;
        *) echo "Unknown argument: $2" ;;
    esac
fi

# HELPERS

get_docker_context() {
    local query="SELECT id, server FROM users WHERE username = '${USERNAME}';"
    user_info=$(mysql -se "$query")
    user_id=$(echo "$user_info" | awk '{print $1}')
    context=$(echo "$user_info" | awk '{print $2}')
    
    if [ -z "$user_id" ]; then
        echo "FATAL ERROR: user $USERNAME does not exist."
        exit 1
    fi
    if [ -z "$context" ]; then
        echo "FATAL ERROR: docker context for user $USERNAME does not exist."
        exit 1
    fi
    DENY_IPS_FILE="/etc/openpanel/caddy/deny/$context.ips"
}

is_valid_ip_or_cidr() {
    local ip="$1"
    if [[ "$ip" =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}(\/([0-9]|[1-2][0-9]|3[0-2]))?$ ]]; then
        IFS='/'; read -r ip_only cidr <<< "$ip"
        IFS='.' read -r o1 o2 o3 o4 <<< "$ip_only"
        for octet in $o1 $o2 $o3 $o4; do
            if ((octet < 0 || octet > 255)); then
                return 1
            fi
        done
        return 0
    fi
    return 1
}

reload_caddy() {
    nohup docker --context=default exec caddy sh -c "caddy validate && caddy reload" >/dev/null 2>&1 &
    disown
}


# MAIN
get_docker_context

# DELETE
if [ "$delete_all" = true ]; then
  > "$DENY_IPS_FILE"
  reload_caddy
  exit 0
fi

# LIST
if [ "$delete_all" = true ]; then
  if [ -f "$DENY_IPS_FILE" ]; then
    grep -oP 'remote_ip\s+\K[\d./]+' "$DENY_IPS_FILE"
  fi
  exit 0
fi

# UPDATE
if [ -n "$list" ]; then

  # 1. validate IP
  IFS=' ' read -r -a ip_array <<< "$list"
  valid_ips=()
  
  for ip in "${ip_array[@]}"; do
      if is_valid_ip_or_cidr "$ip"; then
          echo "Blocking IP $ip for user $USERNAME..."
          valid_ips+=("$ip")
      else
          echo "Warning: Invalid IP or CIDR skipped: $ip"
      fi
  done

  # 2. create caddy include file 
  mkdir -p "$(dirname "$DENY_IPS_FILE")" # <1.7.47
  {
      echo "@blocked {"
      for ip in "${valid_ips[@]}"; do
          echo "    remote_ip $ip"
      done
      echo "}"
      echo
      echo 'respond @blocked "Access Denied" 403'
  } > "$DENY_IPS_FILE"

  # 3. include in all domains for user
  domain_list=$(opencli domains-user "$USERNAME")
  domain_regex='\.'
  user_has_domains=false

  while IFS= read -r domain_name; do
      if [[ ! $domain_name =~ $domain_regex ]]; then
          continue
      fi
      domain_file="${CADDY_VHOST_DIR}/${domain_name}.conf"
      if [[ -f "$domain_file" ]]; then
        user_has_domains=true
        if ! grep -q "import $DENY_IPS_FILE" "$domain_file"; then
            sed -i "/route {/ {a\\
    # ip-blocker\\
    import $DENY_IPS_FILE
            }" "$domain_file"           
            echo "Blocklist applied for domain '$domain_name'"
        else
            echo "Blocklist already applied to domain '$domain_name' - skipping."
        fi
      else
        echo "Warning: No caddy file for domain $domain_name"
      fi
  done <<< "$domain_list"
  # 4. reload caddy
  if [ "$user_has_domains" = true ]; then
    reload_caddy
  fi
fi
