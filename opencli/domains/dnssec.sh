#!/bin/bash
################################################################################
# Script Name: domains/dnssec.sh
# Description: Enable DNSSEC for a domain and re-sign after changes in the zone.
# Usage: opencli domains-dnssec <DOMAIN> [--update | --check]
# Author: Stefan Pejcic
# Created: 09.07.2024
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

# Variables
PDIR=$(pwd)
ZONEDIR="/etc/bind/zones"
ZONE=$1
ZONEFILE="/etc/bind/zones/${ZONE}.zone"
CONFIG_FILE="/etc/bind/named.conf.local"
CONTAINER="openpanel_dns"

# Functions
error_exit() {
  echo "Error: $1"
  cd $PDIR
  exit 1
}

sign_and_reload() {
  docker exec $CONTAINER bash -c "cd $ZONEDIR && dnssec-signzone -A -3 $(head -c 1000 /dev/random | sha1sum | cut -b 1-16) -N INCREMENT -P -o ${ZONE} -t ${ZONEFILE} >/dev/null 2>&1" || error_exit 'Failed to sign the zone file'
  opencli domains-dns reload $ZONE >/dev/null 2>&1 || error_exit "Failed to reload the DNS zone"
}

setup_zone() {
  # Check if the zone file exists
  if [ ! -f "$ZONEFILE" ]; then
    error_exit "Zone file $ZONEFILE does not exist"
  fi

  # Change to the zone directory
  cd $ZONEDIR >/dev/null 2>&1 || error_exit "Failed to change directory to $ZONEDIR"

  # Generate key pairs
  docker exec $CONTAINER bash -c "dnssec-keygen -a NSEC3RSASHA1 -b 2048 -n ZONE ${ZONE} >/dev/null 2>&1" || error_exit 'Failed to generate 2048-bit key'
  docker exec $CONTAINER bash -c "dnssec-keygen -a NSEC3RSASHA1 -b 4096 -n ZONE ${ZONE} >/dev/null 2>&1" || error_exit 'Failed to generate 4096-bit key'

docker cp $CONTAINER:


  # Allow bind group to read the keys
  docker exec $CONTAINER bash -c "chgrp bind K${ZONE}.* >/dev/null 2>&1" || error_exit 'Failed to change group of key files'
  docker exec $CONTAINER bash -c "chmod g=r,o= K${ZONE}.* >/dev/null 2>&1" || error_exit 'Failed to set permissions on key files'

  # Include keys to the zone file
  for key in K${ZONE}.*.key; do
    
    echo "\$INCLUDE $key" >> ${ZONEFILE}
  done

  docker exec openpanel_dns bash -c "cd $ZONEDIR && dnssec-signzone -A -3 $(head -c 1000 /dev/random | sha1sum | cut -b 1-16) -N INCREMENT -P -o ${ZONE} -t ${ZONEFILE} >/dev/null 2>&1" || error_exit 'Failed to sign the zone file'

  # Use sed to append .signed to the filename on the specific line containing the zone
  sed -i "/zone \"${ZONE}\"/,/file/s|\(file \"/etc/bind/zones/${ZONE}\.zone\)|\1.signed|" "$CONFIG_FILE" >/dev/null 2>&1 || error_exit "Failed to update the config file"

  # relaod service
  opencli domains-dns reload $ZONE >/dev/null 2>&1 || error_exit "Failed to reload the DNS zone"

  # Display DS records
  cat dsset-${ZONE}. || error_exit "Failed to display DS records"
}





# Check for required arguments
if [ -z "$ZONE" ]; then
  error_exit "Usage: $0 <DOMAIN> [--update | --check]"
fi

# Parse optional flag
if [ "$2" = "--update" ]; then
  sign_and_reload
  echo "Zone ${ZONE} has been re-signed and DNS service reloaded."
elif [ "$2" = "--check" ]; then
  cat dsset-${ZONE}. || error_exit "Domain $ZONE has no DNSSEC enabled."
else
  setup_zone
fi
