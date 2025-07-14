#!/bin/bash
################################################################################
# Script Name: admin.sh
# Description: Manage OpenAdmin service and Administrators.
# Usage: opencli admin <setting_name> 
# Author: Stefan Pejcic
# Created: 01.11.2023
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

CONFIG_FILE_PATH='/etc/openpanel/openpanel/conf/openpanel.config'
FORBIDDEN_USERNAMES_FILE="/etc/openpanel/openadmin/config/forbidden_usernames.txt"
service_name="admin"
admin_logs_file="/var/log/openpanel/admin/error.log"
admin_access_log="/var/log/openpanel/admin/access.log"
admin_api_log="/var/log/openpanel/admin/api.log"
admin_login_log="/var/log/openpanel/admin/login.log"
admin_failed_login_log="/var/log/openpanel/admin/failed_login.log"
admin_crons_log="/var/log/openpanel/admin/cron.log"
db_file_path="/etc/openpanel/openadmin/users.db"
GREEN='\033[0;32m'
RED='\033[0;31m'
RESET='\033[0m'



# IP SERVERS
SCRIPT_PATH="/usr/local/admin/core/scripts/ip_servers.sh"
if [ -f "$SCRIPT_PATH" ]; then
    source "$SCRIPT_PATH"
else
    IP_SERVER_1=IP_SERVER_2=IP_SERVER_3="https://ip.openpanel.com"
fi



# Display usage information
usage() {
    echo "Usage: opencli admin <command> [options]"
    echo ""
    echo "Commands:"
    echo "  on                                            Enable and start the OpenAdmin service."
    echo "  off                                           Stop and disable the OpenAdmin service."
    echo "  log                                           Display the last 25 lines of the OpenAdmin error log."
    echo "  logs                                          Display live logs for all OpenAdminn services."
    echo "  list                                          List all current admin users."
    echo "  new <user> <pass>                             Add a new user with the specified username and password."
    echo "  password <user> <pass>                        Reset the password for the specified admin user."
    echo "  rename <old> <new>                            Change the admin username."
    echo "  suspend <user>                                Suspend admin user."
    echo "  unsuspend <user>                              Unsuspend admin user."
    echo "  notifications <command> <param> [value]       Control notification preferences."
    echo ""
    echo "  Notifications Commands:"
    echo "    check                                       Check and write notifications."
    echo "    get <param>                                 Get the value of the specified notification parameter."
    echo "    update <param> <value>                      Update the specified notification parameter with the new value."
    echo ""
    echo "Examples:"
    echo "  opencli admin on"
    echo "  opencli admin off"
    echo "  opencli admin log"
    echo "  opencli admin logs"
    echo "  opencli admin list"
    echo "  opencli admin new stefan SuperStrong1"
    echo "  opencli admin password stefan SuperStrong2"
    echo "  opencli admin rename stefan pejcic"
    echo "  opencli admin suspend pejcic"
    echo "  opencli admin unsuspend pejcic"
    echo "  opencli admin notifications check"
    echo "  opencli admin notifications get ssl"
    echo "  opencli admin notifications update ssl true"
    exit 1
}




usage_for_notifications() {
    echo "Usage: opencli admin notifications <get|update> <what> <value>"
    echo ""
    echo " Commands:"
    echo "   get <param>                                 Get the value of the specified notification parameter."
    echo "   update <param> <value>                      Update the specified notification parameter with the new value."
    echo ""
    echo "Examples:"
    echo "  opencli admin notifications get reboot        - if 'yes' you will receive notification when server is restarted."
    echo "  opencli admin notifications get login         - if 'yes' you will receive notification for new logins to OpenAdmin."
    echo "  opencli admin notifications update update yes - receive notification when new OpenPanel update is available."
    echo "  opencli admin notifications update cpu 90     - receive notification when server CPU usage is over 90%."
    echo "  opencli admin notifications update du 95      - receive notification when server disk usage is over 95%."
    echo ""
    echo "List of all available options: https://dev.openpanel.com/cli/admin.html#Options"
    exit 1


reboot=yes
attack=yes
limit=yes
backup=yes
update=yes
login=yes
services=panel,admin,nginx,docker,mysql,named,csf,certbot
load=20
cpu=90
ram=85
du=85
swap=40




    
}

read_config() {
    config=$(awk -F '=' '/\[DEFAULT\]/{flag=1; next} /\[/{flag=0} flag{gsub(/^[ \t]+|[ \t]+$/, "", $1); gsub(/^[ \t]+|[ \t]+$/, "", $2); print $1 "=" $2}' $CONFIG_FILE_PATH)
    echo "$config"
}

get_ssl_status() {
    config=$(read_config)
    ssl_status=$(echo "$config" | grep -i 'ssl' | cut -d'=' -f2)
    [[ "$ssl_status" == "yes" ]] && echo true || echo false
}

get_admin_url() {
    caddyfile="/etc/openpanel/caddy/Caddyfile"
    CADDY_CERT_DIR="/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/"

    domain_block=$(awk '/# START HOSTNAME DOMAIN #/{flag=1; next} /# END HOSTNAME DOMAIN #/{flag=0} flag {print}' "$caddyfile")
    domain=$(echo "$domain_block" | sed '/^\s*$/d' | grep -v '^\s*#' | head -n1)
    domain=$(echo "$domain" | sed 's/[[:space:]]*{//' | xargs)
    domain=$(echo "$domain" | sed 's|^http[s]*://||')
        
    if [ -z "$domain" ] || [ "$domain" = "example.net" ]; then
        ip=$(get_public_ip)
        admin_url="http://${ip}:2087/"
    else
        if [ -f "${CADDY_CERT_DIR}/${domain}/${domain}.crt" ] && [ -f "${CADDY_CERT_DIR}/${domain}/${domain}.key" ]; then
            admin_url="https://${domain}:2087/"
        else
            ip=$(get_public_ip)
            admin_url="http://${ip}:2087/"
        fi
    fi
    
    echo "$admin_url"    
}


get_public_ip() {
    ip=$(curl --silent --max-time 2 -4 $IP_SERVER_1 || wget --timeout=2 -qO- $IP_SERVER_2 || curl --silent --max-time 2 -4 $IP_SERVER_3)
        
    # Check if IP is empty or not a valid IPv4
    if [ -z "$ip" ] || ! [[ "$ip" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        ip=$(hostname -I | awk '{print $1}')
    fi
    echo "$ip"
}


detect_service_status() {
    if systemctl is-active --quiet $service_name; then
        admin_url=$(get_admin_url)
        echo -e "${GREEN}●${RESET} OpenAdmin is running and is available on: $admin_url"
    else
         echo -e "${RED}×${RESET} OpenAdmin is not running. To enable it run 'opencli admin on' "
    fi
}




add_new_user() {
    local username="$1"
    local password="$2"
    local flag="$3"
    local password_hash=$(/usr/local/admin/venv/bin/python3 /usr/local/admin/core/users/hash "$password")    
    local user_exists=$(sqlite3 "$db_file_path" "SELECT COUNT(*) FROM user WHERE username='$username';")

    if [ "$user_exists" -gt 0 ]; then
        echo -e "${RED}Error${RESET}: Username '$username' already exists."
    else
    # Define the SQL commands
    
    create_table_sql="CREATE TABLE IF NOT EXISTS user (id INTEGER PRIMARY KEY, username TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL, role TEXT NOT NULL DEFAULT 'user', is_active BOOLEAN DEFAULT 1 NOT NULL);"
    if [ "$flag" == "--reseller" ]; then
        role="reseller"
    elif [ "$flag" == "--super" ]; then
        admin_check_sql="SELECT COUNT(*) FROM user WHERE role = 'admin';"
        admin_count=$(sqlite3 "$db_file_path" "$admin_check_sql")
        if [ "$admin_count" -eq 0 ]; then
            role="admin"
        else
            echo "An Super Admin user already exists. Cannot create another super admin."
            exit 1
        fi
    else
        role="user"
    fi

    insert_user_sql="INSERT INTO user (username, password_hash, role) VALUES ('$username', '$password_hash', '$role');"

    # Execute the SQL commands
    output=$(sqlite3 "$db_file_path" "$create_table_sql" "$insert_user_sql" 2>&1)
        if [ $? -ne 0 ]; then
            # Output the error and the exact SQL command that failed
            echo "User not created: $output"
            # TODO: on debug only! echo "Failed SQL Command: $insert_user_sql"
        else
            if [ "$flag" == "--reseller" ]; then
	    	local resellers_template="/etc/openpanel/openadmin/config/reseller_template.json"
	    	local resellers_dir="/etc/openpanel/openadmin/resellers"
	    	mkdir -p $resellers_dir
      		cp $resellers_template $resellers_dir/$username.json
                echo "Reseller user '$username' created."
            elif [ "$flag" == "--super" ]; then
                echo "Super Administrator '$username' created."
            else
                echo "Admin User '$username' created."
            fi
        fi
    fi
}






# Function to update the username for provided user
update_username() {
    local old_username="$1"
    local new_username="$2"
    local user_exists=$(sqlite3 "$db_file_path" "SELECT COUNT(*) FROM user WHERE username='$old_username';")
    local new_user_exists=$(sqlite3 "$db_file_path" "SELECT COUNT(*) FROM user WHERE username='$new_username';")

    if [ "$user_exists" -gt 0 ]; then
        if [ "$new_user_exists" -gt 0 ]; then
            echo -e "${RED}Error${RESET}: Username '$new_username' already taken."
        else
       
            sqlite3 $db_file_path "UPDATE user SET username='$new_username' WHERE username='$old_username';"
	    
	    sed -i "s/\b$old_username\b/$new_username/g" /var/log/openpanel/admin/login.log   > /dev/null 2>&1
     
            echo "User '$old_username' renamed to '$new_username'."
            
            local reseller_limits_dir="/etc/openpanel/openadmin/resellers"
	    mv $reseller_limits_dir/$old_username.json $reseller_limits_dir/$new_username.json  > /dev/null 2>&1
   
        fi
    else
        echo -e "${RED}Error${RESET}: User '$old_username' not found."
    fi
}   

# Function to update the password for provided user
update_password() {
    local username="$1"
    local user_exists=$(sqlite3 "$db_file_path" "SELECT COUNT(*) FROM user WHERE username='$username';")    
    local password_hash=$(/usr/local/admin/venv/bin/python3 /usr/local/admin/core/users/hash "$new_password")

    if [ "$user_exists" -gt 0 ]; then
        sqlite3 $db_file_path "UPDATE user SET password_hash='$password_hash' WHERE username='$username';"        
        echo "Password for user '$username' changed."
        echo ""
        printf "=%.0s"  $(seq 1 63)
        echo ""
        detect_service_status
        echo ""
        echo "- username: $username"
        echo "- password: $new_password"
        echo ""
        printf "=%.0s"  $(seq 1 63)
        echo ""
    else
        echo -e "${RED}Error${RESET}: User '$username' not found."
    fi
}



list_current_users() {
users=$(sqlite3 "$db_file_path" "SELECT username, role, is_active FROM user;")
echo "$users"
}

suspend_user() {
    local username="$1"
    local user_exists=$(sqlite3 "$db_file_path" "SELECT COUNT(*) FROM user WHERE username='$username';")
    local is_admin=$(sqlite3 "$db_file_path" "SELECT COUNT(*) FROM user WHERE username='$username' AND role='admin';")

    if [ "$user_exists" -gt 0 ]; then
        if [ "$is_admin" -gt 0 ]; then
            echo -e "${RED}Error${RESET}: Cannot suspend user '$username' with 'admin' role."
        else
            sqlite3 $db_file_path "UPDATE user SET is_active='0' WHERE username='$username';"
            echo "User '$username' suspended successfully."

            #echo ""
            #echo "Suspending accounts owned by the reseller $username"
            query_for_usernames="SELECT username FROM users WHERE owner='$username';"
            usernames=$(mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "$query_for_usernames" -se)
            if [ $? -eq 0 ]; then
                for user in $usernames; do
                    echo "User: $user"
                    opencli user-suspend "$user"
                done
            fi         
           
        fi
    else
        echo -e "${RED}Error${RESET}: User '$username' does not exist."
    fi

}

unsuspend_user() {
    local username="$1"
    local user_exists=$(sqlite3 "$db_file_path" "SELECT COUNT(*) FROM user WHERE username='$username';")

    if [ "$user_exists" -gt 0 ]; then
            sqlite3 $db_file_path "UPDATE user SET is_active='1' WHERE username='$username';"
            echo "User '$username' unsuspended successfully."

            #echo ""
            #echo "Unsuspending accounts owned by the reseller $username"
            query_for_usernames="SELECT username FROM users WHERE owner='$username';"
            usernames=$(mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "$query_for_usernames" -se)
            if [ $? -eq 0 ]; then
                for user in $usernames; do
                    echo "User: $user"
                    opencli user-unsuspend "$user"
                done
            fi         
    else
        echo -e "${RED}Error${RESET}: User '$username' does not exist."
    fi
}

delete_existing_users() {
    local username="$1"
    local user_exists=$(sqlite3 "$db_file_path" "SELECT COUNT(*) FROM user WHERE username='$username';")
    local is_admin=$(sqlite3 "$db_file_path" "SELECT COUNT(*) FROM user WHERE username='$username' AND role='admin';")

    if [ "$user_exists" -gt 0 ]; then
        if [ "$is_admin" -gt 0 ]; then
            echo -e "${RED}Error${RESET}: Cannot delete user '$username' with 'admin' role."
        else
        
            local reseller_limits_file="/etc/openpanel/openadmin/resellers/$username.json"
			rm $reseller_limits_file  > /dev/null 2>&1
        
            sqlite3 $db_file_path "DELETE FROM user WHERE username='$username';"            
            echo "User '$username' deleted successfully."
        fi
    else
        echo -e "${RED}Error${RESET}: User '$username' does not exist."
    fi
}



config_file="/etc/openpanel/openadmin/config/notifications.ini"

# Function to get the current configuration value for a parameter
get_config() {
    param_name="$1"
    param_value=$(grep "^$param_name=" "$config_file" | cut -d= -f2-)
    
    if [ -n "$param_value" ]; then
        echo "$param_value"
    elif grep -q "^$param_name=" "$config_file"; then
        echo "Parameter $param_name has no value."
    else
        echo "Parameter $param_name does not exist. Docs: https://dev.openpanel.com/cli/admin.html#Options"
    fi
}

# Function to update a configuration value
update_config() {
    param_name="$1"
    new_value="$2"

    # Check if the parameter exists in the config file
    if grep -q "^$param_name=" "$config_file"; then
        # Update the parameter with the new value
        sed -i "s/^$param_name=.*/$param_name=$new_value/" "$config_file"
        echo "Updated $param_name to $new_value"
        
    else
        echo "Parameter $param_name not found in the configuration file. Docs: https://dev.openpanel.com/cli/admin.html#Update"
    fi
}

# added validation (only letters and numbers) in 0.2.8
validate_password_and_username() {
    local input="$1"
    local field_name="$2"

    # https://github.com/stefanpejcic/OpenPanel/issues/442
    if [[ "$input" =~ ^\$\$ ]]; then
        echo "ERROR: $field_name cannot start with '\$\$' as it may cause unintended behavior."
	echo "       docs: https://openpanel.com/docs/articles/accounts/forbidden-usernames/#openadmin"
        exit 1
    fi

    
    # Check if input contains only letters and numbers
    if [[ "$input" =~ ^[a-zA-Z0-9_]{5,30}$ ]]; then
        :
    else
        echo "ERROR: $field_name is invalid. It must contain only letters and numbers, and be between 5 and 30 characters."
        echo "       docs: https://openpanel.com/docs/articles/accounts/forbidden-usernames/#openadmin"
        exit 1
    fi
    
    : '
    # TODO: we will at some point include dictionary checks from lists:
    # https://weakpass.com/wordlist
    # https://github.com/steveklabnik/password-cracker/blob/master/dictionary.txt
    #
    DICTIONARY="dictionary.txt"
    # Convert input to lowercase for dictionary check
    local input_lower=$(echo "$input" | tr '[:upper:]' '[:lower:]')

    # Check if input contains any common dictionary word
    if grep -qi "^$input_lower$" "$DICTIONARY"; then
        echo "ERROR: $field_name is invalid. It contains a common dictionary word, which is not allowed."
        exit 1
    fi
    '
}






multitail_admin_logs(){
    check_multitail() {
        if command -v multitail &> /dev/null; then
            return 0
        else
            return 1
        fi
    }
    
    # Function to install multitail
    install_multitail() {
        if command -v apt &> /dev/null; then
            echo "Installing multitail using apt..."
            apt update  > /dev/null 2>&1
            sudo apt install -y multitail > /dev/null 2>&1
        elif command -v dnf &> /dev/null; then
            echo "Installing multitail using dnf..."
            dnf install -y multitail > /dev/null 2>&1
        else
            echo "ERROR: Neither apt nor dnf package manager found. Cannot install multitail."
            exit 1
        fi
    }

        if ! check_multitail; then
            install_multitail
        fi

        all_files_exist=true
        
        # List of required files
        required_files=(
            "$admin_logs_file"
            "$admin_access_log"
            "$admin_login_log"
            "$admin_failed_login_log"
            "$admin_crons_log"
        )
        
        key_value=$(grep "^key=" $CONFIG_FILE_PATH | cut -d'=' -f2-)
        if [ -z "$key_value" ]; then
            required_files+=("$admin_api_log")
        fi
                
        for file in "${required_files[@]}"; do
            if [ ! -f "$file" ]; then
                touch "$file"
            fi
        done
        
        if [ -n "$key_value" ]; then
            multitail "$admin_logs_file" "$admin_access_log" "$admin_login_log" "$admin_failed_login_log" "$admin_crons_log"
        else
            multitail "$admin_logs_file" "$admin_access_log" "$admin_api_log" "$admin_login_log" "$admin_failed_login_log" "$admin_crons_log"
        fi    
}


case "$1" in
    "on")
        # Enable and check
        echo "Enabling the OpenAdmin..."
	rm /root/openadmin_is_disabled > /dev/null 2>&1
        systemctl enable --now $service_name > /dev/null 2>&1
        detect_service_status
        ;;
    "log")
        # tail logs
        echo "Restarting OpenAdmin service:"
        systemctl enable --now $service_name > /dev/null 2>&1
        echo "tail -f 25 $admin_logs_file"
        echo ""
        service admin restart
        tail -25 $admin_logs_file
        echo ""
        ;;
    "logs")
        # tail logs
        multitail_admin_logs
        ;;
    "off")
        # Disable admin panel service
        echo "Disabling the OpenAdmin..."
        systemctl disable --now $service_name > /dev/null 2>&1
	touch /root/openadmin_is_disabled
        detect_service_status
        ;;
    "help")
        # Show usage
        usage
        ;;
    "password")
        # Reset password for admin user
        user_flag="$2"
        new_password="$3"
        # Check if the file exists
        if [ -f "$db_file_path" ]; then
            if [ "$new_password" ]; then
                # valdiate password
                validate_password_and_username "$new_password" "Password"
                # Use provided password
                update_password "$user_flag"
            fi
        else
            echo "Error: File $db_file_path does not exist, password not changed for user."
        fi
                
        ;;
    "rename")
        # Change username
        old_username="$2"
        new_username="$3"
        validate_password_and_username "$new_username" "New Username"
        update_username "$old_username" "$new_username"
        ;;
    "list")
        # List users
        list_current_users
        ;;
    "suspend")
        # List users
        username="$2"
        suspend_user "$username"
        ;;   
    "unsuspend")
        # List users
        username="$2"
        unsuspend_user "$username"
        ;;       
    "new")
        # Add a new user
        new_username="$2"
        new_password="$3"
        reseller="$4"
        # Check if $2 and $3 are provided
        if [ -z "$new_username" ] || [ -z "$new_password" ]; then
            #echo "Error: Missing parameters for new admin command."
            echo "ERROR: Invalid 'opencli admin new' command - please provide username and password."
            usage
            exit 1
        fi
        validate_password_and_username "$new_username" "Username"
        validate_password_and_username "$new_password" "Password"
        add_new_user "$new_username" "$new_password" "$reseller"
        ;;
    "notifications")
        # COntrol notification preferences
        command="$2"
        param_name="$3"
        if [ "$command" != "check" ]; then
            if [ -z "$command" ] || [ -z "$param_name" ]; then
                usage_for_notifications
                exit 1
            fi
        fi
        
        case "$command" in
            check)
                bash  /usr/local/admin/service/notifications.sh
                exit 0
                ;;
            get)
                get_config "$param_name"
                ;;
            update)
                if [ "$#" -ne 4 ]; then
                    echo "ERROR: Usage: opencli admin notifications update <parameter_name> <new_value>"
                    exit 1
                fi
                new_value="$4"
                update_config "$param_name" "$new_value"
                
                case "$param_name" in
                    ssl)
                        update_ssl_config "$new_value"
                        ;;
                    port)
                        update_port_config "$new_value"
                        ;;
                    openpanel_proxy)
                        update_openpanel_proxy_config "$new_value"
                        service nginx reload
                        ;;
                esac
                ;;
            *)
                echo "ERROR: Invalid command."
                usage
                exit 1
                ;;
        esac        
        ;;
        
    "delete")
        # Add a new user
        username="$2"
        delete_existing_users "$username"
        ;;
    *)
    if [ -z "$1" ]; then
        # No argument provided, show service status
        detect_service_status
    else
        # Unknown argument provided, show error
        echo "ERROR: Unknown command: '$1'"
        usage
        exit 1
    fi
        ;;
esac

exit 0
