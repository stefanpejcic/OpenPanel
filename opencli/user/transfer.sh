#!/bin/bash
################################################################################
# Script Name: user/transfer.sh
# Description: Transfers a single user account from this server to another.
# Usage: opencli user-transfer -h <DESTINATION_IP> --user <OPENPANEL_USERNAME> --password <DESTINATION_SSH_PASSWORD>
# Author: Stefan Pejcic
# Created: 28.06.2025
# Last Modified: 14.07.2025
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
Usage: opencli user-transfer -h <remote_host> -u <account_name> [--password <password>] [--force]
'

USERNAME=""
REMOTE_HOST=""
REMOTE_USER="root"
REMOTE_PASS=""
FORCE=0

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--host)
            REMOTE_HOST="$2"
            shift 2
            ;;
        -u|--username)
            USERNAME="$2"
            shift 2
            ;;
        --password)
            REMOTE_PASS="$2"
            shift 2
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

if [[ -z "$REMOTE_HOST" || -z "$USER" ]]; then
    echo "Usage: opencli user-transfer -h <remote_host> -u <OPENPANEL_USERNAME> [--password <ssh_password>] [--force]"
    exit 1
fi

RSYNC_OPTS="-az" #--progress

install_sshpass() {
	if [[ -x "$(command -v apt-get)" ]]; then
	    DEBIAN_FRONTEND=noninteractive apt-get update -qq
	    DEBIAN_FRONTEND=noninteractive apt-get install -y -qq sshpass
	elif [[ -x "$(command -v dnf)" ]]; then
	    sudo dnf install -y sshpass
	elif [[ -x "$(command -v yum)" ]]; then
	    sudo yum install -y epel-release && sudo yum install -y sshpass
	elif [[ -x "$(command -v pacman)" ]]; then
	    sudo pacman -Sy sshpass
	else
	    echo "Package manager not supported. Please install sshpass manually."
	    exit 2
	fi
}


format_commands() {
	# If a password is provided, use sshpass for rsync/scp
	if [[ -n "$REMOTE_PASS" ]]; then
	    if ! command -v sshpass &>/dev/null; then
	        echo "sshpass not found. Installing..."
	 	install_sshpass
	    fi
	    RSYNC_CMD="sshpass -p '$REMOTE_PASS' rsync $RSYNC_OPTS -e 'ssh -o StrictHostKeyChecking=no'"
	    SSH_CMD="sshpass -p $REMOTE_PASS ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=ERROR ${REMOTE_USER}@${REMOTE_HOST}"
	else
	    RSYNC_CMD="rsync $RSYNC_OPTS"
	    SSH_CMD="ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=ERROR ${REMOTE_USER}@${REMOTE_HOST}" # for ssh keys!
	fi

	# test
	echo "Testing SSH connection to $REMOTE_USER@$REMOTE_HOST..."
	if $SSH_CMD "echo 'SSH connection established, starting transfer process..'" >/dev/null 2>&1; then
	    echo "SSH connection established, starting transfer process.."
	else
	    echo "[✘] SSH connection to $REMOTE_HOST failed. Please check credentials or SSH keys."
            echo ""
	    echo "Command used:"
     	    echo $SSH_CMD
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


PANEL_CONFIG_FILE="/etc/openpanel/openpanel/conf/openpanel.config"
key_value=$(grep "^key=" $PANEL_CONFIG_FILE | cut -d'=' -f2-)

get_users_count_on_destination() {
	  user_count_query="SELECT COUNT(*) FROM users"
    user_count=$($SSH_CMD "mysql --defaults-extra-file=$config_file -D $mysql_database -e \"$user_count_query\" -sN")
 
        if [ $? -ne 0 ]; then
            echo "[✘] ERROR: Unable to check users from remote server. Is OpenPanel installed?"
            exit 1
        fi

    if [ -n "$key_value" ]; then
    :
    else
        if [ "$user_count" -gt 2 ]; then
            echo "[✘] ERROR: OpenPanel Community edition has a limit of 3 user accounts - which should be enough for private use."
	          echo "If you require more than 3 accounts, please consider purchasing the Enterprise version that allows unlimited number of users and domains/websites."
            exit 1
        fi
    fi
}


# Function to check if username already exists in the database
check_username_exists() {
    username_exists_query="SELECT COUNT(*) FROM users WHERE username = '$USERNAME'"
    user_count=$($SSH_CMD "mysql --defaults-extra-file=$config_file -D $mysql_database -e \"$username_exists_query\" -sN")
 
    # Check if successful
    if [ $? -ne 0 ]; then
        echo "[✘] Error: Unable to check username existence in the database. Is mysql running?"
        exit 1
    fi

    # Return the count of usernames found
    echo "$user_count"
}



copy_user_account() {
    USERNAME="$1"
    TMPDIR=$(mktemp -d)
    echo "Creating system user on remote server ..."
    awk -F: -v user="$USERNAME" '$1 == user {print}' /etc/passwd > "$TMPDIR/passwd.user"
    awk -F: -v user="$USERNAME" 'BEGIN{gid=""}
        $1 == user {gid=$4}
        $3 == gid {print}
        $1 == user {print}' /etc/group > "$TMPDIR/group.user"
    grep -F -w "^$USERNAME:" /etc/shadow > "$TMPDIR/shadow.user"

    eval $RSYNC_CMD "$TMPDIR/passwd.user" "$TMPDIR/group.user" "$TMPDIR/shadow.user" ${REMOTE_USER}@${REMOTE_HOST}:/root/
    rm -rf "$TMPDIR" >/dev/null

    $SSH_CMD <<'EOF' >/dev/null 2>&1
USER_PASSWD="/root/passwd.user"
USER_GROUP="/root/group.user"
USER_SHADOW="/root/shadow.user"

user_exists() {
    id "$1" &>/dev/null
}

get_used_uids() {
    getent passwd | cut -d: -f3
}

# Find a free UID >= 1000 and not in use
find_free_uid() {
    used_uids=$(get_used_uids)
    uid=$1
    while echo "$used_uids" | grep -qw "$uid"; do
        uid=$((uid + 1))
    done
    echo "$uid"
}

# Find a free username by appending number suffix

# Add groups (same as before)
cut -d: -f1,3 "$USER_GROUP" | while IFS=: read -r group gid; do
    if ! getent group "$group" > /dev/null; then
        groupadd -g "$gid" "$group"
    fi
done

# Read original user info, create user with new username and free UID
while IFS=: read -r user uid gid comment home shell; do

    free_uid=$(find_free_uid "$uid")

    if user_exists "$user"; then
        echo "System user $user already exists, skipping creation."
    else
        useradd -u "$free_uid" -g "$gid" -c "$comment" -d "$home" -s "$shell" "$user"
    fi
done < <(cut -d: -f1,3,4,5,6,7 "$USER_PASSWD")

# Set password for the new user(s)
cut -d: -f1,2 "$USER_SHADOW" | while IFS=: read -r user hash; do
    if [ -n "$hash" ]; then
        usermod -p "$hash" "$user"
    fi
done

rm -rf $USER_PASSWD $USER_GROUP $USER_SHADOW

EOF

}


store_running_containers_for_user() {
output_file="/tmp/docker_containers_names.txt"
> "$output_file"  # clear the file

compose_file="/home/$USERNAME/docker-compose.yml"
if [ -f "$compose_file" ]; then
    echo "Checking docker context ...."
    containers=$(docker --context="$USERNAME" ps -a --format "{{.Names}}" 2>/dev/null)
    if [ -n "$containers" ]; then
        containers_single_line=$(echo "$containers" | tr '\n' ' ' | sed 's/ $//')
        echo "$USERNAME: $containers_single_line" >> "$output_file"
    else
        echo "$USERNAME: no containers" >> "$output_file"
    fi
fi

eval $RSYNC_CMD $output_file ${REMOTE_USER}@${REMOTE_HOST}:$output_file
}

restore_running_containers_for_user() {
output_file="/tmp/docker_containers_names.txt"

# Open the file on FD 3 to avoid stdin conflicts
exec 3<"$output_file"

while IFS=: read -r username containers <&3; do
    username=$(echo "$username" | xargs)

    if [[ -z "$username" ]] || [[ "$containers" =~ no\ containers ]]; then
        echo "No containers running for user"
        continue
    fi

    echo "Starting containers inside docker context on remote server ..."
    $SSH_CMD "docker --context=$username compose -f /home/$username/docker-compose.yml down >/dev/null 2>&1 && docker --context=$username compose -f /home/$username/docker-compose.yml up -d $containers >/dev/null 2>&1"
done

# Close FD 3
exec 3<&-
}

import_mysql() {
  $SSH_CMD bash -s <<EOF
set -e

export mysql_database="$mysql_database"
export USERNAME="$USERNAME"
CONFIG_FILE="/etc/my.cnf"

if [[ -z "\$mysql_database" ]]; then
  echo "[ERROR] mysql_database is not set"
  exit 1
fi
if [[ -z "\$USERNAME" ]]; then
  echo "[ERROR] USERNAME is not set"
  exit 1
fi

cd "/tmp/user_import/" || { echo "[ERROR] Directory /tmp/user_import/ not found"; exit 1; }

# Fix trailing commas in SQL
for f in plan_\${USERNAME}_autoinc.sql user_\${USERNAME}_autoinc.sql domains_\${USERNAME}_autoinc.sql sites_\${USERNAME}_autoinc.sql; do
  [[ -f "\$f" ]] && sed -i -E ':a;N;\$!ba;s/,\s*;\s*/;/g' "\$f"
done

PLAN_NAME=\$(awk -F"'" '/INSERT INTO plans/ {getline; print \$2; exit}' "plan_\${USERNAME}_autoinc.sql")

EXISTING_PLAN_ID=\$(mysql --defaults-extra-file="\$CONFIG_FILE" -D "\$mysql_database" -N -s \
  -e "SELECT id FROM plans WHERE name = '\$PLAN_NAME' LIMIT 1;")

if [[ -n "\$EXISTING_PLAN_ID" ]]; then
  echo "Plan already exists (ID: \$EXISTING_PLAN_ID)"
else
  echo "Importing new plan..."
  (echo "USE \\\`\$mysql_database\\\`;" && cat "plan_\${USERNAME}_autoinc.sql") | mysql --defaults-extra-file="\$CONFIG_FILE"
  EXISTING_PLAN_ID=\$(mysql --defaults-extra-file="\$CONFIG_FILE" -D "\$mysql_database" -N -s \
    -e "SELECT id FROM plans WHERE name = '\$PLAN_NAME' LIMIT 1;")
fi

sed -E "s/,[[:space:]]*[0-9]+\);$/,\$EXISTING_PLAN_ID);/" "user_\${USERNAME}_autoinc.sql" > tmp_user.sql
sed -i "s/'NULL'/NULL/g" tmp_user.sql

echo "Importing user into database..."
(echo "USE \\\`\$mysql_database\\\`;" && cat tmp_user.sql) | mysql --defaults-extra-file="\$CONFIG_FILE"
rm -f tmp_user.sql

USER_ID=\$(mysql --defaults-extra-file="\$CONFIG_FILE" -D "\$mysql_database" -N -s \
  -e "SELECT id FROM users WHERE username = '\$USERNAME';")

if [[ -z "\$USER_ID" ]]; then
  echo "[ERROR] Failed to import user!"
  exit 1
fi


if [[ -f "sites_\${USERNAME}_autoinc.sql" ]]; then
  echo "Importing sites..."
  tail -n +2 "sites_\${USERNAME}_autoinc.sql" | sed "s/),/)\n/g" | while read -r line; do
    clean_line=\$(echo "\$line" | sed "s/[()']//g" | sed 's/,$//')
    SITE_NAME=\$(echo "\$clean_line" | cut -d',' -f1)
    DOMAIN_URL=\$(echo "\$clean_line" | cut -d',' -f2)
    ADMIN_EMAIL=\$(echo "\$clean_line" | cut -d',' -f3)
    VERSION=\$(echo "\$clean_line" | cut -d',' -f4)
    TYPE=\$(echo "\$clean_line" | cut -d',' -f6)
    PORTS=\$(echo "\$clean_line" | cut -d',' -f7)
    PATH=\$(echo "\$clean_line" | cut -d',' -f8)

    DOMAIN_ID=\$(mysql --defaults-extra-file="\$CONFIG_FILE" -D "\$mysql_database" -N -s \
      -e "SELECT domain_id FROM domains WHERE domain_url = '\$DOMAIN_URL' AND user_id = \$USER_ID LIMIT 1;")

    if [[ -n "\$DOMAIN_ID" ]]; then
      mysql --defaults-extra-file="\$CONFIG_FILE" -D "\$mysql_database" -e "
        INSERT INTO sites (site_name, domain_id, admin_email, version, type, ports, path)
        VALUES ('\$SITE_NAME', \$DOMAIN_ID, '\$ADMIN_EMAIL', '\$VERSION', '\$TYPE', \$PORTS, '\$PATH');"
      echo "Site imported: \$SITE_NAME"
    else
      echo "[ERROR] Domain not found for site: \$DOMAIN_URL"
    fi
  done
else
  echo "No sites found to import."
fi

EOF
}


export_mysql() {

TMP_DIR="/tmp/export_${USERNAME}_$RANDOM"
mkdir -p "$TMP_DIR"

# Get user ID
USER_ID=$(mysql --defaults-extra-file=$config_file -D $mysql_database -N -s \
  -e "SELECT id FROM users WHERE username = '$USERNAME';")

if [[ -z "$USER_ID" ]]; then
  echo "[ERROR] No user found with username '$USERNAME'"
  exit 1
fi

# Get plan ID
PLAN_ID=$(mysql --defaults-extra-file=$config_file -D $mysql_database -N -s \
  -e "SELECT plan_id FROM users WHERE id = $USER_ID;")

# Get plan name
PLAN_NAME=$(mysql --defaults-extra-file=$config_file -D $mysql_database -N -s \
  -e "SELECT name FROM plans WHERE id = $PLAN_ID;")

### EXPORT PLAN (no ID)
mysql --defaults-extra-file=$config_file -D $mysql_database -N -s -e "
  SELECT name, description, domains_limit, websites_limit, email_limit, ftp_limit,
         disk_limit, inodes_limit, db_limit, cpu, ram, bandwidth, feature_set
  FROM plans WHERE id = $PLAN_ID;" > "$TMP_DIR/plan.tsv"

awk -v table="plans" '
BEGIN {
  FS="\t";
  print "INSERT INTO plans (name, description, domains_limit, websites_limit, email_limit, ftp_limit, disk_limit, inodes_limit, db_limit, cpu, ram, bandwidth, feature_set) VALUES"
}
{
  printf "('\''%s'\'', '\''%s'\'', %s, %s, %s, %s, '\''%s'\'', %s, %s, '\''%s'\'', '\''%s'\'', %s, '\''%s'\''),\n",
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
}
END { print ";" }
' "$TMP_DIR/plan.tsv" > "$TMP_DIR/plan_${USERNAME}_autoinc.sql"


### EXPORT USER (no ID)
mysql --defaults-extra-file=$config_file -D $mysql_database -N -s -e "
  SELECT username, password, email, owner, user_domains, twofa_enabled, otp_secret,
         plan, registered_date, server, plan_id
  FROM users WHERE id = $USER_ID;" > "$TMP_DIR/user.tsv"

awk '
BEGIN {
  FS="\t";
  print "INSERT INTO users (username, password, email, owner, user_domains, twofa_enabled, otp_secret, plan, registered_date, server, plan_id) VALUES"
}
{
  printf "('\''%s'\'', '\''%s'\'', '\''%s'\'', '\''%s'\'', '\''%s'\'', %s, '\''%s'\'', '\''%s'\'', '\''%s'\'', '\''%s'\'', %s),\n",
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
}
END { print ";" }
' "$TMP_DIR/user.tsv" > "$TMP_DIR/user_${USERNAME}_autoinc.sql"

### EXPORT SITES (if any domains exist)
DOMAIN_IDS=$(mysql --defaults-extra-file=$config_file -D $mysql_database -N -s \
  -e "SELECT domain_id FROM domains WHERE user_id = $USER_ID;")

if [[ -z "$DOMAIN_IDS" ]]; then
  :
else
  DOMAIN_ID_LIST=$(echo "$DOMAIN_IDS" | paste -sd "," -)
	mysql --defaults-extra-file=$config_file -D $mysql_database -N -s -e "
	  SELECT site_name, domain_id, admin_email, version, created_date, type, ports, path
	  FROM sites WHERE domain_id IN ($DOMAIN_ID_LIST);" > "$TMP_DIR/sites.tsv"
	

  awk '
  BEGIN {
    FS="\t";
    print "INSERT INTO sites (site_name, domain_id, admin_email, version, created_date, type, ports, path) VALUES"
  }
  {
    printf "('\''%s'\'', %s, '\''%s'\'', '\''%s'\'', '\''%s'\'', '\''%s'\'', %s, '\''%s'\''),\n",
    $1, $2, $3, $4, $5, $6, $7, $8
  }
  END { print ";" }
  ' "$TMP_DIR/sites.tsv" > "$TMP_DIR/sites_${USERNAME}_autoinc.sql"
fi

eval $RSYNC_CMD $TMP_DIR/plan_${USERNAME}_autoinc.sql ${REMOTE_USER}@${REMOTE_HOST}:/tmp/user_import/
eval $RSYNC_CMD $TMP_DIR/user_${USERNAME}_autoinc.sql ${REMOTE_USER}@${REMOTE_HOST}:/tmp/user_import/
eval $RSYNC_CMD $TMP_DIR/plan_${USERNAME}_autoinc.sql ${REMOTE_USER}@${REMOTE_HOST}:/tmp/user_import/
[[ -f "sites_${USERNAME}_autoinc.sql" ]] && eval $RSYNC_CMD $TMP_DIR/sites_${USERNAME}_autoinc.sql ${REMOTE_USER}@${REMOTE_HOST}:/tmp/user_import/

}



rsync_files_for_user() {
    echo "Syncing files for user ..."
    RSYNC_OUTPUT=$(eval $RSYNC_CMD /home/$USERNAME "${REMOTE_USER}@${REMOTE_HOST}:/home/" 2>&1)
    RSYNC_EXIT=$?
    echo "$RSYNC_OUTPUT"
    if [[ $RSYNC_EXIT -eq 0 ]]; then
        :
	#todo: chown to $free_uid only if it changed!
    else
        echo "[ERROR] Rsync failed! Output:"
        echo "$RSYNC_OUTPUT"
        exit 1
    fi

    CADDY_STATS="/var/log/caddy/stats/$USERNAME"
    if [ -d "$CADDY_STATS" ]; then
		eval $RSYNC_CMD $CADDY_STATS ${REMOTE_USER}@${REMOTE_HOST}:/var/log/caddy/stats/
    fi

    ALL_DOMAINS=$(opencli domains-user "$USERNAME" --docroot --php_version)
    while IFS=$'\t ' read -r domain docroot php_version; do
    whoowns_output=$($SSH_CMD "opencli domains-whoowns $domain")
    owner=$(echo "$whoowns_output" | awk -F "Owner of '$domain': " '{print $2}')

    if [ -z "$owner" ]; then
	    	# add domain on remote
	        $SSH_CMD "opencli domains-add $domain $USERNAME --docroot $docroot --php_version $php_version --skip_caddy --skip_vhost --skip_containers --skip_dns"
	        if [ $? -ne 0 ]; then
		    echo "[✘] ERROR: Failed to import domain $domain"
		    exit 1
		fi
    
	        DOMAIN_CADDY_CONF="/etc/openpanel/caddy/domains/$domain.conf"
	        if [ -f "$DOMAIN_CADDY_CONF" ]; then
				eval $RSYNC_CMD $DOMAIN_CADDY_CONF ${REMOTE_USER}@${REMOTE_HOST}:/etc/openpanel/caddy/domains/
		fi
	
	
	        DOMAIN_CADDY_LOG="/var/log/caddy/domlogs/$domain"
	        if [ -f "$DOMAIN_CADDY_LOG" ]; then
				eval $RSYNC_CMD $DOMAIN_CADDY_LOG ${REMOTE_USER}@${REMOTE_HOST}:/var/log/caddy/domlogs/
		fi
	
	        DOMAIN_CADDY_WAF="/var/log/caddy/coraza_waf/$domain.log"
	        if [ -f "$DOMAIN_CADDY_WAF" ]; then
				eval $RSYNC_CMD $DOMAIN_CADDY_WAF ${REMOTE_USER}@${REMOTE_HOST}:/var/log/caddy/coraza_waf/
		fi

		DOMAIN_ZONE_FILE="/etc/bind/zones/$domain.zone"

		if [ -f "$DOMAIN_ZONE_FILE" ]; then
		    eval $RSYNC_CMD "$DOMAIN_ZONE_FILE" ${REMOTE_USER}@${REMOTE_HOST}:/etc/bind/zones/
      
		    $SSH_CMD "sed -i 's/$current_ip/$REMOTE_HOST/g' /etc/bind/zones/$domain.zone"
      
		    $SSH_CMD <<EOF > /dev/null 2>&1
grep -q "$domain" /etc/bind/named.conf.local || \
echo 'zone "$domain" IN { type master; file "/etc/bind/zones/$domain.zone"; };' >> /etc/bind/named.conf.local
EOF
		fi


	        DOMAIN_CADDY_SSL="/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/$domain"
	        if [ -d "$DOMAIN_CADDY_SSL" ]; then
				eval $RSYNC_CMD $DOMAIN_CADDY_SSL ${REMOTE_USER}@${REMOTE_HOST}:/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/
		    fi
	
	
	        DOMAIN_CADDY_CUSTOM_SSL="/etc/openpanel/caddy/ssl/certs/$domain"
	        if [ -d "$DOMAIN_CADDY_CUSTOM_SSL" ]; then
				eval $RSYNC_CMD $DOMAIN_CADDY_CUSTOM_SSL ${REMOTE_USER}@${REMOTE_HOST}:/etc/openpanel/caddy/ssl/certs/
		fi
 	fi
 
 done <<< "$ALL_DOMAINS"

}





copy_docker_context() {

    eval $RSYNC_CMD /run/user/$USERNAME ${REMOTE_USER}@${REMOTE_HOST}:/run/user/$USERNAME >/dev/null 2>&1
    eval $RSYNC_CMD /etc/apparmor.d/home.$USERNAME.bin.rootlesskit ${REMOTE_USER}@${REMOTE_HOST}:/etc/apparmor.d/ >/dev/null 2>&1
    # TODO!

    $SSH_CMD "systemctl restart apparmor.service" >/dev/null 2>&1
    
	awk -F: '$3 >= 1000 && $3 < 65534 {print $1 ":" $3}' /etc/passwd > /tmp/userlist.txt

	# Open the file on FD 3
	exec 3</tmp/userlist.txt
	
    SRC="/home/$USERNAME/.docker"
    if [[ -d "$SRC" ]]; then
        REMOTE_UID=$($SSH_CMD "id -u $USERNAME" 2>/dev/null)

        if [[ -z "$REMOTE_UID" ]]; then
            echo "FATAL ERROR: Failed to get UID for user $USERNAME on remote server"
            exit 1
        else
            echo "Creating Docker context: $USERNAME on destination ..."
            $SSH_CMD "docker context create $USERNAME --docker 'host=unix:///hostfs/run/user/${REMOTE_UID}/docker.sock' --description '$USERNAME'" >/dev/null 2>&1 || \
                echo "Failed context for $USERNAME"
        fi

        echo "Configuring docker service ..."

        $SSH_CMD "loginctl enable-linger $USERNAME" \
            >/dev/null 2>&1 || echo "Failed to enable linger for $USERNAME"

        $SSH_CMD "machinectl shell ${USERNAME}@ /bin/bash -c 'systemctl --user daemon-reload'" \
            >/dev/null 2>&1 || echo "Failed to reload daemon for $USERNAME"

        $SSH_CMD "machinectl shell ${USERNAME}@ /bin/bash -c 'systemctl --user --quiet restart docker'" \
            >/dev/null 2>&1 || echo "Failed to restart docker for $USERNAME"
    else
        echo "No .docker directory for $USERNAME on source!"
        exit 1
    fi
	# Close FD 3
	exec 3<&-
}

restart_services_on_target() {
            echo "Reloading services on ${REMOTE_HOST} server ..."
	    $SSH_CMD "cd /root && docker compose up -d openpanel bind9 caddy >/dev/null 2>&1 && systemctl restart admin >/dev/null 2>&1"

	if [[ $COMPOSE_START_MAIL -eq 1 ]]; then
            echo "Reloading mailserver and webmail on ${REMOTE_HOST} server ..."
            $SSH_CMD "cd /usr/local/mail/openmail && docker --context default compose up -d mailserver roundcube >/dev/null 2>&1"  
	fi

	#todo: ftp, clamav 
  
}

refresh_quotas() {
            echo "Recalculating disk and inodes usage for all users on ${REMOTE_HOST} ..."
            $SSH_CMD "quotacheck -avm >/dev/null 2>&1 && repquota -u / > /etc/openpanel/openpanel/core/users/repquota"
}



# MAIN
DB_CONFIG_FILE="/usr/local/opencli/db.sh"
. "$DB_CONFIG_FILE"

ssh-keygen -f '/root/.ssh/known_hosts' -R $REMOTE_HOST > /dev/null
format_commands # creates rsync and sshpass commands, installs sshpass if missing
get_server_ipv4 
get_users_count_on_destination

username_exists_count=$(check_username_exists)
if [ "$username_exists_count" -gt 0 ]; then\
    if [[ $FORCE -eq 0 ]]; then
      echo "[✘] Error: Username '$USERNAME' is already taken on destination server."
      exit 1
    else
      echo "[!] Warning: Username '$USERNAME' is already taken on destination server but will be overwritten due to the --force flag."
    fi
fi


export_mysql
import_mysql
copy_user_account $USERNAME
rsync_files_for_user
copy_docker_context # create context on dest, start service
$SSH_CMD "systemctl daemon-reload" 
$SSH_CMD "opencli user-quota $USERNAME >/dev/null 2>&1" # set quotas
$SSH_CMD "mkdir -p /var/log/caddy/stats/ /var/log/caddy/domlogs/ /var/log/caddy/coraza_waf/ /etc/openpanel/caddy/domains/ /etc/bind/zones/ /etc/openpanel/caddy/ssl/certs/ /etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/ /etc/openpanel/openpanel/core/users/"

# todo: folder just for that user!
if [ -n "$key_value" ]; then
  echo "Syncing /var/mail ..."
  eval $RSYNC_CMD /usr/local/mail/openmail ${REMOTE_USER}@${REMOTE_HOST}:/usr/local/mail/openmail
  COMPOSE_START_MAIL=1
fi

# logs and stuff
eval $RSYNC_CMD /etc/openpanel/openpanel/core/users/$USERNAME/ ${REMOTE_USER}@${REMOTE_HOST}:/etc/openpanel/openpanel/core/users/$USERNAME

# todo: mysql data for user!!!!!


store_running_containers_for_user         # export running contianers on source and copy to dest
restore_running_containers_for_user       # start containers on dest
restart_services_on_target                # restart openpanel, webserver and admin on dest
refresh_quotas                            # recalculate user usage on dest

echo "[✔] Transfer for user $USERNAME complete"
