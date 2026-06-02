#!/bin/bash
pid=$$
script_dir=$(dirname "$0")
timestamp="$(date +'%Y-%m-%d_%H-%M-%S')"
start_time=$(date +%s)
DEBUG=true



# ======================================================================
# START HELPER FUNCTIONS

usage() {
    echo "Usage: $0 --backup-location='<path>' --plan-name='<plan_name>' [--dry-run] [--reseller=<reseller>]"
    echo
    echo "Example: $0 --backup-location='/home/backup-7.29.2024_13-22-32_pejcic.tar.gz' --plan-name='Standard plan' --dry-run"
    exit 1
}

log() {
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

    for arg in "$@"; do
        case $arg in
            --backup-location=*) backup_location="${arg#*=}" ;;
            --plan-name=*)       plan_name="${arg#*=}" ;;
            --dry-run)           DRY_RUN=true ;;
            --post-hook=*)       post_hook="${arg#*=}" ;;
			--reseller=*)        reseller=$(strip_quotes "${arg#*=}") ;;
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
    commands=(jq)

    [[ "$backup_file" == *.zip ]] && commands+=(unzip)
    [[ "$backup_file" == *.tar || "$backup_file" == *.tar.gz || "$backup_file" == *.tgz || "$backup_file" == *.gz ]] && commands+=(tar pigz)

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

validate_plan_exists(){
	opencli plan-list --json | grep -qw "$plan_name" || { log "FATAL ERROR: Plan name '$plan_name' does not exist."; exit 1; }
}

# END HELPER FUNCTIONS
# ======================================================================



# ======================================================================
# CHECK BACKUP FILE EXTENSION AND DETERMINE SIZE NEEDED FOR RESTORE
check_if_valid_cp_backup() {
    local backup_location="$1"
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
extract_cyberpanel_backup() {
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
        tar --use-compress-program=pigz --checkpoint="$zero_one_percent" --checkpoint-action=dot -xf "$backup_location" -C "$backup_dir" 
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

get_mysql_type_cnf_and_socket() {
    mysql_type=$(grep '^MYSQL_TYPE=' /home/$cyberpanel_username/.env | cut -d '=' -f2 | tr -d '"')
	mysql_socket="/home/$cyberpanel_username/sockets/mysqld/mysqld.sock"
	mysql_cnf="/home/$cyberpanel_username/my.cnf"
}

reload_user_quotas() {
	nohup bash -c 'opencli user-quota' >/dev/null 2>&1 &
	disown
}

collect_stats() {
    nohup bash -c "opencli docker-collect_stats '$cyberpanel_username'" >/dev/null 2>&1 &
	disown
}

# ======================================================================
# PARSE CYBERPANEL BACKUP METADATA FOR ACCOUNT AND SERVICE INFORMATION
parse_metadata() {
    log "Starting to parse CyberPanel metadata from meta.xml..."

    META_FILE="${real_backup_files_path}/meta.xml"

    if [ ! -f "$META_FILE" ]; then
        log "WARNING: meta.xml file $META_FILE not found. Using default values."
        main_domain=""
        cyberpanel_email=""
        php_version="inherit"
		db_count=0
		email_count=0
		domains_count=0
    else

        get_xml_value() {
            local tag="$1"
            grep -oPm1 "(?<=<$tag>)[^<]+" "$META_FILE"
        }
		
		cyberpanel_username=$(get_xml_value "userName")
        main_domain=$(get_xml_value "masterDomain")
        cyberpanel_email=$(get_xml_value "email")
		php_version=$(get_xml_value "phpSelection")
		hashed_password=$(get_xml_value "userPassword")
		php_version="${php_version#PHP }"
		db_count=$(grep -o "<dbName>" "$META_FILE" | wc -l)
		email_count=$(grep -o "<emailAccount>" "$META_FILE" | wc -l)
		domains_count=$(grep -o "<path>" "$META_FILE" | wc -l)
    fi

    main_domain="${main_domain:-Not found}"
    cyberpanel_email="${cyberpanel_email:-admin@$main_domain}"
    php_version="${php_version:-inherit}"
	db_count="${db_count:-0}"
	email_count="${email_count:-0}"
	domains_count="${domains_count:-0}"

    log "Username:             $cyberpanel_username"
    log "Main Domain:          $main_domain"
    log "Email:                $cyberpanel_email"
    log "PHP Version:          $php_version"

    log "Additional metadata parsed:"
	log "Addon Domains Count:  $domains_count"
	log "Database Count:       $db_count"
    log "Email Account Count:  $email_count"

    log "Finished parsing CyberPanel metadata."
}

# ======================================================================
# CHECK USERNAME AVIABILITY BEFORE SGTARTING THE EXPORT PROCESS
check_if_user_exists(){
    log "Username: $cyberpanel_username"

    local existing_user=""
	existing_user=$(opencli user-list --json | jq -r '.data[] | select(.username == "'"$cyberpanel_username"'") | .id')

	if [ -n "$existing_user" ]; then
        log "FATAL ERROR: $cyberpanel_username already exists."
		cleanup
        exit 1
	fi

    log "Username $cyberpanel_username is available"
	log "Starting import process.."
}

# ======================================================================
# CREATE OPENPANEL USER
create_new_user() {
    local username="$1"
    local email="$3"
    local plan_name="$4"
	local reseller_arg=""

	dry_run "Would create user $username${reseller:+ for reseller $reseller} with email $email and plan $plan_name" && return

	if [ -n "$reseller" ]; then
	    log "Creating openpanel user for reseller $reseller and configuring services.."
	    reseller_arg="--reseller=$reseller"
	else
	    log "Creating openpanel user and configuring services.."
	fi

	create_user_command=$(opencli user-add "$cyberpanel_username" generate "$email" "$plan_name" $reseller_arg --no-sentinel 2>&1)
    while IFS= read -r line; do
        log "$line"
    done <<< "$create_user_command"

    if echo "$create_user_command" | grep -q "Successfully added user"; then
        if [ -f "$shadow_file" ]; then
            . /usr/local/opencli/db.sh
            
            safe_hashed_password=$(printf "%s" "$hashed_password" | sed "s/'/''/g")
            safe_username=$(printf "%s" "$username" | sed "s/'/''/g")
            mysql_query="UPDATE users SET password='$safe_hashed_password' WHERE username='$safe_username';"
            mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$mysql_query"
            if [ $? -eq 0 ]; then
                echo "Imported SHA-256 crypt password hash from CyberPanel (will be automatically converted to pbkdf2:sha256 on first user login)"
            else
                echo "Failed to import SHA-256 crypt password hash from CyberPanel"
            fi
        fi       
    else
        log "FATAL ERROR: User addition failed. Response did not contain the expected success message."
        exit 1
    fi
}

# ======================================================================
# PHP VERSION
restore_php_version() {
    local php_version="$1"

    dry_run "Would set default PHP version $php_version for user $cyberpanel_username" && return

    # 'inherit' = OpenPanel default
    if [ "$php_version" == "inherit" ]; then
        log "PHP version is set to inherit. No changes will be made."
    else
        # custom version
        log "Setting PHP $php_version as the default version for all new domains."
        output=$(opencli php-default "$cyberpanel_username" --update "$php_version" 2>&1)
        while IFS= read -r line; do
            log "$line"
        done <<< "$output"
    fi
}

# ======================================================================
# MYSQL
restore_mysql() {
    local sandbox_warning_logged=false

    log "Restoring MySQL databases for user $cyberpanel_username"

    dry_run "Would restore MySQL databases for user $cyberpanel_username" && return

    # Workaround for MariaDB sandbox mode bug
    apply_sandbox_workaround() {
        local db_file="$1"
        local text_to_check='enable the sandbox mode'
        local first_line

        first_line=$(head -n 1 "${real_backup_files_path}/$db_file")
        if echo "$first_line" | grep -q "$text_to_check"; then
            if [ "$sandbox_warning_logged" = false ]; then
                log "WARNING: Database dumps were created on a MariaDB server with '--sandbox' mode. Applying workaround for backwards compatibility to MySQL (BUG: https://jira.mariadb.org/browse/MDEV-34183)"
                sandbox_warning_logged=true
            fi
            tail -n +2 "${real_backup_files_path}/$db_file" > "${real_backup_files_path}/${db_file}.workaround" && \
            mv "${real_backup_files_path}/${db_file}.workaround" "${real_backup_files_path}/$db_file"
        fi
    }


	declare -A db_users_passwords  # key: "dbName:dbUser" -> password
	declare -A db_users_hosts      # key: "dbName:dbUser" -> host
	databases_array=()             # list of database names
	
	current_db=""
	current_user=""
	current_host=""
	current_pass=""
	
	while IFS= read -r line; do
	    case "$line" in
	        *"<dbName>"*)
	            current_db=$(echo "$line" | sed -E 's/.*<dbName>([^<]+)<\/dbName>.*/\1/')
	            databases_array+=("$current_db")
	            ;;
	        *"<dbUser>"*)
	            current_user=$(echo "$line" | sed -E 's/.*<dbUser>([^<]+)<\/dbUser>.*/\1/')
	            ;;
	        *"<dbHost>"*)
	            current_host=$(echo "$line" | sed -E 's/.*<dbHost>([^<]+)<\/dbHost>.*/\1/')
	            ;;
	        *"<password>"*)
	            current_pass=$(echo "$line" | sed -E 's/.*<password>([^<]+)<\/password>.*/\1/')
	            # Save info in associative arrays
	            key="${current_db}:${current_user}"
	            db_users_passwords["$key"]="$current_pass"
	            db_users_hosts["$key"]="$current_host"
	            # Reset user-specific values for next <databaseUsers> block
	            current_user=""
	            current_host=""
	            current_pass=""
	            ;;
	    esac
	done < "$META_FILE"
	

	if compgen -G "$real_backup_files_path/*.sql" > /dev/null; then

		# STEP 1: Start MySQL container
        if [ "$mysql_type" = "mysql" ]; then
            mysql_version="8.0"
            sed -i 's/^MYSQL_VERSION=.*/MYSQL_VERSION="8.0"/' /home/"$cyberpanel_username"/.env
        fi
        log "Initializing $mysql_type service for user"
        cd "/home/$cyberpanel_username/" && docker --context="$cyberpanel_username" compose up -d "$mysql_type" >/dev/null 2>&1

        # STEP 3: Wait for MySQL to be ready (max 300 seconds)
		local max_wait=300
        log "Waiting for MySQL service to start... (max ${max_wait}s)"
        waited=0
        while ! mysql --defaults-file="$mysql_cnf" --socket="$mysql_socket" --execute="SELECT 1;" >/dev/null 2>&1; do
            sleep 2
            waited=$((waited + 2))
            if [ "$waited" -ge "$max_wait" ]; then
                log "ERROR: $mysql_type did not become ready after $max_wait seconds"
                exit 1
            fi
        done
        log "$mysql_type is ready after $waited seconds"

        # STEP 4: Create and import databases
        total_databases=$(ls "$real_backup_files_path"/*.sql 2>/dev/null | wc -l)
        if [ "$total_databases" -gt 0 ]; then
			log "Starting import for $total_databases MySQL databases"
            current_db=1

			for db_name in "${databases_array[@]}"; do
                log "Creating database: $db_name (${current_db}/${total_databases})"
				mysql --defaults-file="$mysql_cnf" --socket="$mysql_socket" --execute="CREATE DATABASE IF NOT EXISTS \`$db_name\`;"

			    sql_file="${real_backup_files_path}/${db_name}.sql"
			    if [[ -f "$sql_file" ]]; then
			        log "Importing tables for database: $db_name"
					[ "$mysql_type" = "mysql" ] && apply_sandbox_workaround "$db_name.sql"
					mysql --defaults-file="$mysql_cnf" --socket="$mysql_socket" "$db_name" < "${real_backup_files_path}/$db_name.sql" && echo "Database: $db_name - Import tables OK" || echo "Database: $db_name - Import tables FAILED"
	    		fi

			    for key in "${!db_users_passwords[@]}"; do
			        if [[ $key == "$db_name:"* ]]; then
			            user="${key#*:}"

					    if [[ -z "$user" ]]; then
					        continue
					    fi

			            if [[ "$user" == "$cyberpanel_username" ]]; then
			                continue
			            fi
						
			            host="${db_users_hosts[$key]}"
			            password="${db_users_passwords[$key]}"
			            log "Creating user '$user'@'%' with access to $db_name"
						mysql --defaults-file="$mysql_cnf" --socket="$mysql_socket" --execute="CREATE USER IF NOT EXISTS '$user'@'%' IDENTIFIED WITH 'mysql_native_password' AS '$password'; GRANT ALL PRIVILEGES ON \`$db_name\`.* TO '$user'@'%'; FLUSH PRIVILEGES;" \
						&& echo "User: $user - create OK" || echo "User: $user - create FAILED"
			        fi
			    done
				current_db=$((current_db + 1))
			done
            log "Finished processing $((current_db - 1)) databases"
        else
            log "WARNING: No MySQL databases found"
        fi

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
    log "Restoring SSL certificates for user $username"
    # apache_tls/ dir has LE certs, custom are in ssl/
	if compgen -G "$real_backup_files_path/*.cert.pem" > /dev/null; then
        dest_dir="/home/$username/docker-data/volumes/${username}_html_data/_data/"
        for cert_file in "$real_backup_files_path"/*.cert.pem; do
            local domain=$(basename "$cert_file" .cert.pem)
            local key_file="$real_backup_files_path/$domain.privkey.pem"
            local new_cert_file="$dest_dir/$domain.crt"
            local new_key_file="$dest_dir/$domain.key"
            if [ -f "$key_file" ]; then
                cp "$key_file" "$new_key_file"
                cp "$cert_file" "$new_cert_file"             
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


create_home_mountpoint() {
    dry_run "Would create a symlink from html_data volume to /home/$cyberpanel_username/" && return
    
sed -i '/^[[:space:]]*volumes:[[:space:]]*$/{
  N
  /- html_data:\/var\/www\/html\// s|$|\n      - html_data:/home/${CONTEXT}/|
}' /home/$cyberpanel_username/docker-compose.yml

}

# ======================================================================
# HOME DIR
restore_files() {
    dry_run "Would restore files from /home/$cyberpanel_username/public_html/ to html_data volume" && return

    du_needed_for_home=$(du -sh "$real_backup_files_path/public_html" | cut -f1)
    log "Restoring public_html ($du_needed_for_home) to html_data volume"
    #rm -rf "$real_backup_files_path/public_html/.trash"
	WWW_DIR="/home/$cyberpanel_username/docker-data/volumes/${cyberpanel_username}_html_data/_data/"
	mkdir -p "$WWW_DIR"	
	shopt -s dotglob
	mv "${real_backup_files_path}public_html/"* "$WWW_DIR"
	shopt -u dotglob
}

# ======================================================================
# PERMISSIONS
fix_perms(){
    local verbose="" #-v
    log "Changing permissions for all files and folders in user home directory /home/$cyberpanel_username/"
    dry_run "Would change permissions with command: find /home/$cyberpanel_username -print0 | xargs -0 chown $verbose $cyberpanel_username:$cyberpanel_username" && return
	chown -R $cyberpanel_username:$cyberpanel_username /home/$cyberpanel_username    
}

# ======================================================================
# WORDPRESS SITES
restore_wordpress() {
    dry_run "Would restore WordPress sites for user $cyberpanel_username" && return

    log "Checking user files for WordPress installations to add to Site Manager interface.."
    output=$(opencli websites-scan $cyberpanel_username)
        while IFS= read -r line; do
            log "$line"
        done <<< "$output"    
}

# ======================================================================
# DOMAINS
restore_domains() {
		domains_count=$((domains_count + 1))
        log "Detected a total of $domains_count domains for user."

        current_domain_count=0
		
		declare -A addon_php_selection  # PHP version per domain
		declare -A addon_paths          # path per domain
		addon_domains_array=()           # list of addon domains
		
		# Read the metadata file line by line
		domain=""
		php_sel=""
		path=""
	
		while IFS= read -r line; do
		    case "$line" in
		        *"<domain>"*)
		            domain=$(echo "$line" | sed -E 's/.*<domain>([^<]+)<\/domain>.*/\1/')
		            ;;
		        *"<phpSelection>"*)
		            php_sel=$(echo "$line" | sed -E 's/.*<phpSelection>([^<]+)<\/phpSelection>.*/\1/')
		            ;;
		        *"<path>"*)
		            path=$(echo "$line" | sed -E 's/.*<path>([^<]+)<\/path>.*/\1/')
		            # Store values in arrays
		            addon_domains_array+=("$domain")
		            addon_php_selection["$domain"]="$php_sel"
		            addon_paths["$domain"]="$path"
		            # Reset variables for next block
		            domain=""
		            php_sel=""
		            path=""
		            ;;
		    esac
		done < "$META_FILE"


        create_domain(){
            domain="$1"
            original_docroot="$2"
			original_php="$3"
			php="${original_php#PHP }"

            current_domain_count=$((current_domain_count + 1))
            if [[ $domain == \*.* ]]; then
                log "WARNING: Skipping wildcard domain $domain"
            else
				if [ -n "$original_docroot" ]; then
					docroot="${original_docroot#/home/$main_domain/}"
					docroot="/var/www/html/$docroot"
				else
					docroot="/var/www/html/$domain"
				fi

                dry_run "Would restore $domain with --docroot ${docroot:-N/A}" && return

                if opencli domains-whoowns "$domain" | grep -q "not found in the database."; then
                    if [ -n "$php" ]; then
                        output=$(opencli domains-add "$domain" "$cyberpanel_username" --docroot "$docroot" --php_version "$php" 2>&1)
                        while IFS= read -r line; do
                            log "$line"
                        done <<< "$output"
                    else
                        output=$(opencli domains-add "$domain" "$cyberpanel_username" --docroot "$docroot" 2>&1)
                        while IFS= read -r line; do
                            log "$line"
                        done <<< "$output"                        
                    fi
                else
                    log "WARNING: $domain already exists and will not be added to this user."
                fi
            fi
        }

        log "Processing main (primary) domain.."
        create_domain "$main_domain" "/home/$main_domain/public_html/"

        if [ "${#addon_domains_array[@]}" -eq 0 ]; then
            log "No addon domains detected."
        else
        	log "Processing ${#addon_domains_array[@]} child/addon domain(s):"
	        for d in "${addon_domains_array[@]}"; do
                create_domain "$d" "${addon_paths[$d]}" "${addon_php_selection[$d]}"
            done
        fi

        log "Finished importing $domains_count domains"
}


# ======================================================================
# POST-IMPORT HOOK
run_custom_post_hook() {
    if [ -n "$post_hook" ]; then
        if [ -x "$post_hook" ]; then
            log "Executing post-hook script.."
            "$post_hook" "$cyberpanel_username"
        else
            log "WARNING: Post-hook file '$post_hook' is not executable or not found."
            exit 1
        fi
    fi
}

create_tmp_dir_and_path() {
    backup_filename="${backup_filename%.*}"
    backup_dir=$(mktemp -d /tmp/cyberpanel_import_XXXXXX)
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

    dry_run "import process for user $cyberpanel_username completed" && return

    log "SUCCESS: Import for user $cyberpanel_username completed successfully."

    nohup opencli sentinel --action=user_create --title="User account '$cyberpanel_username' imported from CyberPanel backup" --message="User account '$cyberpanel_username' has been successfully imported from backup file '$backup_filename'" >/dev/null 2>&1 &
	disown
}

log_paths_are() {
    log "Log file: $LOG_FILE"
    log "PID: $pid"
}

start_message() {
    echo -e "
---------------- STARTING CYBERPANEL ACCOUNT IMPORT ----------------
--------------------------------------------------------------------

Currently supported features:

├─ DOMAINS:
│  ├─ Primary domain, Addons, Aliases and Subdomains
│  └─ SSL certificates
├─ WEBSITES:
│  └─ WordPress instalations
├─ DATABASES:
│    └─ MySQL databases and users
├─ PHP:
│    └─ PHP versions per domain
├─ FILES
└─ ACCOUNT
    └─ CyberPanel account password

***emails, crons, dns, ftp, nodejs/python, postgres are not yet supported***

--------------------------------------------------------------------
  if you experience any errors with this script, please report to
    https://github.com/stefanpejcic/CyberPanel-to-OpenPanel/issues
--------------------------------------------------------------------
"
}


# ======================================================================
# FTP
ftp_accounts_import() {

    if [ -f "$ftp_conf" ]; then
        log "WARNING: Importing PureFTPD accounts is not yet supported"
    fi
}

# ======================================================================
# EMAILS
import_email_accounts_and_data() {
        log "WARNING: Importing Email accounts is not yet supported"

        # TODO:
        # - check setting from openamdin where mails are stored
        # for each email check domain owner is the new user
        # mv email data to domain based dir
        # mv messages to domain based dir
        # list emails for user to confirm import
}


# ======================================================================

write_import_activity() {
    echo "$(date '+%Y-%m-%d %H:%M:%S')  $new_ip  Administrator ROOT user imported CyberPanel backup file" > /etc/openpanel/openpanel/core/users/$cyberpanel_username/activity.log
}



# MAIN
main() {
    start_message                                                              # what will be imported
    log_paths_are                                                              # where will we store the progress
    
    # STEP 1. PRE-RUN CHECKS
    check_if_valid_cp_backup "$backup_location"                                # is it?
    check_if_disk_available                                                    # calculate du needed for extraction
    validate_plan_exists                                                       # check if provided plan exists
    install_dependencies                                                       # install commands we will use for this script
    
    # STEP 2. EXTRACT
    create_tmp_dir_and_path                                                    # create /tmp/.. dir and set the path
    extract_cyberpanel_backup "$backup_location" "${backup_dir}"               # extract the archive
    parse_metadata                                                             # get data and configurations
    check_if_user_exists                                                       # only after extract we have username!

    # STEP 3. IMPORT
	restore_files                                                              # homedir
    create_new_user "$cyberpanel_username" "random" "$cyberpanel_email" "$plan_name"   # create user data and container
    setquota -u $cyberpanel_username 0 0 0 0 /                                     # set unlimited quota while we do import!
    #create_home_mountpoint                                                     # mount /var/www/html/ to /home/USERNAME 
    get_mysql_type_cnf_and_socket                                              # mysql or mariadb, path to socket and my.cnf logins
    fix_perms &                                                                 # fix permissions for all files
    restore_php_version "$php_version" &                                        # php v needs to run before domains 
	wait
    restore_domains                                                            # add domains
    #restore_dns_zones
    restore_mysql                                                              # mysql databases, users and grants
    restore_ssl "$cyberpanel_username"                                         # ssl certs
    restore_wordpress                                                          # import wp sites to sitemanager
    opencli user-quota --update $cyberpanel_username                           # restore quota

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

# ======================================================================
# ENTRYPOINT
define_data_and_log "$@"
