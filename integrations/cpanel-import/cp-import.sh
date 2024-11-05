#!/bin/bash
pid=$$
script_dir=$(dirname "$0")
timestamp="$(date +'%Y-%m-%d_%H-%M-%S')" #used by log file name
start_time=$(date +%s) #used to calculate elapsed time at the end

set -eo pipefail

DEBUG=true



# wget cp backup
# cd /home && git clone https://github.com/stefanpejcic/cPanel-to-OpenPanel
# bash /home/cPanel-to-OpenPanel/cp-import.sh --backup-location /home/backup-10.23.2024_14-49-42_pejcic.tar.gz --plan-name ubuntu_nginx_mysql



















###############################################################
# HELPER FUNCTIONS

usage() {
    echo "Usage: $0 --backup-location <path> --plan-name <plan_name> [--dry-run]"
    echo ""
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

    # Parse command-line arguments
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

    # Validate required parameters
    if [ -z "$backup_location" ] || [ -z "$plan_name" ]; then
        usage
    fi

    # Format log file
    base_name="$(basename "$backup_location")"
    base_name_no_ext="${base_name%.*}"
    local log_dir="/var/log/openpanel/admin/imports"
    mkdir -p $log_dir
    log_file="$log_dir/${base_name_no_ext}_${timestamp}.log"

    # Run the main function
    echo "Import started, log file: $log_file"

    main
}



command_exists() {
    command -v "$1" >/dev/null 2>&1
}



install_dependencies() {
    log "Checking and installing dependencies..."

    install_needed=false

    # needed commands
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

    # If installation is needed, update package list and install missing packages
    if [ "$install_needed" = true ]; then
        log "Updating package manager..."

        # Hold kernel packages to prevent upgrades
        apt-mark hold linux-image-generic linux-headers-generic >/dev/null 2>&1

        # Update package list without upgrading
        apt-get update -y >/dev/null 2>&1

        for cmd in "${!commands[@]}"; do
            if ! command_exists "$cmd"; then
                log "Installing ${commands[$cmd]}"
                # Install package without upgrading or installing recommended packages
                apt-get install -y --no-upgrade --no-install-recommends "${commands[$cmd]}" >/dev/null 2>&1
            fi
        done

        # Unhold kernel packages
        apt-mark unhold linux-image-generic linux-headers-generic >/dev/null 2>&1

        log "Dependencies installed successfully."
    else
        log "All required dependencies are already installed."
    fi
}



get_server_ipv4(){
    # Get server ipv4 from ip.openpanel.co or ifconfig.me
    new_ip=$(curl --silent --max-time 2 -4 https://ip.openpanel.co || wget --timeout=2 -qO- https://ip.openpanel.co || curl --silent --max-time 2 -4 https://ifconfig.me)

    # if no internet, get the ipv4 from the hostname -I
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

    # Identify the backup type
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

    # Get available space in /tmp and home directories in bytes
    AVAILABLE_TMP=$(df --output=avail "$TMP_DIR" | tail -n 1)
    AVAILABLE_HOME=$(df --output=avail "$HOME_DIR" | tail -n 1)

    AVAILABLE_TMP=$(($AVAILABLE_TMP * 1024))
    AVAILABLE_HOME=$(($AVAILABLE_HOME * 1024))

    # Check if there's enough space
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

    # Extract the backup
    if [ "$extraction_command" = "unzip" ]; then
        $extraction_command "$backup_location" -d "$backup_dir"
    elif [ "$extraction_command" = "tar -xzf" ]; then
        # Get the size of the backup in bytes
        backup_size=$(stat -c %s "${backup_location}")
        
        # Calculate 1% of the backup size (in bytes)
        zero_one_percent=$((backup_size / 1000000))
       
        # Extract the tar file with progress for every 1%
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

    # Try to locate the key directories
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

        # Check for Cloudlinux PHP Selector
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

    # Ensure we have at least an empty string for each variable
    main_domain="${main_domain:-}"
    cpanel_email="${cpanel_email:-}"
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
        :
    else
        log "FATAL ERROR: User addition failed. Response did not contain the expected success message."
        exit 1
    fi
}



# PHP VERSION
restore_php_version() {
    local php_version="$1"

    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: Would check/install PHP version $php_version for user $cpanel_username"
        return
    fi

    # Check if php_version is "inherit"
    if [ "$php_version" == "inherit" ]; then
        log "PHP version is set to inherit. No changes will be made."
    else
        log "Checking if current PHP version installed matches the version from backup"
        local current_version=$(opencli php-default_version "$cpanel_username" | sed 's/Default PHP version for user.*: //')
        if [ "$current_version" != "$php_version" ]; then
            local installed_versions=$(opencli php-installed_versions "$cpanel_username")
            if ! echo "$installed_versions" | grep -q "$php_version"; then
                log "Default PHP version $php_version from backup is not present in the container, installing.."
                output=$(opencli php-install_version "$cpanel_username" "$php_version" 2>&1)
                while IFS= read -r line; do
                    log "$line"
                done <<< "$output"

                # Set as default PHP version
                log "Setting newly installed PHP $php_version as the default version for all new domains."
                output=$(opencli php-default_version "$cpanel_username" --update "$php_version" 2>&1)
                while IFS= read -r line; do
                    log "$line"
                done <<< "$output"
            fi
        else
            log "Default PHP version in backup file ($php_version) matches the installed PHP version: ($current_version)"
        fi
    fi
}



# PHPMYADMIN
grant_phpmyadmin_access() {
    local username="$1"

    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: Would grant phpMyAdmin access to all databases for user $username"
        return
    fi

    log "Granting phpMyAdmin access to all databases for user $username"
    # https://github.com/stefanpejcic/OpenPanel/blob/148b5e482f7bde4850868ba5cf85717538770882/docker/apache/phpmyadmin/pma.php#L13C44-L13C54
    phpmyadmin_user="phpmyadmin"
    sql_command="GRANT ALL ON *.* TO 'phpmyadmin'@'localhost'; FLUSH PRIVILEGES;"
    grant_commands=$(docker exec $username mysql -N -e "$sql_command")

    log "Access granted to phpMyAdmin user for all databases of $username"

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

    #https://jira.mariadb.org/browse/MDEV-34183
    apply_sandbox_workaround() {
        local db_file="$1"
        text_to_check='enable the sandbox mode'
        local first_line

        first_line=$(head -n 1 ${real_backup_files_path}/mysql/$db_file)
        if echo "$first_line" | grep -q "$text_to_check"; then
            if [ "$sandbox_warning_logged" = false ]; then
                log "WARNING: Database dumps were created on a MariaDB server with '--sandbox' mode. Applying workaround for backwards compatibility to MySQL (BUG: https://jira.mariadb.org/browse/MDEV-34183)"
                sandbox_warning_logged=true
            fi
            # Remove the first line and save the changes to the same file
            tail -n +2 "${real_backup_files_path}/mysql/$db_file" > "${real_backup_files_path}/mysql/${db_file}.workaround" && mv "${real_backup_files_path}/mysql/${db_file}.workaround" "${real_backup_files_path}/mysql/$db_file"
        fi
    }

    if [ -d "$mysql_dir" ]; then
        # STEP 1. get old server ip and replace it in the mysql.sql file that has import permissions
        old_ip=$(grep -oP 'IP=\K[0-9.]+' ${real_backup_files_path}/cp/$cpanel_username)
        log "Replacing old server IP: $old_ip with '%' in database grants"
        sed -i "s/$old_ip/%/g" $mysql_conf

        old_hostname=$(cat ${real_backup_files_path}/meta/hostname)
        log "Removing old hostname $old_hostname from database grants"
        sed -i "/$old_hostname/d" "$mysql_conf"


        # STEP 2. start mysql for user
        log "Initializing MySQL service for user"
        docker exec $cpanel_username bash -c "service mysql start >/dev/null 2>&1"
        docker exec "$cpanel_username" sed -i 's/CRON_STATUS="off"/CRON_STATUS="on"/' /etc/entrypoint.sh

        # STEP 3. create and import databases
        total_databases=$(ls "$mysql_dir"/*.create | wc -l)
        log "Starting import for $total_databases MySQL databases"
        if [ "$total_databases" -gt 0 ]; then
            current_db=1
            for db_file in "$mysql_dir"/*.create; do
                local db_name=$(basename "$db_file" .create)

                log "Creating database: $db_name (${current_db}/${total_databases})"
                apply_sandbox_workaround "$db_name.create" # Apply the workaround if it's needed
                docker cp ${real_backup_files_path}/mysql/$db_name.create $cpanel_username:/tmp/${db_name}.create  >/dev/null 2>&1
                docker exec $cpanel_username bash -c "mysql < /tmp/${db_name}.create && rm /tmp/${db_name}.create"

                log "Importing tables for database: $db_name"
                apply_sandbox_workaround "$db_name.sql" # Apply the workaround if it's needed
                docker cp ${real_backup_files_path}/mysql/$db_name.sql $cpanel_username:/tmp/$db_name.sql >/dev/null 2>&1
                docker exec $cpanel_username bash -c "mysql ${db_name} < /tmp/${db_name}.sql && rm /tmp/${db_name}.sql"
                current_db=$((current_db + 1))
            done
            log "Finished processing $current_db databases"
        else
            log "WARNING: No MySQL databases found"
        fi
        # STEP 4. import grants and flush privileges
        log "Importing database grants"
        python3 $script_dir/mysql/json_2_sql.py ${real_backup_files_path}/mysql.sql ${real_backup_files_path}/mysql.TEMPORARY.sql >/dev/null 2>&1

        docker cp ${real_backup_files_path}/mysql.TEMPORARY.sql $cpanel_username:/tmp/mysql.TEMPORARY.sql >/dev/null 2>&1
        docker exec $cpanel_username bash -c "mysql < /tmp/mysql.TEMPORARY.sql && mysql -e 'FLUSH PRIVILEGES;' && rm /tmp/mysql.TEMPORARY.sql"

        # STEP 5. Grant phpMyAdmin access
        grant_phpmyadmin_access "$cpanel_username"

    else
        log "No MySQL databases found to restore"
    fi
}



# SSL CACHE
refresh_ssl_file() {
    local username="$1"

    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: Would refresh SSL file for user $username"
        return
    fi

    log "Creating a list of SSL certificates for user interface"
    output=$(opencli ssl-user "$cpanel_username")
        while IFS= read -r line; do
            log "$line"
        done <<< "$output"
}




# SSL CERTIFICATES
restore_ssl() {
    local username="$1"

    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: Would restore SSL certificates for user $username"
        return
    fi


    # TODO: edit to cove certs/ keys/ 
    log "Restoring SSL certificates for user $username"
    if [ -d "$real_backup_files_path/ssl" ]; then
        for cert_file in "$real_backup_files_path/ssl"/*.crt; do
            local domain=$(basename "$cert_file" .crt)
            local key_file="$real_backup_files_path/ssl/$domain.key"
            if [ -f "$key_file" ]; then
                log "Installing SSL certificate for domain: $domain"
                opencli ssl install --domain "$domain" --cert "$cert_file" --key "$key_file"
            else
                log "SSL key file not found for domain: $domain"
            fi
        done

        # Refresh the SSL file after restoring certificates
        refresh_ssl_file "$username"
    else
        log "No SSL certificates found to restore"
    fi
}





# SSH PASSWORD

# openpanel does not use pam for authentication so we can not se the user password to be same as on cp, but we can set ssh password to remain the same!
restore_ssh_password() {
    local username="$1"
    local shell_file="$real_backup_files_path/shadow"
    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: Would set SSH password for user: $username from: $shell_file"
        return
    fi

    if [ -f "$shell_file" ]; then
        log "Restoring SSH password for user $username"
        password_hash=$(cat $shell_file)
        if [ -z "$password_hash" ]; then
            echo "WARNNG: Failed to retrieve password hash for user from $shell_file file."
        else
            scaped_password_hash=$(echo "$password_hash" | sed 's/\$/\\\$/g')
            docker exec "$username" bash -c "sed -i 's|^$username:[^:]*:|$username:$escaped_password_hash:|' /etc/shadow"
            if [ $? -eq 0 ]; then
                echo "Successfully set the password for ssh user $username in contianer to be the same as in cpanel backup file."
            else
                echo "Failed to set ssh password for user $username to eb the same as in cpanel backup file."
            fi
        fi
    else
        echo "WARNNG: Failed to retrieve password shadow for user from $shell_file file."
    fi
}




# SSH KEYS
restore_ssh() {
    local username="$1"

    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: Would restore SSH access for user $username"
        return
    fi

    log "Restoring SSH access for user $username"
    local shell_access=$(grep -oP 'shell: \K\S+' "$real_backup_files_path/userdata/main")
    if [ "$shell_access" == "/bin/bash" ]; then
        opencli user-ssh enable "$username"
        if [ -f "$real_backup_files_path/.ssh/id_rsa.pub" ]; then
            mkdir -p "/home/$username/.ssh"
            cp "$real_backup_files_path/.ssh/id_rsa.pub" "/home/$username/.ssh/authorized_keys"
            chown -R "$username:$username" "/home/$username/.ssh"
        fi
    fi
}




# DNS ZONES
restore_dns_zones() {
    log "Restoring DNS zones for user $cpanel_username"




            #domain_file="$real_backup_files_path/userdata/$domain"
            #domain=$(basename "$domain_file")



    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: Would restore DNS zones for user $cpanel_username"
        return
    fi

    if [ -d "$real_backup_files_path/dnszones" ]; then
        for zone_file in "$real_backup_files_path/dnszones"/*; do
            local zone_name=$(basename "${zone_file%.db}")

            # Check if the destination zone file exists, if not, it was probably a subdomain that had no dns zone and
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
                log "Replacing old server IP: $old_ip with new IP: $new_ip in DNS zone file for domain: $zone_name"
                sed -i "s/$old_ip/$new_ip/g" $zone_file
            fi

            # Temporary files to store intermediate results
            temp_file_of_original_zone=$(mktemp)
            temp_file_of_created_zone=$(mktemp)

            # Remove all lines after the last line that starts with '@'
            log "Editing original zone for domain $zone_name to temporary file: $temp_file_of_original_zone"
            awk '/^@/ { found=1; last_line=NR } { if (found && NR > last_line) exit } { print }' "$zone_file" > "$temp_file_of_original_zone"

            # Remove all lines from the beginning until the line that has 'NS' and including that line
            log "Editing created zone for domain $zone_name to temporary file: $temp_file_of_created_zone"
            awk '/NS/ { found=1; next } found { print }' "/etc/bind/zones/${zone_name}.zone" > "$temp_file_of_created_zone"

            # Append the processed second file to the first
            log "Merging the DNS zone records from  $temp_file_of_created_zone with $temp_file_of_original_zone"
            cat "$temp_file_of_created_zone" >> "$temp_file_of_original_zone"

            # Move the merged content to the final file
            log "Replacing the created zone /etc/bind/zones/${zone_name}.zone with updated records."
            mv "$temp_file_of_original_zone" "/etc/bind/zones/${zone_name}.zone"

            # Clean up
            rm "$temp_file_of_created_zone"

            log "DNS zone file for $zone_name has been imported."
        done
    else
        log "No DNS zones found to restore"
    fi
}



# HOME DIR
restore_files() {
    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: Would restore files to /home/$cpanel_username/"
        return
    fi

    du_needed_for_home=$(du -sh "$real_backup_files_path/homedir" | cut -f1)
    log "Restoring home directory ($du_needed_for_home) to /home/$cpanel_username/"

    mv $real_backup_files_path/homedir /home/$cpanel_username

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

    # Move all files from public_html to main domain dir
    log "Moving main domain files from public_html to $main_domain directory."
    mv /home/$cpanel_username/public_html /home/$cpanel_username/$main_domain # openpanel has no concept of 'primary' domain
    rm /home/$cpanel_username/www                                             # since www is just a symlink to public_html
}



# PERMISSIONS
fix_perms(){
    log "Changing permissions for all files and folders in user home directory /home/$cpanel_username/"
    if [ "$DRY_RUN" = true ]; then
        log 'DRY RUN: Would change permissions with command: docker exec $cpanel_username bash -c "chown -R 1000:33 /home/$cpanel_username"'
        return
    fi
    docker exec $cpanel_username bash -c "chown -R 1000:33 /home/$cpanel_username"
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
        # Initialize variables
        main_domain=""
        parked_domains=""
        sub_domains=""
        addon_domains=""

        # Read the file line by line
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
                    # Avoid adding invalid entries and trailing colons
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
    
                if [ "$DRY_RUN" = true ]; then
                    log "DRY RUN: Would restore $type $domain"
                elif opencli domains-whoowns "$domain" | grep -q "not found in the database."; then
                    output=$(opencli domains-add "$domain" "$cpanel_username" 2>&1)
                    while IFS= read -r line; do
                        log "$line"
                    done <<< "$output"
                else
                    log "WARNING: $type $domain already exists and will not be added to this user."
                fi
            fi

        }

        # Process the domains
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
                # TODO: create record in DNS zone instead of separate domain if only DNS zone and no folder!
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
        # exclude shell and email variables from file!
        sed -i '1,2d' "$real_backup_files_path/cron/$cpanel_username"

        output=$(docker cp $real_backup_files_path/cron/$cpanel_username $cpanel_username:/var/spool/cron/crontabs/$cpanel_username 2>&1)
        while IFS= read -r line; do
            log "$line"
        done <<< "$output"

        output=$(docker exec $cpanel_username bash -c "crontab -u $cpanel_username /var/spool/cron/crontabs/$cpanel_username" 2>&1)
        while IFS= read -r line; do
            log "$line"
        done <<< "$output"

        output=$(docker exec $cpanel_username bash -c "service cron restart" 2>&1)
        while IFS= read -r line; do
            log "$line"
        done <<< "$output"

        docker exec "$cpanel_username" sed -i 's/CRON_STATUS="off"/CRON_STATUS="on"/' /etc/entrypoint.sh  >/dev/null 2>&1
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
├─ SSH
│   ├─ Remote SSH access
│   ├─ SSH password
│   └─ SSH keys
└─ ACCOUNT
    ├─ Notification preferences
    ├─ cPanel account creation date
    └─ locale

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
        : '
        #cat proftpdpasswd
        pejcic:$6$cv9wnxSLeD1VEk.U$dm84PcqygxOWqT/uyMjrICKUPFeAQwOimJ8frihDCxjRfa1BKf6bnHIhWrbfmLrLn2YBSMnNatW09ZZMAS7GT/:1030:1034:pejcic:/home/pejcic:/bin/bash
        neko@pcx3.com:$6$7GZJXVYlO53hV.M7$750UVg6zKmX.Uj8cmWUxkRnNXxjuZfcm6BxnJceiFD5Zl80sB7jZL0UeHIpw2a3aQRWh.BMH9WuCPdqwj8zxG.:1030:1034:pejcic:/home/pejcic/folder:/bin/ftpsh
        whmcsmybekap@openpanel.co:$6$rDNAW7GZEAJ6zHJm$wYqg.H6USldSPCNz4jbgEi55tJ8hgeDzQCAmhSHfAPyzkJeP1u9E.LaLflQ.7kUbuRtBED7I70.QoCNRlxzEy0:1030:1034:pejcic:/home/pejcic/WHMC_MY_OPENPANEL_DB_BEKAP:/bin/ftpsh
        pejcic_logs:$6$cv9wnxSLeD1VEk.U$dm84PcqygxOWqT/uyMjrICKUPFeAQwOimJ8frihDCxjRfa1BKf6bnHIhWrbfmLrLn2YBSMnNatW09ZZMAS7GT/:1030:1034:pejcic:/etc/apache2/logs/domlogs/pejcic:/bin/ftpsh
        '
    fi
}




import_domlogs() {

    import_domlogs_for_domain() {
        local ssl_log_file="$1"
        local domain="$2"
        local destination_file="/var/log/nginx/domlogs/${domain}.log"
    
        # Check if the source file exists
        if [[ -e "$ssl_log_file" ]]; then
            # Move the file to the destination
            mv "$ssl_log_file" "$destination_file"
            log "Imported logs from file $ssl_log_file to $destination_file"
        else
            log "WARNING: Error importing ssl logs from file: $ssl_log_file"
        fi
    }
    
    if [[ -d "$domain_logs" ]]; then
        ALL_DOMAINS_OWNED_BY_USER=$(opencli domains-user "$cpanel_username")
        for domain in $ALL_DOMAINS_OWNED_BY_USER; do
            ssl_log_file="$domain_logs/$domain-ssl_log"
            if [[ -e "$ssl_log_file" ]]; then
                log  "Importing SSL logs for domain $domain from file: $ssl_log_file"
                import_domlogs_for_domain "$ssl_log_file" "$domain"
            else
                log "SSL logs not available for domain $domain - Skipping"
            fi
        done
    else
        log "WARNING: SSL logs not detected for domains and will not be imported."
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
    fix_perms                                                                  # fix permissions for all files
    restore_php_version "$php_version"                                         # php v needs to run before domains 
    restore_domains                                                            # add domains
    restore_dns_zones                                                          # add dns 
    import_domlogs                                                             # import ssl logs for domains
    restore_mysql "$mysqldir"                                                  # mysql databases, users and grants
    restore_cron                                                               # cronjob
    restore_ssl "$cpanel_username"                                             # ssl certs
    restore_ssh "$cpanel_username"                                             # enable remote ssh for user
    restore_ssh_password "$cpanel_username"                                    # set ssh password same as in backup
    restore_wordpress "$real_backup_files_path" "$cpanel_username"             # import wp sites to sitemanager
    restore_notifications "$real_backup_files_path" "$cpanel_username"         # notification preferences from cp
    restore_startdate "$real_backup_files_path" "$cpanel_username"             # cp account creation date
    
    # STEP 4. IMPORT ENTERPRISE FEATURES
    import_email_accounts_and_data                                             # import emails, filters, forwarders..
    ftp_accounts_import                                                        # import ftp accounts

    # STEP 5. DELETE TMP FILES
    cleanup                                                                    # delete extracter files after import

    # STEP 6. NOTIFY USER
    success_message                                                            # have a 🍺

    # STEP 7. RUN ANY CUSTOM SCRIPTS
    run_custom_post_hook                                                       # any script to run after the import? example: edit dns on cp server, run tests, notify user, etc.

}

# MAIN FUNCTION
define_data_and_log "$@"
