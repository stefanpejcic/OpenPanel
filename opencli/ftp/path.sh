#!/bin/bash
################################################################################
# Script Name: ftp/path.sh
# Description: Change FTP path for a user.
# Usage: opencli ftp-path <username> <path> <openpanel_username> [--debug]
# Docs: https://docs.openpanel.co/
# Author: Stefan Pejcic
# Created: 10.09.2024
# Last Modified: 13.07.2025
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
        --debug)
            DEBUG=true
            ;;
        *)
            ;;
    esac
done


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

        docker exec openadmin_ftp sh -c "usermod -d ${path} ${username}"
        
            if [ $? -eq 0 ]; then
                sed -i "/^${username}|/s|/var/www/html/[^|]*|${path}|" /etc/openpanel/ftp/users/$openpanel_username/users.list
                
                # TODO EDIT THIS USER FILE ALSO!
                echo "Success: FTP path for user '$username' changed successfully."
            else
                if [ "$DEBUG" = true ]; then
                    echo "ERROR: Failed to change FTP path with sed command:"
                    echo ""
                    echo "Run the command manually to check for errors."
                else
                    echo "ERROR: Failed to change FTP path. To debug, run this command on terminal: opencli ftp-path $username $path $openpanel_username --debug"
                fi
                exit 1
            fi
}

# Ensure the paths.list file exists
mkdir -p /etc/openpanel/ftp/users/${openpanel_username}

validate_path
change_path
