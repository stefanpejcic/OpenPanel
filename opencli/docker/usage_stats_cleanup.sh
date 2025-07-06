#!/bin/bash
################################################################################
# Script Name: usage_stats_cleanup.sh
# Description: Rotates resource usage logs for all users according to the resource_usage_retention setting.
# Use: opencli docker-usage_stats_cleanup
# Author: Stefan Pejcic
# Created: 01.10.2023
# Last Modified: 04.07.2025
# Company: openpanel.co
# Copyright (c) openpanel.co
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

# Define the directory where the JSON files are stored
stats_dir="/etc/openpanel/openpanel/core/stats"

# Get the resource_usage_retention value from panel.config
panel_config="/etc/openpanel/openpanel/conf/openpanel.config"
resource_usage_retention=$(grep -Eo "resource_usage_retention=[0-9]+" "$panel_config" | cut -d'=' -f2)

if [[ -z $resource_usage_retention ]]; then
  echo "Error: Could not determine resource_usage_retention value in panel.config"
  exit 1
fi

# Loop through the directories
for user_dir in "$stats_dir"/*; do
  if [ -d "$user_dir" ]; then
    user=$(basename "$user_dir")
    file_count=$(find "$user_dir" -name "*.json" | wc -l)
    
    # Calculate the number of files to delete
    files_to_delete=$((file_count - resource_usage_retention))

    # Delete the oldest files if necessary
    if [ "$files_to_delete" -gt 0 ]; then
      cd "$user_dir" || exit 1
      find . -name "*.json" -type f -printf '%T@ %p\n' | sort -n | head -n "$files_to_delete" | cut -d' ' -f2 | xargs rm
      echo "Deleted $files_to_delete files in $user_dir"
    fi
  fi
done

exit 0
