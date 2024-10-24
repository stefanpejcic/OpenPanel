#!/bin/bash
pid=$$
script_dir=$(dirname "$0")
timestamp="$(date +'%Y-%m-%d_%H-%M-%S')" #used by log file name
start_time=$(date +%s) #used to calculate elapsed time at the end

set -eo pipefail

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

DEBUG=true  # Set this to true to enable debug logging

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

define_data_and_log() {
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

    # Format log file and ensure log directory exists
    base_name="$(basename "$backup_location")"
    base_name_no_ext="${base_name%.*}"
    local log_dir="/var/log/openpanel/admin/imports"
    mkdir -p "$log_dir"
    log_file="$log_dir/${base_name_no_ext}_$(date +'%Y-%m-%d_%H-%M-%S').log"

    # Run the main function and pass plan_name
    echo "Import started, log file: $log_file"

    main "$plan_name"
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
        sudo apt-mark hold linux-image-generic linux-headers-generic

        # Update package list without upgrading
        sudo apt-get update -y >/dev/null 2>&1

        for cmd in "${!commands[@]}"; do
            if ! command_exists "$cmd"; then
                log "Installing ${commands[$cmd]}"
                # Install package without upgrading or installing recommended packages
                sudo apt-get install -y --no-upgrade --no-install-recommends "${commands[$cmd]}" >/dev/null 2>&1
            fi
        done

        # Unhold kernel packages
        sudo apt-mark unhold linux-image-generic linux-headers-generic

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
    log "Extracting backup from $backup_location to $backup_dir"
    mkdir -p "$backup_dir"

    # Extract the backup
    if [ "$extraction_command" = "unzip" ]; then
        $extraction_command "$backup_location" -d "$backup_dir"
    else
        $extraction_command "$backup_location" -C "$backup_dir"
    fi
    if [ $? -eq 0 ]; then
        log "Backup extracted successfully."
    else
        log "FATAL ERROR: Backup extraction failed."
        cleanup
        exit 1
    fi

    # Handle nested archives (common in some cPanel backups)
    for nested_archive in "$backup_dir"/*.tar.gz "$backup_dir"/*.tgz; do
        if [ -f "$nested_archive" ]; then
            log "Found nested archive: $nested_archive"
            tar -xzf "$nested_archive" -C "$backup_dir"
            rm "$nested_archive"  # Remove the nested archive after extraction
        fi
    done
}

# LOCATE FILES IN EXTRACTED BACKUP
locate_backup_directories() {
    log "Locating important files in the extracted backup"

    if [ -d "$real_backup_files_path/homedir" ]; then
        homedir="$real_backup_files_path/homedir"
        log "Home directory: $homedir"
    else
        log "ERROR: Home directory not found in backup."
        exit 1
    fi

    if [ -d "$real_backup_files_path/mysql" ]; then
        mysqldir="$real_backup_files_path/mysql"
        log "MySQL directory: $mysqldir"
    else
        log "ERROR: MySQL directory not found in backup."
        exit 1
    fi

    if [ -f "$real_backup_files_path/mysql" ]; then
        log "MySQL grants: $real_backup_files_path/mysql"
    else
        log "MySQL grants file not found."
    fi

    if [ -f "$real_backup_files_path/cp/$cpanel_username" ]; then
        cpconfig="$real_backup_files_path/cp/$cpanel_username"
        log "cPanel configuration: $cpconfig"
    else
        log "ERROR: cPanel configuration file not found."
        exit 1
    fi

    log "Backup directories located successfully"
}

# CPANEL BACKUP METADATA
parse_cpanel_metadata() {
    log "Starting to parse cPanel metadata..."

    cp_file="${real_backup_files_path}/cp/${cpanel_username}"
    log "DEBUG: Attempting to parse metadata from file: $cp_file"

    if [ ! -f "$cp_file" ]; then
        log "WARNING: cp file $cp_file not found. Using default values."
        main_domain=""
        cpanel_email=""
        php_version="inherit"
    else
        # Function to get value from cp file with default
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

        # Check for CloudLinux PHP Selector
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
        log "IP Address: $ip_address"
        log "Plan: $plan"
        log "Max Addon Domains: $max_addon"
        log "Max FTP Accounts: $max_ftp"
        log "Max SQL Databases: $max_sql"
        log "Max Email Accounts: $max_pop"
        log "Max Subdomains: $max_sub"
    fi

    # Ensure we have at least an empty string for each variable
    main_domain="${main_domain:-}"
    cpanel_email="${cpanel_email:-}"
    php_version="${php_version:-inherit}"

    log "cPanel metadata parsed:"
    log "Main Domain: ${main_domain:-Not found}"
    log "Email: ${cpanel_email:-Not found}"
    log "PHP Version: $php_version"
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
        log "Username $cpanel_username is available, starting import.."
    else
        log "FATAL ERROR: $cpanel_username already exists."
        exit 1
    fi
}

# CREATE NEW USER
create_new_user() {
    local username="$1"
    local password="$2"
    local email="$3"
    local plan_name="$4"

    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: Would create user $username with password $password, email $email, and plan $plan_name"
        return
    fi

    # Validate the parameters
    if [ -z "$username" ] || [ -z "$password" ] || [ -z "$email" ] || [ -z "$plan_name" ]; then
        log "ERROR: Missing required parameters. Username, password, email, and plan must be provided."
        return
    fi

    # Log the creation attempt
    log "Attempting to create user $username with email $email, plan $plan_name, and password $password"

    # Construct the correct opencli command
    local cli_command="opencli user-add $username $password $email $plan_name"

    # Log the exact command that will be run
    log "Running command: $cli_command"

    # Execute the OpenCLI command
    create_user_command=$($cli_command 2>&1)

    # Process the output and log it
    while IFS= read -r line; do
        log "$line"
    done <<< "$create_user_command"

    # Check if the user was added successfully
    if echo "$create_user_command" | grep -q "Successfully added user"; then
        log "User $username was added successfully."
    else
        log "ERROR: Failed to create user $username. Command output: $create_user_command"
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
        local current_version=$(opencli php-default_php_version "$cpanel_username" | sed 's/Default PHP version for user.*: //')
        if [ "$current_version" != "$php_version" ]; then
            local installed_versions=$(opencli php-enabled_php_versions "$cpanel_username")
            if ! echo "$installed_versions" | grep -q "$php_version"; then
            log "Default PHP version $php_version from backup is not present in the container, installing.."
                output=$(opencli php-install_php_version "$cpanel_username" "$php_version" 2>&1)
                while IFS= read -r line; do
                    log "$line"
                done <<< "$output"

                #SET AS DEFAULT PHP VERSION
                log "Setting newly installed PHP $php_version as the default version for all new domains."
                output=$(opencli php-default_php_version "$cpanel_username" --update "$php_version" 2>&1)
                while IFS= read -r line; do
                    log "$line"
                done <<< "$output"
            fi
        else
        log "Default PHP version in backup file ($php_version) matches the installed PHP version: ($current_version) "
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
    local mysql_backup_dir="$1"
    local mysql_grants_file="$real_backup_files_path/mysql"

    log "Starting MySQL database restoration..."
    log "MySQL backup directory: $mysql_backup_dir"

    if [ ! -d "$mysql_backup_dir" ]; then
        log "No MySQL backup directory found at $mysql_backup_dir."
        return
    fi

    # Enable nullglob to handle the case when no .sql files are found
    shopt -s nullglob
    sql_files=("$mysql_backup_dir"/*.sql)
    shopt -u nullglob

    if [ ${#sql_files[@]} -eq 0 ]; then
        log "No SQL files found in $mysql_backup_dir."
    else
        for sql_file in "${sql_files[@]}"; do
            db_name=$(basename "$sql_file" .sql)
            log "Restoring database: $db_name from $sql_file"

            # Drop the database if it exists
            log "Dropping database $db_name if it exists..."
            if mysql -e "DROP DATABASE IF EXISTS \`$db_name\`;" 2>&1 | while read -r line; do log "$line"; done; then
                log "Database $db_name dropped (if it existed)."
            else
                log "ERROR: Failed to drop database $db_name."
                continue
            fi

            # Create the database
            log "Creating database $db_name..."
            if mysql -e "CREATE DATABASE \`$db_name\`;" 2>&1 | while read -r line; do log "$line"; done; then
                log "Database $db_name created successfully."
            else
                log "ERROR: Failed to create database $db_name."
                continue
            fi

            # Import the SQL file
            log "Importing SQL file $sql_file into database $db_name..."
            if mysql "$db_name" < "$sql_file" 2>&1 | while read -r line; do log "$line"; done; then
                log "Database $db_name restored successfully."
            else
                log "ERROR: Failed to import SQL file into database $db_name."
                continue
            fi
        done
    fi

    # Restore MySQL users and grants from mysql grants file if available
    if [ -f "$mysql_grants_file" ]; then
        log "Restoring MySQL users and grants from $mysql_grants_file"

        # Check MySQL version to adjust grant file syntax if necessary
        mysql_version=$(mysql -V | awk '{print $5}' | awk -F'.' '{print $1"."$2}')
        log "Detected MySQL version: $mysql_version"

        temp_grants_file="/tmp/processed_mysql_grants.sql"

        if (( $(echo "$mysql_version >= 8.0" | bc -l) )); then
            log "Modifying grants file for MySQL 8.0 compatibility..."
            # Replace 'IDENTIFIED BY PASSWORD' with 'IDENTIFIED WITH mysql_native_password AS'
            sed "s/IDENTIFIED BY PASSWORD/IDENTIFIED WITH mysql_native_password AS/g" "$mysql_grants_file" > "$temp_grants_file"
        else
            log "Using grants file as-is for MySQL version < 8.0"
            cp "$mysql_grants_file" "$temp_grants_file"
        fi

        # Execute the grants file
        if mysql < "$temp_grants_file" 2>&1 | while read -r line; do log "$line"; done; then
            log "MySQL users and grants restored successfully."
        else
            log "ERROR: Failed to restore MySQL users and grants."
        fi

        # Remove the temporary grants file
        rm -f "$temp_grants_file"
    else
        log "No MySQL grants file found at $mysql_grants_file."
    fi

    log "MySQL database restoration completed."
}
# SSL CACHE
refresh_ssl_file() {
    local username="$1"

    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: Would refresh SSL file for user $username"
        return
    fi

    log "Creating a list of SSL certificates for user interface"
    opencli ssl-user "$cpanel_username"

}

# SSL CERTIFICATES
restore_ssl() {
    local username="$1"

    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: Would restore SSL certificates for user $username"
        return
    fi

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
    log "Restoring files ($backup_size) to /home/$cpanel_username/"
    log "real_backup_files_path is set to: $real_backup_files_path"

    if [ ! -d "/home/$cpanel_username" ]; then
        log "/home/$cpanel_username does not exist. Creating directory."
        mkdir -p "/home/$cpanel_username"
        log "Successfully created /home/$cpanel_username."
    else
        log "/home/$cpanel_username exists."
    fi

    # Optionally comment out the listing of homedir contents
    # log "Contents of $real_backup_files_path/homedir before moving:"
    # ls -alh "$real_backup_files_path/homedir" | while IFS= read -r line; do
    #     log "$line"
    # done

    log "Attempting to sync $real_backup_files_path/homedir to /home/$cpanel_username/"
    rsync -a "$real_backup_files_path/homedir/" "/home/$cpanel_username/"
    if [ $? -eq 0 ]; then
        log "Successfully synced homedir to /home/$cpanel_username."
    else
        log "ERROR: Failed to sync homedir to /home/$cpanel_username."
    fi

    # Moving main domain files
    log "Moving main domain files from public_html to $main_domain directory using rsync."
    if [ ! -d "/home/$cpanel_username/$main_domain" ]; then
        mkdir -p "/home/$cpanel_username/$main_domain"
    fi
    rsync -a "/home/$cpanel_username/public_html/" "/home/$cpanel_username/$main_domain/"
    if [ $? -eq 0 ]; then
        log "Successfully synced public_html to $main_domain."
    else
        log "ERROR: Failed to sync public_html to $main_domain."
    fi

    # Remove public_html if needed
    rm -rf "/home/$cpanel_username/public_html"
}
# PERMISSIONS
fix_perms(){
    log "Changing permissions for all files and folders in user home directory /home/$cpanel_username/"
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

    log "Restoring WordPress sites for user $username"
    if [ -d "$real_backup_files_path/wptoolkit" ]; then
        for wp_file in "$real_backup_files_path/wptoolkit"/*.json; do
            log "Importing WordPress site from: $wp_file"
            opencli wp-import "$username" "$wp_file"
        done
    else
        log "No WordPress data found to restore"
    fi
}




# DOMAINS
create_domain() {
    local domain="$1"
    local type="$2"
    local username="$3"

    log "DEBUG: domain='$domain', type='$type', username='$username'"

    current_domain_count=$((current_domain_count + 1))
    log "Restoring $type $domain (${current_domain_count}/${domains_total_count})"

    # Log the command being executed
    log "Running command: opencli domains-add '$domain' '$username'"

    if [ "$DRY_RUN" = true ]; then
        log "DRY RUN: Would restore $type $domain"
    else
        # Check if the domain already exists
        domain_check_output=$(opencli domains-whoowns "$domain" 2>&1)
        if echo "$domain_check_output" | grep -q "not found in the database"; then
            # Proceed to add the domain
            output=$(opencli domains-add "$domain" "$username" 2>&1)
            log "opencli domains-add output: $output"
            if echo "$output" | grep -q "Successfully added domain"; then
                log "Domain $domain added successfully."
            else
                log "ERROR: Failed to add domain $domain. Command output: $output"
            fi
        else
            log "WARNING: $type $domain already exists and will not be added to this user."
        fi
    fi
}
restore_domains() {
    log "Starting domain restoration process..."

    if [ -f "$real_backup_files_path/userdata/main" ]; then
        file_path="$real_backup_files_path/userdata/main"

        # Log the content of the userdata/main file for debugging
        log "Contents of userdata/main:"
        while IFS= read -r line; do
            log "$line"
        done < "$file_path"

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
            log "Addon domains ($addon_domains_count): ${addon_domains_array[@]}"
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
            log "Subdomains ($filtered_sub_domains_count): ${filtered_sub_domains[@]}"
        fi

        domains_total_count=$((main_domain_count + addon_domains_count + parked_domains_count + filtered_sub_domains_count))

        log "Detected a total of $domains_total_count domains for user."

        current_domain_count=0

        # Process the domains
        log "Processing main (primary) domain..."
        create_domain "$main_domain" "main domain" "$cpanel_username"

        if [ "$parked_domains_count" -eq 0 ]; then
            log "No parked (alias) domains detected."
        else
            log "Processing parked (alias) domains..."
            for parked in "${parked_domains_array[@]}"; do
                create_domain "$parked" "alias domain" "$cpanel_username"
            done
        fi

        if [ "$addon_domains_count" -eq 0 ]; then
            log "No addon domains detected."
        else
            log "Processing addon domains..."
            for addon in "${addon_domains_array[@]}"; do
                create_domain "$addon" "addon domain" "$cpanel_username"
            done
        fi

        if [ "$filtered_sub_domains_count" -eq 0 ]; then
            log "No subdomains detected."
        else
            log "Processing subdomains..."
            for filtered_sub in "${filtered_sub_domains[@]}"; do
                create_domain "$filtered_sub" "subdomain" "$cpanel_username"
                # TODO: create record in DNS zone instead of separate domain if only DNS zone and no folder!
            done
        fi

        log "Finished importing $domains_total_count domains."

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

# Main execution
# Main execution
main() {
    local plan_name="$1"
    # ... previous steps ...

    parse_cpanel_metadata

    # Now that we have cpanel_username and cpanel_email, create the user
    create_new_user "$cpanel_username" "random" "$cpanel_email" "$plan_name"

    # Restore domains (this creates domain directories)
    restore_domains

    # Restore files (now that domain directories exist)
    restore_files

    # Adjust permissions
    fix_perms

    # Restore PHP version
    restore_php_version "$php_version"

    # Restore MySQL databases and users
    restore_mysql "$mysqldir"

    # ... remaining steps ...

    # Cleanup
    cleanup

    end_time=$(date +%s)
    elapsed=$(( end_time - start_time ))
    hours=$(( elapsed / 3600 ))
    minutes=$(( (elapsed % 3600 ) / 60 ))
    seconds=$(( elapsed % 60 ))

    log "Elapsed time: ${hours}h ${minutes}m ${seconds}s"

    log "SUCCESS: Import for user $cpanel_username completed successfully."

    # Run post-hook if provided
    if [ -n "$post_hook" ]; then
        if [ -x "$post_hook" ]; then
            log "Executing post-hook script..."
            "$post_hook" "$cpanel_username"
        else
            log "WARNING: Post-hook file '$post_hook' is not executable or not found."
            exit 1
        fi
    fi
}

# MAIN FUNCTION
define_data_and_log "$@"