#!/bin/bash
################################################################################
# Script Name: user/delete.sh
# Description: Delete user account and permanently remove all their data.
# Usage: opencli user-delete <username> [-y]
# Author: Stefan Pejcic
# Created: 01.10.2023
# Last Modified: 27.03.2026
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
node_ip_address=""



# ======================================================================
# Validations
if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
    echo "Usage: opencli user-delete <username> [-y]"
    exit 1
fi

confirm_action() {
    if [ "$skip_confirmation" = true ]; then
        return 0
    fi

    read -r -p "This will permanently delete user '$USERNAME' and all associated data. Confirm? [Y/n]: " response
    response=${response,,} # to lowercase
    if [[ ! $response =~ ^(yes|y| ) ]]; then
        echo "Operation canceled for user '$USERNAME'."
        exit 0
    fi
}



source /usr/local/opencli/db.sh


# ======================================================================
# Functions

get_user_info() {
	# 1. get context and ID
    read -r user_id context <<< $(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -e "
        SELECT id, server FROM users 
        WHERE username='$USERNAME'
        UNION ALL
        SELECT id, server FROM users 
        WHERE username LIKE 'SUSPENDED_%_$USERNAME'
        LIMIT 1;
    ")

    [ -n "$user_id" ] || { echo "ERROR: User '$USERNAME' not found in the database."; exit 1; }

    # 2. check if remote
    if [[ "$context" == ssh://* ]]; then
        ssh_host=${context#ssh://}
        node_ip_address=${ssh_host#*@}
    else
        node_ip_address=""
    fi
}

delete_user_from_database() {
    openpanel_username="$1"
	
    # 1. Get all domain IDs and URLs
	read -r domain_ids domain_urls < <(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -e "
	SELECT 
	    GROUP_CONCAT(domain_id) AS ids,
	    GROUP_CONCAT(domain_url) AS urls
	FROM domains
	WHERE user_id='$user_id';
	")

	# 2. Prepare SQL queries and execute all at once
    sql=""
	[ -n "$domain_ids" ] && sql+="DELETE FROM sites WHERE domain_id IN ($domain_ids); "
    sql+="DELETE FROM domains WHERE user_id='$user_id'; "
    sql+="DELETE FROM active_sessions WHERE user_id='$user_id'; "
	sql+="DELETE FROM users WHERE username='$openpanel_username' OR username LIKE 'SUSPENDED_%_$openpanel_username';"
	[ -n "$sql" ] && mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$sql"
	# 3. delete domain files, emails and reload Caddy
	ionice -c3 rm -rf "/var/log/caddy/stats/$openpanel_username"            # goaccess reports
    if [ -n "$domain_urls" ]; then
		local email_storage_path=$(grep -E '^email_storage_location=' /etc/openpanel/openadmin/config/admin.ini | cut -d'=' -f2- | xargs)
	    IFS=',' read -ra domains_array <<< "$domain_urls"
        paths_to_delete=()
        for domain in "${domains_array[@]}"; do
		    if [[ "$email_storage_path" == /* && "$email_storage_path" != "user_dir" ]]; then
		        paths_to_delete+=("$email_storage_path/$domain")
		    fi
           paths_to_delete+=(
           "/etc/openpanel/caddy/domains/$domain.conf"
           "/etc/openpanel/caddy/suspended_domains/$domain.conf"
           "/var/log/caddy/domlogs/$domain"
           "/var/log/caddy/coraza_waf/$domain.log"
           "/etc/openpanel/caddy/ssl/custom/$domain"
           )
        done
		[[ ${#paths_to_delete[@]} -gt 0 ]] && ionice -c3 rm -rf "${paths_to_delete[@]}"
        nohup docker --context default exec caddy caddy reload --config /etc/caddy/Caddyfile >/dev/null 2>&1 &
        disown
    fi
}

postfwd_setup(){
    # Delete hourly email limits for the user
	opencli email-ratelimit --delete-user="$1" >/dev/null 2>&1		
}

delete_email_users() {
    openpanel_username="$1"
    local email_file="/etc/openpanel/openpanel/core/users/$openpanel_username/emails.yml"
    if [ -f "$email_file" ]; then    
        mapfile -t emails < <(awk 'NF {print $2}' "$email_file")
        if [ "${#emails[@]}" -gt 0 ]; then
            opencli email-setup email del -y "${emails[@]}"
        fi
    fi
}

delete_ftp_users() {
    openpanel_username="$1"
    users_dir="/etc/openpanel/ftp/users"
    users_file="${users_dir}/${openpanel_username}/users.list"

    if [[ -d "${users_dir}/${openpanel_username}" ]]; then
        if [[ -f "$users_file" ]]; then
            local max_jobs=5
            local job_count=0
            while IFS='|' read -r username password directories; do
                opencli ftp-delete "$username" "$openpanel_username" &
                (( ++job_count ))
                if (( job_count >= max_jobs )); then
                    wait -n 2>/dev/null || wait
                    (( --job_count ))
                fi
            done < "$users_file"
            wait
        fi
        ionice -c3 rm -rf "${users_dir:?}/${openpanel_username:?}"
    fi
}

delete_all_user_files() {
    if [ -n "$node_ip_address" ]; then
		# 1. delete from node
        ssh "root@$node_ip_address" bash -c "'
            pkill -u $context -9 2>/dev/null || true
            deluser --remove-home "$context" >/dev/null 2>&1 || true
            [ -d /home/"$context" ] && ionice -c3 rm -rf /home/$context 
        '"
		# 2. unmount from master
		umount "/home/$context" >/dev/null 2>&1
    fi
	# 3. delete on master 
	pkill -u "$context" -9 2>/dev/null || true
    deluser --remove-home "$context" >/dev/null 2>&1 || true
    [ -d /home/"$context" ] && rm -rf "/home/${context:?}"
    [ -d /etc/openpanel/openpanel/core/users/"$context" ] && rm -rf "/etc/openpanel/openpanel/core/users/$context"

}

delete_context() {
    docker context rm "$context"  > /dev/null 2>&1
}

refresh_resellers_data() {
	local reseller_files="/etc/openpanel/openadmin/resellers"
	# TODO: optimize: now it checks all resellers, instead should check just the account owner!
	if [ -d "$reseller_files" ]; then
		for json_file in "$reseller_files"/*.json; do
			if [ -f "$json_file" ]; then
				reseller=$(basename "$json_file" .json)
				query_for_owner="SELECT COUNT(*) FROM users WHERE owner='$reseller';"
				current_accounts=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -e "$query_for_owner")
				mysql_exit=$?
				if [ $mysql_exit -eq 0 ]; then				
					jq ".current_accounts = $current_accounts" $json_file > /tmp/${reseller}_config.json && mv /tmp/${reseller}_config.json $json_file
				fi
			fi
		done
	fi
}

# ======================================================================
# Parse args
USERNAME="$1"

if [ "$2" = "-y" ]; then
	skip_confirmation=true
fi

# ======================================================================
# Main

if [ -z "$USERNAME" ]; then
	echo "ERROR: Username is required."
	exit 1
fi

USERNAME="${USERNAME##*_}"

# 1. -y
confirm_action "$USERNAME"

# 2. get docker context and user ID from database
get_user_info

# 3. in parallel: delete emails, delete ftp accounts, user/sites/domains from database, homedir, postfwd limits, docker context
delete_email_users "$USERNAME" &
delete_ftp_users "$USERNAME" &
delete_user_from_database "$USERNAME" &
delete_all_user_files &
postfwd_setup "$USERNAME" &
delete_context &

# 4. wait for all of the above functions to finish
wait

# 5. notify
echo "User $USERNAME deleted successfully."
nohup opencli sentinel --action=user_delete --title="User account deleted" --message="User account '$USERNAME' has been deleted." >/dev/null 2>&1 &
disown

# 6. refresh resellers usage
refresh_resellers_data &
nohup bash -c 'quotacheck -avm >/dev/null 2>&1; repquota -u / > /etc/openpanel/openpanel/core/users/repquota' >/dev/null 2>&1 &
disown
