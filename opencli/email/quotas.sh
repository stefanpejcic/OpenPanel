#!/bin/bash
################################################################################
# Script Name: quotas.sh
# Description: Fixes email permission issues for all domains.
# Usage: opencli email-quotas
# Author: Stefan Pejcic
# Created: 03.12.2025
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
##################################################################################

readonly ENTERPRISE="/usr/local/opencli/enterprise.sh"
readonly PANEL_CONFIG_FILE="/etc/openpanel/openpanel/conf/openpanel.config"
key_value=$(grep "^key=" $PANEL_CONFIG_FILE | cut -d'=' -f2-)

if [ -n "$key_value" ]; then
    :
else
    echo "Error: OpenPanel Community edition does not support emails. Please consider purchasing the Enterprise version that allows unlimited number of email addresses."
    # shellcheck source=/usr/local/opencli/enterprise.sh
    source "$ENTERPRISE"
    echo "$ENTERPRISE_LINK"
    exit 1
fi

readonly CONFIG_FILE="/etc/openpanel/openadmin/config/admin.ini"
readonly LOG_FILE="/var/log/openpanel/admin/email-quotas.log"

log() {
    local message="$1"
    echo "$(date '+%Y-%m-%d %H:%M:%S') : $message" | tee -a "$LOG_FILE"
}


STORE_EMAILS_IN=$(grep -E '^email_storage_location=' /etc/openpanel/openadmin/config/admin.ini | cut -d'=' -f2- | xargs)

if [[ "$STORE_EMAILS_IN" == /* ]]; then
  FOLDER="$STORE_EMAILS_IN"
else
  FOLDER=/home/*/mail/
fi

FOLDERS=($FOLDER/*)
TOTAL=${#FOLDERS[@]}

log "Total mail domains in "$FOLDER": [$TOTAL]"

COUNT=0

for FOLDER_NAME in "$FOLDER"/*; do
  ((COUNT++))
  DOMAIN=$(basename "$(dirname "$FOLDER_NAME")")
  log "[*] DOMAIN: $DOMAIN [$COUNT/$TOTAL]"
  log "Checking username and docker context for domain.."
  whoowns_output=$(opencli domains-whoowns "$DOMAIN" --context)
  owner=$(log "$whoowns_output" | awk -F "Owner of '$DOMAIN': " '{print $2}')
  if [ -z "$owner" ]; then
    log "$FOLDER_NAME skipped - No owner detected for domain: $DOMAIN"
  else
    USERNAME=$(log "$owner" | sed -n '1p')
    CONTEXT=$(log "$owner" | sed -n '2p')
    log "USERNAME: $USERNAME - CONTEXT: $CONTEXT"
    log "Setting permissions to: $CONTEXT:$CONTEXT for all mails in: $FOLDER_NAME"
    chown -R "$CONTEXT:$CONTEXT" "$FOLDER_NAME"
  fi
done

log "Finished processing a total of $COUNT domains."
