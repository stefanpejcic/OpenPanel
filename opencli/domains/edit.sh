#!/bin/bash
################################################################################
# Script Name: domains/docroot.sh
# Description: AEnter docroot for a domain.
# Usage: opencli domains-edit <DOMAIN_NAME>
# Author: Stefan Pejcic
# Created: 20.08.2024
# Last Modified: 10.11.2025
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

if [ "$#" -lt 1 ]; then
    echo "Usage: opencli domains-edit <DOMAIN_NAME>"
    exit 1
fi

DOMAIN="$1"
USE_WS=0
if [ "$2" == "--ws" ]; then
    USE_WS=1 # webserver dir instead of docroot
fi

OUTPUT=$(opencli domains-whoowns "$DOMAIN" --context --docroot)
USERNAME=$(echo "$OUTPUT" | awk '{print $1}')
CONTEXT=$(echo "$OUTPUT" | awk '{print $2}')
DOCROOT=$(echo "$OUTPUT" | awk '{print $3}')


get_webserver_for_user(){
	  output=$(opencli webserver-get_webserver_for_user $USERNAME)		
		ws=$(echo "$output" | grep -Eo 'nginx|openresty|apache|openlitespeed|litespeed' | head -n1)
}


if [ "$USE_WS" -eq 1 ]; then
    # webserver
    FINAL_PATH="/home/${CONTEXT}/docker-data/volumes/${CONTEXT}_webserver_data/_data/${DOMAIN}.conf"
    if [ -f "$FINAL_PATH" ]; then
        echo "Opening VirtualHosts file in nano: $FINAL_PATH"
        nano "$FINAL_PATH"
        echo "Restarting $ws webserver to save changes.."
        get_webserver_for_user
        docker --context=$CONTEXT restart $ws
    else
        echo "ERROR: VirtualHosts for domain $DOMAIN does not exist: $FINAL_PATH"
        exit 1
    fi
else
    # docroot
    DOCROOT_PATH=${DOCROOT#/var/www/html/}
    FINAL_PATH="/home/${CONTEXT}/docker-data/volumes/${CONTEXT}_html_data/_data/${DOCROOT_PATH}"
    if [ -d "$FINAL_PATH" ]; then
        echo "Directory: $FINAL_PATH"
    else
        echo "ERROR: Directory does not exist: $FINAL_PATH"
        exit 1
    fi
fi
