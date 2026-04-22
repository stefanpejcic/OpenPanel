#!/bin/bash
################################################################################
# Script Name: port.sh
# Description: View and change port for accessing openpanel.
# Usage: opencli port [set <port>] 
# Author: Stefan Pejcic
# Created: 17.02.2025
# Last Modified: 21.04.2026
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

ENV_FILE="/root/.env"
COMPOSE_DIR="/root"
REDIRECTS_FILE="/etc/openpanel/caddy/redirects.conf"
PROXY_FILE="/etc/openpanel/nginx/vhosts/openpanel_proxy.conf"



get_current_port() {
  if [ -f /.dockerenv ] && [ -f /etc/openpanel/no_port ]; then
      current_port=443
  fi
  if [ -z "$current_port" ]; then
    current_port="$(grep -oP '(?<=^PORT=").*(?=")' $ENV_FILE | head -n 1)"
  fi
  echo "$current_port"
}



usage() {

  echo "Usage:"
  echo ""
  echo "opencli port                        - displays current port  "
  echo "opencli port set 2090               - set 2090 as port for user panel"
  echo "opencli port default                - set 2083 as port for user panel"
}



success_msg() {
  echo "$new_port is now set for accessing the OpenPanel interface."             
}

# for docker compose
update_env() {
  sed -i "s/^PORT=\"[^\"]*\"/PORT=\"$new_port\"/" "$ENV_FILE"
}

# for redirects!
do_reload() {
  if [[ "$3" != '--no-restart' ]]; then
    # restart caddy and openpanel
    cd $COMPOSE_DIR
    nohup docker compose restart caddy > /dev/null 2>&1 < /dev/null &
    nohup bash -c "docker --context=default compose down openpanel && docker --context=default compose up -d openpanel" > /dev/null 2>&1 &

    # open port on csf
    local csf_conf="/etc/csf/csf.conf"
    [ -f "$csf_conf" ] || return 0
    port_opened=$(grep "TCP_IN = .*${new_port}" "$csf_conf")
    if [ -z "$port_opened" ]; then
        sed -i "s/TCP_IN = \"\(.*\)\"/TCP_IN = \"\1,${new_port}\"/" "$csf_conf"
        nohup csf -r > /dev/null 2>&1 < /dev/null &
    fi
   fi
}


update_redirects() {
  sed -i "/redir @openpanel/s/:[0-9]\+/:$new_port/g" $REDIRECTS_FILE
}

update_proxy_file() {
  sed -i "/# openpanel/,/# openadmin/ s/:[0-9]\+/:$new_port/g" "$PROXY_FILE" 
}



update_port() {
  current_port=$(get_current_port)
  if [ "$current_port" != "$new_port" ]; then
    update_env
    update_redirects
    update_proxy_file
    if [ "$new_port" == '443' ]; then
        sed -i "s#\${PORT}:2083/tcp#2083:2083/tcp#g" /root/docker-compose.yml
    else
        sed -i "s#2083:2083/tcp#\${PORT}:2083/tcp#g" /root/docker-compose.yml
    fi
    do_reload
    success_msg
  else
    echo "Port ${new_port} s already set for accessing the user panel."
    exit 0
  fi
}




if [ -z "$1" ]; then
    get_current_port

elif [[ "$1" == 'set' && -n "$2" ]]; then
    if [[ "$2" =~ ^[0-9]+$ ]] && [ "$2" -ge 443 ]; then
        new_port=$2
        update_port $new_port
    else
        echo "Invalid port format. Please provide a valid port like '2083'."
        usage
    fi    
elif [[ "$1" == 'default' ]]; then
        new_port="2083"
        update_port $new_port
else
    usage
fi


