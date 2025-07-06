#!/bin/bash
################################################################################
# Script Name: domains/dns.sh
# Description: Manage DNS for a domain.
# Usage: opencli domains-dns <DOMAIN>
# Author: Stefan Pejcic
# Created: 31.08.2024
# Last Modified: 04.07.2025
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

# COLORS
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
RESET='\033[0m'

usage() {
  command="opencli domains-dns"
  echo "Usage:"
  echo -e " ${GREEN}$command reconfig${RESET}               - Load new DNS zones into bind server."
  echo -e " ${GREEN}$command check <DOMAIN>${RESET}         - Check and validate dns zone for a domain."
  echo -e " ${GREEN}$command reload <DOMAIN>${RESET}        - Reload DNS zone for a single domain."
  echo -e " ${GREEN}$command show <DOMAIN>${RESET}          - Display DNS zone for a single domain."
  echo -e " ${GREEN}$command list${RESET}                   - List all domains with DNS zones on the server."
  echo -e " ${GREEN}$command create <DOMAIN>${RESET}        - Create DNS zone for a domain."
  echo -e " ${GREEN}$command delete <DOMAIN>${RESET}        - Delete DNS zone for a domain."
  echo -e " ${GREEN}$command default <DOMAIN>${RESET}       - Restart default DNS zone for a domain."
  echo -e " ${GREEN}$command count${RESET}                  - Display total number of DNS zones present on the server."
  echo -e " ${GREEN}$command config${RESET}                 - Check main bind configuration file for syntax errros."
  echo -e " ${GREEN}$command start${RESET}                  - Start the DNS server."
  echo -e " ${GREEN}$command restart${RESET}                - Soft restart of bind9 docker container."
  echo -e " ${GREEN}$command hard-restart${RESET}           - Hard restart: terminates container and start again."
  echo -e " ${GREEN}$command stop${RESET}                   - Stop the DNS server."
  exit 1
}

# Check if at least one argument is provided
if [ "$#" -lt 1 ]; then
    usage
fi











######## START MAIN FUNCTION #######



reconfig_command(){
  echo "Loading new DNS zones.."
  docker --context default exec openpanel_dns rndc reconfig
  exit 0
}



check_named_main_conf(){
  echo "Checking /etc/bind/named.conf configuration:"
  docker --context default exec openpanel_dns named-checkconf  /etc/bind/named.conf
  exit 0
}



reload_one_or_all_dns_zone(){
  DOMAIN=$1
  if [[ -n "$DOMAIN" ]]; then
    echo "Reloading DNS zone for domain: $DOMAIN"
    docker --context default exec openpanel_dns rndc reload $DOMAIN
  else
    echo "Reloading all DNS zones.."
    docker --context default exec openpanel_dns rndc reload
  fi
  exit 0
}


list_all_zones(){
  echo "Displaying all DNS zones on the server:"
  find /etc/bind/zones/ -type f -name "*.zone"
  exit 0
}

list_dns_records(){
  DOMAIN=$1
  if [[ -n "$DOMAIN" ]]; then
    echo "DNS zone for domain: $DOMAIN - file: /etc/bind/zones/$DOMAIN.zone"
    cat /etc/bind/zones/$DOMAIN.zone
    exit 0
  else
    echo "ERROR: domain name is needed to list its zone."
    exit 1
  fi
}


show_count(){
  count=$(find /etc/bind/zones/ -type f -name "*.zone" | wc -l)
  echo -e "Total number of DNS zones on the server: ${GREEN}$count${RESET}"
  exit 0
}



delete_dns_zone(){
  DOMAIN=$1
  YES="$2"


  delete_zone_file(){
    # backup zone
    mkdir -p /etc/bind/zones/backups/
    mv /etc/bind/zones/$DOMAIN.zone /etc/bind/zones/backups/$DOMAIN.zone.$(date +%Y%m%d%H%M%S)
    #TODO: reload zones!
  }


  if [[ -n "$DOMAIN" ]]; then
    if [[ "$YES" == "-y" ]]; then
      echo "Deleting DNS zone for domain: $DOMAIN"
      delete_zone_file
      exit 0
    else
      # wait for confirmation
      read -t 10 -p "Are you sure you want to delete the existing DNS zone for domain: $DOMAIN ? (y/n): " CONFIRM
      if [[ $? -ne 0 ]]; then
        echo "Timed out."
        exit 1
      fi
      if [[ "$CONFIRM" == "y" || "$CONFIRM" == "Y" ]]; then
        echo "Deleting DNS zone for domain: $DOMAIN"
        delete_zone_file
        exit 0
      else
        echo "Canceled."
        exit 1
      fi
    fi
  else
    echo "Error: No domain provided."
    exit 1
  fi 
}











restore_zone_to_default(){
  DOMAIN=$1
  YES="$2"

  edit_zone_to_default(){
  domain_name="$1"
  
  # backup zone
  mkdir -p /etc/bind/zones/backups/
  mv /etc/bind/zones/$domain_name.zone /etc/bind/zones/backups/$domain_name.zone.$(date +%Y%m%d%H%M%S)
  
  # todo: add cron to crear zones daily, older than 24hrs.
  create_dns_zone_for_domain "$domain_name"

  }

  
  if [[ -n "$DOMAIN" ]]; then
    if [[ "$YES" == "-y" ]]; then
      echo "Deleting DNS zone for domain: $DOMAIN  and restoring default zone.."
      edit_zone_to_default "$DOMAIN"
      exit 0
    else
      # wait for confirmation
      read -t 10 -p "Are you sure you want to delete the existing DNS zone for domain: $DOMAIN and restore default records? (y/n): " CONFIRM
      if [[ $? -ne 0 ]]; then
        echo "Timed out."
        exit 1
      fi
      if [[ "$CONFIRM" == "y" || "$CONFIRM" == "Y" ]]; then
        echo "Restoring DNS zone for domain: $DOMAIN to default..."
        edit_zone_to_default "$DOMAIN"
        exit 0
      else
        echo "Canceled."
        exit 1
      fi
    fi
  else
    echo "Error: No domain provided."
    exit 1
  fi
}



create_dns_zone_for_domain(){
  domain_name=$1
  if [[ -f "/etc/bind/zones/$domain_name.zone" ]]; then
    echo "Error: DNS zone already exists for domain: $domain_name"
    exit 1
  fi
  
  if [[ -n "$domain_name" ]]; then
    echo "Creating DNS zone for domain: $domain_name"
  fi

    
  # generate new zone
    ZONE_TEMPLATE_PATH='/etc/openpanel/bind9/zone_template.txt'
    ZONE_FILE_DIR='/etc/bind/zones/'
    CONFIG_FILE='/etc/openpanel/openpanel/conf/openpanel.config'

    zone_template=$(<"$ZONE_TEMPLATE_PATH")

	# Function to extract value from config file
	get_config_value() {
	    local key="$1"
	    grep -E "^\s*${key}=" "$CONFIG_FILE" | sed -E "s/^\s*${key}=//" | tr -d '[:space:]'
	}



    ns1=$(get_config_value 'ns1')
    ns2=$(get_config_value 'ns2')

    # Fallback
    if [ -z "$ns1" ]; then
        ns1='ns1.openpanel.co'
    fi

    if [ -z "$ns2" ]; then
        ns2='ns2.openpanel.co'
    fi

    # Create zone content
    timestamp=$(date +"%Y%m%d")

	whoowns_output=$(opencli domains-whoowns "$domain_name")
	owner=$(echo "$whoowns_output" | awk -F "Owner of '$domain_name': " '{print $2}')
	IP_SERVER_1="https://ip.openpanel.com"
	IP_SERVER_2="https://ip.openpanel.com"
	IP_SERVER_3="https://ip.openpanel.com"
	
	 
	if [ -n "$owner" ]; then
	    USERNAME="$owner"
	    JSON_FILE="/etc/openpanel/openpanel/core/users/$USERNAME/ip.json"
	    SERVER_IP=$(curl --silent --max-time 2 -4 $IP_SERVER_1 || wget --timeout=2 -qO- $IP_SERVER_2 || curl --silent --max-time 2 -4 $IP_SERVER_3)
	        if [ -e "$JSON_FILE" ]; then
	            current_ip=$(jq -r '.ip' "$JSON_FILE")
	        else
	            current_ip="$SERVER_IP"
	        fi
     	else
	        current_ip=$(curl --silent --max-time 2 -4 $IP_SERVER_1 || wget --timeout=2 -qO- $IP_SERVER_2 || curl --silent --max-time 2 -4 $IP_SERVER_3)
		if [ -z "$current_ip" ]; then
		    current_ip=$(ip addr|grep 'inet '|grep global|head -n1|awk '{print $2}'|cut -f1 -d/)
		fi     
     	fi




    
    # Replace placeholders in the template
	  zone_content=$(echo "$zone_template" | sed -e "s/{domain}/$domain_name/g" \
	                                           -e "s/{ns1}/$ns1/g" \
	                                           -e "s/{ns2}/$ns2/g" \
	                                           -e "s/{server_ip}/$current_ip/g" \
	                                           -e "s/YYYYMMDD/$timestamp/g")

    # Ensure the directory exists
    mkdir -p "$ZONE_FILE_DIR"

    # Write the zone content to the zone file
    echo "$zone_content" > "$ZONE_FILE_DIR$domain_name.zone"

    # Reload BIND service
    docker --context default exec openpanel_dns rndc reconfig >/dev/null 2>&1
    cd /root && docker --context default compose up -d bind9  >/dev/null 2>&1
  exit 0

}



check_single_dns_zone(){
  DOMAIN=$1
  if [[ -n "$DOMAIN" ]]; then
    echo "Checking DNS zone for domain: $DOMAIN"
  fi
  docker --context default exec openpanel_dns named-checkzone  $DOMAIN /etc/bind/zones/$DOMAIN.zone
  exit 0
}

start_dns_server(){
  echo "Starting DNS service.."
  cd /root && docker --context default compose up -d bind9
  exit 0
}

stop_dns_server(){
  echo "Stopping DNS service.."
  docker stop openpanel_dns && docker rm openpanel_dns
  exit 0
}



soft_reset(){
  echo "Restarting DNS service.."
  docker restart openpanel_dns
  exit 0
}


hard_reset(){
  stop_dns_server
  start_dns_server
  exit 0
}



######## END MAIN FUNCTIONS ########













while [[ $# -gt 0 ]]; do
  case $1 in
    reconfig)
      reconfig_command
      shift
      ;;
    check)
      if [[ -z "$2" ]]; then
        echo "Error: 'check' command requires a zone argument."
        usage
      else
        check_single_dns_zone "$2"
        shift 2
      fi
      ;;
    list)
      list_all_zones
      ;; 
    show)
      if [[ -z "$2" ]]; then
        echo "Error: 'show' command requires a zone argument."
        usage
      else
        list_dns_records "$2"
        shift 2
      fi
      ;; 
    create)
      create_dns_zone_for_domain "$2"
      shift 2
      ;; 
    reload)
      reload_one_or_all_dns_zone "$2"
      shift 2
      ;; 
    delete)
      if [[ -n "$2" && -n "$3" ]]; then
        delete_dns_zone "$2" "$3"
        shift 3
      elif [[ -n "$2" ]]; then
        delete_dns_zone "$2"
        shift 2
      else
        echo "Error: 'delete' command requires a domain name."
        usage
      fi
      ;; 
    default)
      if [[ -n "$2" && -n "$3" ]]; then
        restore_zone_to_default "$2" "$3"
        shift 3
      elif [[ -n "$2" ]]; then
        restore_zone_to_default "$2"
        shift 2
      else
        echo "Error: 'default' command requires a domain name."
        usage
      fi
      ;; 
    count)
      show_count
      shift
      ;;
    restart)
      soft_reset
      shift
      ;;
    config)
      check_named_main_conf
      shift
      ;;
    start)
      start_dns_server
      shift
      ;;
    stop)
      stop_dns_server
      shift
      ;;
    hard-restart)
      hard_reset
      shift
      ;;
    *)
      if [[ -z "$DOMAIN" ]]; then
        DOMAIN=$1
      else
        echo "Unknown option: $1"
        usage
      fi
      shift
      ;;
  esac
done
