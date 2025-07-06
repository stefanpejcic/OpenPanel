#!/bin/bash
################################################################################
# Script Name: ftp/password.sh
# Description: Change password for FTP sub-user of openpanel user.
# Usage: opencli ftp-password <username> <new_password> <openpanel_username> [--debug]
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 22.05.2024
# Last Modified: 04.07.2025
# Company: openpanel.co
# Copyright (c) openpanel.co
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

# Check for the correct number of arguments
if [ "$#" -lt 3 ] || [ "$#" -gt 4 ]; then
    echo "Usage: opencli ftp-password <username> <new_password> <openpanel_username> [--debug]"
    exit 1
fi

username="${1,,}"
new_password="$2"
openpanel_username="$3"
DEBUG=false  # Default value for DEBUG

# Parse optional flags to enable debug mode when needed!
for arg in "$@"; do
    case $arg in
        --debug)
            DEBUG=true
            ;;
        *)
            ;;
    esac
done

# Function to update the user's password
update_password() {
    docker exec openadmin_ftp sh -c "echo -e '${new_password}\n${new_password}' | passwd ${username}"
    
    if [ $? -eq 0 ]; then
        # Update the password in the users.list file
        sed -i "/^$username|/d" /etc/openpanel/ftp/users/${openpanel_username}/users.list
        echo "$username|$new_password|/home/$openpanel_username/$username" >> /etc/openpanel/ftp/users/${openpanel_username}/users.list
        echo "Success: FTP user '$username' password updated successfully."
    else
        if [ "$DEBUG" = true ]; then
            echo "ERROR: Failed to update FTP user password with command:"
            echo ""
            echo "docker exec openadmin_ftp sh -c 'echo -e ${new_password}\n${new_password} | passwd ${username}'"
            echo ""
            echo "Run the command manually to check for errors."
        else
            echo "ERROR: Failed to update FTP user password. To debug run this command on terminal: opencli ftp-password $username $new_password $openpanel_username --debug"
        fi
        exit 1
    fi
}

# Check if the FTP user exists
user_exists() {
    local user="$1"
    grep -Fq "$user|" /etc/openpanel/ftp/users/${openpanel_username}/users.list
}

mkdir -p /etc/openpanel/ftp/users/${openpanel_username}
touch /etc/openpanel/ftp/users/${openpanel_username}/users.list

# Check if user exists
if ! user_exists "$username"; then
    echo "Error: FTP User '$username' does not exist."
    exit 1
fi

# Check if password length is at least 8 characters
if [ ${#new_password} -lt 8 ]; then
    echo "ERROR: New password is too short. It must be at least 8 characters long."
    exit 1
fi

# Check if password contains at least one uppercase letter
if ! [[ $new_password =~ [A-Z] ]]; then
    echo "ERROR: New password must contain at least one uppercase letter."
    exit 1
fi

# Check if password contains at least one lowercase letter
if ! [[ $new_password =~ [a-z] ]]; then
    echo "ERROR: New password must contain at least one lowercase letter."
    exit 1
fi

# Check if password contains at least one digit
if ! [[ $new_password =~ [0-9] ]]; then
    echo "ERROR: New password must contain at least one digit."
    exit 1
fi

# Check if password contains at least one special character
if ! [[ $new_password =~ [[:punct:]] ]]; then
    echo "ERROR: New password must contain at least one special character."
    exit 1
fi

# Call the function to update the password
update_password
