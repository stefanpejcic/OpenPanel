#!/bin/bash
################################################################################
# Script Name: domains/delete.sh
# Description: Delete a domain name.
# Usage: opencli domains-delete <DOMAIN_NAME> --debug
# Author: Stefan Pejcic
# Created: 07.11.2024
# Last Modified: 06.02.2026
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

[ $# -lt 1 ] && { echo "Usage: opencli domains-delete <DOMAIN_NAME> [--debug]"; exit 1; }

# ======================================================================
# Constants
readonly PANEL_CONFIG_FILE='/etc/openpanel/openpanel/conf/openpanel.config'

# ======================================================================
# Variables
domain_name="$1"
onion_domain=false
debug_mode=false

[[ $2 == --debug ]] && debug_mode=true

# ======================================================================
# Functions
log() {
    if $debug_mode; then
        echo "$1"
    fi
}

pre_flight_checks() {
    log "Checking owner for domain $domain_name"
    user=$(opencli domains-whoowns "$domain_name" | awk -F "Owner of '$domain_name': " '{print $2}')
    if [[ -z "$user" ]]; then
        echo "❌ No owner found for '$domain_name'. Ensure the domain is assigned and MySQL is running."
        exit 1
    fi

    domain_id=$(mysql -Nse "SELECT domain_id FROM domains WHERE domain_url = '$domain_name';")
    if [[ -z "$domain_id" ]]; then
        echo "❌ Domain ID not found for '$domain_name'. Ensure it exists and MySQL is running."
        exit 1
    fi

    if [[ "$domain_name" =~ ^[a-zA-Z0-9]{16}\.onion$ ]]; then
        onion_domain=true
        log "Detected .onion address — Tor website will be removed from configuration."
    fi
}


remove_onion_files() {
    if $onion_domain; then
        hostfs_path="/home/$context/tor"
        onion_dir=$(grep -rlE '\.onion$' "$hostfs_path" | grep '/hostname$' | xargs -n1 dirname)
        for dir in $onion_dir; do
            dir_name=$(basename "$dir")

            if [[ "$dir_name" == "hidden_service" ]]; then
                sed -i '/^HiddenServiceDir.*\/hidden_service$/{
                    N
                    /\nHiddenServicePort/d
                    d
                }' "$hostfs_path/torrc"

            elif [[ "$dir_name" =~ ^[0-9]+$ ]]; then
                sed -i "/^HiddenServiceDir.*\/hidden_service\/$dir_name$/{
                    N
                    /\nHiddenServicePort/d
                    d
                }" "$hostfs_path/torrc"
            fi
        done
	rm -rf $onion_dir
    fi
}



restart_tor_for_user() {
	if [ $(docker --context $context ps -q -f name=tor) ]; then
 	    log "Tor service is running, restarting to remove configuration.."
 	    docker --context $context restart tor >/dev/null 2>&1
     	fi  
}


get_webserver_for_user(){
	    log "Checking webserver configuration"
	    output=$(opencli webserver-get_webserver_for_user $user)		
		ws=$(echo "$output" | grep -Eo 'nginx|openresty|apache|openlitespeed|litespeed' | head -n1)
}

rm_domain_to_clamav_list(){	
	local domains_list="/etc/openpanel/clamav/domains.list"
 	local domain_path="/home/$user/$domain_name"
	# from 0.3.4 we have optional script to run clamav scan for all files in domains dirs, this adds new domains to list of directories to monitor
 	if [ -f $domains_list ]; then
      	log "ClamAV Upload Scanner is enabled - Removing $domain_path for monitoring"
	sed -i "\|$domain_path|d" "$domains_list"
 	fi
}

get_slave_dns_option() {
	BIND_CONFIG_FILE="/etc/bind/named.conf.options"

    # 1. check if cluster enabled, if it is, both lines are uncommented!
    if ! grep -qP '^(?!\s*//).*allow-transfer\s+\{[^}]*\}' "$BIND_CONFIG_FILE" \
        || ! grep -qP '^(?!\s*//).*also-notify\s+\{[^}]*\}' "$BIND_CONFIG_FILE"; then
        return
    fi

	# 2. get IP(s) for slave servers, line formats are: allow-transfer {EXAMPLE;ANOTHER;};  AND also-notify {EXAMPLE;ANOTHER;};
    ALLOW_TRANSFER=$(grep -oP '^(?!\s*//).*allow-transfer\s+\{\s*\K[^}]*' "$BIND_CONFIG_FILE" | tr -d '[:space:]')
    ALSO_NOTIFY=$(grep -oP '^(?!\s*//).*also-notify\s+\{\s*\K[^}]*' "$BIND_CONFIG_FILE" | tr -d '[:space:]')

    IFS=';' read -r -a ALLOW_TRANSFER_IPS <<< "$ALLOW_TRANSFER"
    IFS=';' read -r -a ALSO_NOTIFY_IPS <<< "$ALSO_NOTIFY"

    # 3. For each slave server we execute notify_slave()
    for ip in "${ALLOW_TRANSFER_IPS[@]}"; do
        [[ -z "$ip" ]] && continue
        if [[ " ${ALSO_NOTIFY_IPS[*]} " == *" $ip "* ]]; then
            SLAVE_IP=$ip
            notify_slave
        fi
    done
}

notify_slave(){
	echo "Notifying Slave DNS server ($SLAVE_IP) to remove zone for domain $domain_name"
	timeout 5 ssh -q -o LogLevel=ERROR -o ConnectTimeout=5 -T root@$SLAVE_IP <<EOF >/dev/null 2>&1
    if grep -q "$domain_name.zone" /etc/bind/named.conf.local; then
        sed -i "/zone \"$domain_name\" {/,/};/d" /etc/bind/named.conf.local
		rm -f "/etc/bind/zones/$domain_name.zone"
        echo "Zone $domain_name deleted from slave server."
    fi
EOF

	timeout 5 ssh -q -o LogLevel=ERROR -o ConnectTimeout=5 -T root@$SLAVE_IP <<EOF >/dev/null 2>&1
    docker --context default exec openpanel_dns rndc reconfig >/dev/null 2>&1
EOF
}


get_user_info() {
    local user="$1"
    local query="SELECT id, server FROM users WHERE username = '${user}';"
    
    # Retrieve both id and context
    user_info=$(mysql -se "$query")
    
    # Extract user_id and context from the result
    user_id=$(echo "$user_info" | awk '{print $1}')
    context=$(echo "$user_info" | awk '{print $2}')
    
    echo "$user_id,$context"
}




vhost_files_delete() {
	
	vhost_in_docker_file="/etc/$ws/sites-available/${domain_name}.conf"
 	vhost_ln_in_docker_file="/etc/$ws/sites-enabled/${domain_name}.conf"

result=$(get_user_info "$user")
user_id=$(echo "$result" | cut -d',' -f1)
context=$(echo "$result" | cut -d',' -f2)

if [ -z "$user_id" ]; then
    echo "FATAL ERROR: user $user does not exist."
    exit 1
fi



	log "Deleting $vhost_in_docker_file"	
	vhost_in_docker_file="/home/$context/docker-data/volumes/${context}_webserver_data/_data/${domain_name}.conf"
	rm $vhost_in_docker_file >/dev/null 2>&1

 	log "Restarting $ws to apply changes"
	docker --context=$context $user restart $ws >/dev/null 2>&1
}



# Function to validate and reload Caddy service
check_and_add_to_enabled() {
    # Validate the Caddyfile
    if docker exec caddy caddy validate --config /etc/caddy/Caddyfile 2>&1 | grep -q "Valid configuration"; then
        # If validation is successful, reload the Caddy service
        docker --context default compose exec caddy caddy reload --config /etc/caddy/Caddyfile >/dev/null 2>&1
        return 0
    else
        # If validation fails, revert the domains file and return an error
        echo "Validation failed, reverting changes."
        return 1
    fi
}



delete_domain_file() {
  log "Removing domain from the proxy"
 
	mkdir -p /etc/openpanel/openpanel/core/users/$user/
	domains_file="/etc/openpanel/caddy/domains/$domain_name.conf" 
 	rm -rf $domains_file
  
	if [ $(docker ps -q -f name=caddy) ]; then
 	    log "Caddy is running, reloading configuration"
	    check_and_add_to_enabled
	fi   
	 #stats and logs
 	rm -rf /var/log/caddy/domlogs/$domain_name/
  	rm -rf /var/log/caddy/stats/$user/$domain_name.html
   	rm -rf /var/log/caddy/coraza_waf/${domain_name}.log > /dev/null 2>&1
}


update_named_conf() {
    ZONE_FILE_DIR='/etc/bind/zones/'
    NAMED_CONF_LOCAL='/etc/bind/named.conf.local'
    local config_line="zone \"$domain_name\" IN { type master; file \"$ZONE_FILE_DIR$domain_name.zone\"; };"

    # Check if the domain exists in named.conf.local
    if grep -q "zone \"$domain_name\"" "$NAMED_CONF_LOCAL"; then
        log "Removing zone information from the server."
        sed -i "/zone \"$domain_name\"/d" "$NAMED_CONF_LOCAL"
    fi
}


remove_dns_entries_from_apex_zone() {
    local tld_file="/etc/openpanel/openpanel/conf/public_suffix_list.dat"

    tld="${domain_name##*.}"
    tld_lower=$(echo "$tld" | tr '[:upper:]' '[:lower:]')
    update_tlds=false

    if [[ ! -f "$tld_file" ]]; then
        log "TLD list not found, downloading from IANA..."
        update_tlds=true
    elif [[ $(find "$tld_file" -mtime +6 2>/dev/null) ]]; then
        update_tlds=true
    fi

    if [[ "$update_tlds" == true ]]; then
        mkdir -p "$(dirname "$tld_file")"
        wget --timeout=5 --tries=3 -q --inet4-only -O "$tld_file" "https://publicsuffix.org/list/public_suffix_list.dat"
        if [[ $? -ne 0 ]]; then
            log "Failed to download TLD list from IANA"
        fi
    fi

    if grep -qx "$tld_lower" "$tld_file"; then
        domain_base="${domain_name%.*}"
        if [[ "$domain_base" == *.* ]]; then
            is_subdomain=true
            apex_tld="${domain_name##*.}"
            second_level="${domain_name%.*}"
            second_level="${second_level##*.}"
            apex_domain="${second_level}.${apex_tld}"
            subdomain="${domain_name%%.$apex_domain}"

            log "Detected subdomain: $subdomain of apex domain: $apex_domain"

            zone_file="/etc/bind/zones/${apex_domain}.zone"
            if [[ -f "$zone_file" ]]; then
                log "Removing DNS records for subdomain $subdomain from $zone_file"
				escaped_domain_name=$(printf '%s' "$domain_name" | sed 's/[.[\*^$/]/\\&/g')
				sed -i "/^${escaped_domain_name}[[:space:]]/d" "$zone_file"
				sed -i "/^${escaped_domain_name}\.[[:space:]]/d" "$zone_file"

                if docker ps -q -f name=openpanel_dns >/dev/null 2>&1; then
                    log "Reloading BIND DNS to apply changes"
                    docker exec openpanel_dns rndc reconfig >/dev/null 2>&1
                fi
            else
                log "Zone file for apex domain $apex_domain not found."
            fi
		else
            zone_file="/etc/bind/zones/${domain_name}.zone"
            if [[ -f "$zone_file" ]]; then
                log "Deleting DNS zone $zone_file"
				rm -rf $zone_file

                if docker ps -q -f name=openpanel_dns >/dev/null 2>&1; then
                    log "Reloading BIND DNS to apply changes"
                    docker exec openpanel_dns rndc reconfig >/dev/null 2>&1
                fi
            else
                log "Zone file for domain $domain_name not found."
            fi
        fi
    else
        echo "ERROR: Invalid domain or unrecognized TLD: .$tld_lower"
        exit 1
    fi
}


# Function to create a zone file
delete_zone_file() {
    ZONE_FILE_DIR='/etc/bind/zones/'
    log "Removing DNS zone file: $ZONE_FILE_DIR$domain_name.zone"
    rm "$ZONE_FILE_DIR$domain_name.zone"

    # Reload BIND service
    if [ $(docker ps -q -f name=openpanel_dns) ]; then
        log "DNS service is running, reloading the zones"
      	docker exec openpanel_dns rndc reconfig >/dev/null 2>&1
    fi
}




# add mountpoint and reload mailserver
# todo: need better solution!
delete_mail_mountpoint(){
    key_value=$(grep "^key=" "$PANEL_CONFIG_FILE" | cut -d'=' -f2-)

    if [ -n "$key_value" ]; then
        DOMAIN_DIR="/home/$user/mail/$domain_name/"
        COMPOSE_FILE="/usr/local/mail/openmail/compose.yml"
        mount_path="/var/mail/$domain_name/"
        volume_to_remove="$DOMAIN_DIR:$mount_path"

        if [ -f "$COMPOSE_FILE" ]; then
            log "Removign volume: $volume_to_remove from mailserver"

            # Escape slashes for sed
            escaped_volume=$(printf '%s\n' "  - $volume_to_remove" | sed 's/[\/&]/\\&/g')

            # Remove the exact volume line if it exists in the mailserver service block
            sed -i "/^  mailserver:/,/^[^[:space:]]/ {
                /^    volumes:/,/^[^[:space:]]/ {
                    /^\s*- $escaped_volume/d
                }
            }" "$COMPOSE_FILE"

            log "Reloading mailserver to apply changes"
            (cd /usr/local/mail/openmail/ && docker-compose up -d --force-recreate mailserver) > /dev/null 2>&1 & disown
        fi
    fi
}



delete_websites() {
    log "Removing any websites associated with domain (ID: $domain_id)"
    delete_sites_query="DELETE FROM sites WHERE domain_id = '$domain_id';"
    mysql -e "$delete_sites_query"
}


delete_emails() {
  local username="$1"
  local domain="$2"
  local email_file="/etc/openpanel/openpanel/core/users/$username/emails.yml"

  # Check if the email file exists
  if [ -f "$email_file" ]; then
    	  log "Deleting @$domain_name email accounts"
	  # Loop through each line in the email file
	  while IFS= read -r line; do
	    # Extract the email address from each line
	    email=$(echo "$line" | awk '{print $2}')
	    if [[ -n "$email" ]]; then
	      echo "- deleting: $email"
	      opencli email-setup email del "$email"
	    fi
	  done < "$email_file"
  fi
}




delete_domain_from_mysql(){
    local domain_name="$1"
    log "Removing $domain_name from the domains database"
    local delete_query="DELETE from domains where domain_url = '$domain_name';"
    mysql -e "$delete_query"
}



dns_stuff() {
    enabled_features_line=$(grep '^enabled_modules=' "$PANEL_CONFIG_FILE")
    if [[ $enabled_features_line == *"dns"* ]]; then  
        delete_zone_file                             # create zone
        update_named_conf                            # include zone
		get_slave_dns_option                         # slaves
    fi
}


delete_domain() {
    local user="$1"
    local domain_name="$2"
    
    delete_websites $domain_name                     # delete sites associated with domain id
    # TODO: delete pm2 apps associated with it.
    delete_domain_from_mysql $domain_name            # delete

	local verify_query="SELECT COUNT(*) FROM domains WHERE domain_url = '$domain_name';"
    local result=$(mysql -N -e "$verify_query")

    if [ "$result" -eq 0 ]; then
        get_webserver_for_user                       #
        vhost_files_delete                           # delete file in container
		remove_dns_entries_from_apex_zone	         # subdomain-specific DNS cleanup
		if $onion_domain; then
			remove_onion_files		                 # delete conf files
			restart_tor_for_user		             # restart if running
	 	else
		    delete_domain_file                       # remove caddy files
		 	dns_stuff			                     # remove dns files
	 	fi
        delete_mail_mountpoint                       # delete mountpoint to mailserver
        delete_emails  $user $domain_name            # delete mails for the domain
        rm_domain_to_clamav_list                     # added in 0.3.4    
        echo "Domain $domain_name deleted successfully"
    else
        log "Deleting domain $domain_name failed! Contact administrator to check if the mysql database is running."
        echo "Failed to delete domain $domain_name"
    fi
}


# ======================================================================
# Main
pre_flight_checks
delete_domain "$user" "$domain_name"
