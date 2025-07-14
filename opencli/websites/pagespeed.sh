#!/bin/bash
################################################################################
# Script Name: websites/pagespeed.sh
# Description: Check Google PageSpeed data for website(s)
# Usage: opencli websites-pagespeed <DOMAIN> [-all]
# Author: Stefan Pejcic
# Created: 27.06.2024
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

# Check if jq is installed
if ! command -v jq &> /dev/null; then
  echo "jq could not be found, please install jq to proceed."
  exit 1
fi

usage() {
  echo "Usage: opencli websites-pagespeed <website> OR opencli websites-pagespeed -all"
  exit 1
}


get_api_key() {
  actual_domain="${website_url%%/*}"
  actual_domain="${actual_domain#*://}"
  
  whoowns_output=$(opencli domains-whoowns "$actual_domain")
  owner=$(echo "$whoowns_output" | awk -F "Owner of '$actual_domain': " '{print $2}')
  
if [ -n "$owner" ]; then
  source /usr/local/opencli/db.sh
  query="SELECT server FROM users WHERE username = '$owner'"
  context=$(mysql -D "$mysql_database" -e "$query" -sN)
  if [ -z "$context" ]; then
    api_key=""
  else
    api_key=$(cat /etc/openpanel/openpanel/service/pagespeed.api)
    user_key_path="/hostfs/home/$context/docker-data/volumes/${context}_html_data/_data/pagespeed_api_key.txt"
    global_key_path="/hostfs/etc/openpanel/openpanel/service/pagespeed.api"
    
    if [[ -s "$user_key_path" ]]; then
      api_key=$(cat "$user_key_path")
    elif [[ -s "$global_key_path" ]]; then
      api_key=$(cat "$global_key_path")
    else
      api_key=""
    fi
  fi
fi


}

get_page_speed() {
  local website_url=$1
  local strategy=$2
  local encoded_domain=$(printf '%s' "$website_url" | jq -s -R -r @uri)
  

  local api_url="https://www.googleapis.com/pagespeedonline/v5/runPagespeed?url=http://$encoded_domain&strategy=$strategy"

  get_api_key
  
  [[ -n "$api_key" ]] && api_url+="&key=$api_key"

  local api_response=$(curl -s "$api_url")
  
  # Check for error in API response
  local error_message=$(echo "$api_response" | jq -r '.error.message // empty')
  if [[ -n "$error_message" ]]; then
    echo "API Error: $error_message"
    return 0  # Return 0 so `$(...)` doesn't break; use output to detect error
  fi

  local performance_score=$(echo "$api_response" | jq '.lighthouseResult.categories.performance.score')
  local first_contentful_paint=$(echo "$api_response" | jq -r '.lighthouseResult.audits."first-contentful-paint".displayValue')
  local speed_index=$(echo "$api_response" | jq -r '.lighthouseResult.audits."speed-index".displayValue')
  local interactive=$(echo "$api_response" | jq -r '.lighthouseResult.audits.interactive.displayValue')
  
  echo "{\"performance_score\": $performance_score, \"first_contentful_paint\": \"$first_contentful_paint\", \"speed_index\": \"$speed_index\", \"interactive\": \"$interactive\"}"
}


# Function to generate report for a domain
generate_report() {
  local website=$1
  local desktop_speed=$(get_page_speed "$website" "desktop")
  if echo "$desktop_speed" | grep -q "API Error"; then
    echo "$desktop_speed"
    echo "Aborting due to API error."
    return 1
  fi
  local mobile_speed=$(get_page_speed "$website" "mobile")
  local timestamp=$(date +"%Y-%m-%d %H:%M:%S")
  local filename="/etc/openpanel/openpanel/websites/$(echo "$website" | sed 's|https\?://||' | sed 's|/|_|g').json"
  
  cat <<EOF > "$filename"
{
  "timestamp": "$timestamp",
  "website": "$website",
  "desktop_speed": $desktop_speed,
  "mobile_speed": $mobile_speed
}
EOF

  echo "Google PageSpeed data saved to $filename"
}

mkdir -p /etc/openpanel/openpanel/websites

if [ $# -eq 0 ]; then
  usage
elif [[ "$1" == "-all" || "$1" == "--all" ]]; then

  websites=$(opencli websites-all)

  if [[ -z "$websites" || "$websites" == "No sites found in the database." ]]; then
    echo "No sites found in the database or opencli command error."
    exit 1
  fi

  for website in $websites; do
    generate_report "$website"
  done
elif [ $# -eq 1 ]; then
  generate_report "$1"
else
  usage  
fi
