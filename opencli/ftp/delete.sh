#!/bin/bash
################################################################################
# Script Name: ftp/delete.sh
# Description: Delete FTP sub-user for openpanel user.
# Usage: opencli ftp-delete <username> <openpanel_username> [--debug]
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 22.05.2024
# Last Modified: 10.06.2026
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

# validate our op user owns the domain and get context
get_docker_context_for_user

mkdir -p /etc/openpanel/ftp/users/${context}
touch /etc/openpanel/ftp/users/${context}/users.list

docker --context=default exec openadmin_ftp sh -c "deluser $username"
if [ $? -eq 0 ]; then
    sed -i "/^$username|/d" /etc/openpanel/ftp/users/${context}/users.list
    nohup opencli sentinel --action=ftp_delete --title="FTP account deleted" --message="FTP account '$username' has been deleted." >/dev/null 2>&1 &
    disown
    echo "Success: FTP user '$username' deleted successfully."
else
    echo "ERROR: Failed to delete FTP user. To debug run this command on terminal: opencli -x ftp-delete $username $openpanel_username"
    exit 1
fi
