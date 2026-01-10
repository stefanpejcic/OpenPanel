#!/bin/bash
################################################################################
# Script Name: files/calculate_resellers_storage.sh
# Description: Calculates total disk usage for all resellers.
# Usage: opencli files-calculate_resellers_storage
# Author: Stefan Pejcic
# Created: 24.09.2025
# Last Modified: 09.01.2026
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

readonly REPQUOTA_PATH="/etc/openpanel/openpanel/core/users/repquota"
readonly DB_CONFIG="/usr/local/opencli/db.sh"

# Source database configuration
if [[ ! -f "$DB_CONFIG" ]]; then
    echo "Error: Database configuration file not found: $DB_CONFIG" >&2
    exit 1
fi

# shellcheck source=/usr/local/opencli/db.sh
source "$DB_CONFIG"

# Global variables
declare -g mysql_database config_file


validate_db_config() {
    if [[ -z "${mysql_database:-}" ]]; then
        log_error "Database name not configured"
        return 1
    fi
    
    if [[ -z "${config_file:-}" ]]; then
        log_error "Database config file not specified"
        return 1
    fi
    
    if [[ ! -f "$config_file" ]]; then
        log_error "Database config file not found: $config_file"
        return 1
    fi
    
    return 0
}



process_resellers() {
	resellers=$(opencli admin list | awk -F'|' '$2=="reseller" && $3=="1" {print $1}')
	resellers_count=$(echo "$resellers" | wc -w)	 
	echo "Total resellers: $resellers_count"
	echo
	
	for reseller in $resellers; do
  
	    users=$(mysql --defaults-extra-file="$config_file" -N -D "$mysql_database" \
		-e "SELECT username FROM users WHERE owner='$reseller';")
	    
	    if [ -z "$users" ]; then
	        echo "$reseller has no user accounts - skipping.."
		continue
	    else
	    	count=$(echo "$users" | wc -w)	    
	        echo "$reseller has ($count) accounts - total disk blocks usage:"
	    fi

	    total_blocks_used=0

	    reseller_limits_file="/etc/openpanel/openadmin/resellers/$reseller.json"
	    if [[ -f "$reseller_limits_file" ]]; then
	    	total_blocks_hard=$(jq -r '.max_disk_blocks // 0' "$reseller_limits_file")
	    else
	    	echo "Warning: Reseller config file not found: $reseller_limits_file"
	    	total_blocks_hard=0
	    fi


	    # Collect usage from repquota
	    for user in $users; do
      		read blocks_used < <(
      		    awk -v u="$user" '$1==u {print $3}' "$REPQUOTA_PATH"
      		)
		
      		if [ -n "$blocks_used" ]; then
      		    total_blocks_used=$((total_blocks_used + blocks_used))
      		fi
	    done
	    
	    total_blocks_used=${total_blocks_used:-0}
	    total_blocks_hard=${total_blocks_hard:-0}

  		jq --argjson current_disk_blocks "$total_blocks_used" \
  		   --argjson max_disk_blocks "$total_blocks_hard" \
  		   '.current_disk_blocks = $current_disk_blocks | .max_disk_blocks = $max_disk_blocks' \
  		   "$reseller_limits_file" > "/tmp/${reseller}_config.json" \
  		   && mv "/tmp/${reseller}_config.json" "$reseller_limits_file"
	    
	    echo "${total_blocks_used} / ${total_blocks_hard} (blocks)"
	    echo
	    echo "------------------------------------------------"


	done
}


main() {
    if ! validate_db_config; then
        exit 1
    fi
    process_resellers
    exit 0
}


main
