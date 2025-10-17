#!/bin/bash
################################################################################
# Script Name: user/varnish.sh
# Description: Enable/disable Varnish Caching for user and display current status.
# Usage: opencli user-varnish <USERNAME> [on|off]
# Author: Stefan Pejcic
# Created: 20.03.2025
# Last Modified: 16.10.2025
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
# todo: check context!

ENV_FILE="/home/$USER/.env"

if [ -z "$USER" ]; then
    echo "Usage: opencli user-varnish <user> [enable|disable]"
    exit 1
fi

if [ ! -f "$ENV_FILE" ]; then
    echo "Error: $ENV_FILE not found!"
    exit 1
fi


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


get_webserver_for_user(){
	    log "Checking webserver configuration"
		output=$(opencli webserver-get_webserver_for_user "$USER")
		ws=$(echo "$output" | grep -Eo 'nginx|openresty|apache|openlitespeed|litespeed' | head -n1)
}


stop_webserver(){
  cd /home/$USER && docker --context $USER compose down $ws > /dev/null;
}

stop_varnish(){
  cd /home/$USER && docker --context $USER compose down varnish #> /dev/null;
}

start_varnish(){
  cd /home/$USER && docker --context $USER compose up -d varnish #> /dev/null;
}

start_webserver(){
  cd /home/$USER && docker --context $USER compose up -d $ws > /dev/null;
}

check_varnish() {
  local action="$1"
  local running=1

  if cd "/home/$USER" && docker --context "$USER" compose ps -q varnish | grep -q .; then
    running=0
  fi

  case "$action" in
    start)
      [[ $running -eq 0 ]] && echo "Varnish Cache is now enabled." || echo "Failed to enable Varnish Cache."
      ;;
    stop)
      [[ $running -ne 0 ]] && echo "Varnish Cache is now disabled." || echo "Failed to disable Varnish Cache."
      ;;
    status)
      [[ $running -eq 0 ]] && echo "Varnish Cache is enabled." || echo "Varnish Cache is disabled."
      ;;
    *)
      echo "Usage: status_varnish {start|stop|status}"
      return 1
      ;;
  esac
}



if [ -n "$ACTION" ]; then
    get_webserver_for_user
    case "$ACTION" in
        enable)
            stop_webserver
            sed -i 's/^#PROXY_HTTP_PORT=/PROXY_HTTP_PORT=/' "$ENV_FILE"
            start_webserver
            start_varnish
            check_varnish start
            ;;
        disable)
            stop_webserver
            sed -i 's/^PROXY_HTTP_PORT=/#PROXY_HTTP_PORT=/' "$ENV_FILE"
            stop_varnish
            start_webserver
			check_varnish stop
            ;;
        status)
			check_varnish status
            ;;			
        *)
            echo "Invalid action. Use 'status', 'enable' or 'disable'."
            exit 1
            ;;
    esac
else
  check_status
fi
