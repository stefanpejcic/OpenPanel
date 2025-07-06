#!/bin/bash
################################################################################
# Script Name: ftp/add.sh
# Description: List FTP sub-users for openpanel user.
# Usage: opencli ftp-list <OPENPANEL_USERNAME>
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 10.09.2024
# Last Modified: 04.07.2025
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


openpanel_username="$1"
users_dir="/etc/openpanel/ftp/users"
users_file="${users_dir}/${openpanel_username}/users.list"

# Check if the file exists
if [ ! -f "$users_file" ]; then
    echo "ERROR: Users list file does not exist for '$openpanel_username'."
    exit 1
fi

# List users from the file
echo "FTP sub-ssers for '$openpanel_username':"
cat "$users_file" | while IFS='|' read -r username password directory; do
    echo "$username | $directory"
done
