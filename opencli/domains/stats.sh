#!/bin/bash
################################################################################
# Script Name: domains/stats.sh
# Description: Parse caddy access logs for users domains and generate static html
# Usage: opencli domains-stats
#        opencli domains-stats --debug
#        opencli domains-stats <USERNAME>
#        opencli domains-stats <USERNAME> --debug
# Author: Radovan Jecmenica
# Created: 14.12.2023
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

DEBUG=false # Default value for DEBUG
SINGLE_USER=false
OPENPANEL_CONF_DIR="/etc/openpanel/goaccess"

# Parse optional flags to enable debug mode when needed!
for arg in "$@"; do
    case $arg in
        --debug)
            DEBUG=true
            ;;
        *)
            SINGLE_USER=true
            username="$arg"
            ;;
    esac
done



check_if_reports_are_enabled() {
    enabled_modules=$(grep '^enabled_modules=' "/etc/openpanel/openpanel/conf/openpanel.config" | cut -d'=' -f2)
    # Check if 'domains_visitors' is in the list of enabled modules
    if echo "$enabled_modules" | grep -q 'goaccess'; then
        # 'goaccess' is enabled
        :
    else
        # 'domains_visitors' is not enabled
        echo "'goaccess' module is not enabled. Skipping report generation."
        exit 1
    fi
}



configure_goaccess() {
    # GoAccess
    tar -xzvf "${OPENPANEL_CONF_DIR}/GeoLite2-City_20231219.tar.gz" -C "${OPENPANEL_CONF_DIR}/" > /dev/null
    mkdir -p /usr/local/share/GeoIP/GeoLite2-City_20231219
    cp -r "${OPENPANEL_CONF_DIR}/GeoLite2-City_20231219/"* /usr/local/share/GeoIP/GeoLite2-City_20231219 
}


process_logs() {
    local username="$1"
    local excluded_ips_file="/etc/openpanel/openpanel/core/users/$username/domains/excluded_ips_for_goaccess"
    #local container_ip=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $username)
    local excluded_ips=""

    if [ -f "$excluded_ips_file" ] && [ -s "$excluded_ips_file" ]; then
        excluded_ips=$(<"$excluded_ips_file")
    fi

    local domains=$(opencli domains-user "$username")

    if [[ "$domains" == *"No domains found for user '$username'"* ]]; then
        echo "No domains found for user $username. Skipping."
    else

        for domain in $domains; do
            local log_file="/var/log/caddy/domlogs/${domain}/access.log"
            local output_dir="/var/log/caddy/stats/${username}/"
            local html_output="${output_dir}/${domain}.html"
            local sed_command="s/Dashboard/$domain/g"
    
            mkdir -p "$output_dir"
            if [ -s "$log_file" ]; then
                docker run --memory="256m" --cpus="0.5" \
                   -v /usr/local/share/GeoIP/GeoLite2-City_20231219/GeoLite2-City.mmdb:/GeoLite2-City.mmdb \
                   -v ${log_file}:/var/log/caddy/access.log \
                   -v ${output_dir}:${output_dir}\
                   --rm -i -e LANG=EN allinurl/goaccess \
                   -e "$excluded_ips" --ignore-panel=KEYPHRASES -a -o html \
                   --log-format=CADDY /var/log/caddy/access.log \
                   > "$html_output"
              
                
                sed -i "$sed_command" "$html_output" > /dev/null 2>&1
        
                if [ "$DEBUG" = true ]; then
                    echo "Processed domain $domain for user $username with IP exclusions"
                else
                    echo "Processed domain $domain for user $username"
                fi
            else
                echo "Skipped empty log file for domain $domain of user $username"
            fi
        done
        
    fi
}



process_single_or_all_users() {
    if [ "$SINGLE_USER" = true ]; then
        process_logs "$username"
    else
        usernames=$(opencli user-list --json | grep -v 'SUSPENDED' | awk -F'"' '/username/ {print $4}')
        for username in $usernames; do
            process_logs "$username"
        done
    fi
}


# use flock to disable script running multiple times
(
 flock -s 200
 check_if_reports_are_enabled   # added in 0.2.8
 configure_goaccess             # setup database and data for container
 process_single_or_all_users    # actual generation
)200>/var/lock/caddy_stats.lock
