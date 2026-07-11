#!/bin/bash
################################################################################
# Script Name: websites/wp_autodownload.sh
# Description: Pre-download latest 10 WordPress releases to be available in WP Manager.
# Usage: opencli websites-wp_autodownload
# Author: Stefan Pejcic
# Created: 13.03.2026
# Last Modified: 10.07.2026
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

set -euo pipefail

if [[ "$(opencli config get wp_autodownload)" != "yes" ]]; then
    echo "[SKIP] WordPress auto-download is disabled, to enable it run: opencli config update wp_autodownload yes"
    exit 0
fi

# Number of WordPress releases to pre-download (default: 10)
DOWNLOAD_COUNT="${1:-10}"

# Validate input
if ! [[ "$DOWNLOAD_COUNT" =~ ^[0-9]+$ ]] || [[ "$DOWNLOAD_COUNT" -lt 1 ]]; then
    echo "[ERROR] Invalid download count: $DOWNLOAD_COUNT"
    echo "Usage: $0 [number_of_versions]"
    exit 1
fi

API="https://api.wordpress.org/core/stable-check/1.0/"
ARCHIVE_DIR="/etc/openpanel/wordpress/archives"
BASE_URL="https://wordpress.org"

mkdir -p "$ARCHIVE_DIR"
mapfile -t versions < <(curl -fsSL "$API" | jq -r 'keys[]' | sort -V | tail -n "$DOWNLOAD_COUNT")

for wordpress_version in "${versions[@]}"; do
    archive_name="wordpress-${wordpress_version}.tar.gz"
    archive_path="${ARCHIVE_DIR}/${archive_name}"

    [[ -f "$archive_path" ]] && { echo "[OK] Already exists: $archive_path"; continue; }

    url="${BASE_URL}/${archive_name}"

    echo "[GET] $url"

    if curl -fLs --retry 3 -o "$archive_path" "$url"; then
        echo "[OK] Downloaded $archive_path"
    else
        echo "[FAIL] Could not download $url"
        rm -f "$archive_path"
    fi
done
