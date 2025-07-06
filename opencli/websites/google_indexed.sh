#!/bin/bash
################################################################################
# Script Name: webistes/google_index.sh
# Description: Check if website is indexed on Google and monitor results.
# Usage: opencli webistes/google_index --domain [DOMAIN]
# Author: Stefan Pejcic
# Created: 03.06.2025
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

MAX_JOBS=5  # keep lowAdd commentMore actions
mkdir -p "/etc/openpanel/openpanel/websites/"

check_index() {
  local domain=$1
  local timestamp=$(date +"%Y-%m-%d %H:%M:%S")
  local filename="/etc/openpanel/openpanel/websites/$(echo "$domain" | sed 's|https\?://||' | sed 's|/|_|g').google_index.json"
  local result page_count indexed error_msg=""

  # Fetch search results page from Google
  result=$(curl -s -A "Mozilla/5.0" "https://www.google.com/search?q=site:$domain")

  # Determine indexed status and results count
  if echo "$result" | grep -q "did not match any documents"; then
    indexed=false
    page_count=0
  else
    indexed=true
    page_count=$(echo "$result" | grep -oP 'About [\d,]+ results' | head -1 | sed 's/About //; s/ results//; s/,//g')
    if ! [[ "$page_count" =~ ^[0-9]+$ ]]; then
      page_count=0
    fi
  fi

  # Compare with previous results if file exists
  if [ -f "$filename" ]; then
    prev_indexed=$(jq -r '.indexed' "$filename")
    prev_count=$(jq -r '.results_count' "$filename" | sed 's/,//g')
    if ! [[ "$prev_count" =~ ^[0-9]+$ ]]; then
      prev_count=0
    fi

    if [ "$prev_indexed" = "true" ] && [ "$indexed" = "false" ]; then
      error_msg="Site was indexed before but now NOT indexed."
    fi

    if [ "$indexed" = "true" ] && [ "$prev_count" -gt 0 ]; then
      drop=$((prev_count - page_count))
      drop_percent=$(( drop * 100 / prev_count ))
      if [ "$drop_percent" -ge 10 ]; then
        error_msg="Results count dropped by ${drop_percent}% compared to previous check."
      fi
    fi
  fi

  # Write JSON with or without error field
  if [ -n "$error_msg" ]; then
    cat > "$filename" <<EOF
{
  "timestamp": "$timestamp",
  "domain": "$domain",
  "indexed": $indexed,
  "results_count": "$page_count",
  "error": "$error_msg"
}
EOF
  else
    cat > "$filename" <<EOF
{
  "timestamp": "$timestamp",
  "domain": "$domain",
  "indexed": $indexed,
  "results_count": "$page_count"
}
EOF
  fi

  # Print JSON to console
  cat "$filename"
}

run_parallel() {
  local domains=("$@")
  local count=0

  for domain in "${domains[@]}"; do
    check_index "$domain" &
    ((count++))

    if (( count % MAX_JOBS == 0 )); then
      wait
    fi
  done
  wait
}

DOMAIN=""

# Argument parsing
while [[ $# -gt 0 ]]; do
  case $1 in
    --domain)
      DOMAIN="$2"
      shift 2
      ;;
    *)
      echo "Unknown argument: $1"
      echo "Usage: $0 [--domain <DOMAIN>]"
      exit 1
      ;;
  esac
done

if [ -z "$DOMAIN" ]; then
  # Run for all domains from opencli websites-all
  mapfile -t domains < <(opencli websites-all)
  run_parallel "${domains[@]}"
else
  check_index "$DOMAIN"
fi
