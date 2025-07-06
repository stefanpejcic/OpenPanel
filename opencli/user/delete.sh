#!/bin/bash
################################################################################
# Script Name: user/delete.sh
# Description: Delete user account and permanently remove all their data.
# Usage: opencli user-delete <username> [-y] [--all]
# Author: Stefan Pejcic
# Created: 01.10.2023
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

# Check if the correct number of command-line arguments is provided
if [ "$#" -lt 1 ] || [ "$#" -gt 3 ]; then
    echo "Usage: opencli user-delete <username> [-y] [--all]"
    exit 1
fi


# Default values
skip_confirmation=false
delete_all=false
node_ip_address=""

# Parse arguments
for arg in "$@"; do
    case $arg in
        --all)
            delete_all=true
            ;;
        -y)
            skip_confirmation=true
            ;;
        *)
            provided_username="$arg"
            ;;
    esac
done

# ----------------- Confirmation Function -----------------
confirm_action() {
    if [ "$skip_confirmation" = true ]; then
        return 0
    fi

    # Ask for confirmation
    read -r -p "This will permanently delete user '$1' and all associated data. Confirm? [Y/n]: " response
    response=${response,,} # Convert to lowercase
    if [[ ! $response =~ ^(yes|y| ) ]]; then
        echo "Operation canceled for user '$1'."
        exit 0
    fi
}


source /usr/local/opencli/db.sh


get_docker_context_for_user(){
    # GET CONTEXT NAME FOR DOCKER COMMANDS
    context=$(mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "SELECT server FROM users WHERE username='$provided_username';" -N)
    context_flag="--context $context"

    context_type=$(docker context inspect "$context" --format '{{.Endpoints.docker.Host}}')

    # The Host usually looks like ssh://user@ip or something else
    if [[ "$context_type" == ssh://* ]]; then
        ssh_host=${context_type#ssh://}
        if [[ "$ssh_host" == *@* ]]; then
            node_ip_address=${ssh_host#*@}
        else
            node_ip_address=$ssh_host
        fi
        #echo "Remote SSH Docker context detected. Node IP address: $node_ip_address"
    else
        node_ip_address=""
    fi
    
}


get_userid_from_db() {
    # Get the user_id from the 'users' table
    user_id=$(mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "SELECT id FROM users WHERE username='$provided_username';" -N)

    if [ -z "$user_id" ]; then
        suspended_user=$(mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "SELECT username FROM users WHERE username LIKE 'SUSPENDED_%_$provided_username';" -N)
    
        if [ -z "$suspended_user" ]; then
            echo "ERROR: User '$username' not found in the database."
            exit 1
        else
            user_id=$(mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "SELECT id FROM users WHERE username='$suspended_user';" -N)
        fi
    else
        suspended_user=$provided_username
    fi
}


    reload_user_quotas() {
    	quotacheck -avm >/dev/null 2>&1
    	repquota -u / > /etc/openpanel/openpanel/core/users/repquota 
    }

# TODO: delete on remote nginx server!

# Delete all users domains vhosts files from Caddy
delete_vhosts_files() {

    local all_user_domains=$(opencli domains-user $username)
    deleted_count=0 
    
    for domain in $all_user_domains; do
        rm /etc/openpanel/caddy/domains/${domain}.conf >/dev/null 2>&1
        deleted_count=$((deleted_count + 1))
    done
    
    docker --context default exec caddy caddy reload --config /etc/caddy/Caddyfile >/dev/null 2>&1
    
    echo "Configuration (VirutalHosts) for $deleted_count domain(s) of user '$username' deleted successfully."
}

# Function to delete user from the database
delete_user_from_database() {

    # Step 2: Get all domain_ids associated with the user_id from the 'domains' table
    domain_ids=$(mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "SELECT domain_id FROM domains WHERE user_id='$user_id';" -N)

    # Step 3: Delete rows from the 'sites' table based on the domain_ids
    for domain_id in $domain_ids; do
        mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "DELETE FROM sites WHERE domain_id='$domain_id';"
    done

    # Step 4: Delete rows from the 'domains' table based on the user_id
    mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "DELETE FROM domains WHERE user_id='$user_id';"

    # Step 5: Delete the user_id from the 'active_sessions' table
    mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "DELETE FROM active_sessions WHERE user_id='$user_id';"

    # Step 6: Delete the user from the 'users' table
    mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "DELETE FROM users WHERE username='$suspended_user';"

    #echo "User '$username' deleted successfully."
}



delete_ftp_users() {
    openpanel_username="$1"
    users_dir="/etc/openpanel/ftp/users"
    users_file="${users_dir}/${openpanel_username}/users.list"

    # Check if the users file exists
    if [[ -f "$users_file" ]]; then
        echo "Checking and removing user's FTP sub-accounts"
        # Loop through each line in the users.list file
        while IFS='|' read -r username password directories; do
            # Run the opencli command for each username
            echo "Deleting FTP user: $username"
            opencli ftp-delete "$username" "$openpanel_username"
        done < "$users_file"
    fi
}


# Function to delete bandwidth limit settings for a user
delete_bandwidth_limits() {
	# TODO
        ip_address=$(docker $context_flag container inspect -f '{{ .NetworkSettings.IPAddress }}' "$username")
        if [ -n "$node_ip_address" ]; then
            # TODO: INSTEAD OF ROOT USER SSH CONFIG OR OUR CUSTOM USER!
            ssh "root@$node_ip_address" "tc qdisc del dev docker0 root && tc class del dev docker0 parent 1: classid 1:1 && tc filter del dev docker0 parent 1: protocol ip prio 16 u32 match ip dst $ip_address"
        else
              tc qdisc del dev docker0 root 2>/dev/null
              tc class del dev docker0 parent 1: classid 1:1 2>/dev/null
              tc filter del dev docker0 parent 1: protocol ip prio 16 u32 match ip dst "$ip_address" 2>/dev/null
        fi
}

edit_firewall_rules(){
    # CSF
    if command -v csf >/dev/null 2>&1; then
        FIREWALL="CSF"
        container_ports=("22" "3306" "7681" "8080")
        #we use range, so not need to rm rules for account delete..
    fi
}


delete_all_user_files() {
        if [ -n "$node_ip_address" ]; then
	    umount /home/$username
            ssh "root@$node_ip_address" rm -rf /home/$username > /dev/null 2>&1  
            ssh "root@$node_ip_address" killall -u $username -9  > /dev/null 2>&1
            ssh "root@$node_ip_address" deluser --remove-home $username  > /dev/null 2>&1
	fi
            rm -rf /home/$username > /dev/null 2>&1
            killall -u $username -9  > /dev/null 2>&1
            deluser --remove-home $username  > /dev/null 2>&1
            rm -rf /etc/openpanel/openpanel/core/stats/$username
            rm -rf /etc/openpanel/openpanel/core/users/$username
}


delete_context() {
    docker context rm $context  > /dev/null 2>&1
}

refresh_resellers_data() {

	    local reseller_files="/etc/openpanel/openadmin/resellers"

        if [ -d "$reseller_files" ]; then
            for json_file in "$reseller_files"/*.json; do
                if [ -f "$json_file" ]; then
                    reseller=$(basename "$json_file" .json)  # Extract reseller name from filename
        
                    # Query the database for the number of users owned by the reseller
                    query_for_owner="SELECT COUNT(*) FROM users WHERE owner='$reseller';"
                    current_accounts=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$query_for_owner" -se)
                    if [ $? -eq 0 ]; then
                        jq ".current_accounts = $current_accounts" $json_file > /tmp/${reseller}_config.json && mv /tmp/${reseller}_config.json $json_file
                    fi
                fi
            done
        fi
}

# MAIN



delete_user() {
    # Get username from a command-line argument
    local provided_username="$1"
    
    if [[ "$provided_username" == SUSPENDED_* ]]; then
        username="${provided_username##*_}"
    else    
        username=$provided_username
    fi
    
    confirm_action "$username"
    
    # MAIN EXECUTION
    get_userid_from_db                       # check if user exists
    get_docker_context_for_user              # on which server is the container running
    delete_vhosts_files                      # delete nginx conf files from that server
    edit_firewall_rules                      # close user ports on firewall
    #delete_bandwidth_limits                  # delete bandwidth limits for private ip
    delete_ftp_users $provided_username
    delete_user_from_database                # delete user from database
    delete_all_user_files                    # permanently delete data
    delete_context  
    refresh_resellers_data                   # count users for all resellers
    reload_user_quotas
    echo "User $username deleted successfully."           # if we made it
}



# ----------------- Delete All Users -----------------
if [ "$delete_all" = true ]; then
  all_users=$(opencli user-list --json | grep -v 'SUSPENDED' | awk -F'"' '/username/ {print $4}')

  # Check if no sites found
  if [[ -z "$all_users" || "$all_users" == "No users." ]]; then
    echo "No users found in the database."
    exit 1
  fi

    total_users=$(echo "$all_users" | wc -w)

    # Loop through all users and delete them

    current_user_index=1
    for user in $all_users; do
        echo "- $user ($current_user_index/$total_users)"
        delete_user "$user"
    echo "------------------------------"
    ((current_user_index++))
    done
    echo "DONE."
    echo "$((current_user_index - 1)) users have been deleted."
    exit 0
fi


# ----------------- Single User Deletion -----------------
if [ -z "$provided_username" ]; then
    echo "Error: Username is required unless --all is specified."
    exit 1
fi


delete_user "$provided_username"
