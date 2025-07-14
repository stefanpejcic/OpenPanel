#!/bin/bash
################################################################################
# Script Name: user/rename.sh
# Description: Rename username.
# Usage: opencli user-rename <old_username> <new_username>
# Author: Radovan Jecmenica
# Created: 23.11.2023
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
# Check if the correct number of arguments is provided
if [ "$#" -lt 2 ] || [ "$#" -gt 3 ]; then
    echo "Usage: opencli user-rename <old_username> <new_username>"
    exit 1
fi

old_username="$1"
new_username="$2"
DEBUG=false
RENAME_CONTEXT=false
FORBIDDEN_USERNAMES_FILE="/etc/openpanel/openadmin/config/forbidden_usernames.txt"


# Parse optional flags to enable debug mode when needed!
for arg in "$@"; do
    case $arg in
        --debug)
            DEBUG=true
            ;;
        --context)
            RENAME_CONTEXT=true
            ;;     
        *)
            ;;
    esac
done

ensure_jq_installed() {
    # Check if jq is installed
    if ! command -v jq &> /dev/null; then
        # Detect the package manager and install jq
        if command -v apt-get &> /dev/null; then
            sudo apt-get update > /dev/null 2>&1
            sudo apt-get install -y -qq jq > /dev/null 2>&1
        elif command -v yum &> /dev/null; then
            sudo yum install -y -q jq > /dev/null 2>&1
        elif command -v dnf &> /dev/null; then
            sudo dnf install -y -q jq > /dev/null 2>&1
        else
            echo "Error: No compatible package manager found. Please install jq manually and try again."
            exit 1
        fi

        # Check if installation was successful
        if ! command -v jq &> /dev/null; then
            echo "Error: jq installation failed. Please install jq manually and try again."
            exit 1
        fi
    fi
}








check_username_is_valid() {
    is_username_forbidden() {
        local check_username="$1"
        readarray -t forbidden_usernames < "$FORBIDDEN_USERNAMES_FILE"

        # Check against forbidden usernames
        for forbidden_username in "${forbidden_usernames[@]}"; do
            if [[ "${check_username,,}" == "${forbidden_username,,}" ]]; then
                return 0
            fi
        done
    
        return 1
    }


    is_username_valid() {
        local check_username="$1"
    
        # Check if the username meets all criteria
        if [[ "$check_username" =~ [[:space:]] ]] || [[ "$check_username" =~ [-_] ]] || \
           [[ ! "$check_username" =~ ^[a-zA-Z0-9]+$ ]] || \
           (( ${#check_username} < 3 || ${#check_username} > 20 )); then
            return 0
        fi
    
        return 1
    }


    
    # Validate username
    if is_username_valid "$new_username"; then
        echo "Error: The username '$new_username' is not valid. Ensure it is a single word with no hyphens or underscores, contains only letters and numbers, and has a length between 3 and 20 characters."
        echo "       docs: https://openpanel.com/docs/articles/accounts/forbidden-usernames/#openpanel"
        exit 1
    elif is_username_forbidden "$new_username"; then
        echo "Error: The username '$new_username' is not allowed."
        echo "       docs: https://openpanel.com/docs/articles/accounts/forbidden-usernames/#reserved-usernames"
        exit 1
    fi
}



check_if_exists_in_db() {
    
    # DB
    source /usr/local/opencli/db.sh
    
    # Check if the username already exists in the users table
    username_exists_query="SELECT COUNT(*) FROM users WHERE username = '$new_username'"
    username_exists_count=$(mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "$username_exists_query" -sN)
    
    # Check if successful
    if [ $? -ne 0 ]; then
        echo "Error: Unable to check username existence in the database."
        exit 1
    fi
    
    # count > 0) show error and exit
    if [ "$username_exists_count" -gt 0 ]; then
        echo "Error: Username '$new_username' already exists."
        exit 1
    fi

    
    context_exists_query="SELECT COUNT(*) FROM users WHERE server = '$new_username'"
    context_exists_count=$(mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "$context_exists_query" -sN)
    
    # count > 0) show error and exit
    if [ "$context_exists_count" -gt 0 ]; then
        echo "Error: Context '$new_username' already exists."
        exit 1
    fi
    
}




get_context() {


# get user ID from the database
get_user_info() {
    local user="$1"
    local query="SELECT id, server FROM users WHERE username = '${user}';"
    
    # Retrieve both id and context
    user_info=$(mysql -se "$query")
    
    # Extract user_id and context from the result
    user_id=$(echo "$user_info" | awk '{print $1}')
    context=$(echo "$user_info" | awk '{print $2}')
    
    echo "$user_id,$context"
}


result=$(get_user_info "$old_username")
user_id=$(echo "$result" | cut -d',' -f1)
context=$(echo "$result" | cut -d',' -f2)

if [ -z "$user_id" ]; then
    echo "ERROR: user $old_username does not exist."
    exit 1
fi

if ! docker context inspect "${context}" &>/dev/null; then
    echo "ERROR: Context '$context' not found."
    exit 1
fi

}

mv_user_data() {
        mv /etc/openpanel/openpanel/core/users/"$old_username" /etc/openpanel/openpanel/core/users/"$new_username" > /dev/null 2>&1
        rm /etc/openpanel/openpanel/core/users/$new_username/data.json > /dev/null 2>&1
	mv /var/log/caddy/stats/$old_username/ /var/log/caddy/stats/$new_username/ > /dev/null 2>&1
}


# Function to rename user in the database
rename_user_in_db() {
    OLD_USERNAME=$1
    NEW_USERNAME=$2
    
    # Update the username in the database with the suspended prefix
    mysql_query="UPDATE users SET username='$NEW_USERNAME' WHERE username='$OLD_USERNAME';"
    
    mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "$mysql_query"

    if [ $? -eq 0 ]; then
        echo "User '$OLD_USERNAME' successfully renamed to '$NEW_USERNAME'."
    else
        echo "Error: Changing username in database failed!"
    fi
}

rename_env(){
		sed -i -E "s/^USERNAME=.*/USERNAME=\"${NEW_USERNAME}\"/" "$file"
}


# MAIN
check_username_is_valid                                                    # validate username first
check_if_exists_in_db                                                      # check in mysql db
get_context "$old_username"
mv_user_data                                                               # /etc/openpanel/openpanel/{core|stats}
ensure_jq_installed                                                        # just helper for parsing json
rename_user_in_db "$old_username" "$new_username"                          # rename username in mysql db
# rename_env USERNAME is no longer used!
reload_user_quotas
#TODO: rename ftp accounts suffix!
# rename paths

exit 0
