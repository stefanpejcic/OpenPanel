#!/bin/bash
################################################################################
# Script Name: user/varnish.sh
# Description: Enable/disable Varnish Caching for user and display current status.
# Usage: opencli user-varnish <USERNAME> [on|off]
# Author: Stefan Pejcic
# Created: 20.03.2025
# Last Modified: 22.04.2026
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

USER=$1
ACTION=$2

[ -z "$USER" ] && { echo "Usage: opencli user-varnish <user> [enable|disable]"; exit 1; }

get_docker_context(){
	IFS=',' read -r user_id CONTEXT <<< "$(mysql -se "SELECT CONCAT(id, ',', server) FROM users WHERE username='${user}';")"
	if [ -z "$user_id" ] || [ -z "$CONTEXT" ]; then
	    echo "FATAL ERROR: Missing user ID or context for user $user."
	    exit 1
	fi
}

ENV_FILE="/home/$CONTEXT/.env"
[ -f "$ENV_FILE" ] || { echo "Error: $ENV_FILE not found!"; exit 1; }

check_status() {
  if grep -q "^#PROXY_HTTP_PORT=" "$ENV_FILE"; then
      STATUS="Off"
  elif grep -q "^PROXY_HTTP_PORT=" "$ENV_FILE"; then
      STATUS="On"
  else
      STATUS="Unknown"
  fi

  echo "Current status: $STATUS"
}

toggle_service() { cd /home/$CONTEXT && docker --context="$CONTEXT" compose "$@" 2>/dev/null; }

check_varnish_status() {
  local action="$1"
  local running=1

  if cd "/home/$CONTEXT" && docker --context="$CONTEXT" compose ps -q varnish | grep -q .; then
    running=0
  fi

  case "$action" in
    start) [[ $running -eq 0 ]] && echo "Varnish Cache is now enabled." || echo "Failed to enable Varnish Cache." ;;
    stop) [[ $running -ne 0 ]] && echo "Varnish Cache is now disabled." || echo "Failed to disable Varnish Cache." ;;
    status) [[ $running -eq 0 ]] && echo "Varnish Cache is enabled." || echo "Varnish Cache is disabled." ;;
    *) echo "Usage: status_varnish {start|stop|status}"; return 1 ;;
  esac
}

# Main
if [ -n "$ACTION" ]; then
    WEB_SERVER=$(grep "^WEB_SERVER=" "$ENV_FILE" | awk -F'=' '{print $2}' | tr -d '[:space:]' | sed 's/^"\(.*\)"$/\1/' | grep -Eo 'nginx|openresty|apache|openlitespeed|litespeed' | head -n1);
    case "$ACTION" in
        enable)
            toggle_service down "$WEB_SERVER"
            sed -i 's/^#PROXY_HTTP_PORT=/PROXY_HTTP_PORT=/' "$ENV_FILE"
            toggle_service up -d "$WEB_SERVER" varnish
            check_varnish_status start
            ;;
        disable)
            toggle_service down "$WEB_SERVER"
            sed -i 's/^PROXY_HTTP_PORT=/#PROXY_HTTP_PORT=/' "$ENV_FILE"
            toggle_service down varnish
            toggle_service up -d "$WEB_SERVER"
			check_varnish_status stop
            ;;
        status) check_varnish_status status ;;			
        *) echo "Invalid action. Use 'status', 'enable' or 'disable'."; exit 1 ;;
    esac
else
  check_status
fi
