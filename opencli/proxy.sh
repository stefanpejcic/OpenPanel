#!/bin/bash
################################################################################
# Script Name: proxy.sh
# Description: View and change proxy path '/openpanel' for accessing openpanel.
# Usage: opencli port [set <path>] 
# Author: Stefan Pejcic
# Created: 17.02.2025
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


COMPOSE_DIR="/root"
REDIRECTS_FILE="/etc/openpanel/caddy/redirects.conf"

get_current_path() {
  current_path=$(awk '/^@openpanel path/ {match($0, /\/([^*]*)\*/ , arr); print arr[1]}' "$REDIRECTS_FILE")
  echo "$current_path"
}


usage() {

echo "Usage:"
echo ""
echo "opencli proxy                        - displays current path"
echo "opencli proxy set /custompath        - set /custompath for access"
echo "opencli proxy default                - set /openpanel for access"
}



success_msg() {
  echo "/$new_path is now set for accessing the OpenPanel interface."             
}

# for docker compose
update_redirects() {
  sed -i 's|\(@openpanel path /\)[^*]*\*|\1'"$new_path"'*|' "$REDIRECTS_FILE"
}

# for redirects!
do_reload() {
  if [[ "$3" != '--no-restart' ]]; then
    cd $COMPOSE_DIR
    nohup docker compose restart openpanel > /dev/null 2>&1 &
   fi
}



update_path() {
  current_path=$(get_current_path)
      if [ "$current_path" != "$new_path" ]; then
        update_redirects
        do_reload
        success_msg
      else
        echo "Path /${new_path} is already set for accessing the user panel."
        exit 0
      fi

}




if [ -z "$1" ]; then
    get_current_path
elif [[ "$1" == 'set' && -n "$2" ]]; then
    if [[ "$2" =~ ^(/?[a-zA-Z0-9_-]+)$ ]]; then
        new_path="${2#/}"
        update_path $new_path
    else
        echo "Invalid path format. Please provide a path like '/openpanel' or '/hosting'."
        usage
    fi    
elif [[ "$1" == 'default' ]]; then
        new_path="openpanel"
        update_path $new_path
else
    usage
fi
