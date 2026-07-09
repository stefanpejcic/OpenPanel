#!/bin/bash
################################################################################
# Script Name: ftp/add.sh
# Description: List FTP sub-users for openpanel user.
# Usage: opencli ftp-list <OPENPANEL_USERNAME> [--json]
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 10.09.2024
# Last Modified: 08.07.2026
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

# Parse args
JSON_OUTPUT=0
ARGS=()
for arg in "$@"; do
    case "$arg" in
        --json)
            JSON_OUTPUT=1
            ;;
        *)
            ARGS+=("$arg")
            ;;
    esac
done
set -- "${ARGS[@]}"

json_escape() {
    local s="$1"
    s="${s//\\/\\\\}"
    s="${s//\"/\\\"}"
    printf '%s' "$s"
}

if [ -z "$1" ]; then
    # ALL USERS
    if [ "$JSON_OUTPUT" -eq 1 ]; then
        json_out="["
        first=1
        for context in /etc/openpanel/ftp/users/*; do
            if [ -d "$context" ]; then
                user_dir_name=$(basename "$context")
                users_file="${context}/users.list"
                if [ ! -f "$users_file" ]; then
                    continue
                fi
                while IFS='|' read -r username password directory uid gid; do
                    [ -z "$username" ] && continue
                    if [ "$first" -eq 0 ]; then
                        json_out+=","
                    fi
                    first=0
                    json_out+="{\"openpanel_user\":\"$(json_escape "$user_dir_name")\",\"username\":\"$(json_escape "$username")\",\"password\":\"$(json_escape "$password")\",\"directory\":\"$(json_escape "$directory")\",\"uid\":\"$(json_escape "$uid")\",\"gid\":\"$(json_escape "$gid")\"}"
                done < "$users_file"
            fi
        done
        json_out+="]"
        echo "$json_out"
    else
        for context in /etc/openpanel/ftp/users/*; do
            if [ -d "$context" ]; then
                user_dir_name=$(basename "$context")
                users_file="${context}/users.list"
                if [ ! -f "$users_file" ]; then
                    continue
                fi
                echo "FTP sub-users for '$user_dir_name':"
                while IFS='|' read -r username password directory uid gid; do
                    echo "$username | $directory"
                done < "$users_file"
                echo
            fi
        done
    fi
else
    # SINGLE USER
    context="$1"
    user_dir_name="$context"
    users_file="/etc/openpanel/ftp/users/${context}/users.list"
    if [ ! -f "$users_file" ]; then
        if [ "$JSON_OUTPUT" -eq 1 ]; then
            echo "[]"
        else
            echo "No FTP sub-users for OpenPanel user: '$user_dir_name'."
        fi
        exit 1
    fi

    if [ "$JSON_OUTPUT" -eq 1 ]; then
        json_out="["
        first=1
        while IFS='|' read -r username password directory uid gid; do
            [ -z "$username" ] && continue
            if [ "$first" -eq 0 ]; then
                json_out+=","
            fi
            first=0
            json_out+="{\"openpanel_user\":\"$(json_escape "$user_dir_name")\",\"username\":\"$(json_escape "$username")\",\"password\":\"$(json_escape "$password")\",\"directory\":\"$(json_escape "$directory")\",\"uid\":\"$(json_escape "$uid")\",\"gid\":\"$(json_escape "$gid")\"}"
        done < "$users_file"
        json_out+="]"
        echo "$json_out"
    else
        echo "FTP sub-users for '$context':"
        while IFS='|' read -r username password directory uid gid; do
            echo "$username | $directory"
        done < "$users_file"
    fi
fi
