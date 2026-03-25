#!/bin/bash
pid=$$
script_dir=$(dirname "$0")
timestamp="$(date +'%Y-%m-%d_%H-%M-%S')"
start_time=$(date +%s)
DEBUG=true



# ======================================================================
# START HELPER FUNCTIONS

usage() {
    echo "Usage: $0 --backup-location='<path>' --plan-name='<plan_name>' [--dry-run]"
    echo
    echo "Example: $0 --backup-location='/home/backup-7.29.2024_13-22-32_pejcic.tar.gz' --plan-name='Standard plan' --dry-run"
    exit 1
}

log() {
    if [[ -z "${1// /}" ]]; then
        return
    fi
    local timestamp=$(date +'%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] $1" | tee -a "$LOG_FILE"
}

debug_log() {
    if [ "$DEBUG" = true ]; then
        log "DEBUG: $1"
    fi
}

dry_run() {
    local msg="$1"
    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: $msg"
        return 0
    fi
    return 1
}

handle_error() {
    log "FATAL ERROR: An error occurred in function '$1' on line $2"
    cleanup
    exit 1
}

trap 'handle_error "${FUNCNAME[-1]}" "$LINENO"' ERR

cleanup() {
    log "Cleaning up temporary files and directories"
    rm -rf "$backup_dir"
}

define_data_and_log(){
    local backup_location=""
    plan_name=""
    DRY_RUN=false

	strip_quotes() {
	    local str="$1"
	    str="${str%\"}"  # remove trailing double quote
	    str="${str#\"}"  # remove leading double quote
	    str="${str%\'}"  # remove trailing single quote
	    str="${str#\'}"  # remove leading single quote
	    echo "$str"
	}
	
	for arg in "$@"; do
	    case $arg in
	        --backup-location=*) backup_location=$(strip_quotes "${arg#*=}") ;;
	        --plan-name=*)       plan_name=$(strip_quotes "${arg#*=}") ;;
	        --post-hook=*)       post_hook=$(strip_quotes "${arg#*=}") ;;
	        --dry-run)           DRY_RUN=true ;;
	        *)                   usage ;;
	    esac
	done

	[[ -z "$backup_location" || -z "$plan_name" ]] && usage

    base_name="$(basename "$backup_location")"
    base_name_no_ext="${base_name%.*}"
    local log_dir="/var/log/openpanel/admin/imports"
    mkdir -p $log_dir
    LOG_FILE="$log_dir/${base_name_no_ext}_${timestamp}.log"
    echo "Import started, log file: $LOG_FILE"
    main
}

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

install_dependencies() {
    log "Checking and installing dependencies..."
    install_needed=false
    commands=(tar unzip jq pigz wget curl rsync)

    for cmd in "${commands[@]}"; do
        command_exists "$cmd" || { install_needed=true; break; }
    done

    if [ "$install_needed" = true ]; then
        if command_exists apt-get; then
            log "Detected APT package manager. Updating..."
            apt-mark hold linux-image-generic linux-headers-generic >/dev/null 2>&1
            apt-get update -y >/dev/null 2>&1
            for cmd in "${commands[@]}"; do
                if ! command_exists "$cmd"; then
                    log "Installing $cmd (APT)"
                    apt-get install -y --no-upgrade --no-install-recommends "$cmd" >/dev/null 2>&1
                fi
            done
            apt-mark unhold linux-image-generic linux-headers-generic >/dev/null 2>&1
        elif command_exists dnf; then
            log "Detected DNF package manager. Updating..."
            dnf -y makecache >/dev/null 2>&1
            for cmd in "${commands[@]}"; do
                if ! command_exists "$cmd"; then
                    log "Installing $cmd (DNF)"
                    dnf install -y "$cmd" >/dev/null 2>&1
                fi
            done
        else
            log "Error: Unsupported package manager. Please install dependencies manually."
            exit 1
        fi
        log "Dependencies installed successfully."
    else
        log "All required dependencies are already installed."
    fi
}

get_server_ipv4(){
    new_ip=$(curl --silent --max-time 2 -4 https://ip.openpanel.com || curl --silent --max-time 2 -4 https://ifconfig.me)
    if [ -z "$new_ip" ]; then
        new_ip=$(ip addr|grep 'inet '|grep global|head -n1|awk '{print $2}'|cut -f1 -d/)
    fi
}

validate_plan_exists(){
	opencli plan-list --json | grep -qw "$plan_name" || { log "FATAL ERROR: Plan name '$plan_name' does not exist."; exit 1; }
}

# END HELPER FUNCTIONS
# ======================================================================



# ======================================================================
# CHECK BACKUP FILE EXTENSION AND DETERMINE SIZE NEEDED FOR RESTORE
check_if_valid_cp_backup() {
    local backup_location="$1"
    local backup_filename
    backup_filename=$(basename "$backup_location")

    ARCHIVE_SIZE=$(stat -c%s "$backup_location")

    extraction_command="tar -xzf"
    multiplier=2

    case "$backup_filename" in
        cpmove-*.tar.gz) log "Identified cpmove backup" ;;
        backup-*.tar.gz|*.tar.gz) log "Identified gzipped tar backup" ;;
        *.tgz) log "Identified tgz backup"; multiplier=3 ;;
        *.tar) log "Identified tar backup"; extraction_command="tar -xf"; multiplier=3 ;;
        *.zip) log "Identified zip backup"; extraction_command="unzip"; multiplier=3 ;;
        *) log "FATAL ERROR: Unrecognized backup format: $backup_filename"; exit 1 ;;
    esac

    EXTRACTED_SIZE=$((ARCHIVE_SIZE * multiplier))
}

# ======================================================================
# CHECK AVAILABLE DISK SPACE ON THE OPENPANEL SERVER
check_if_disk_available(){
    TMP_DIR="/tmp"
    HOME_DIR="/home"
    AVAILABLE_TMP=$(df -B1 --output=avail "$TMP_DIR" | tail -n 1)
    AVAILABLE_HOME=$(df -B1 --output=avail "$HOME_DIR" | tail -n 1)

    if [[ $AVAILABLE_TMP -ge $EXTRACTED_SIZE && $AVAILABLE_HOME -ge $EXTRACTED_SIZE ]]; then
        log "There is enough disk space to extract the archive and copy it to the home directory."
    else
        log "FATAL ERROR: Not enough disk space."
        if [[ $AVAILABLE_TMP -lt $EXTRACTED_SIZE ]]; then
            log "Insufficient space in the '/tmp' partition."
            log "Available: $AVAILABLE_TMP - Needed: $EXTRACTED_SIZE"
        fi
        if [[ $AVAILABLE_HOME -lt $EXTRACTED_SIZE ]]; then
            log "Insufficient space in the '/home' directory."
            log "Available: $AVAILABLE_HOME - Needed: $EXTRACTED_SIZE"
        fi
        exit 1
    fi
}

# ======================================================================
# EXTRACT BACKUP TO TMP LOCATION
extract_cpanel_backup() {
    backup_location="$1"
    backup_dir="$2"
    backup_dir="${backup_dir%.*}"
    log "Extracting backup from $backup_location to $backup_dir"
    mkdir -p "$backup_dir"

    if [ "$extraction_command" = "unzip" ]; then
        $extraction_command "$backup_location" -d "$backup_dir"
    elif [ "$extraction_command" = "tar -xzf" ]; then
        backup_size=$(stat -c %s "${backup_location}")
        zero_one_percent=$((backup_size / 1000000))
        tar --use-compress-program=pigz \
            --checkpoint="$zero_one_percent" \
            --checkpoint-action=dot \
            -xf "$backup_location" \
            -C "$backup_dir" 
    else
        $extraction_command "$backup_location" -C "$backup_dir"
    fi
    
    if [ $? -eq 0 ]; then
        log "Backup extracted successfully."
        log "Extracted backup folder: $real_backup_files_path"
    else
        log "FATAL ERROR: Backup extraction failed."
        cleanup
        exit 1
    fi
}

# ======================================================================
# LOCATE FILES IN EXTRACTED BACKUP
locate_backup_directories() {
    log "Locating important files in the extracted backup"

    homedir=$(find "$backup_dir" -type d -name "homedir" | head -n 1)
    if [ -z "$homedir" ]; then
        homedir=$(find "$backup_dir" -type d -name "public_html" -printf '%h\n' | head -n 1)
    fi
	[[ -n $homedir ]] || { log "FATAL ERROR: Unable to locate home directory in the backup"; exit 1; }

    mysqldir="$real_backup_files_path/mysql"
	[[ -d $mysqldir ]] || log "WARNING: Unable to locate MySQL directory in the backup"

    mysql_conf="$real_backup_files_path/mysql.sql"
	[[ -f $mysql_conf ]] || log "WARNING: Unable to locate MySQL grants file in the backup"

    psqldir="$real_backup_files_path/psql"
    psql_grants="$real_backup_files_path/psql_grants.sql"
    psql_users="$real_backup_files_path/psql_users.sql"
    ftp_conf="$real_backup_files_path/proftpdpasswd"

    domain_logs="$real_backup_files_path/logs/"
	[[ -d $domain_logs ]] || log "WARNING: Unable to locate apache domlogs in the backup"

    cp_file="$real_backup_files_path/cp/$cpanel_username"
	[[ -f $cp_file ]] || { log "FATAL ERROR: Unable to locate cp/$cpanel_username file in the backup"; exit 1; }

    log "Backup directories and configuration files located:"
    log "- Home directory:       $homedir"
    log "- MySQL directory:      $mysqldir"
    log "- MySQL grants:         $mysql_conf"
    [[ -d $psqldir ]] && log "- PostgreSQL directory: $psqldir"
	[[ -f $psql_grants ]] && log "- PostgreSQL grants:    $psql_grants"
	[[ -f $psql_users ]] && log "- PostgreSQL users:     $psql_users"
	[[ -f $ftp_conf ]] && log "- PureFTPD users:       $ftp_conf"
    log "- Domain logs:          $domain_logs"
    log "- cPanel configuration: $cp_file"
}

get_mariadb_or_mysql_for_user() {
    mysql_type=$(grep '^MYSQL_TYPE=' /home/$cpanel_username/.env | cut -d '=' -f2 | tr -d '"')
}

reload_user_quotas() {
	nohup bash -c 'quotacheck -avm && repquota -u / > /etc/openpanel/openpanel/core/users/repquota' >/dev/null 2>&1 &
	disown
}

collect_stats() {
    nohup bash -c "opencli docker-collect_stats '$cpanel_username'" >/dev/null 2>&1 &
	disown
}

# ======================================================================
# PARSE CPANEL BACKUP METADATA FOR ACCOUNT AND SERVICE INFORMATION
parse_cpanel_metadata() {
    log "Starting to parse cPanel metadata..."

    cp_file="${real_backup_files_path}/cp/${cpanel_username}"

    if [ ! -f "$cp_file" ]; then
        log "WARNING: cp file $cp_file not found. Using default values."
        main_domain=""
        cpanel_email=""
        php_version="inherit"
    else

        get_cp_value() {
            local key="$1"
            local default="$2"
            local value
            value=$(grep "^$key=" "$cp_file" | cut -d'=' -f2-)
            if [ -z "$value" ]; then
                echo "$default"
            else
                echo "$value"
            fi
        }

		# TODO: reuse for webhook_url in openpanel
        PUSHBULLET_ACCESS_TOKEN=$(get_cp_value "PUSHBULLET_ACCESS_TOKEN" "")

        main_domain=$(get_cp_value "DNS" "")
        cpanel_email=$(get_cp_value "CONTACTEMAIL" "")
        [ -z "$cpanel_email" ] && cpanel_email=$(get_cp_value "CONTACTEMAIL2" "")
        [ -z "$cpanel_email" ] && cpanel_email=$(get_cp_value "EMAIL" "")

        # Cloudlinux PHP Selector
        cfg_file="${real_backup_files_path}/homedir/.cl.selector/defaults.cfg"
        if [ -f "$cfg_file" ]; then
            php_version=$(grep '^php\s*=' "$cfg_file" | awk -F '=' '{print $2}' | tr -d '[:space:]')
            [ -z "$php_version" ] && php_version="inherit"
        else
            php_version="inherit"
        fi

        ip_address=$(get_cp_value "IP" "")
        plan=$(get_cp_value "PLAN" "default")
        max_addon=$(get_cp_value "MAXADDON" "0")
        max_ftp=$(get_cp_value "MAXFTP" "unlimited")
        max_sql=$(get_cp_value "MAXSQL" "unlimited")
        max_pop=$(get_cp_value "MAXPOP" "unlimited")
        max_sub=$(get_cp_value "MAXSUB" "unlimited")

        log "Additional metadata parsed:"
        log "IP Address:           $ip_address"
        log "Plan:                 $plan"
        log "Max Addon Domains:    $max_addon"
        log "Max FTP Accounts:     $max_ftp"
        log "Max SQL Databases:    $max_sql"
        log "Max Email Accounts:   $max_pop"
        log "Max Subdomains:       $max_sub"
    fi

    main_domain="${main_domain:-}"
    cpanel_email="${cpanel_email:-admin@$main_domain}"
    php_version="${php_version:-inherit}"

    log "Email:                ${cpanel_email:-Not found}"
    log "PHP Version:          $php_version"
    log "Finished parsing cPanel metadata."
}

# ======================================================================
# CHECK USERNAME AVIABILITY BEFORE SGTARTING THE EXPORT PROCESS
check_if_user_exists(){
    backup_filename=$(basename "$backup_location")
    cpanel_username="${backup_filename##*_}"
    cpanel_username="${cpanel_username%%.*}"
    log "Username: $cpanel_username"

    local existing_user=""
	existing_user=$(opencli user-list --json | jq -r '.data[] | select(.username == "'"$cpanel_username"'") | .id')

	if [ -n "$existing_user" ]; then
        log "FATAL ERROR: $cpanel_username already exists."
        exit 1
	fi

    log "Username $cpanel_username is available"
	log "Starting import process.."
}

# ======================================================================
# CREATE OPENPANEL USER
create_new_user() {
    local username="$1"
    local email="$3"
    local plan_name="$4"

    dry_run "Would create user $username with email $email and plan $plan_name" && return

    create_user_command=$(opencli user-add "$cpanel_username" generate "$email" "$plan_name" --no-sentinel >/dev/null 2>&1)
    while IFS= read -r line; do
        log "$line"
    done <<< "$create_user_command"

	if [ -f "/home/$cpanel_username/docker-compose.yml" ]; then
        shadow_file="$real_backup_files_path/shadow"
        if [ -f "$shadow_file" ]; then
            . /usr/local/opencli/db.sh

            hashed_password=$(cat "$shadow_file")
            safe_hashed_password=$(printf "%s" "$hashed_password" | sed "s/'/''/g")
            safe_username=$(printf "%s" "$username" | sed "s/'/''/g")
            mysql_query="UPDATE users SET password='$safe_hashed_password' WHERE username='$safe_username';"
            mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$mysql_query"
            if [ $? -eq 0 ]; then
                echo "Imported SHA-512 crypt password hash from cpanel (will be automatically converted to pbkdf2:sha256 on first user login)"
            else
                echo "Failed to import SHA-512 crypt password hash from cpanel"
            fi
        fi       
    else
        log "FATAL ERROR: User addition failed."
        exit 1
    fi
}

# ======================================================================
# PHP VERSION
restore_php_version() {
    local php_version="$1"

    dry_run "Would set default PHP version $php_version for user $cpanel_username" && return

    # 'inherit' = OpenPanel default
    if [ "$php_version" == "inherit" ]; then
        log "PHP version is set to inherit. No changes will be made."
    else
        # cPanel custom version
        #log "Setting PHP $php_version as the default version for all new domains."
        output=$(opencli php-default "$cpanel_username" --update "$php_version" 2>&1)
        while IFS= read -r line; do
            log "$line"
        done <<< "$output"
    fi
}

# ======================================================================
# POSTGRESQL
restore_psql() {
	local psql_dir="$1"
    log "Restoring PostgreSQL databases"
    dry_run "Would restore PostgreSQL databases for user $cpanel_username" && return

	if [ -d "$psql_dir" ]; then
        # STEP 1: Start psql container
        log "Initializing postgres service for user"
        cd "/home/$cpanel_username/" && docker --context="$cpanel_username" compose up -d postgres >/dev/null 2>&1

        # STEP 2: Wait for PostgreSQL to be ready (max 300 seconds)
		local max_wait=300
        log "Waiting for PostgreSQL service to start... (max {$max_wait}s)"		
        waited=0
		while ! docker --context="$cpanel_username" exec postgres psql -U postgres -d "$postgres" -c "SELECT 1;" >/dev/null 2>&1; do
            sleep 2
            waited=$((waited + 2))
            if [ "$waited" -ge "$max_wait" ]; then
                log "ERROR: postgres did not respond to 'SELECT 1' after $max_wait seconds - no database or users are imported"			
                return 1
            fi
        done
        log "postgres is ready after $waited seconds"

        # STEP 3: Create and import databases
        total_databases=$(ls "$psql_dir"/*.tar 2>/dev/null | wc -l)
        log "Starting import for $total_databases PostgreSQL databases"
        if [ "$total_databases" -gt 0 ]; then
            current_db=1
            for db_file in "$psql_dir"/*.tar; do
                db_name=$(basename "$db_file" .tar)

                log "Creating database: $db_name (${current_db}/${total_databases})"
				docker --context="$cpanel_username" exec postgres psql -U postgres -c "CREATE DATABASE \"$db_name\";" #>/dev/null 2>&1

                log "Importing tables for database: $db_name"
                docker --context="$cpanel_username" cp "${real_backup_files_path}/psql/$db_name.tar" "$mysql_type:/tmp/${db_name}.tar" >/dev/null 2>&1		
				docker --context="$cpanel_username" exec postgres bash -c "cd /tmp && tar -xf /tmp/${db_name}.tar restore.sql && psql -U postgres -d \"$db_name\" -f restore.sql && rm restore.sql /tmp/${db_name}.tar"
                current_db=$((current_db + 1))
            done
            log "Finished processing $((current_db - 1)) databases"
        else
            log "WARNING: No PostgreSQL databases found"
        fi

		# psql_users.sql
		if [[ -f "$psql_users" && -s "$psql_users" ]]; then
			log "Importing database users"
		    while IFS= read -r line || [[ -n "$line" ]]; do
		        [[ -z "$line" ]] && continue
		        docker --context="$cpanel_username" exec postgres psql -U postgres -d postgres -c "$line"
		    done < "$psql_users"
		else
		    echo "WARNING: No PostgreSQL USERS detected"
		fi

        # psql_grants.sql
		if [[ -f "$psql_grants" && -s "$psql_grants" ]]; then
			log "Importing database grants"
		    while IFS= read -r line || [[ -n "$line" ]]; do
		        [[ -z "$line" ]] && continue
		        docker --context="$cpanel_username" exec postgres psql -U postgres -d postgres -c "$line"
		    done < "$psql_grants"
		else
		    echo "WARNING: No PostgreSQL GRANTS detected"
		fi

    else
        log "No PostgreSQL databases found to restore"
    fi


}

# ======================================================================
# MYSQL
restore_mysql() {
    local mysql_dir="$1"
    local sandbox_warning_logged=false

    log "Restoring MySQL databases"
    dry_run "Would restore MySQL databases for user $cpanel_username" && return

    # Workaround for MariaDB sandbox mode bug
    apply_sandbox_workaround() {
        local db_file="$1"
        local text_to_check='enable the sandbox mode'
        local first_line

        first_line=$(head -n 1 "${real_backup_files_path}/mysql/$db_file")
        if echo "$first_line" | grep -q "$text_to_check"; then
            if [ "$sandbox_warning_logged" = false ]; then
                log "WARNING: Database dumps were created on a MariaDB server with '--sandbox' mode. Applying workaround for backwards compatibility to MySQL (BUG: https://jira.mariadb.org/browse/MDEV-34183)"
                sandbox_warning_logged=true
            fi
            tail -n +2 "${real_backup_files_path}/mysql/$db_file" > "${real_backup_files_path}/mysql/${db_file}.workaround" && \
            mv "${real_backup_files_path}/mysql/${db_file}.workaround" "${real_backup_files_path}/mysql/$db_file"
        fi
    }

	# STEP 1: Enable remote access
	if [ -f "${real_backup_files_path}/mysql_host_notes.json" ]; then
		key_count=$(jq 'keys | length' "${real_backup_files_path}/mysql_host_notes.json")
		if [ "$key_count" -gt 0 ]; then
			log "Enabling remote MySQL access ($key_count hosts detected)"
			sed -i '/^MYSQL_PORT=/ s/127\.0\.0\.1://g' /home/$cpanel_username/.env
		fi
	fi

    if [ -d "$mysql_dir" ]; then
        # STEP 2: Replace old IP and hostname
        old_ip=$(grep -oP 'IP=\K[0-9.]+' "${real_backup_files_path}/cp/$cpanel_username")
        log "Replacing old server IP: $old_ip with '%' in database grants"
        sed -i "s/$old_ip/%/g" "$mysql_conf"

        old_hostname=$(cat "${real_backup_files_path}/meta/hostname")
        log "Removing old hostname $old_hostname from database grants"
        sed -i "/$old_hostname/d" "$mysql_conf"
        
        # STEP 3: Start MySQL container
        if [ "$mysql_type" = "mysql" ]; then
            mysql_version="8.0"
            sed -i 's/^MYSQL_VERSION=.*/MYSQL_VERSION="8.0"/' /home/"$cpanel_username"/.env
        fi
        log "Initializing $mysql_type service for user"
        cd "/home/$cpanel_username/" && docker --context="$cpanel_username" compose up -d "$mysql_type" >/dev/null 2>&1

        # STEP 4: Wait for MySQL to be ready (max 300 seconds)
		local max_wait=300
        log "Waiting for MySQL service to start... (max {$max_wait}s)"
        waited=0
        while ! docker --context="$cpanel_username" exec "$mysql_type" $mysql_type -e "SELECT 1" >/dev/null 2>&1; do
            sleep 2
            waited=$((waited + 2))
            if [ "$waited" -ge "$max_wait" ]; then
                log "ERROR: $mysql_type did not respond to 'SELECT 1' after $max_wait seconds - no database or users are imported"
                return 1
            fi
        done
        log "$mysql_type is ready after $waited seconds"

        # STEP 5: Create and import databases
        total_databases=$(ls "$mysql_dir"/*.create 2>/dev/null | wc -l)
        log "Starting import for $total_databases MySQL databases"
        if [ "$total_databases" -gt 0 ]; then
            current_db=1
            for db_file in "$mysql_dir"/*.create; do
                db_name=$(basename "$db_file" .create)

                log "Creating database: $db_name (${current_db}/${total_databases})"
                if [ "$mysql_type" = "mysql" ]; then
					apply_sandbox_workaround "$db_name.create"
				fi
                docker --context="$cpanel_username" cp "${real_backup_files_path}/mysql/$db_name.create" "$mysql_type:/tmp/${db_name}.create" >/dev/null 2>&1
                docker --context="$cpanel_username" exec "$mysql_type" bash -c "$mysql_type < /tmp/${db_name}.create && rm /tmp/${db_name}.create"

                log "Importing tables for database: $db_name"
				if [ "$mysql_type" = "mysql" ]; then
                	apply_sandbox_workaround "$db_name.sql"
				fi
                docker --context="$cpanel_username" cp "${real_backup_files_path}/mysql/$db_name.sql" "$mysql_type:/tmp/${db_name}.sql" >/dev/null 2>&1
                docker --context="$cpanel_username" exec "$mysql_type" bash -c "$mysql_type $db_name < /tmp/${db_name}.sql && rm /tmp/${db_name}.sql"

                current_db=$((current_db + 1))
            done
            log "Finished processing $((current_db - 1)) databases"
        else
            log "WARNING: No MySQL databases found"
        fi

        # STEP 6: Import grants
        log "Importing database grants"
        python3 "$script_dir/mysql/json_2_sql.py" "${real_backup_files_path}/mysql.sql" "${real_backup_files_path}/mysql.TEMPORARY.sql" >/dev/null 2>&1

        docker --context="$cpanel_username" cp "${real_backup_files_path}/mysql.TEMPORARY.sql" "$mysql_type:/tmp/mysql.TEMPORARY.sql" >/dev/null 2>&1
        docker --context="$cpanel_username" exec "$mysql_type" bash -c "$mysql_type < /tmp/mysql.TEMPORARY.sql && $mysql_type -e 'FLUSH PRIVILEGES;' && rm /tmp/mysql.TEMPORARY.sql"

		# STEP 7: Start phpMyAdmin
		log "Starting phpMyAdmin service"
		nohup cd "/home/$cpanel_username/" && docker --context="$cpanel_username" compose up -d phpmyadmin >/dev/null 2>&1 &
		disown

    else
        log "No MySQL databases found to restore"
    fi
}

# ======================================================================
# SSL CERTIFICATES
restore_ssl() {
    local username="$1"

    dry_run "Would restore SSL certificates for user $username" && return

    # TODO: edit to cover certs/ keys/ 
    log "Restoring SSL certificates"
    # apache_tls/ dir has LE certs, custom are in ssl/
    if [ -d "$real_backup_files_path/ssl" ]; then
        dest_dir="/home/$username/docker-data/volumes/${username}_html_data/_data/"
		shopt -s nullglob
        for cert_file in "$real_backup_files_path/ssl"/*.crt; do
            local domain=$(basename "$cert_file" .crt)
            local key_file="$real_backup_files_path/ssl/$domain.key"
            local new_cert_file="$dest_dir/$domain.crt"
            local new_key_file="$dest_dir/$domain.key"
            if [ -f "$key_file" ]; then
                cp "$key_file" "$new_key_file"
                cp "$cert_file" "$new_cert_file"         
			    if [[ "$domain" == *.cpanel.site ]]; then
			        echo "Skipping cPanel temporary domain: $domain"
			        continue
			    fi
				log "Installing SSL certificate for domain: $domain"
                opencli domains-ssl "$domain" custom "/var/www/html/$domain.key" "/var/www/html/$domain.crt"
				rm -rf "$new_cert_file" "$new_key_file"
            else
                log "SSL key file not found for domain: $domain"
            fi
        done

        nohup docker --context=default exec caddy caddy reload --config /etc/caddy/Caddyfile > /dev/null 2>&1 &
        disown

    else
        log "No SSL certificates found to restore"
    fi
}

# ======================================================================
# DNS ZONES
restore_dns_zones() {
    log "Restoring DNS zones"

    dry_run "Would restore DNS zones for user $cpanel_username" && return

    if [ -d "$real_backup_files_path/dnszones" ]; then
        for zone_file in "$real_backup_files_path/dnszones"/*; do
            local zone_name=$(basename "${zone_file%.db}")

            # Check if the destination zone file exists, if not, it was probably a subdomain that had no dns zone
            if [ ! -f "/etc/bind/zones/${zone_name}.zone" ]; then
                log "DNS zone file /etc/bind/zones/${zone_name}.zone does not exist. Skipping import for $zone_name."
                continue
			elif [[ "$zone_name" == *.cpanel.site ]]; then
			    log "Skipping cPanel temporary domain: $zone_name"
			    continue
            fi

            old_ip=$(grep -oP 'IP=\K[0-9.]+' ${real_backup_files_path}/cp/$cpanel_username)
            if [ -z "$old_ip" ]; then
                log "WARNING: old server ip address not detected in file ${real_backup_files_path}/cp/$cpanel_username - records will not be automatically updated to new ip address."
            else
                sed -i "s/$old_ip/$new_ip/g" $zone_file
            fi
            temp_file_of_original_zone=$(mktemp)
            temp_file_of_created_zone=$(mktemp)

            awk '/^@/ { found=1; last_line=NR } { if (found && NR > last_line) exit } { print }' "$zone_file" > "$temp_file_of_original_zone"
            awk '/NS/ { found=1; next } found { print }' "/etc/bind/zones/${zone_name}.zone" > "$temp_file_of_created_zone"
            cat "$temp_file_of_created_zone" >> "$temp_file_of_original_zone"
            mv "$temp_file_of_original_zone" "/etc/bind/zones/${zone_name}.zone"
            rm "$temp_file_of_created_zone"

            opencli domains-update_ns ${zone_name} >/dev/null 2>&1
            log "DNS zone file for $zone_name has been imported."
        done
    else
        log "No DNS zones found to restore"
    fi
}

create_home_mountpoint() {
    dry_run "Would create a symlink from html_data volume to /home/$cpanel_username/" && return
    
sed -i '/^[[:space:]]*volumes:[[:space:]]*$/{
  N
  /- html_data:\/var\/www\/html\// s|$|\n      - html_data:/home/${CONTEXT}/|
}' /home/$cpanel_username/docker-compose.yml

}

# ======================================================================
# HOME DIR
restore_files() {
    dry_run "Would restore files from /home/$cpanel_username/ to html_data volume" && return

    du_needed_for_home=$(du -sh "$real_backup_files_path/homedir" | cut -f1)
    log "Restoring home directory ($du_needed_for_home) to html_data volume"
    mkdir -p /home/$cpanel_username/docker-data/volumes/${cpanel_username}_html_data/
    rm -rf "$real_backup_files_path"/homedir/{.cpanel,.trash,wordpress-backups}
    mv $real_backup_files_path/homedir /home/$cpanel_username/docker-data/volumes/${cpanel_username}_html_data/_data
}

# ======================================================================
# PERMISSIONS
fix_perms(){
    local verbose="" #-v
    log "Changing permissions for all files and folders in user home directory /home/$cpanel_username/"

    dry_run "Would change permissions with command: find /home/$cpanel_username -print0 | xargs -0 chown $verbose $cpanel_username:$cpanel_username" && return
    
    if ! timeout 600 find /home/$cpanel_username -print0 | xargs -0 chown $verbose $cpanel_username:$cpanel_username > /dev/null 2>&1; then
        if [ $? -eq 124 ]; then
            log "ERROR: Timeout reached while changing permissions (10 minutes)."
        else
            log "ERROR: Failed to change permissions."
        fi
            log "       Make sure to change permissions manually from terminal with: find /home/$cpanel_username -print0 | xargs -0 chown -v $cpanel_username:$cpanel_username"
    fi
    
}

# ======================================================================
# WORDPRESS SITES
restore_wordpress() {
    local real_backup_files_path="$1"
    local username="$2"

    dry_run "Would restore WordPress sites for user $username" && return

    log "Checking user files for WordPress installations to add to Site Manager interface.."
    output=$(opencli websites-scan $cpanel_username)
        while IFS= read -r line; do
            log "$line"
        done <<< "$output"    
}

# LOCAE
restore_locale() {
    if [ -f "$real_backup_files_path/cp/$openpanel_username" ]; then
        local file_path="$real_backup_files_path/cp/$openpanel_username"
    	local locale_code=$(grep '^LOCALE=' "$file_path" | cut -d'=' -f2 | cut -c1-2 | tr '[:upper:]' '[:lower:]')
	    if [ "$locale_code" = "en" ] || [ -d "/etc/openpanel/openpanel/translations/$locale_code" ]; then
			log "Setting locale:$locale_code for OpenPanel UI"
			echo "$locale_code" > /home/$openpanel_username/locale
		else
			log "Skipping user locale: '$locale_code' is not available"
		fi
	fi
}

# ======================================================================
# DOMAINS
restore_domains() {
    if [ -f "$real_backup_files_path/userdata/main" ]; then
        file_path="$real_backup_files_path/userdata/main"
        main_domain=""
        parked_domains=""
        sub_domains=""
        addon_domains=""

		parked_domains_section=false
		sub_domains_section=false
		addon_domains_section=false


        while IFS= read -r line; do
            if [[ "$line" =~ ^main_domain: ]]; then
                main_domain=$(echo "$line" | awk '{print $2}')
			elif [[ "$line" =~ ^parked_domains: ]]; then
			    parked_domains_section=true
			    sub_domains_section=false
			    addon_domains_section=false
			    continue
			elif [[ "$line" =~ ^sub_domains: ]]; then
			    sub_domains_section=true
			    parked_domains_section=false
			    addon_domains_section=false
			    continue
			elif [[ "$line" =~ ^addon_domains: ]]; then
			    addon_domains_section=true
			    parked_domains_section=false
			    sub_domains_section=false
			    continue
            fi

            if [[ "$sub_domains_section" == true ]]; then
                if [[ "$line" =~ ^[[:space:]]+- ]]; then
                    sub_domains+=$(echo "$line" | awk '{print $2}')$'\n'
                else
                    sub_domains_section=false
                fi
            fi

            if [[ "$parked_domains_section" == true ]]; then
                if [[ "$line" =~ ^[[:space:]]+- ]]; then
                    parked_domains+=$(echo "$line" | awk '{print $2}')$'\n'
                else
                    parked_domains_section=false
                fi
            fi

            if [[ "$addon_domains_section" == true ]]; then
                if [[ "$line" =~ ^[[:space:]]*[^:]+:[[:space:]]*[^[:space:]]+$ ]]; then
                    domain=$(echo "$line" | awk -F: '{print $1}' | tr -d '[:space:]')
                    if [[ -n "$domain" && "$domain" != "main_domain" && "$domain" != "parked_domains" ]]; then
                        addon_domains+="$domain"$'\n'
                    fi
                else
                    addon_domains_section=false
                fi
            fi
        done < "$file_path"

        sub_domains_array=()
        addon_domains_array=()
		parked_domains_array=()

        # Parse sub_domains
        while IFS= read -r domain; do
            if [[ -n "$domain" ]]; then
                sub_domains_array+=("$domain")
            fi
        done <<< "$sub_domains"

        # Parse addon_domains
        while IFS= read -r domain; do
            if [[ -n "$domain" ]]; then
                addon_domains_array+=("$domain")
            fi
        done <<< "$addon_domains"

        # Parse parked_domains
        while IFS= read -r domain; do
            if [[ -n "$domain" ]]; then
                parked_domains_array+=("$domain")
            fi
        done <<< "$parked_domains"

        # Filter out subdomains that are essentially addon_domain.$main_domain
        filtered_sub_domains=()
        for sub_domain in "${sub_domains_array[@]}"; do
            trimmed_sub_domain=$(echo "$sub_domain" | xargs)
            is_addon=false
            for addon in "${addon_domains_array[@]}"; do
                if [[ "$trimmed_sub_domain" == "$addon.$main_domain" ]]; then
                    is_addon=true
                    break
                fi
            done
            if [ "$is_addon" = false ]; then
                filtered_sub_domains+=("$trimmed_sub_domain")
            fi
        done

		main_domain_count=1
		log "Primary domain ($main_domain_count): $main_domain"

        addon_domains_count=${#addon_domains_array[@]}
        if [ "${#addon_domains_array[@]}" -eq 1 ] && [ -z "${addon_domains_array[0]}" ]; then
            addon_domains_count=0
            log "No addon domains detected."
        else
            log "Addon domains  ($addon_domains_count): ${addon_domains_array[@]}"
        fi

        parked_domains_count=${#parked_domains_array[@]}
        if [ "${#parked_domains_array[@]}" -eq 1 ] && [ -z "${parked_domains_array[0]}" ]; then
            parked_domains_count=0
            log "No parked domains detected."
        else
            log "Parked domains ($parked_domains_count): ${parked_domains_array[@]}"
        fi

        filtered_sub_domains_count=${#filtered_sub_domains[@]}
        if [ "${#filtered_sub_domains[@]}" -eq 1 ] && [ -z "${filtered_sub_domains[0]}" ]; then
            filtered_sub_domains_count=0
            log "No subdomains detected."
        else
            log "Subdomains     ($filtered_sub_domains_count): ${filtered_sub_domains[@]}"
        fi


        domains_total_count=$((main_domain_count + addon_domains_count + parked_domains_count + filtered_sub_domains_count))

        log "Detected a total of $domains_total_count domains"

        current_domain_count=0

        create_domain(){
            domain="$1"
            type="$2"

            current_domain_count=$((current_domain_count + 1))

		    if [[ "$domain" == *.cpanel.site ]]; then
		        log "Skipping cPanel temporary domain: $domain"
		        return 0
		    fi
			
            if [[ $domain == \*.* ]]; then
                log "WARNING: Skipping wildcard domain $domain"
				return 0
            fi
			log "Restoring $type $domain (${current_domain_count}/${domains_total_count})"

			local userdata_file="$real_backup_files_path/userdata/$domain"
			local addon_userdata_file="$real_backup_files_path/userdata/$domain.$main_domain"
			file_to_use="$userdata_file"
			[ ! -f "$file_to_use" ] && [ -f "$addon_userdata_file" ] && file_to_use="$addon_userdata_file"
			
			docroot=""
			if [ -f "$file_to_use" ]; then
				original_docroot=$(awk -F': ' '/^documentroot:/ {print $2}' "$file_to_use" | xargs)
			    docroot="${original_docroot#/home/$cpanel_username/}"
			    docroot="/var/www/html/$docroot"
			else
				log "WARNING: userdata file not found for $domain. Using default docroot."
			fi

			dry_run "Would restore $type $domain with --docroot ${docroot:-N/A}" && return

			if opencli domains-whoowns "$domain" | grep -q "not found in the database."; then
				if [ -n "$docroot" ]; then
					output=$(opencli domains-add "$domain" "$cpanel_username" --docroot "$docroot" 2>&1)
					while IFS= read -r line; do
						log "$line"
					done <<< "$output"
				else
					output=$(opencli domains-add "$domain" "$cpanel_username" 2>&1)
					while IFS= read -r line; do
						log "$line"
					done <<< "$output"                        
				fi
				if [ -f "$real_backup_files_path/userdata/$domain" ]; then
					local secruleengineoff=$(grep '^secruleengineoff=' "$real_backup_files_path/userdata/$domain" | cut -d'=' -f2 | cut -c1-2 | tr '[:upper:]' '[:lower:]')
					if [ "$secruleengineoff" == "1" ]; then
						log "Disabling WAF because ModSecurity is turned off for this domain in cPanel."
						output=$(opencli waf domain "$domain" "disable" 2>&1)
						while IFS= read -r line; do
							log "$line"
						done <<< "$output"
					fi
				fi
			else
				log "WARNING: $type $domain already exists and will not be added to this user."
			fi
        }

        log "Processing main (primary) domain.."
        create_domain "$main_domain" "main domain"

        if [ "$parked_domains_count" -eq 0 ]; then
            log "No parked (alias) domains detected."
        else
            log "Processing parked (alias) domains.."
            for parked in "${parked_domains_array[@]}"; do
                create_domain "$parked" "alias domain"
            done
        fi

        if [ "$addon_domains_count" -eq 0 ]; then
            log "No addon domains detected."
        else
            log "Processing addon domains.."
            for addon in "${addon_domains_array[@]}"; do
                create_domain "$addon" "addon domain"
            done
        fi

        if [ "$filtered_sub_domains_count" -eq 0 ]; then
            log "No subdomains detected."
        else
            log "Processing sub-domains.."
            for filtered_sub in "${filtered_sub_domains[@]}"; do
                create_domain "$filtered_sub" "subdomain"
            done
        fi

        log "Finished importing $domains_total_count domains"

    else
        log "FATAL ERROR: domains file userdata/main is missing in backup file."
        exit 1
    fi
}

# ======================================================================
# CRONJOB
restore_cron() {
    log "Restoring cron jobs"
    
    dry_run "Would restore cron jobs for user $cpanel_username" && return

    if [ -f "$real_backup_files_path/cron/$cpanel_username" ]; then
        sed -i '1,2d' "$real_backup_files_path/cron/$cpanel_username"
        ofelia_cron_path="/home/${cpanel_username}/crons.ini"
        > "$ofelia_cron_path"

        job_index=1
        job_found=false
        while IFS= read -r cron_line; do
            [[ -z "$cron_line" || "$cron_line" =~ ^# ]] && continue

            job_found=true
            schedule="* $(echo "$cron_line" | awk '{print $1, $2, $3, $4, $5}')"
            command=$(echo "$cron_line" | cut -d' ' -f6-)

            if [[ "$command" == *mysql* || "$command" == *mariadb* ]]; then
                container_name="$mysql_type"
                comment_prefix=""
            elif [[ "$command" == *php* ]]; then
                container_name="php-fpm-$php_version"
                comment_prefix=""
            else
                container_name=""
                comment_prefix="# "
            fi

            {
                echo "${comment_prefix}[job-exec \"${cpanel_username}_job_$job_index\"]"
                echo "${comment_prefix}schedule = $schedule"
                if [[ -n "$container_name" ]]; then
                    echo "${comment_prefix}container = $container_name"
                fi
                echo "${comment_prefix}command = $command"
                echo
            } >> "$ofelia_cron_path"

            ((job_index++))
        done < "$real_backup_files_path/cron/$cpanel_username"

        if [ "$job_found" = true ]; then
            log "Converted crontab to Ofelia config at: $ofelia_cron_path"
            log "Starting Cron service"
            output=$(cd /home/$cpanel_username && docker --context=$cpanel_username compose up -d cron >/dev/null 2>&1)           
            while IFS= read -r line; do
                log "$line"
            done <<< "$output"
        else
            log "No cron jobs found in file, not starting cron service"
            rm -f "$ofelia_cron_path"
        fi

    else
        log "No cron jobs found to restore"
    fi
}

# ======================================================================
# POST-IMPORT HOOK
run_custom_post_hook() {
    if [ -n "$post_hook" ]; then
        if [ -x "$post_hook" ]; then
            log "Executing post-hook script.."
            "$post_hook" "$cpanel_username"
        else
            log "WARNING: Post-hook file '$post_hook' is not executable or not found."
            exit 1
        fi
    fi
}

create_tmp_dir_and_path() {
    backup_filename="${backup_filename%.*}"
    backup_dir=$(mktemp -d /tmp/cpanel_import_XXXXXX)
    log "Created temporary directory: $backup_dir"
    real_backup_files_path="${backup_dir}/${backup_filename%.*}"
}

success_message() {
    end_time=$(date +%s)
    elapsed=$(( end_time - start_time ))
    hours=$(( elapsed / 3600 ))
    minutes=$(( (elapsed % 3600) / 60 ))
    seconds=$(( elapsed % 60 ))

    log "Elapsed time: ${hours}h ${minutes}m ${seconds}s"

    dry_run "import process for user $cpanel_username completed" && return

    log "SUCCESS: Import for user $cpanel_username completed successfully."

    nohup opencli sentinel --action=user_create --title="User account '$cpanel_username' imported from cPanel backup" --message="User account '$cpanel_username' has been successfully imported from backup file '$backup_filename'" >/dev/null 2>&1 &
	disown
}

log_paths_are() {
    log "Log file: $LOG_FILE"
    log "PID: $pid"
}

start_message() {
    echo -e "
------------------ STARTING CPANEL ACCOUNT IMPORT ------------------
--------------------------------------------------------------------

Currently supported features:

├─ DOMAINS:
│  ├─ Primary domain, Addons, Aliases and Subdomains
│  ├─ DNS zones
│  ├─ SSL certificates
│  ├─ Modsecurity status
│  └─ Access logs (Apache domlogs)
├─ WEBSITES:
│  └─ WordPress instalations from WP Toolkit & Softaculous 
├─ DATABASES:
│    ├─ MySQL databases, users and grants
│    ├─ PostgreSQL databases, users and grants
│    └─ Remote access to MySQL
├─ PHP:
│    └─ Installed version from Cloudlinux PHP Selector
├─ FILES
├─ CRONS
└─ ACCOUNT
    ├─ Account Password
    ├─ Notification preferences
    ├─ Creation date
    └─ Locale (Language)

***ftp accounts and nodejs/python apps are not yet supported***

--------------------------------------------------------------------
  if you experience any errors with this script, please report to
    https://github.com/stefanpejcic/cPanel-to-OpenPanel/issues
--------------------------------------------------------------------
"
}


# ======================================================================
# FTP
ftp_accounts_import() {

    if [ -f "$ftp_conf" ]; then
        log "WARNING: Importing PureFTPD accounts is not yet supported"
        # this is cpanel's format:
        : '
        #cat proftpdpasswd
        pejcic:$6$cv9wnxSLeD1VEk.U$dm84PcqygxOWqT/uyMjrICKUPFeAQwOimJ8frihDCxjRfa1BKf6bnHIhWrbfmLrLn2YBSMnNatW09ZZMAS7GT/:1030:1034:pejcic:/home/pejcic:/bin/bash
        neko@pcx3.com:$6$7GZJXVYlO53hV.M7$750UVg6zKmX.Uj8cmWUxkRnNXxjuZfcm6BxnJceiFD5Zl80sB7jZL0UeHIpw2a3aQRWh.BMH9WuCPdqwj8zxG.:1030:1034:pejcic:/home/pejcic/folder:/bin/ftpsh
        whmcsmybekap@openpanel.co:$6$rDNAW7GZEAJ6zHJm$wYqg.H6USldSPCNz4jbgEi55tJ8hgeDzQCAmhSHfAPyzkJeP1u9E.LaLflQ.7kUbuRtBED7I70.QoCNRlxzEy0:1030:1034:pejcic:/home/pejcic/WHMC_MY_OPENPANEL_DB_BEKAP:/bin/ftpsh
        pejcic_logs:$6$cv9wnxSLeD1VEk.U$dm84PcqygxOWqT/uyMjrICKUPFeAQwOimJ8frihDCxjRfa1BKf6bnHIhWrbfmLrLn2YBSMnNatW09ZZMAS7GT/:1030:1034:pejcic:/etc/apache2/logs/domlogs/pejcic:/bin/ftpsh
        '
    fi
}


# ======================================================================
# EMAILS
import_email_accounts_and_data() {
    local cpanel_username="$1"

    log "Restoring email accounts and mailboxes"

    local base_dir="/home/$cpanel_username/docker-data/volumes/${cpanel_username}_html_data/_data/etc"
	
	dry_run "Would restore email accounts from $base_dir" && return

	postfix_file="/usr/local/mail/openmail/docker-data/dms/config/postfix-accounts.cf"
	if [ ! -f "$postfix_file" ]; then
		log "WARNING: Skipping email imports due to mailserver not configured."
		return 0
	fi

	# 1: check where openadmin is configured to store mail data
	STORE_EMAILS_IN=$(grep -E '^email_storage_location=' /etc/openpanel/openadmin/config/admin.ini | cut -d'=' -f2- | xargs)

	if [[ "$STORE_EMAILS_IN" == /* ]]; then
		: # keep as is
	#elif [[ "$STORE_EMAILS_IN" == "user_dir" ]]; then
	else # TODO: this is a fallback for <1.7.3
		STORE_EMAILS_IN="domain"
	fi


    if [ -f "$real_backup_files_path/cp/$openpanel_username" ]; then
    	local mailbox_format=$(grep '^MAILBOX_FORMAT=' "$real_backup_files_path/cp/$openpanel_username" | cut -d'=' -f2 | cut -c1-2 | tr '[:upper:]' '[:lower:]')
	    if [ "$mailbox_format" == "maildir" ]; then
			:
		elif [ "$mailbox_format" == "mbox" ]; then
			log "WARNING: Emails will not be imported because the cPanel account uses 'mbox' format, while OpenPanel uses 'maildir'. Emails remain available in /var/www/html/mail/."
		else
			log "WARNING: Emails will not be imported because the mailbox format could not be detected. Emails remain available in /var/www/html/mail/."
		fi
	fi



	# Loop through each folder in the base dir
    for dir_path in "$base_dir"/*/; do
	    shadow_file="$dir_path/shadow"
	    if [[ -f "$shadow_file" && -s "$shadow_file" ]]; then
	        domain=$(basename "$dir_path")
	        owner=$(opencli domains-whoowns "$domain" | awk -F': ' '{print $2}')
	        if [[ "$owner" == "$cpanel_username" ]]; then
				# cpanel format: emailtest:$6$7XrOu5w5Iou8b1wj$dHcNUF0017EMLtue2X/nM2AlEoU8OS5TkyCR9QDEG8FcUOePTASdbDRhsU6ImxbGGiL7OdpJkNksWYqlvNSam/:20536::::::
				while IFS=: read -r username password_hash rest; do
				    [[ -z "$username" || -z "$password_hash" ]] && continue
				    email="${username}@${domain}"
				    log "Importing mailbox: $email"
				    opencli email-setup email add "$email" tempPassword123 >/dev/null 2>&1
					# openpanel format: emailtest@openpanel.org|{SHA512-CRYPT}$6$yspsXbUo.nkxXIs6$4x.rqdVe8dGaLWKZhlbmO5xFEgverG/ESS8.Cz3w9qH1GP6coXu7qs1CBFSE1co6cYHuVIqFS9bJR0PUcH3EZ0
					sed -i.bak "/^${email}|/c\\
${email}|{SHA512-CRYPT}${password_hash}
" "$postfix_file"
				done < "$shadow_file"

				# 2. move mails
				# openpanel storage: $STORE_EMAILS_IN/stefantestira.rs/emailtest2 OR /home/stefan/mail/stefantestira.rs/emailtest2
				if [ "$mailbox_format" == "maildir" ]; then
					if [ "$STORE_EMAILS_IN" == "domain" ]; then
					    STORE_EMAILS_IN="/home/$cpanel_username/mail/$domain/"
					fi
					# cpanel storage: extract/backup-3.24.2026_14-03-06_stefantestira/homedir/mail/stefantestira.rs/emailtest2
					if [ -d "/home/$cpanel_username/docker-data/volumes/${cpanel_username}_html_data/_data/mail/$domain/$username" ]; then
						log "Restoring mailboxes to $STORE_EMAILS_IN/$domain/"
						rsync -av --remove-source-files "/home/$cpanel_username/docker-data/volumes/${cpanel_username}_html_data/_data/mail/$domain/." "$STORE_EMAILS_IN/$domain/"
					else
						log "Failed restoring mailbox to $STORE_EMAILS_IN - $base_dir/mail/$domain/$username does not exist"
					fi
				fi
	        else
	            log "Skipping $domain: not owned by user $cpanel_username."
	        fi
	    else
	        log "Skipping $dir_path (shadow file missing or empty)"
	    fi
    done
}


# ======================================================================
# CREATED DATE IN OPENADMIN FROM WHM
restore_startdate() {
    real_backup_files_path="$1"
    cpanel_username="$2"
    cp_file_path="$real_backup_files_path/cp/$cpanel_username"
    STARTDATE=$(grep -oP 'STARTDATE=\K\d+' "$cp_file_path")
    
    if [ -n "$STARTDATE" ]; then
      human_readable_date=$(date -d @"$STARTDATE" +"%Y-%m-%d %H:%M:%S")
      log "Updating account creation date to reflect cpanel date: $human_readable_date"
      update_timestamp="UPDATE users SET registered_date = '$human_readable_date' WHERE username = '$cpanel_username';"
      mysql -e "$update_timestamp"
    fi
}

# ======================================================================
# EMAIL NOTIFICATIONS
restore_notifications() {
    local real_backup_files_path="$1"
    local cpanel_username="$2"
    notifications_cp_file="$real_backup_files_path/cp/$cpanel_username"
    notifications_op_file="/etc/openpanel/openpanel/core/users/$cpanel_username/notifications.yaml"
   
    if [ -z "$notifications_cp_file" ]; then
        log "WARNING: Unable to access $notifications_cp_file for notification preferences - Skipping"
    else
        dry_run "Would restore notification preferences from $notifications_cp_file" && return

        grep "notify_" $notifications_cp_file > $notifications_op_file
        cat_notifications_file=$(cat $notifications_op_file 2>&1)
        while IFS= read -r line; do
            log "$line"
        done <<< "$cat_notifications_file"
    fi
}

write_import_activity() {
    echo "$(date '+%Y-%m-%d %H:%M:%S')  $new_ip  Administrator ROOT user imported cpanel backup file" > /etc/openpanel/openpanel/core/users/$cpanel_username/activity.log
}



# MAIN
main() {
    start_message                                                              # what will be imported
    log_paths_are                                                              # where will we store the progress
    
    # STEP 1. PRE-RUN CHECKS
    check_if_valid_cp_backup "$backup_location"                                # is it?
    check_if_disk_available                                                    # calculate du needed for extraction
    check_if_user_exists                                                       # make sure we dont overwrite user!
    validate_plan_exists                                                       # check if provided plan exists
    install_dependencies                                                       # install commands we will use for this script
    get_server_ipv4                                                            # used in mysql grants
    
    # STEP 2. EXTRACT
    create_tmp_dir_and_path                                                    # create /tmp/.. dir and set the path
    extract_cpanel_backup "$backup_location" "${backup_dir}"                   # extract the archive

    # STEP 3. IMPORT
    locate_backup_directories                                                  # get paths from backup
    parse_cpanel_metadata                                                      # get data and configurations
    restore_files                                                              # homedir
    create_new_user "$cpanel_username" "random" "$cpanel_email" "$plan_name"   # create user data and container
    setquota -u $cpanel_username 0 0 0 0 /                                     # set unlimited quota while we do import!
    create_home_mountpoint                                                     # mount /var/www/html/ to /home/USERNAME 
    get_mariadb_or_mysql_for_user                                              # mysql or mariadb
    #NOT NEEDED ON CPANEL #fix_perms                                           # fix permissions for all files
    restore_php_version "$php_version"                                         # php v needs to run before domains 
    restore_domains                                                            # add domains
    restore_dns_zones                                                          # add dns 
    restore_mysql "$mysqldir"                                                  # mysql databases, users and grants
	restore_psql "$psqldir"                                                    # postgresql databases, users and grants
    restore_cron                                                               # cronjob
    restore_ssl "$cpanel_username"                                             # ssl certs
    restore_wordpress "$real_backup_files_path" "$cpanel_username"             # import wp sites to sitemanager
    restore_notifications "$real_backup_files_path" "$cpanel_username"         # notification preferences from cp
    restore_startdate "$real_backup_files_path" "$cpanel_username"             # cp account creation date
    opencli user-quota $cpanel_username                                        # restore quota

    # STEP 4. IMPORT ENTERPRISE FEATURES
    import_email_accounts_and_data "$cpanel_username"                          # import emails, filters, forwarders..
    ftp_accounts_import                                                        # import ftp accounts

	restore_locale                                                             # language
    reload_user_quotas                                                         # refresh du and inodes
    collect_stats                                                              # get cpu and ram usage
    write_import_activity

    # STEP 5. DELETE TMP FILES
    cleanup                                                                    # delete extracter files after import

    # STEP 6. NOTIFY USER
    success_message                                                            # have a 🍺

    # STEP 7. RUN ANY CUSTOM SCRIPTS
    run_custom_post_hook                                                       # any script to run after the import? example: edit dns on cp server, run tests, notify user, etc.
}

# ======================================================================
# ENTRYPOINT
define_data_and_log "$@"
