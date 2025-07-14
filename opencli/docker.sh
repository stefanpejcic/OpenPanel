#!/bin/bash
################################################################################
# Script Name: docker.sh
# Description: Manage OpenPanel system or user containers with lazydocker.
# Usage: opencli docker [username]
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 07.03.2025
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

if [ -z "$1" ]; then
  socket="/var/run/docker.sock"
else
  # If $1 is provided, do A
  USERNAME=$1
  CONTEXT=$(mysql -e "SELECT server FROM users WHERE username = '$USERNAME';" -sN) 

  if [ -z "$CONTEXT" ]; then
    echo "ERROR: No context found for username: $USERNAME"
    exit 1
  fi
  
  USER_UID=$(id -u "$CONTEXT" 2>/dev/null)
  if ! [[ "$USER_UID" =~ ^[0-9]+$ ]]; then
    echo "ERROR: Invalid USER_UID: $USER_UID"
    exit 1
  fi
  
  socket="/hostfs/run/user/${USER_UID}/docker.sock"
fi

    exec docker run --rm -it \
        --cpus="0.1" \
        --memory="100m" \
        --pids-limit="10" \
        --security-opt no-new-privileges \
        -v ${socket}:/var/run/docker.sock \
        lazyteam/lazydocker

exit 0
