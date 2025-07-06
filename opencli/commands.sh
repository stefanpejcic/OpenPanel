#!/bin/bash
################################################################################
# Script Name: commands.sh
# Description: Lists all available OpenCLI commands.
# Usage: opencli commands
# Author: Stefan Pejcic
# Created: 15.11.2023
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

SCRIPTS_DIR="/usr/local/opencli"
ALIAS_FILE="$SCRIPTS_DIR/aliases.txt"

# delete exisitng aliases first
> "$ALIAS_FILE"

GREEN='\033[0;32m'
RESET='\033[0m'

# Loop through all scripts from https://github.com/stefanpejcic/openpanel-docker-cli/
find "$SCRIPTS_DIR" -type f \
  ! -path "$SCRIPTS_DIR/.git/*" \
  ! -path "$SCRIPTS_DIR/.github/*" \
  ! -name "error.py" \
  ! -name "db.sh" \
  ! -name "aliases.txt" \
  ! -name "send_mail.sh" \
  ! -name "LICENSE.md" \
  ! -name "README.md" \
  ! -name "*README.md" \
  ! -name "*TODO*" | while read -r script; do
        script_name=$(basename "$script" | sed 's/\(\.sh\|\.py\)$//') # rm extensions
        dir_name=$(dirname "$script" | sed 's:.*/::') # folder name without the full path

        if [ "$dir_name" = "opencli" ]; then
            dir_name=""
        else
            dir_name="${dir_name}-"
        fi

        alias_name="${dir_name}${script_name}"
        full_alias="opencli $alias_name"
	description=$(grep -E "^# Description:" "$script" | sed 's/^# Description: //') # extract description if available
	usage=$(grep -E "^# Usage:" "$script" | sed 's/^# Usage: //') # extract usage if available
 
	echo -e "${GREEN}$full_alias${RESET}` #for $script`"

	if [ -n "$description" ]; then
		echo "Description: $description"
	fi
	if [ -n "$usage" ]; then
		echo "Usage: $usage"
	fi
 
	echo "------------------------"
 
	echo "$full_alias" >> "$ALIAS_FILE" # add to file
done

# Sort the aliases in the file by names
sort -o "$ALIAS_FILE" "$ALIAS_FILE"
