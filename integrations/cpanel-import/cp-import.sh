#!/bin/bash
pid=$$
script_dir=$(dirname "$0")
timestamp="$(date +'%Y-%m-%d_%H-%M-%S')"
start_time=$(date +%s)
DEBUG=true

###############################################################
# HELPER FUNCTIONS

usage() {
    echo "Usage: $0 --backup-location <path> --plan-name <plan_name> [--dry-run]"
    echo
    echo "Example: $0 --backup-location /home/backup-7.29.2024_13-22-32_stefan.tar.gz --plan-name default_plan_nginx --dry-run"
    exit 1
}

log() {
    local message="$1"
    local timestamp=$(date +'%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] $message" | tee -a "$log_file"
}

debug_log() {
    if [ "$DEBUG" = true ]; then
        log "DEBUG: $1"
    fi
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
    local plan_name=""
    DRY_RUN=false

    while [ "$1" != "" ]; do
        case $1 in
            --backup-location ) shift
                                backup_location=$1
                                ;;
            --plan-name )       shift
                                plan_name=$1
                                ;;
            --dry-run )         DRY_RUN=true
                                ;;
            --post-hook )       shift
                                post_hook=$1
                                ;;
            * )                 usage
        esac
        shift
    done

    if [ -z "$backup_location" ] || [ -z "$plan_name" ]; then
        usage
    fi

    base_name="$(basename "$backup_location")"
    base_name_no_ext="${base_name%.*}"
    local log_dir="/var/log/openpanel/admin/imports"
    mkdir -p $log_dir
    log_file="$log_dir/${base_name_no_ext}_${timestamp}.log"
    echo "Import started, log file: $log_file"
    main
}

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

install_dependencies() {
    log "Checking and installing dependencies..."
    install_needed=false
    declare -A commands=(
        ["tar"]="tar"
        ["parallel"]="parallel"
        ["rsync"]="rsync"
        ["unzip"]="unzip"
        ["jq"]="jq"
        ["pigz"]="pigz"
        ["mysql"]="mysql-client"
        ["wget"]="wget"
        ["curl"]="curl"
    )

    for cmd in "${!commands[@]}"; do
        if ! command_exists "$cmd"; then
            install_needed=true
            break
        fi
    done

    if [ "$install_needed" = true ]; then
        if command_exists apt-get; then
            log "Detected APT package manager. Updating..."
            apt-mark hold linux-image-generic linux-headers-generic >/dev/null 2>&1
            apt-get update -y >/dev/null 2>&1

            for cmd in "${!commands[@]}"; do
                if ! command_exists "$cmd"; then
                    log "Installing ${commands[$cmd]} (APT)"
                    apt-get install -y --no-upgrade --no-install-recommends "${commands[$cmd]}" >/dev/null 2>&1
                fi
            done

            apt-mark unhold linux-image-generic linux-headers-generic >/dev/null 2>&1

        elif command_exists dnf; then
            log "Detected DNF package manager. Updating..."
            dnf -y makecache >/dev/null 2>&1

            for cmd in "${!commands[@]}"; do
                if ! command_exists "$cmd"; then
                    log "Installing ${commands[$cmd]} (DNF)"
                    dnf install -y "${commands[$cmd]}" >/dev/null 2>&1
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
    new_ip=$(curl --silent --max-time 2 -4 https://ip.openpanel.com || wget --timeout=2 -qO- https://ip.openpanel.com || curl --silent --max-time 2 -4 https://ifconfig.me)
    if [ -z "$new_ip" ]; then
        new_ip=$(ip addr|grep 'inet '|grep global|head -n1|awk '{print $2}'|cut -f1 -d/)
    fi
}

validate_plan_exists(){
    if ! opencli plan-list --json | grep -qw "$plan_name"; then
        log "FATAL ERROR: Plan name '$plan_name' does not exist."
        exit 1
    fi
}







###############################################################
# MAIN FUNCTIONS

# CHECK EXTENSION AND DETERMINE SIZE
check_if_valid_cp_backup(){
    local backup_location="$1"
    ARCHIVE_SIZE=$(stat -c%s "$backup_location")
    local backup_filename=$(basename "$backup_location")
    extraction_command=""

    case "$backup_filename" in
        cpmove-*.tar.gz)
            log "Identified cpmove backup"
            extraction_command="tar -xzf"
            EXTRACTED_SIZE=$(($ARCHIVE_SIZE * 2))
        ;;
        backup-*.tar.gz)
            log "Identified full or partial cPanel backup"
            extraction_command="tar -xzf"
            EXTRACTED_SIZE=$(($ARCHIVE_SIZE * 2))
            ;;
        *.tar.gz)
            log "Identified gzipped tar backup"
            extraction_command="tar -xzf"
            EXTRACTED_SIZE=$(($ARCHIVE_SIZE * 2))
            ;;
        *.tgz)
            log "Identified tgz backup"
            extraction_command="tar -xzf"
            EXTRACTED_SIZE=$(($ARCHIVE_SIZE * 3))
            ;;
        *.tar)
            log "Identified tar backup"
            extraction_command="tar -xf"
            EXTRACTED_SIZE=$(($ARCHIVE_SIZE * 3))
            ;;
        *.zip)
            log "Identified zip backup"
            extraction_command="unzip"
            EXTRACTED_SIZE=$(($ARCHIVE_SIZE * 3))
            ;;
        *)
            log "FATAL ERROR: Unrecognized backup format: $backup_filename"
            exit 1
            ;;
    esac
}

# CHECK DISK USAGE
check_if_disk_available(){
    TMP_DIR="/tmp"
    HOME_DIR="/home"
    AVAILABLE_TMP=$(df --output=avail "$TMP_DIR" | tail -n 1)
    AVAILABLE_HOME=$(df --output=avail "$HOME_DIR" | tail -n 1)

    AVAILABLE_TMP=$(($AVAILABLE_TMP * 1024))
    AVAILABLE_HOME=$(($AVAILABLE_HOME * 1024))

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

# EXTRACT
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

# LOCATE FILES IN EXTRACTED BACKUP
locate_backup_directories() {
    log "Locating important files in the extracted backup"

    homedir=$(find "$backup_dir" -type d -name "homedir" | head -n 1)
    if [ -z "$homedir" ]; then
        homedir=$(find "$backup_dir" -type d -name "public_html" -printf '%h\n' | head -n 1)
    fi
    if [ -z "$homedir" ]; then
        log "FATAL ERROR: Unable to locate home directory in the backup"
        exit 1
    fi

    mysqldir="$real_backup_files_path/mysql"
    if [ -z "$mysqldir" ]; then
        log "WARNING: Unable to locate MySQL directory in the backup"
    fi

    mysql_conf="$real_backup_files_path/mysql.sql"
    if [ -z "$mysql_conf" ]; then
        log "WARNING: Unable to locate MySQL grants file in the backup"
    fi

    ftp_conf="$real_backup_files_path/proftpdpassword"
    if [ -z "$ftp_conf" ]; then
        log "WARNING: Unable to locate ProFTPD users file file in the backup"
    fi

    domain_logs="$real_backup_files_path/logs/"
    if [ -z "$domain_logs" ]; then
        log "WARNING: Unable to locate apache domlogs in the backup"
    fi

    cp_file="$real_backup_files_path/cp/$cpanel_username"
    if [ -z "$cp_file" ]; then
        log "FATAL ERROR: Unable to locate cp/$cpanel_username file in the backup"
        exit 1
    fi

    log "Backup directories and configuration files located successfully"
    log "- Home directory:       $homedir"
    log "- MySQL directory:      $mysqldir"
    log "- MySQL grants:         $mysql_conf"
    log "- PureFTPD users:       $ftp_conf"
    log "- Domain logs:          $domain_logs"
    log "- cPanel configuration: $cp_file"
}

get_mariadb_or_mysql_for_user() {
    mysql_type=$(grep '^MYSQL_TYPE=' /home/$cpanel_username/.env | cut -d '=' -f2 | tr -d '"')

    if [[ "$mysql_type" != "mariadb" && "$mysql_type" != "mysql" ]]; then
        echo "Unsupported MYSQL_TYPE: $mysql_type"
        exit 1
    fi

}

reload_user_quotas() {
    nohup bash -c '
        quotacheck -avm >/dev/null 2>&1
        repquota -u / > /etc/openpanel/openpanel/core/users/repquota
    ' >/dev/null 2>&1 &
}

collect_stats() {
    nohup bash -c "opencli docker-collect_stats '$cpanel_username'" >/dev/null 2>&1 &
}

# CPANEL BACKUP METADATA
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

        # Additional metadata
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

    log "Main Domain:          ${main_domain:-Not found}"
    log "Email:                ${cpanel_email:-Not found}"
    log "PHP Version:          $php_version"
    log "Finished parsing cPanel metadata."
}


# CHECK BEFORE EXPORT
check_if_user_exists(){
    backup_filename=$(basename "$backup_location")
    cpanel_username="${backup_filename##*_}"
    cpanel_username="${cpanel_username%%.*}"
    log "Username: $cpanel_username"

    local existing_user=""
    if opencli user-list --json > /dev/null 2>&1; then
        existing_user=$(opencli user-list --json | jq -r ".[] | select(.username == \"$cpanel_username\") | .id")
    fi
    if [ -z "$existing_user" ]; then
        log "Username $cpanel_username is available"
        if [ "$DRY_RUN" = false ]; then
            log "Starting import process.."
        fi
    else
        log "FATAL ERROR: $cpanel_username already exists."
        exit 1
    fi
}


# CREATE NEW USER
create_new_user() {
    local username="$1"
    local email="$3"
    local plan_name="$4"

    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: Would create user $username with email $email and plan $plan_name"
        return
    fi
        
    create_user_command=$(opencli user-add "$cpanel_username" generate "$email" "$plan_name" 2>&1)
    while IFS= read -r line; do
        log "$line"
    done <<< "$create_user_command"

    if echo "$create_user_command" | grep -q "Successfully added user"; then
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
        log "FATAL ERROR: User addition failed. Response did not contain the expected success message."
        exit 1
    fi
}



# PHP VERSION
restore_php_version() {
    local php_version="$1"

    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: Would sed default PHP version $php_version for user $cpanel_username"
        return
    fi

    # if 'inherit' we will keep the default of op
    if [ "$php_version" == "inherit" ]; then
        log "PHP version is set to inherit. No changes will be made."
    else
        # if custom, set it
        log "Setting PHP $php_version as the default version for all new domains."
        output=$(opencli php-default "$cpanel_username" --update "$php_version" 2>&1)
        while IFS= read -r line; do
            log "$line"
        done <<< "$output"
    fi
}



# PHPMYADMIN
grant_root_access() {
    local username="$1"
    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: Would grant root user access to all databases for user $username"
        return
    fi
    log "Granting root access to all databases for user $username"
    phpmyadmin_user="root"
    sql_command="GRANT ALL ON *.* TO 'root'@'localhost'; FLUSH PRIVILEGES;"
    grant_commands=$(docker --context=$username exec mysql mysql -N -e "$sql_command")
    log "Access granted to root user for all databases of $username"
}


# MYSQL
restore_mysql() {
    local mysql_dir="$1"
    local sandbox_warning_logged=false

    log "Restoring MySQL databases for user $cpanel_username"

    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: Would restore MySQL databases for user $cpanel_username"
        return
    fi

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

    if [ -d "$mysql_dir" ]; then
        # STEP 1: Replace old IP and hostname
        old_ip=$(grep -oP 'IP=\K[0-9.]+' "${real_backup_files_path}/cp/$cpanel_username")
        log "Replacing old server IP: $old_ip with '%' in database grants"
        sed -i "s/$old_ip/%/g" "$mysql_conf"

        old_hostname=$(cat "${real_backup_files_path}/meta/hostname")
        log "Removing old hostname $old_hostname from database grants"
        sed -i "/$old_hostname/d" "$mysql_conf"
        
        # STEP 2: Start MySQL container
        if [ "$mysql_type" = "mysql" ]; then
            mysql_version="8.0"
            sed -i 's/^MYSQL_VERSION=.*/MYSQL_VERSION="8.0"/' /home/"$cpanel_username"/.env
        fi
        log "Initializing $mysql_type $mysql_version service for user"
        cd "/home/$cpanel_username/" && docker --context="$cpanel_username" compose up -d "$mysql_type" >/dev/null 2>&1

        # STEP 3: Wait for MySQL to be ready (max 90 seconds)
        log "Waiting for MySQL service to be ready..."
        max_wait=90
        waited=0
        while ! docker --context="$cpanel_username" exec "$mysql_type" $mysql_type -e "SELECT 1" >/dev/null 2>&1; do
            sleep 2
            waited=$((waited + 2))
            if [ "$waited" -ge "$max_wait" ]; then
                log "ERROR: MySQL did not become ready after $max_wait seconds"
                exit 1
            fi
        done

        # STEP 4: Create and import databases
        total_databases=$(ls "$mysql_dir"/*.create 2>/dev/null | wc -l)
        log "Starting import for $total_databases MySQL databases"
        if [ "$total_databases" -gt 0 ]; then
            current_db=1
            for db_file in "$mysql_dir"/*.create; do
                db_name=$(basename "$db_file" .create)

                log "Creating database: $db_name (${current_db}/${total_databases})"
                apply_sandbox_workaround "$db_name.create"
                docker --context="$cpanel_username" cp "${real_backup_files_path}/mysql/$db_name.create" "$mysql_type:/tmp/${db_name}.create" >/dev/null 2>&1
                docker --context="$cpanel_username" exec "$mysql_type" bash -c "mysql < /tmp/${db_name}.create && rm /tmp/${db_name}.create"

                log "Importing tables for database: $db_name"
                apply_sandbox_workaround "$db_name.sql"
                docker --context="$cpanel_username" cp "${real_backup_files_path}/mysql/$db_name.sql" "$mysql_type:/tmp/${db_name}.sql" >/dev/null 2>&1
                docker --context="$cpanel_username" exec "$mysql_type" bash -c "mysql $db_name < /tmp/${db_name}.sql && rm /tmp/${db_name}.sql"

                current_db=$((current_db + 1))
            done
            log "Finished processing $((current_db - 1)) databases"
        else
            log "WARNING: No MySQL databases found"
        fi

        # STEP 5: Import grants
        log "Importing database grants"
        python3 "$script_dir/mysql/json_2_sql.py" "${real_backup_files_path}/mysql.sql" "${real_backup_files_path}/mysql.TEMPORARY.sql" >/dev/null 2>&1

        docker --context="$cpanel_username" cp "${real_backup_files_path}/mysql.TEMPORARY.sql" "$mysql_type:/tmp/mysql.TEMPORARY.sql" >/dev/null 2>&1
        docker --context="$cpanel_username" exec "$mysql_type" bash -c "$mysql_type < /tmp/mysql.TEMPORARY.sql && $mysql_type -e 'FLUSH PRIVILEGES;' && rm /tmp/mysql.TEMPORARY.sql"

        # STEP 6: Grant root user all access
        grant_root_access "$cpanel_username"

    else
        log "No MySQL databases found to restore"
    fi
}




# SSL CERTIFICATES
restore_ssl() {
    local username="$1"

    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: Would restore SSL certificates for user $username"
        return
    fi

    # TODO: edit to cover certs/ keys/ 
    log "Restoring SSL certificates for user $username"
    # apache_tls/ dir has LE certs, custom are in ssl/
    if [ -d "$real_backup_files_path/ssl" ]; then
        for cert_file in "$real_backup_files_path/ssl"/*.crt; do
            local domain=$(basename "$cert_file" .crt)
            local key_file="$real_backup_files_path/ssl/$domain.key"
            
            # todo: move keys to var/www/html first!            
            if [ -f "$key_file" ]; then
                log "Installing SSL certificate for domain: $domain"
                opencli domains-ssl "$domain" "$cert_file" "$key_file"
            else
                log "SSL key file not found for domain: $domain"
            fi
        done

        # TODO: reload caddy
    else
        log "No SSL certificates found to restore"
    fi
}




# DNS ZONES
restore_dns_zones() {
    log "Restoring DNS zones for user $cpanel_username"

    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: Would restore DNS zones for user $cpanel_username"
        return
    fi

    if [ -d "$real_backup_files_path/dnszones" ]; then
        for zone_file in "$real_backup_files_path/dnszones"/*; do
            local zone_name=$(basename "${zone_file%.db}")

            # Check if the destination zone file exists, if not, it was probably a subdomain that had no dns zone
            if [ ! -f "/etc/bind/zones/${zone_name}.zone" ]; then
                log "DNS zone file /etc/bind/zones/${zone_name}.zone does not exist. Skipping import for $zone_name."
                continue
            else
                log "Importing DNS zone: $zone_name"
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

# creates symlink of /var/www/html/ to /home/$cpanel_username so all paths in files keep working!
create_home_mountpoint() {
    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: Would create a symlink from html_data volume to /home/$cpanel_username/"
        return
    fi
    
sed -i '/^[[:space:]]*volumes:[[:space:]]*$/{
  N
  /- html_data:\/var\/www\/html\// s|$|\n      - html_data:/home/${CONTEXT}/|
}' /home/$cpanel_username/docker-compose.yml

}

# HOME DIR
restore_files() {
    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: Would restore files from /home/$cpanel_username/ to html_data volume"
        return
    fi

    du_needed_for_home=$(du -sh "$real_backup_files_path/homedir" | cut -f1)
    log "Restoring home directory ($du_needed_for_home) to html_data volume"
    mkdir -p /home/$cpanel_username/docker-data/volumes/${cpanel_username}_html_data/
    #rm -rf "$real_backup_files_path"/homedir/{.cpanel,.trash,wordpress-backups}
    mv $real_backup_files_path/homedir /home/$cpanel_username/docker-data/volumes/${cpanel_username}_html_data/_data

    : '
    # LEAVE THIS FOR CLUSTERING FEATURE
    rsync -Prltvc --info=progress2 "$real_backup_files_path/homedir/" "/home/$cpanel_username/" 2>&1 | while IFS= read -r line; do
        log "$line"
    done

    log "Finished transferring files, comparing to source.."
    original_size=$(du -sb "$real_backup_files_path/homedir" | cut -f1)
    copied_size=$(du -sb "/home/$cpanel_username/" | cut -f1)

    if [[ "$original_size" -eq "$copied_size" ]]; then
        log "The original and target directories have the same size."
    else
        log "WARNING: The original and target directories differ in size after restore."
        log "Original size: $original_size bytes"
        log "Target size:   $copied_size bytes"
    fi
    '

}



# PERMISSIONS
fix_perms(){
    local verbose="" #-v
    log "Changing permissions for all files and folders in user home directory /home/$cpanel_username/"
    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: Would change permissions with command: find /home/$cpanel_username -print0 | xargs -0 chown $verbose $cpanel_username:$cpanel_username"
        return
    fi
    
    if ! timeout 600 find /home/$cpanel_username -print0 | xargs -0 chown $verbose $cpanel_username:$cpanel_username > /dev/null 2>&1; then
        if [ $? -eq 124 ]; then
            log "ERROR: Timeout reached while changing permissions (10 minutes)."
        else
            log "ERROR: Failed to change permissions."
        fi
            log "       Make sure to change permissions manually from terminal with: find /home/$cpanel_username -print0 | xargs -0 chown -v $cpanel_username:$cpanel_username"
    fi
    
}


# WORDPRESS SITES
restore_wordpress() {
    local real_backup_files_path="$1"
    local username="$2"

    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: Would restore WordPress sites for user $username"
        return
    fi

    log "Checking user files for WordPress installations to add to Site Manager interface.."
    output=$(opencli websites-scan $cpanel_username)
        while IFS= read -r line; do
            log "$line"
        done <<< "$output"    
}


# DOMAINS
restore_domains() {
    if [ -f "$real_backup_files_path/userdata/main" ]; then
        file_path="$real_backup_files_path/userdata/main"
        main_domain=""
        parked_domains=""
        sub_domains=""
        addon_domains=""

        while IFS= read -r line; do
            if [[ "$line" =~ ^main_domain: ]]; then
                main_domain=$(echo "$line" | awk '{print $2}')
            elif [[ "$line" =~ ^parked_domains: ]]; then
                parked_domains=$(echo "$line" | awk '{print $2}' | tr -d '[]')
            elif [[ "$line" =~ ^sub_domains: ]]; then
                sub_domains_section=true
                continue
            elif [[ "$line" =~ ^addon_domains: ]]; then
                addon_domains_section=true
                continue
            fi

            if [[ "$sub_domains_section" == true ]]; then
                if [[ "$line" =~ ^[[:space:]]+- ]]; then
                    sub_domains+=$(echo "$line" | awk '{print $2}')$'\n'
                else
                    sub_domains_section=false
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

        # Parse parked_domains
        if [[ -z "$parked_domains" ]]; then
            parked_domains_array=()
        else
            IFS=',' read -r -a parked_domains_array <<< "$parked_domains"
        fi

        sub_domains_array=()
        addon_domains_array=()

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

        log "Detected a total of $domains_total_count domains for user."

        current_domain_count=0

        create_domain(){
            domain="$1"
            type="$2"

            current_domain_count=$((current_domain_count + 1))
            if [[ $domain == \*.* ]]; then
                log "WARNING: Skipping wildcard domain $domain"
            else
                log "Restoring $type $domain (${current_domain_count}/${domains_total_count})"

                userdata_file="$real_backup_files_path/userdata/$domain"
                docroot=""
                if [ -f "$userdata_file" ]; then
                    original_docroot=$(awk -F': ' '/^documentroot:/ {print $2}' "$userdata_file" | xargs)
                    docroot="${original_docroot#/home/$cpanel_username/}"
                    docroot="/var/www/html/$docroot"
                else
                    log "WARNING: userdata file not found for $domain. Using default docroot."
                fi
            
                if [ "$DRY_RUN" = true ]; then
                    log "DRY RUN: Would restore $type $domain with --docroot ${docroot:-N/A}"
                elif opencli domains-whoowns "$domain" | grep -q "not found in the database."; then
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
                else
                    log "WARNING: $type $domain already exists and will not be added to this user."
                fi
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




# CRONJOB
restore_cron() {
    log "Restoring cron jobs for user $cpanel_username"

    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: Would restore cron jobs for user $cpanel_username"
        return
    fi

    if [ -f "$real_backup_files_path/cron/$cpanel_username" ]; then
        sed -i '1,2d' "$real_backup_files_path/cron/$cpanel_username"
        ofelia_cron_path="/home/${cpanel_username}/crons.ini"
        > "$ofelia_cron_path"

        job_index=1
        job_found=false
        while IFS= read -r cron_line; do
            [[ -z "$cron_line" || "$cron_line" =~ ^# ]] && continue

            job_found=true
            schedule=$(echo "$cron_line" | awk '{print $1, $2, $3, $4, $5}')
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

run_custom_post_hook() {
    if [ -n "$post_hook" ]; then
        if [ -x "$post_hook" ]; then
            log "Executing post-hool script.."
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

    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: import process for user $cpanel_username completed."
    else
        log "SUCCESS: Import for user $cpanel_username completed successfully."
    fi
}

log_paths_are() {
    log "Log file: $log_file"
    log "PID: $pid"
}

start_message() {
    echo -e "
------------------ STARTING CPANEL ACCOUNT IMPORT ------------------
--------------------------------------------------------------------

Currently supported features:

├─ DOMAINS:
│  ├─ Primary domain, Addons, Aliases and Subdomains
│  ├─ SSL certificates
│  ├─ Domains access logs (Apache domlogs)
│  └─ DNS zones
├─ WEBSITES:
│  └─ WordPress instalations from WPToolkit & Softaculous 
├─ DATABASES:
│    ├─ Remote access to MySQL
│    └─ MySQL databases, users and grants
├─ PHP:
│    └─ Installed version from Cloudlinux PHP Selector
├─ FILES
├─ CRONS
└─ ACCOUNT
    ├─ Notification preferences
    ├─ cPanel account creation date
    └─ cPanel account password

***emails, ftp, nodejs/python, postgres are not yet supported***

--------------------------------------------------------------------
  if you experience any errors with this script, please report to
    https://github.com/stefanpejcic/cPanel-to-OpenPanel/issues
--------------------------------------------------------------------
"
}


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

import_email_accounts_and_data() {
        log "WARNING: Importing Email accounts is not yet supported"
}


# timestamp in openadmin/whm
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

# EMAIL NOTIFICATIONS
restore_notifications() {
    local real_backup_files_path="$1"
    local cpanel_username="$2"
    notifications_cp_file="$real_backup_files_path/cp/$cpanel_username"
    notifications_op_file="/etc/openpanel/openpanel/core/users/$cpanel_username/notifications.yaml"
   
    if [ -z "$notifications_cp_file" ]; then
        log "WARNING: Unable to access $notifications_cp_file for notification preferences - Skipping"
    else
        if [ "$DRY_RUN" = true ]; then
            log "DRY RUN: Would restore notification preferences from $notifications_cp_file"
            check_notifications=$(grep "notify_" $notifications_cp_file)
            while IFS= read -r line; do
                log "$line"
            done <<< "$check_notifications"
            return
        fi  
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


###################################### MAIN SCRIPT EXECUTION ######################################
###################################################################################################
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
    fix_perms                                                                  # fix permissions for all files
    restore_php_version "$php_version"                                         # php v needs to run before domains 
    restore_domains                                                            # add domains
    restore_dns_zones                                                          # add dns 
    restore_mysql "$mysqldir"                                                  # mysql databases, users and grants
    restore_cron                                                               # cronjob
    restore_ssl "$cpanel_username"                                             # ssl certs
    restore_wordpress "$real_backup_files_path" "$cpanel_username"             # import wp sites to sitemanager
    restore_notifications "$real_backup_files_path" "$cpanel_username"         # notification preferences from cp
    restore_startdate "$real_backup_files_path" "$cpanel_username"             # cp account creation date
    opencli user-quota $cpanel_username                                        # restore quota

    # STEP 4. IMPORT ENTERPRISE FEATURES
    import_email_accounts_and_data                                             # import emails, filters, forwarders..
    ftp_accounts_import                                                        # import ftp accounts
    
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

# MAIN FUNCTION
define_data_and_log "$@"
