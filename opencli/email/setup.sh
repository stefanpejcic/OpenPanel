#!/bin/bash
################################################################################
# Script Name: email/setup.sh
# Description: Setup email addresses, forwarders, filters..
# Usage: opencli email-setup <COMMAND> <ATTRIBUTES>
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 18.08.2024
# Last Modified: 10.02.2026
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


if [ "$#" -lt 1 ]; then
    echo "Usage: opencli email-setup <command> [<args>...]"
    echo "Docs: https://dev.openpanel.com/cli/email.html#Emails-1"
    exit 1
fi


# ======================================================================
# START Functions

is_valid_email() {
  local email="$1"
  local email_regex='^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
  if [[ $email =~ $email_regex ]]; then
    return 0
  else
    return 1
  fi
}


validate_first() {
    local key_value
    key_value=$(grep "^key=" -- "/etc/openpanel/openpanel/conf/openpanel.config" | cut -d'=' -f2-)

    if [ -n "$key_value" ]; then
        :
    else
        echo "Error: OpenPanel Community edition does not support emails. Please consider purchasing the Enterprise version that allows unlimited number of email addresses."
        local ENTERPRISE="/usr/local/opencli/enterprise.sh"
        # shellcheck disable=SC1090
        . "$ENTERPRISE"
        echo "$ENTERPRISE_LINK"
        exit 1
    fi
}


reload_emails_data_file_for_user() {
    email="$1"
    domain="${email#*@}"
    whoowns_output=$(opencli domains-whoowns "$domain")
    owner=$(echo "$whoowns_output" | awk -F "Owner of '$domain': " '{print $2}')
    if [ -n "$owner" ]; then
        file_to_refresh="/etc/openpanel/openpanel/core/users/$owner/emails.yml"
        ALL_DOMAINS_OWNED_BY_USER=$(opencli domains-user $owner)
        ALL_EMAILS_ON_SERVER=$(opencli email-setup email list)
        
        > "$file_to_refresh"
        for domain in $ALL_DOMAINS_OWNED_BY_USER; do
            echo "$ALL_EMAILS_ON_SERVER" | grep "@${domain}" >> "$file_to_refresh"
        done
    fi
}

# END Functions
# ======================================================================



# ======================================================================
# https://docker-mailserver.github.io/docker-mailserver/latest/config/setup.sh/

validate_first
command="$@" 
docker exec openadmin_mailserver setup $command  
if [[ "$1" == "email" && ("$2" == "add" || "$2" == "del") ]]; then
  if is_valid_email "$3"; then
    reload_emails_data_file_for_user $3
  fi
fi
