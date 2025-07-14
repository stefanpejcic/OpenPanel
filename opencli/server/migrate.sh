#!/bin/bash
################################################################################
# Script Name: server/migrate.sh
# Description: Migrates all data from this server to another.
# Usage: opencli server-migrate -h <DESTINATION_IP> --user root --password <DESTINATION_PASSWORD>
# Author: Stefan Pejcic
# Created: 26.06.2025
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

: '
Usage: opencli server-migrate -h <remote_host> -u <remote_user> [--password <password>] [--exclude-home] [--exclude-logs] [--exclude-mail] [--exclude-bind] [--exclude-openpanel] [--exclude-mysql] [--exclude-stack] [--exclude-postupdate] [--exclude-users]
'

REMOTE_HOST=""
REMOTE_USER=""
REMOTE_PASS=""
EXCLUDE_HOME=0
EXCLUDE_LOGS=0
EXCLUDE_MAIL=0
EXCLUDE_BIND=0
EXCLUDE_CSF=0
EXCLUDE_OPENPANEL=0
EXCLUDE_MYSQL=0
EXCLUDE_STACK=0
EXCLUDE_POSTUPDATE=0
EXCLUDE_USERS=0
EXCLUDE_CONTEXTS=0
FORCE=0
COMPOSE_START_MAIL=0

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--host)
            REMOTE_HOST="$2"
            shift 2
            ;;
        -u|--user)
            REMOTE_USER="$2"
            shift 2
            ;;
        --password)
            REMOTE_PASS="$2"
            shift 2
            ;;
        --exclude-home)
            EXCLUDE_HOME=1
            shift
            ;;
        --exclude-logs)
            EXCLUDE_LOGS=1
            shift
            ;;
	--exclude-logs)
            EXCLUDE_CSF=1
            shift
            ;;        --exclude-mail)
            EXCLUDE_MAIL=1
            shift
            ;;
        --exclude-bind)
            EXCLUDE_BIND=1
            shift
            ;;
        --exclude-openpanel)
            EXCLUDE_OPENPANEL=1
            shift
            ;;
        --exclude-mysql)
            EXCLUDE_MYSQL=1
            shift
            ;;
        --exclude-stack)
            EXCLUDE_STACK=1
            shift
            ;;
        --exclude-postupdate)
            EXCLUDE_POSTUPDATE=1
            shift
            ;;
        --exclude-users)
            EXCLUDE_USERS=1
            shift
            ;;
        --exclude-contexts)
            EXCLUDE_CONTEXTS=1
            shift
            ;;
        --force)
            FORCE=1
            shift
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

if [[ -z "$REMOTE_HOST" || -z "$REMOTE_USER" ]]; then
    echo "Usage: opencli server-migrate -h <remote_host> -u <remote_user> [--password <password>] [--exclude-* options]"
    exit 1
fi

RSYNC_OPTS="-az" #--progress

check_install_sshpass() {
	# If a password is provided, use sshpass for rsync/scp
	if [[ -n "$REMOTE_PASS" ]]; then
		if ! command -v sshpass &>/dev/null; then
		    if [[ -x "$(command -v apt-get)" ]]; then
		        apt-get update &>/dev/null && apt-get install -y sshpass &>/dev/null
		    elif [[ -x "$(command -v dnf)" ]]; then
		        dnf install -y sshpass &>/dev/null
		    elif [[ -x "$(command -v yum)" ]]; then
		        yum install -y epel-release &>/dev/null && yum install -y sshpass &>/dev/null
		    elif [[ -x "$(command -v pacman)" ]]; then
		        pacman -Sy sshpass &>/dev/null
		    else
		        exit 2
		    fi
		fi
  		RSYNC_CMD="sshpass -p '$REMOTE_PASS' rsync $RSYNC_OPTS -e 'ssh -o StrictHostKeyChecking=no'"
	else
	    	RSYNC_CMD="rsync $RSYNC_OPTS"
	fi
}



check_disk_used_on_source() {
    HOME_DIR="/home"
    USED_HOME_ON_SOURCE=$(df --output=used "$HOME_DIR" | tail -n 1)
    USED_HOME_ON_SOURCE_BYTES=$(($USED_HOME_ON_SOURCE * 1024)) #1K blocks
}

check_if_dest_has_space(){
    echo "Checking available disk on destination server ..."

    AVAILABLE_HOME_ON_DEST=$(sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no "${REMOTE_USER}@${REMOTE_HOST}" \
        "df --output=avail $HOME_DIR | tail -n 1")

    AVAILABLE_HOME_ON_DEST_BYTES=$(($AVAILABLE_HOME_ON_DEST * 1024)) #1K blocks

    if [[ $AVAILABLE_HOME_ON_DEST_BYTES -ge $USED_HOME_ON_SOURCE_BYTES ]]; then
        echo "There is enough disk space on destination server."
    else
        echo "FATAL ERROR: Not enough disk space on destination."
        echo "Available: $AVAILABLE_HOME_ON_DEST_BYTES bytes - Needed: $USED_HOME_ON_SOURCE_BYTES bytes"
        exit 1
    fi
}



get_server_ipv4(){
	# Get server ipv4
 
	# list of ip servers for checks
	IP_SERVER_1="https://ip.openpanel.com"
	IP_SERVER_2="https://ipv4.openpanel.com"
	IP_SERVER_3="https://ifconfig.me"

	current_ip=$(curl --silent --max-time 2 -4 $IP_SERVER_1 || \
                 wget --inet4-only --timeout=2 -qO- $IP_SERVER_2 || \
                 curl --silent --max-time 2 -4 $IP_SERVER_3)

	if [ -z "$current_ip" ]; then
	    current_ip=$(ip addr|grep 'inet '|grep global|head -n1|awk '{print $2}'|cut -f1 -d/)
	fi

	    is_valid_ipv4() {
	        local ip=$1
	        # is it ip
	        [[ $ip =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}$ ]] && \
	        # is it private
	        ! [[ $ip =~ ^10\. ]] && \
	        ! [[ $ip =~ ^172\.(1[6-9]|2[0-9]|3[0-1])\. ]] && \
	        ! [[ $ip =~ ^192\.168\. ]]
	    }

	if ! is_valid_ipv4 "$current_ip"; then
	        echo "Invalid or private IPv4 address: $current_ip. OpenPanel requires a public IPv4 address to bind Nginx configuration files."
	fi

}


get_users_count_on_destination() {

	user_count_query="SELECT COUNT(*) FROM users"

    user_count=$(sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no "${REMOTE_USER}@${REMOTE_HOST}" \
    "mysql --defaults-extra-file=$config_file -D $mysql_database -e \"$user_count_query\" -sN")
 
        if [ $? -ne 0 ]; then
            echo "[✘] ERROR: Unable to check users from remote server. Is OpenPanel installed?"
            exit 1
        fi
    
        if [ "$user_count" -gt 0 ]; then
            echo "[✘] ERROR: Migration is possible only to a freshly installed OpenPanel with no existing users."
            exit 1
        fi
}




copy_user_accounts() {
    TMPDIR=$(mktemp -d)
    awk -F: '$3 >= 1000 {print}' /etc/passwd > "$TMPDIR/passwd.users"
    awk -F: '$3 >= 1000 {print}' /etc/group > "$TMPDIR/group.users"
    grep -F -f <(cut -d: -f1 "$TMPDIR/passwd.users") /etc/shadow > "$TMPDIR/shadow.users"
    eval $RSYNC_CMD "$TMPDIR/passwd.users" "$TMPDIR/group.users" "$TMPDIR/shadow.users" ${REMOTE_USER}@${REMOTE_HOST}:/root/
    rm -rf "$TMPDIR" >/dev/null

sshpass -p "$REMOTE_PASS" ssh -q -o LogLevel=ERROR -o StrictHostKeyChecking=no "${REMOTE_USER}@${REMOTE_HOST}" <<'EOF' >/dev/null 2>&1
USER_PASSWD="/root/passwd.users"
USER_GROUP="/root/group.users"
USER_SHADOW="/root/shadow.users"

# Add groups
cut -d: -f1,3 "$USER_GROUP" | while IFS=: read -r group gid; do
    if ! getent group "$group" > /dev/null; then
        groupadd -g "$gid" "$group"
    fi
done

# Add users
cut -d: -f1,3,4,5,6,7 "$USER_PASSWD" | while IFS=: read -r user uid gid comment home shell; do
    if ! id "$user" &>/dev/null; then
        useradd -u "$uid" -g "$gid" -c "$comment" -d "$home" -s "$shell" "$user"
    fi
done

# Set passwords from shadow file
cut -d: -f1,2 "$USER_SHADOW" | while IFS=: read -r user hash; do
    if [ -n "$hash" ]; then
        usermod -p "$hash" "$user"
    fi
done

rm -rf $USER_PASSWD $USER_GROUP $USER_SHADOW

EOF
    
}

store_running_containers_for_users() {
output_file="/tmp/docker_containers_names.txt"
> "$output_file"  # clear the file

for userdir in /home/*; do
    if [ -d "$userdir" ]; then
        username=$(basename "$userdir")
        compose_file="$userdir/docker-compose.yml"

        if [ -f "$compose_file" ]; then
            echo "Checking docker context for user: $username"
            containers=$(docker --context="$username" ps -a --format "{{.Names}}" 2>/dev/null)

            if [ -n "$containers" ]; then
                containers_single_line=$(echo "$containers" | tr '\n' ' ' | sed 's/ $//')
                echo "$username: $containers_single_line" >> "$output_file"
            else
                echo "$username: no containers" >> "$output_file"
            fi
        fi
    fi
done

eval $RSYNC_CMD $output_file ${REMOTE_USER}@${REMOTE_HOST}:$output_file
}

restore_running_containers_for_all_users() {
output_file="/tmp/docker_containers_names.txt"

# Count total lines for progress display
TOTALCOUNT=$(wc -l < "$output_file")
CURRENT=0

# Open the file on FD 3 to avoid stdin conflicts
exec 3<"$output_file"

while IFS=: read -r username containers <&3; do
    CURRENT=$((CURRENT+1))
    username=$(echo "$username" | xargs)

    if [[ -z "$username" ]] || [[ "$containers" =~ no\ containers ]]; then
        #echo "Skipping user $username (no containers or empty line)"
        continue
    fi

    echo "Starting containers for context: $username ($CURRENT/$TOTALCOUNT)..."
    sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no "${REMOTE_USER}@${REMOTE_HOST}" \
        "docker --context=$username compose -f /home/$username/docker-compose.yml down >/dev/null 2>&1 && docker --context=$username compose -f /home/$username/docker-compose.yml up -d $containers"
done

# Close FD 3
exec 3<&-
}

copy_docker_contexts() {

    eval $RSYNC_CMD /run/user/ ${REMOTE_USER}@${REMOTE_HOST}:/run/user/
    eval $RSYNC_CMD /etc/apparmor.d/home.* ${REMOTE_USER}@${REMOTE_HOST}:/etc/apparmor.d/

    sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no "${REMOTE_USER}@${REMOTE_HOST}" \
	"systemctl restart apparmor.service"
    
	awk -F: '$3 >= 1000 && $3 < 65534 {print $1 ":" $3}' /etc/passwd > /tmp/userlist.txt
	TOTALCOUNT=$(wc -l < /tmp/userlist.txt)
	CURRENT=0
	
	# Open the file on FD 3
	exec 3</tmp/userlist.txt
	
	while IFS=: read -r USERNAME USER_ID <&3; do
	    CURRENT=$((CURRENT+1))
	    SRC="/home/$USERNAME/.docker"
	    if [[ -d "$SRC" ]]; then
	        echo "Setting linger for: $USERNAME"
		sshpass -p "$REMOTE_PASS" ssh -tt -o StrictHostKeyChecking=no "${REMOTE_USER}@${REMOTE_HOST}" \
		    "loginctl enable-linger $USERNAME" \
		    >/dev/null 2>&1 || echo "Failed to enable linger for $USERNAME"
     
     		eval $RSYNC_CMD /run/user/$USER_ID ${REMOTE_USER}@${REMOTE_HOST}:/run/user/$USER_ID

		sshpass -p "$REMOTE_PASS" ssh -tt -o StrictHostKeyChecking=no "${REMOTE_USER}@${REMOTE_HOST}" "
		    if [ -d /run/user/$USER_ID ]; then
		        ls -l /run/user/$USER_ID
		    else
		        echo '❌ Transfer failed for /run/user/$USER_ID/'
		        exit 1
		    fi
      "

	        echo "Creating Docker context: $USERNAME ($CURRENT/$TOTALCOUNT) ..."
	        sshpass -p "$REMOTE_PASS" ssh -tt -o StrictHostKeyChecking=no "${REMOTE_USER}@${REMOTE_HOST}" \
	            "docker context create $USERNAME --docker 'host=unix:///hostfs/run/user/${USER_ID}/docker.sock' --description '$USERNAME'" || echo "Failed context for $USERNAME"
	
	        echo "Configuring docker service for: $USERNAME"

		sshpass -p "$REMOTE_PASS" ssh -tt -o StrictHostKeyChecking=no "${REMOTE_USER}@${REMOTE_HOST}" \
		    "machinectl shell ${USERNAME}@ /bin/bash -c 'systemctl --user daemon-reload'" \
		    >/dev/null 2>&1 || echo "Failed to reload daemon for $USERNAME"

		sshpass -p "$REMOTE_PASS" ssh -tt -o StrictHostKeyChecking=no "${REMOTE_USER}@${REMOTE_HOST}" \
		    "machinectl shell ${USERNAME}@ /bin/bash -c 'systemctl --user --quiet restart docker'" \
		    >/dev/null 2>&1 || echo "Failed to restart docker for $USERNAME"
	    else
	        echo "No .docker directory for $USERNAME, skipping."
	    fi
            echo "[OK] Context $USERNAME processed"
	    echo ""
	done
	
	# Close FD 3
	exec 3<&-
}

restart_services_on_target() {
            echo "Restarting services on ${REMOTE_HOST} server ..."
            sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no "${REMOTE_USER}@${REMOTE_HOST}" \
                "cd /root && docker compose compose up -d openpanel bind9 caddy >/dev/null 2>&1 && systemctl restart admin >/dev/null 2>&1"

	if [[ $COMPOSE_START_MAIL -eq 1 ]]; then
            echo "Starting mailserver and webmail on ${REMOTE_HOST} server ..."
            sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no "${REMOTE_USER}@${REMOTE_HOST}" \
                "cd /usr/local/mail/openmail && docker --context default compose up -d mailserver roundcube >/dev/null 2>&1"  
	fi

	#todo: ftp, clamav 
  
}

refresh_quotas() {
            echo "Recalculating disk and inodes usage for all users on ${REMOTE_HOST} ..."
            sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no "${REMOTE_USER}@${REMOTE_HOST}" \
                "quotacheck -avm >/dev/null 2>&1 && repquota -u / > /etc/openpanel/openpanel/core/users/repquota"
}

  
replace_ip_in_zones() {
	sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no "${REMOTE_USER}@${REMOTE_HOST}" bash -c "'
	    zones_dir=\"/etc/bind/zones\"
	
	    for ZONE_CONF in \"\$zones_dir\"/*.zone; do
	        if [ -f \"\$ZONE_CONF\" ]; then
	            domain=\$(basename \"\$ZONE_CONF\" .zone)
	            sed -i \"s/$current_ip/$REMOTE_HOST/g\" \"\$ZONE_CONF\"
	            echo \"Updated DNS zone for domain \$domain - \$ZONE_CONF\"
	        fi
	    done
	'"
}


# MAIN
DB_CONFIG_FILE="/usr/local/opencli/db.sh"
. "$DB_CONFIG_FILE"

ssh-keygen -f '/root/.ssh/known_hosts' -R $REMOTE_HOST >/dev/null 2>&1
check_install_sshpass
get_server_ipv4
check_disk_used_on_source
check_if_dest_has_space


if [[ $FORCE -eq 0 ]]; then
	get_users_count_on_destination
fi

if [[ $EXCLUDE_USERS -eq 0 ]]; then
    echo "Creating system users on remote server ..."
    copy_user_accounts
fi

if [[ $EXCLUDE_HOME -eq 0 ]]; then
    echo "Syncing files (/home directory) ..."
    RSYNC_OUTPUT=$(eval $RSYNC_CMD /home/ "${REMOTE_USER}@${REMOTE_HOST}:/home/" 2>&1)
    RSYNC_EXIT=$?
    echo "$RSYNC_OUTPUT"
    if [[ $RSYNC_EXIT -eq 0 ]]; then
        echo "[OK] Files have been copied to the remote server."
	echo ""
    else
        echo "[ERROR] Rsync failed! Output:"
        echo "$RSYNC_OUTPUT"
        exit 1
    fi
fi

if [[ $EXCLUDE_CONTEXTS -eq 0 ]]; then
    echo "Syncing docker contexts ..."
    copy_docker_contexts # create docker context, start docker, set quotas
fi


sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no "${REMOTE_USER}@${REMOTE_HOST}" \
"systemctl daemon-reload" 



# set quotas
echo "Restoring user quotas ..."
sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no "${REMOTE_USER}@${REMOTE_HOST}" \
	"opencli user-quota --all"


if [[ $EXCLUDE_LOGS -eq 0 ]]; then
    echo "Syncing /var/log/openpanel ..."
    eval $RSYNC_CMD /var/log/openpanel/ ${REMOTE_USER}@${REMOTE_HOST}:/var/log/openpanel/

    echo "Syncing /var/log/caddy/ ..."
    eval $RSYNC_CMD /var/log/caddy/ ${REMOTE_USER}@${REMOTE_HOST}:/var/log/caddy/
fi

if [[ $EXCLUDE_MAIL -eq 0 ]]; then
    PANEL_CONFIG_FILE="/etc/openpanel/openpanel/conf/openpanel.config"
    key_value=$(grep "^key=" $PANEL_CONFIG_FILE | cut -d'=' -f2-)

	if [ -n "$key_value" ]; then
	    if [ -d /usr/local/mail/openmail ]; then
	        echo "Syncing /var/mail ..."
	        eval $RSYNC_CMD /usr/local/mail/openmail ${REMOTE_USER}@${REMOTE_HOST}:/usr/local/mail/openmail
	        COMPOSE_START_MAIL=1
	    fi
	fi


fi

if [[ $EXCLUDE_CSF -eq 0 ]]; then
    echo "Syncing /etc/csf/ ..."
    eval $RSYNC_CMD /etc/csf/ ${REMOTE_USER}@${REMOTE_HOST}:/etc/csf/     
    sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no "${REMOTE_USER}@${REMOTE_HOST}" \
	"csf -a $current_ip > /dev/null && csf -r >/dev/null && systemctl restart lfd"    
fi


if [[ $EXCLUDE_BIND -eq 0 ]]; then
    echo "Syncing /etc/bind ..."
    eval $RSYNC_CMD /etc/bind/ ${REMOTE_USER}@${REMOTE_HOST}:/etc/bind/
    replace_ip_in_zones   
fi

if [[ $EXCLUDE_OPENPANEL -eq 0 ]]; then
    echo "Syncing /etc/openpanel ..."
    eval $RSYNC_CMD /etc/openpanel/ ${REMOTE_USER}@${REMOTE_HOST}:/etc/openpanel/

    echo "Syncing system cronjobs..."
    eval $RSYNC_CMD /etc/cron.d/openpanel ${REMOTE_USER}@${REMOTE_HOST}:/etc/cron.d/
    
fi

if [[ $EXCLUDE_MYSQL -eq 0 ]]; then
    echo "Syncing root_mysql Docker volume ..."
    if [[ -d "/var/lib/docker/volumes/root_mysql/_data" ]]; then
        eval $RSYNC_CMD /var/lib/docker/volumes/root_mysql/_data/ ${REMOTE_USER}@${REMOTE_HOST}:/var/lib/docker/volumes/root_mysql/_data/
    else
        echo "/var/lib/docker/volumes/root_mysql/_data does not exist! Skipping."
    fi
fi

if [[ $EXCLUDE_STACK -eq 0 ]]; then
    echo "Syncing /root/docker-compose.yml and /root/.env ..."
    eval $RSYNC_CMD /root/docker-compose.yml ${REMOTE_USER}@${REMOTE_HOST}:/root/
    eval $RSYNC_CMD /root/.env ${REMOTE_USER}@${REMOTE_HOST}:/root/
fi

if [[ $EXCLUDE_POSTUPDATE -eq 0 ]]; then
    if [[ -e /root/openpanel_run_after_update ]]; then
        echo "Syncing /root/openpanel_run_after_update ..."
        eval $RSYNC_CMD /root/openpanel_run_after_update ${REMOTE_USER}@${REMOTE_HOST}:/root/
    fi
fi


store_running_containers_for_users        # export running contianers on source and copy to dest
restore_running_containers_for_all_users  # start containers per context on dest
restart_services_on_target                # restart openpanel, webserver and admin on dest
refresh_quotas                            # recalculate users usage on dest

echo "[OK] Sync complete"
