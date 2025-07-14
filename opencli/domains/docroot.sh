#!/bin/bash
################################################################################
# Script Name: domains/docroot.sh
# Description: View and change docroot for a domain.
# Usage: opencli domains-docroot <DOMAIN_NAME> [update </var/www/html/>] --debug
# Author: Stefan Pejcic
# Created: 10.02.2025
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

debug_mode=false
args=()

# Process arguments and check for --debug
for arg in "$@"; do
    if [[ "$arg" == "--debug" ]]; then
        debug_mode=true
    else
        args+=("$arg")  # Store non-debug arguments
    fi
done

show_help() {
    echo "Usage: opencli domains-docroot <domain> [update <docroot>] [--debug]"
    exit 1
}


# Ensure at least one non-debug argument is provided
if [[ ${#args[@]} -lt 1 ]]; then
    show_help
    exit 1
fi

domain="${args[0]}"
action="${args[1]:-}"
new_docroot="${args[2]:-}"

log() {
    if $debug_mode; then
        echo "$1"
    fi
}


log "Debug Mode: ON"
log "Domain: $domain"
log "Action: ${action:-None}"
log "New Docroot: ${new_docroot:-None}"


################################## helpers

# get user ID from the database
get_user_info() {
    local user="$1"
    local query="SELECT id, server FROM users WHERE username = '${user}';"
    
    # Retrieve both id and context
    user_info=$(mysql -se "$query")
    
    # Extract user_id and context from the result
    user_id=$(echo "$user_info" | awk '{print $1}')
    context=$(echo "$user_info" | awk '{print $2}')
    
    echo "$user_id,$context"
}


get_user_context() {

  result=$(get_user_info "$user")
  user_id=$(echo "$result" | cut -d',' -f1)
  context=$(echo "$result" | cut -d',' -f2)

  if [ -z "$user_id" ]; then
      echo "FATAL ERROR: user $user does not exist."
      exit 1
  fi
}

validate_docroot() {

  log "Updating document root for domain: $domain to: $new_docroot"
  if [[ -n "$docroot" && ! "$docroot" =~ ^/var/www/html/ ]]; then
      echo "FATAL ERROR: Invalid docroot. It must start with /var/www/html/"
      exit 1
  fi

}


make_folder() {
  # Extract the folder name from new_docroot
  folder_name="${new_docroot##/var/www/html/}"

  # Build the new path
  newnew_docroot="/home/${context}/docker-data/volumes/${context}_html_data/_data/${folder_name}"

  log "Creating document root directory $new_docroot"
  mkdir -p "$newnew_docroot"
  chown "$user:$user" "$newnew_docroot"
}



get_webserver_for_user(){
	    log "Checking webserver configuration"
	    ws=$(opencli webserver-get_webserver_for_user $user)
}


vhost_file_edit() {
	vhost_file=/home/${context}/docker-data/volumes/${context}_webserver_data/_data/${domain}.conf
	sed -i -E 's|(/var/www/html/[^>;]*)|'"$new_docroot"'|g' $vhost_file > /dev/null 2>&1
	docker --context $context restart $ws > /dev/null 2>&1
}


get_user() {
  whoowns_output=$(opencli domains-whoowns "$domain")
  user=$(echo "$whoowns_output" | awk -F "Owner of '$domain': " '{print $2}')
  if [ -n "$user" ]; then
    log "User $user owns domain $domain"
  else
      echo "Failed to determine the owner of the domain '$domain'." >&2
      exit 1
  fi
  
}

main_func() {

  validate_docroot
  get_user
  get_user_context
  

  
  mysql -e "UPDATE domains SET docroot='$new_docroot' WHERE domain_url='$domain';"
  mysql -e "$insert_query"
  result=$(mysql -se "$query")
  local verify_query="SELECT COUNT(*) FROM domains WHERE docroot = '$new_docroot' AND domain_url = '$domain';"
  local result=$(mysql -N -e "$verify_query")
  
  if [ "$result" -eq 1 ]; then
  
    make_folder
    get_webserver_for_user
    vhost_file_edit
    
    echo "Docroot updated to: $new_docroot for domain: $domain"  
  else
      log "Updating docroot for domain $domain failed! Contact administrator to check if the mysql database is running."
      echo "Failed to change docroot for domain $domain"
  fi
  
}


get_current_docroot(){
  local get_docroot="SELECT docroot FROM domains WHERE domain_url = '$domain';"
  result=$(mysql -N -B -e "$get_docroot")
if [[ -z "$result" ]]; then
  echo "Docroot not found for domain: $domain - does the domain exist? run: opencli domains-whoowns $domain"
else
  echo "$result"
fi
}


if [[ -n "$domain" && -z "$action" ]]; then
    get_current_docroot
elif [[ -n "$domain" && "$action" == "update" ]]; then
    if [[ -z "$new_docroot" ]]; then
        echo "Error: Missing new_docroot for update action."
        show_help
        exit 1
    fi
   main_func # this does all other functions!
else
    show_help
    exit 1
fi

