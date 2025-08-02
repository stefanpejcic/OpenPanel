#!/bin/bash

ZONE_DIR="/etc/bind/zones"

# get domain
extract_domain_from_filename() {
  local filename="$1"
  echo "${filename%.zone}"
}


# reload
reload_bind9() {
  local domain="$1"
  echo "Zone file changed. Reloading BIND9 for domain $domain..."
  if docker exec openpanel_dns rndc reload "$domain"; then
    echo "BIND9 reloaded successfully for domain $domain."
  else
    echo "Failed to reload BIND9 for domain $domain." >&2
  fi
}


# Ensure inotifywait is installed
if ! command -v inotifywait &> /dev/null; then

        # Detect the package manager and install inotifywait-tools
        if command -v apt-get &> /dev/null; then
            sudo apt-get update > /dev/null 2>&1
            sudo apt-get install -y -qq inotify-tools > /dev/null 2>&1
        elif command -v yum &> /dev/null; then
            sudo yum install -y -q inotify-tools > /dev/null 2>&1
        elif command -v dnf &> /dev/null; then
            sudo dnf install -y -q inotify-tools > /dev/null 2>&1
        else
            echo "Error: No compatible package manager found. Please install inotify-tools manually and try again."
            exit 1
        fi

        # Check if installation was successful
        if ! command -v inotifywait &> /dev/null; then
            echo "Error: inotifywait installation failed. Please install inotify-tools manually and try again."
            exit 1
        fi
fi


# monitor zones

echo "Monitoring zone files in $ZONE_DIR for changes..."

inotifywait -m -e modify "$ZONE_DIR" --format '%w%f' | while read -r file; do
  # Extract the domain from the zone file name
  DOMAIN=$(extract_domain_from_filename "$(basename "$file")")
  if [[ -n "$DOMAIN" ]]; then
    echo "Detected change for domain $DOMAIN - reloading zone"
    reload_bind9 "$DOMAIN"
  else
    echo "Unable to extract domain from file: $file" >&2
  fi
done
