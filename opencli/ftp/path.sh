#!/bin/bash
################################################################################
# Script Name: ftp/path.sh
# Description: Change FTP path for a user.
# Usage: opencli ftp-path <username> <path> <openpanel_username> [--debug]
# Docs: https://docs.openpanel.com/
# Author: Stefan Pejcic
# Created: 10.09.2024
# Last Modified: 03.06.2026
# Company: openpanel.com
# Copyright (c) openpanel.com
# 
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation (the "Software"), to deal
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

# Check for the correct number of arguments
if [ "$#" -ne 3 ] && [ "$#" -ne 4 ]; then
    echo "Usage: opencli ftp-path <username> <path> <openpanel_username> [--debug]"
    exit 1
fi

username="$1"
path="$2"
openpanel_username="$3"
DEBUG=false  # Default value for DEBUG

# Parse optional flags to enable debug mode when needed
for arg in "$@"; do
    case $arg in
        --debug) DEBUG=true ;;
        *) ;;
    esac
done


get_docker_context_for_user(){
    context=$(mysql -e "SELECT server FROM users WHERE username='$openpanel_username';" -N)   
	context=$(mysql -N -e "
	SELECT u.server
	FROM users u
	WHERE u.username='$openpanel_username'
	AND EXISTS (
	    SELECT 1
	    FROM domains d
	    WHERE d.domain_url = SUBSTRING_INDEX('$username','@',-1)
	      AND d.user_id = u.id
	);
	")

    if [ -z "$context" ]; then
        echo "ERROR: No context found for user '$openpanel_username' - or does not own the domain name. Aborting!"
        exit 1
    fi    
}

# Validate the path
validate_path() {
    if [[ "$path" != /var/www/html/* ]]; then
        echo "ERROR: Invalid path. It must start with /var/www/html/"
        exit 1
    fi

    if [[ "$path" == *".."* ]]; then
        echo "ERROR: Path traversal detected (.. is not allowed)."
        exit 1
    fi

    if [[ "$path" == *"//"* ]]; then
        echo "ERROR: Double slashes are not allowed in the path."
        exit 1
    fi

    if [[ "$path" == *"|"* ]]; then
        echo "ERROR: The path cannot contain the '|' character."
        exit 1
    fi
}




# Function to add or update the FTP path
change_path() {
    real_path="/home/${context}/docker-data/volumes/${context}_html_data/_data/"
    relative_path="${path##/var/www/html/}"
    container_path="${real_path}${relative_path}"

    docker exec openadmin_ftp sh -c "usermod -d '${container_path}' '${username}'"
    
    if [ $? -eq 0 ]; then
        sed -i "/^${username}|/s|/var/www/html/[^|]*|${path}|" /etc/openpanel/ftp/users/$openpanel_username/users.list
        echo "Success: FTP path for user '$username' changed successfully."
    else
        if [ "$DEBUG" = true ]; then
            echo "ERROR: Failed to change FTP path with usermod command:"
            echo "docker exec openadmin_ftp sh -c 'usermod -d ${container_path} ${username}'"
        else
            echo "ERROR: Failed to change FTP path. To debug, run this command on terminal: opencli ftp-path $username $path $openpanel_username --debug"
        fi
        exit 1
    fi
}


mkdir -p /etc/openpanel/ftp/users/${openpanel_username} # Ensure the paths.list file exists
get_docker_context_for_user                             # validate our op user owns the domain and get context
validate_path
change_path
