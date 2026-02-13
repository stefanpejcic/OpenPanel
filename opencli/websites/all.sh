#!/bin/bash
################################################################################
# Script Name: websites/all.sh
# Description: Lists all websites currently hosted on the server.
# Usage: opencli websites-all
# Author: Stefan Pejcic
# Created: 26.10.2023
# Last Modified: 12.02.2026
# Company: openpanel.comm
# Copyright (c) openpanel.comm
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


# DB
source /usr/local/opencli/db.sh

get_all_sites() {
    local site_type="$1"

    if [ ! -f "$config_file" ]; then
        echo "Config file $config_file not found."
        exit 1
    fi

    local query="SELECT site_name FROM sites"

    if [ -n "$site_type" ]; then
        query+=" WHERE type = '$site_type'"
    fi

    local sites
    sites=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$query" -sN)

    if [ -z "$sites" ]; then
        echo "No sites found${site_type:+ for type '$site_type'}."
    else
        echo "$sites"
    fi
}

get_all_sites "$1"
