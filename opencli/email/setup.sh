#!/bin/bash
################################################################################
# Script Name: email/setup.sh
# Description: Setup email addresses, forwarders, filters..
# Usage: opencli email-setup <COMMAND> <ATTRIBUTES>
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 18.08.2024
# Last Modified: 24.06.2026
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
    echo "Docs: https://openpanel.com/docs/articles/opencli/email"
    exit 1
fi

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

    if [ -z "$key_value" ]; then
        echo "Error: OpenPanel Community edition does not support emails. Please consider purchasing the Enterprise version that allows unlimited number of email addresses."
        local ENTERPRISE="/usr/local/opencli/enterprise.sh"
        # shellcheck disable=SC1090
        . "$ENTERPRISE"
        echo "$ENTERPRISE_LINK"
        exit 1
    fi
}

get_openpanel_username_and_uid_for_domain() {
    local domain="${1#*@}"
    local whoowns_output owner
    whoowns_output=$(opencli domains-whoowns "$domain" --context)
    read -r _ owner <<< "$whoowns_output"
    [[ -n "$owner" ]] && OP_UID=$(stat -c '%u' "/home/$owner")
}

reload_emails_data_file_for_user() {
    [[ -z "$owner" ]] && return
    local file_to_refresh="/etc/openpanel/openpanel/core/users/$owner/emails.yml"
    local all_domains all_emails
    all_domains=$(opencli domains-user "$owner")
    sleep 2 # for https://github.com/docker-mailserver/docker-mailserver/blob/cb76075f25e22476e8fdb45adfbea8026d4ea898/target/bin/addmailuser#L16
    all_emails=$(opencli email-setup email list)
    > "$file_to_refresh"
    while IFS= read -r domain; do
        grep "@${domain}" <<< "$all_emails" >> "$file_to_refresh"
    done <<< "$all_domains"
}

# ======================================================================
# Run setup command
validate_first
command="$@" 
# https://docker-mailserver.github.io/docker-mailserver/latest/config/setup.sh/
docker exec openadmin_mailserver setup $command  

if [[ "$1" == "email" && "$2" =~ ^(add|update|del)$ ]] || [[ "$1" == "quota" && "$2" =~ ^(set|del)$ ]]; then
    if is_valid_email "$3"; then
        # get OpenPanel user UID and store/update it in postfix-accounts.cf
        get_openpanel_username_and_uid_for_domain "$3"
        if [[ "$2" =~ ^(add|update)$ && -n "$OP_UID" && "$OP_UID" =~ ^[0-9]+$ ]]; then
            sed -i "/^$3|/ { s/^\([^|]*|[^|]*\).*/\1|$OP_UID/}" "/usr/local/mail/openmail/docker-data/dms/config/postfix-accounts.cf"
            # TODO: after 2.0 edit to only run on 'add' and not on 'update'!
            nohup timeout 300 docker --context=default exec openadmin_mailserver bash -c "chown -R \"${UID_OVERRIDE}:${UID_OVERRIDE}\" \"/var/mail/${DOMAIN}/${USER}\"" &
        fi

        # if email add/del OR quita set/del then we need to reload the cached user file for OpenPanel UI to display to user
        if [[ "$2" != "update" && "$5" != "--wait" ]]; then
            reload_emails_data_file_for_user
        fi
    fi
fi
