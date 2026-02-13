#!/bin/bash
################################################################################
# Script Name: php/domain.sh
# Description: View or change the PHP version used for a single domain name.
# Usage: opencli php-domain <domain_name>
#        opencli php-domain <domain_name> --update <new_php_version>
# Author: Stefan Pejcic
# Created: 07.10.2023
# Last Modified: 12.02.2026
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
# Variables
domain="$1"
update_flag=false
new_php_version=""


# ======================================================================
# Validators

if [ $# -lt 1 ]; then
    echo "Usage: opencli php-domain <domain> [--update <new_php_version>]"
    exit 1
fi

if [ "$2" == "--update" ]; then
    if [ -z "$3" ]; then
        echo "Error: --update flag requires a new PHP version in the format number.number."
        exit 1
    fi
    if [[ ! "$3" =~ ^[0-9]\.[0-9]$ ]]; then
        echo "Invalid PHP version format. Please use the format 'number.number' (e.g., 8.1 or 5.6)."
        exit 1
    fi

    update_flag=true
    new_php_version="$3"
fi


# ======================================================================
# Helpers
get_context_for_user() {
	# shellcheck source=/usr/local/opencli/db.sh
    source /usr/local/opencli/db.sh
	username_query="SELECT server FROM users WHERE username = '$owner'"
	context=$(mysql -D "$mysql_database" -e "$username_query" -sN)
	if [ -z "$context" ]; then
		context=$owner
	fi
}

get_webserver_for_user(){
	    output=$(opencli webserver-get_webserver_for_user "$owner")		
		ws=$(echo "$output" | grep -Eo 'nginx|openresty|apache|openlitespeed|litespeed' | head -n1)
        if [[ "$ws" == "openlitespeed" || "$ws" == "litespeed" ]]; then
            echo "ERROR: PHP version can not be changed on $ws webserver. Instead you need to change the docker image tag for the user."
            echo "Available tags: https://hub.docker.com/r/litespeedtech/openlitespeed/tags"
            exit 0
        fi
}

# ======================================================================
# Main
whoowns_output=$(opencli domains-whoowns "$domain")
owner=$(echo "$whoowns_output" | awk -F "Owner of '$domain': " '{print $2}')
if [ -n "$owner" ]; then
    get_webserver_for_user
    get_context_for_user
    domain_path_in_volume="/home/$context/docker-data/volumes/${context}_webserver_data/_data/$domain.conf"
    php_version=$(grep -o "php-fpm-[0-9.]\+" "$domain_path_in_volume" | grep -o "[0-9.]\+" | head -n 1)

    if [ -n "$php_version" ]; then
        if [ "$update_flag" == true ]; then
            if [ -n "$new_php_version" ]; then
                # 1. domain vhost file
                sed -i "s/php-fpm-[0-9.]\+/php-fpm-$new_php_version/g" "$domain_path_in_volume"

                # 2. start new php version
                nohup sh -c "opencli user-resources $owner --activate=php-fpm-${new_php_version}" </dev/null >nohup.out 2>nohup.err &

                # 3. restart webservers
                nohup sh -c "docker --context $context restart nginx" </dev/null >nohup.out 2>nohup.err &
                nohup sh -c "docker --context $context restart openresty" </dev/null >nohup.out 2>nohup.err &
                nohup sh -c "docker --context $context restart apache" </dev/null >nohup.out 2>nohup.err &

                # 4. save in db
                update_query="UPDATE domains SET php_version='$new_php_version' WHERE domain_url='$domain';"
                mysql -e "$update_query"

                echo "Updated PHP version in the configuration file to $new_php_version"
            else
                echo "Error: new php verison not provied!"
                exit 1
            fi
        else
            echo "Domain '$domain' (owned by user: $owner) uses PHP version: $php_version"
        fi
    else
        echo "Failed to determine the PHP version for the domain '$domain' (owned by user $owner)." >&2
        exit 1
    fi
else
    echo "Failed to determine the owner of the domain '$domain'." >&2
    exit 1
fi
