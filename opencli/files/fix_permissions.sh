#!/bin/bash
################################################################################
# Script Name: files/fix_permissions.sh
# Description: Fix permissions for users /home directory files inside the container.
# Usage: opencli files-fix_permissions <USERNAME> [PATH]
# Author: Stefan Pejcic
# Created: 15.11.2023
# Last Modified: 06.11.2025
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

verbose=""

ensure_jq_installed() {
    if ! command -v jq &> /dev/null; then
        if command -v apt-get &> /dev/null; then
            apt-get update > /dev/null 2>&1
            apt-get install -y -qq jq > /dev/null 2>&1
        elif command -v yum &> /dev/null; then
            yum install -y -q jq > /dev/null 2>&1
        elif command -v dnf &> /dev/null; then
            dnf install -y -q jq > /dev/null 2>&1
        else
            echo "Error: No compatible package manager found. Please install jq manually and try again."
            exit 1
        fi

        if ! command -v jq &> /dev/null; then
            echo "Error: jq installation failed. Please install jq manually and try again."
            exit 1
        fi
    fi
}

check_and_fix_FTP_permissions() {
    local user="$1"
    local base_path="/home/${user}/docker-data/volumes/${user}_html_data/_data"
    local found=0
    opencli ftp-list "$user" \
        | tail -n +2 \
        | cut -d'|' -f2 \
        | sed 's/^ *//;s/ *$//' \
        | grep -v "^/var/www/html$" \
        | while read -r directory; do

            [ -z "$directory" ] && continue
            found=1

            relative_path="${directory#/var/www/html/}"
            relative_path="${relative_path#/var/www/html}"

            [ -z "$relative_path" ] && continue

            new_directory="${base_path}/${relative_path}"

            if [ -d "$new_directory" ]; then
                echo "[*] Fixing ownership for FTP path: $new_directory"
                chown -R "$uid:$uid" "$new_directory"
            fi
        done

    if [ "$found" -eq 1 ]; then
        chmod +rx "/home/$user" \
                  "/home/$user/docker-data" \
                  "/home/$user/docker-data/volumes" \
                  "/home/$user/docker-data/volumes/${user}_html_data" \
                  "/home/$user/docker-data/volumes/${user}_html_data/_data"
    fi
}



# Function to apply permissions and ownership changes within a Docker container
apply_permissions_in_container() {
  local container_name="$1"
  local path="$2"

        get_user_info() {
            local user="$1"
            local query="SELECT id, server FROM users WHERE username = '${user}';"
            user_info=$(mysql -se "$query")
            
            user_id=$(echo "$user_info" | awk '{print $1}')
            context=$(echo "$user_info" | awk '{print $2}')
            
            echo "$user_id,$context"
        }

        
        result=$(get_user_info "$container_name")
        context=$(echo "$result" | cut -d',' -f2)


        if [ -z "$context" ]; then
            echo "FATAL ERROR: user $container_name does not have a valid docker context."
            exit 1
        fi


        if [ -n "$path" ]; then       
            if [[ $path == /var/www/html/* ]]; then
                path="${path#/var/www/html/}"
            fi
    
            directory="/home/${context}/docker-data/volumes/${context}_html_data/_data/$path"
            fake_directory="/var/www/html/$path"
        else   
            directory="/home/${context}/docker-data/volumes/${context}_html_data/_data/"     
            fake_directory="/var/www/html/"
        fi
    

        # get uid first!
        uid="$(grep -E "^${context}:" /hostfs/etc/passwd | cut -d: -f3)"

        # USERNAME OWNER
        #chown -R $verbose $uid:$uid $directory
        find $directory -print0 | xargs -0 chown $verbose $uid:$uid > /dev/null 2>&1
        owner_result=$?
        
        # FILES
        find $directory -type f -print0 | xargs -0 chmod $verbose 644 > /dev/null 2>&1
        files_result=$?
        
        # FOLDERS
        find $directory -type d -print0 | xargs -0 chmod $verbose 775
        folders_result=$?

        check_and_fix_FTP_permissions "$container_name"
        ftp_result=$?

        # CHECK ALL 4
        if [ $owner_result -eq 0 ] && [ $files_result -eq 0 ] && [ $folders_result -eq 0 ] && [ $ftp_result -eq 0 ]; then
            echo "Permissions applied successfully to $fake_directory"
        else
            echo "Error applying permissions to $fake_directory"
        fi
}



# FLAGS
parse_flags() {
  local args=()
  while [ $# -gt 0 ]; do
    case "$1" in
      --debug) verbose="-v" ;;
      *) args+=("$1") ;;
    esac
    shift
  done
  echo "${args[@]}"
}

args=($(parse_flags "$@"))

# ALL USERS
if [ "${args[0]}" == "--all" ]; then
  ensure_jq_installed
  for username in $(opencli user-list --json | jq -r '.[].username'); do
    apply_permissions_in_container "$username"
  done
else
# SINGLE USER
  username="${args[0]}"
  path="${args[1]:-}"
  [ -z "$username" ] && { echo "Username required"; exit 1; }
  apply_permissions_in_container "$username"
fi
