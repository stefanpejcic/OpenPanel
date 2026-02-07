#!/bin/bash
################################################################################
# Script Name: ftp/delete.sh
# Description: Delete FTP sub-user for openpanel user.
# Usage: opencli ftp-delete <username> <openpanel_username> [--debug]
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 22.05.2024
# Last Modified: 06.02.2026
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

# ======================================================================
# Usage
if [ "$#" -lt 2 ] || [ "$#" -gt 3 ]; then
    echo "Usage: opencli ftp-delete <username> <openpanel_username> [--debug]"
    exit 1
fi


# ======================================================================
# Variables
username="${1,,}"
openpanel_username="$2"



mkdir -p /etc/openpanel/ftp/users/${openpanel_username}
touch /etc/openpanel/ftp/users/${openpanel_username}/users.list



# ======================================================================
# Main
docker --context=default exec openadmin_ftp sh -c "deluser $username"
if [ $? -eq 0 ]; then
    sed -i "/^$username|/d" /etc/openpanel/ftp/users/${openpanel_username}/users.list
    echo "Success: FTP user '$username' deleted successfully."
else
    echo "ERROR: Failed to delete FTP user. To debug run this command on terminal: opencli -x ftp-delete $username $openpanel_username"
    exit 1
fi
