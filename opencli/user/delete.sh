#!/bin/bash
################################################################################
# Script Name: user/delete.sh
# Description: Delete user account and permanently remove all their data.
# Usage: opencli user-delete <username> [-y]
# Author: Stefan Pejcic
# Created: 01.10.2023
# Last Modified: 15.06.2026
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
	context_endpoint=$(docker context inspect "$context" --format '{{.Endpoints.docker.Host}}' 2>/dev/null)
    if [[ "$context_endpoint" == ssh://* ]]; then
	    ssh_host=${context_endpoint#ssh://}
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
	# legacy, handle: active_sessions
    sql+="DELETE FROM domains WHERE user_id='$user_id'; "
	sql+="DELETE FROM users WHERE username='$openpanel_username' OR username LIKE 'SUSPENDED_%_$openpanel_username';"
	[ -n "$sql" ] && mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$sql"

	# 3. terminate redis sessions
	# TODO: drop all cache by username!
	session_keys=$(docker --context=default exec openpanel_redis redis-cli --scan --pattern "session:$user_id:*")
	if [ -n "$session_keys" ]; then
		session_count=$(echo "$session_keys" | wc -l | tr -d ' ')
		while IFS= read -r key; do
			docker --context=default exec openpanel_redis redis-cli unlink "$key" > /dev/null
		done <<< "$session_keys"
	fi

	# 4. delete domain files, emails and reload Caddy
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
		   "/etc/bind/zones/$domain.zone"
           )

	        if grep -q "zone \"$domain\"" /etc/bind/named.conf.local; then
	            sed -i "/zone \"$domain\" IN {/,/};/d" /etc/bind/named.conf.local
	        fi
        done
		[[ ${#paths_to_delete[@]} -gt 0 ]] && ionice -c3 rm -rf "${paths_to_delete[@]}"
		# reload webserver
        nohup docker --context default exec caddy caddy reload --config /etc/caddy/Caddyfile >/dev/null 2>&1 &
        disown
		# reload dns
        nohup docker --context default exec openpanel_dns rndc reconfig >/dev/null 2>&1 &
        disown
    fi
}

postfwd_setup(){
    # Delete hourly email limits for the user
	opencli email-ratelimit --delete-user="$1" >/dev/null 2>&1		
}

delete_emails() {
    openpanel_username="$1"

	# email accounts
    local email_file="/etc/openpanel/openpanel/core/users/$openpanel_username/emails.yml"
    if [ -f "$email_file" ]; then
        emails=()
        while read -r _ email; do
            [ -n "$email" ] && emails+=("$email")
        done < "$email_file"

        if [ "${#emails[@]}" -gt 0 ]; then
			nohup opencli email-setup email del -y "${emails[@]}" >/dev/null 2>&1 &
			disown
        fi
    fi

	# aliases
	aliases_file="/etc/openpanel/openpanel/core/users/$openpanel_username/aliases.yml"
	
	if [ -f "$aliases_file" ]; then
	    aliases=()
	    while read -r _ email target; do
            [ -n "$email" ] && aliases+=("$email")
	    done < "$aliases_file"

	    if [ "${#aliases[@]}" -gt 0 ]; then
	        nohup opencli email-setup alias del "${aliases[@]}" >/dev/null 2>&1 &
	        disown
	    fi
	fi

	# regex aliases (default email address)
	regex_aliases_file="/usr/local/mail/openmail/docker-data/dms/config/postfix-regex.cf"
	
	if [ -f "$regex_aliases_file" ]; then
		tmp_file="$(mktemp)"
		changed=0

	    while read -r pattern target; do
	        [ -z "$pattern" ] && continue
			if [[ "$pattern" =~ @${domain//./\\.}\$ ]]; then
				changed=1; continue
	        fi
	        echo "$pattern $target" >> "$tmp_file"
	    done < "$regex_aliases_file"

		[ "$changed" -eq 1 ] && mv "$tmp_file" "$regex_aliases_file"
	fi
}

delete_ftp_users() {
    context="$1"
    users_dir="/etc/openpanel/ftp/users"
    ftp_accounts_file="${users_dir}/${context}/users.list"

    if [[ -d "${users_dir}/${context}" ]]; then
        if [[ -f "$ftp_accounts_file" ]]; then
            local max_jobs=5
            local job_count=0
            while IFS='|' read -r username _; do
                cut -d'|' -f1 "$ftp_accounts_file" | xargs -I{} docker --context=default exec openadmin_ftp deluser {}
            done < "$users_file"
            wait
        fi
        ionice -c3 rm -rf "${users_dir:?}/${context:?}"
    fi
}

delete_all_user_files() {
    if [ -n "$node_ip_address" ]; then
		# 1. delete from node
		ssh "root@$node_ip_address" bash -s -- "$context" <<'EOF'
		user="$1"
		
		pkill -u "$user" -9 2>/dev/null || true
		
		if command -v deluser >/dev/null 2>&1; then
			deluser --remove-home "$user" >/dev/null 2>&1 || true
		elif command -v userdel >/dev/null 2>&1; then
			userdel -r "$user" >/dev/null 2>&1 || true
		fi
		
		[ -d "/home/$user" ] && ionice -c3 rm -rf "/home/$user"
EOF
		# 2. unmount from master
		umount "/home/$context" >/dev/null 2>&1
    fi
	# 3. delete on master 
	pkill -u "$context" -9 2>/dev/null || true
	delete_system_user "$context"
    [ -d /home/"$context" ] && rm -rf "/home/${context:?}"
    [ -d /etc/openpanel/openpanel/core/users/"$context" ] && rm -rf "/etc/openpanel/openpanel/core/users/$context"
}

delete_context() {
    docker context rm "$context"  > /dev/null 2>&1
}

delete_system_user() {
    local user="$1"

    if command -v deluser >/dev/null 2>&1; then
        deluser --remove-home "$user" # Debian
    elif command -v userdel >/dev/null 2>&1; then
        userdel -r "$user"            # RHEL
    else
        echo "ERROR: Neither deluser nor userdel found"
        return 1
    fi
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
delete_emails "$USERNAME"
delete_ftp_users "$context" &
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
nohup opencli user-quota >/dev/null 2>&1 &
disown
