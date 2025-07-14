#!/bin/bash
################################################################################
# Script Name: domains/update_ns.sh
# Description: Change nameservers for a single or all dns zones.
# Usage: opencli domains-update_ns <DOMAIN_NAME>
#        opencli domains-update_ns --all
# Author: Stefan Pejcic
# Created: 20.08.2023
# Last Modified: 13.07.2025
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


# Define the path to the configuration file
CONFIG_FILE="/etc/openpanel/openpanel/conf/openpanel.config"
ZONE_DIR="/etc/bind/zones"
BACKUP_DIR="$ZONE_DIR/backups"

# Function to print usage
print_usage() {
  echo "Usage: $0 [--all | --all -y | domain_name]"
  echo "Options:"
  echo "  --all       Update all zone files."
  echo "  --all -y    Update all zone files without confirmation."
  echo "  domain_name Update the zone file for the specified domain."
  exit 1
}

# Function to create a backup of a zone file
backup_zone_file() {
  local zone_file="$1"
  mkdir -p "$BACKUP_DIR"
  cp "$zone_file" "$BACKUP_DIR/$(basename "$zone_file").bak"
}

# Function to update nameservers in a zone file
update_zone_file() {
  local zone_file="$1"
  local tmp_file=$(mktemp)

  # Backup the file before editing
  backup_zone_file "$zone_file"

  # Remove existing NS records and create new content
  grep -v '^[^;].*IN[[:space:]]*NS[[:space:]]*' "$zone_file" > "$tmp_file"

  # Add new NS records at the top
  for ns in "${NAMESERVERS[@]}"; do
    echo "@ IN NS $ns." >> "$tmp_file"
  done

  # Prepend new NS records and restore original file structure
  cat "$tmp_file" > "$zone_file"
  rm "$tmp_file"
}

# Function to confirm the operation for --all
confirm_all() {
  echo "You are about to update all zone files. This action cannot be undone."
  echo "Proceed? (y/n)"
  read -t 10 -r confirm
  if [[ $confirm != "y" ]]; then
    echo "Operation aborted."
    exit 1
  fi
}

# Check if the configuration file exists
if [ ! -f "$CONFIG_FILE" ]; then
  echo "Configuration file not found: $CONFIG_FILE"
  exit 1
fi

# Extract nameserver values from the configuration file
NS1=$(grep '^ns1=' "$CONFIG_FILE" | cut -d'=' -f2)
NS2=$(grep '^ns2=' "$CONFIG_FILE" | cut -d'=' -f2)
NS3=$(grep '^ns3=' "$CONFIG_FILE" | cut -d'=' -f2)
NS4=$(grep '^ns4=' "$CONFIG_FILE" | cut -d'=' -f2)

# Check if at least ns1 and ns2 are set
if [ -z "$NS1" ] || [ -z "$NS2" ]; then
  echo "ns1 and ns2 must be set in the configuration file."
  exit 1
fi

# Collect nameservers to be used
NAMESERVERS=("$NS1" "$NS2")
[ -n "$NS3" ] && NAMESERVERS+=("$NS3")
[ -n "$NS4" ] && NAMESERVERS+=("$NS4")

case "$1" in
  --all)
    if [ "$2" != "-y" ]; then
      confirm_all
    fi
    for zone_file in "$ZONE_DIR"/*.zone; do
      if [ -f "$zone_file" ]; then
        update_zone_file "$zone_file"
      fi
    done
    ;;
    
  "")
    print_usage
    ;;

  *)
    domain="$1"
    zone_file="$ZONE_DIR/$domain.zone"
    if [ -f "$zone_file" ]; then
      update_zone_file "$zone_file"
    else
      echo "Zone file for $domain not found: $zone_file"
      exit 1
    fi
    ;;
esac


# Reload BIND service
docker --context default exec openpanel_dns rndc reconfig >/dev/null 2>&1
cd /root && docker --context default compose up -d bind9  >/dev/null 2>&1

echo "Nameservers have been updated and BIND9 zones reloaded."
exit 0
