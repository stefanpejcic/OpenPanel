#!/bin/bash
################################################################################
# Script Name: user/delete.sh
# Description: Delete user account and permanently remove all their data.
# Usage: opencli user-delete <username> [-y] [--all]
# Author: Stefan Pejcic
# Created: 01.10.2023
# Last Modified: 02.02.2026
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
# Variables
skip_confirmation=false
delete_all=false
node_ip_address=""



# ======================================================================
# Validations
if [ "$#" -lt 1 ] || [ "$#" -gt 3 ]; then
    echo "Usage: opencli user-delete <username> [-y] [--all]"
    exit 1
fi

confirm_action() {
    if [ "$skip_confirmation" = true ]; then
        return 0
    fi

    read -r -p "This will permanently delete user '$1' and all associated data. Confirm? [Y/n]: " response
    response=${response,,} # to lowercase
    if [[ ! $response =~ ^(yes|y| ) ]]; then
        echo "Operation canceled for user '$1'."
        exit 0
    fi
}



source /usr/local/opencli/db.sh


# ======================================================================
# Functions

get_user_info() {
	# 1. get context and ID
    read user_id context <<< $(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -e "
        SELECT id, server FROM users 
        WHERE username='$provided_username'
        UNION ALL
        SELECT id, server FROM users 
        WHERE username LIKE 'SUSPENDED_%_$provided_username'
        LIMIT 1;
    ")

    if [ -z "$user_id" ]; then
        echo "ERROR: User '$provided_username' not found in the database."
        exit 1
    fi

    # 2. set context
	context_flag="--context $context"

    # 3. check if remote
    if [[ "$context" == ssh://* ]]; then
        ssh_host=${context#ssh://}
        node_ip_address=${ssh_host#*@}
    else
        node_ip_address=""
    fi
}

reload_user_quotas() {
	touch /etc/openpanel/openpanel/core/users/repquota
	quotacheck -avm >/dev/null 2>&1
	repquota -u / > /etc/openpanel/openpanel/core/users/repquota 
}

delete_vhosts_files() {
    local all_user_domains=$(opencli domains-user $username)
    deleted_count=0 
    # TODO: delete on remote nginx server!
    for domain in $all_user_domains; do
        rm /etc/openpanel/caddy/domains/${domain}.conf >/dev/null 2>&1
        deleted_count=$((deleted_count + 1))
    done
    docker --context default exec caddy caddy reload --config /etc/caddy/Caddyfile >/dev/null 2>&1
}

delete_user_from_database() {
    openpanel_username="$1"
	
    # 1. Get all domain IDs
    domain_ids=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e \
        "SELECT domain_id FROM domains WHERE user_id='$user_id';" -N | paste -sd "," -)

    sql=""

    # 2. Delete sites if there are any domain IDs
    if [ -n "$domain_ids" ]; then
        sql+="DELETE FROM sites WHERE domain_id IN ($domain_ids); "
    fi

    # 3. Delete domains for the user
    sql+="DELETE FROM domains WHERE user_id='$user_id'; "

    # 4. Delete active sessions for the user
    sql+="DELETE FROM active_sessions WHERE user_id='$user_id'; "

    # 5. Delete the user
    sql+="DELETE FROM users WHERE username='$openpanel_username';"

    # 6. Execute all 5 in a single call
    if [ -n "$sql" ]; then
        mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$sql"
    fi
}

delete_ftp_users() {
    openpanel_username="$1"
    users_dir="/etc/openpanel/ftp/users"
    users_file="${users_dir}/${openpanel_username}/users.list"

    if [[ -d "${users_dir}/${openpanel_username}" ]]; then
	    if [[ -f "$users_file" ]]; then
		echo "Checking and removing user's FTP sub-accounts"
		while IFS='|' read -r username password directories; do
		    echo "Deleting FTP user: $username"
		    opencli ftp-delete "$username" "$openpanel_username"
		done < "$users_file"
	    fi
     	rm -rf ${users_dir}/${openpanel_username}
    fi
}

delete_all_user_files() {
    if [ -n "$node_ip_address" ]; then
        ssh "root@$node_ip_address" bash -c "'
            pkill -u $context -9 2>/dev/null || true
            deluser --remove-home $context 2>/dev/null || true
            [ -d /home/$context ] && rm -rf /home/$context
        '"
    fi
	pkill -u $context -9 2>/dev/null || true
    deluser --remove-home $context 2>/dev/null || true
    [ -d /home/$context ] && rm -rf /home/$context
    [ -d /etc/openpanel/openpanel/core/stats/$context ] && rm -rf /etc/openpanel/openpanel/core/stats/$context
    [ -d /etc/openpanel/openpanel/core/users/$context ] && rm -rf /etc/openpanel/openpanel/core/users/$context

}

delete_context() {
    docker context rm $context  > /dev/null 2>&1
}

refresh_resellers_data() {
	local reseller_files="/etc/openpanel/openadmin/resellers"
	# TODO: optimize: now it checks all resellers, instead should check just the account owner!
	if [ -d "$reseller_files" ]; then
		for json_file in "$reseller_files"/*.json; do
			if [ -f "$json_file" ]; then
				reseller=$(basename "$json_file" .json)
				query_for_owner="SELECT COUNT(*) FROM users WHERE owner='$reseller';"
				current_accounts=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$query_for_owner" -se)
				if [ $? -eq 0 ]; then
					jq ".current_accounts = $current_accounts" $json_file > /tmp/${reseller}_config.json && mv /tmp/${reseller}_config.json $json_file
				fi
			fi
		done
	fi
}

delete_user() {
    local provided_username="$1"
    if [[ "$provided_username" == SUSPENDED_* ]]; then
        username="${provided_username##*_}"
    else    
        username=$provided_username
    fi
    
    confirm_action "$username"
    get_user_info                                 # get user ID and docker context from db
    delete_vhosts_files                           # delete caddy vhosts files from the server
    delete_ftp_users $provided_username           # delete all ftp sub-usersthat
    delete_user_from_database $provided_username  # delete user from database
    delete_all_user_files                         # permanently delete data
    delete_context  
    refresh_resellers_data                        # count users for all resellers
    reload_user_quotas
    echo "User $username deleted successfully." # if we made it
}



# ======================================================================
# Parse args
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



# ======================================================================
# Main
if [ "$delete_all" = true ]; then
	# ALL USERS
	all_users=$(opencli user-list --json | grep -v 'SUSPENDED' | awk -F'"' '/username/ {print $4}')
	
	if [[ -z "$all_users" || "$all_users" == "No users." ]]; then
		echo "No users found in the database."
		exit 1
	fi

    total_users=$(echo "$all_users" | wc -w)
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
else
	# SINGLE USER
	if [ -z "$provided_username" ]; then
	    echo "Error: Username is required unless --all is specified."
	    exit 1
	fi
	delete_user "$provided_username"
fi
