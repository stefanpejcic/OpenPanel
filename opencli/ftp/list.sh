#!/bin/bash
################################################################################
# Script Name: ftp/add.sh
# Description: List FTP sub-users for openpanel user.
# Usage: opencli ftp-list <OPENPANEL_USERNAME>
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 10.09.2024
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

users_dir="/etc/openpanel/ftp/users"

if [ -z "$1" ]; then
    # No argument provided, loop over all users directories
    for openpanel_username in "$users_dir"/*; do
        # Only directories
        if [ -d "$openpanel_username" ]; then
            user_dir_name=$(basename "$openpanel_username")
            users_file="${openpanel_username}/users.list"
            if [ ! -f "$users_file" ]; then
                echo "ERROR: Users list file does not exist for '$user_dir_name'."
                continue
            fi
            echo "FTP sub-users for '$user_dir_name':"
            while IFS='|' read -r username password directory; do
                echo "$username | $directory"
            done < "$users_file"
            echo
        fi
    done
else
    openpanel_username="$1"
    users_file="${users_dir}/${openpanel_username}/users.list"

    if [ ! -f "$users_file" ]; then
        echo "ERROR: Users list file does not exist for '$openpanel_username'."
        exit 1
    fi

    echo "FTP sub-users for '$openpanel_username':"
    while IFS='|' read -r username password directory; do
        echo "$username | $directory"
    done < "$users_file"
fi
