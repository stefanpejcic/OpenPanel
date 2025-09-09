#!/bin/bash
set -euo pipefail  # Exit on error, undefined variable, or pipeline failure

# Define required OpenLiteSpeed volumes
REQUIRED_VOLUMES=$(cat <<'EOV'
      - ./openlitespeed.conf:/usr/local/lsws/conf/httpd_config.conf # main conf
      - /etc/openpanel/openlitespeed/start.sh:/entrypoint.sh:ro     # chmod +x first!
      - ./sockets/mysqld:/var/run/mysqld                            # MySQL socket
      - webserver_data:/usr/local/lsws/conf/vhosts/                 # vhosts
      - html_data:/var/www/html/                                    # Website files
      - /etc/openpanel/nginx/certs/:/etc/nginx/ssl/                 # SSLs - not used atm
      - /etc/openpanel/wordpress/wp-cli.phar:/usr/bin/wp            # wpcli
      - ./php.ini/7.4.ini:/usr/local/lsws/lsphp74/etc/php/7.4/litespeed/php.ini
      - ./php.ini/8.0.ini:/usr/local/lsws/lsphp80/etc/php/8.0/litespeed/php.ini
      - ./php.ini/8.1.ini:/usr/local/lsws/lsphp81/etc/php/8.1/litespeed/php.ini
      - ./php.ini/8.2.ini:/usr/local/lsws/lsphp82/etc/php/8.2/litespeed/php.ini
      - ./php.ini/8.3.ini:/usr/local/lsws/lsphp83/etc/php/8.3/litespeed/php.ini
      - ./php.ini/8.4.ini:/usr/local/lsws/lsphp84/etc/php/8.4/litespeed/php.ini
EOV
)

# Iterate over all user home directories
for dir in /home/*; do
    file="$dir/docker-compose.yml"
    user=$(basename "$dir")

    # Skip if docker-compose.yml does not exist
    if [[ ! -f "$file" ]]; then
        echo "No docker-compose.yml for user: $user"
        continue
    fi

    echo "🔍 Processing $user ..."

    # Check if openlitespeed service exists in docker-compose
    if ! grep -qE '^\s*openlitespeed:' "$file"; then
        echo "No openlitespeed service, skipping."
        continue
    fi

    # Create a timestamped backup before modifying
    backup="$file.bak.$(date +%Y%m%d_%H%M%S)"
    cp "$file" "$backup"
    echo "  🛡️  Backup saved as $backup"

    # Determine the starting line and indentation of openlitespeed service
    start_line=$(grep -nE '^\s*openlitespeed:' "$file" | cut -d: -f1)
    indent=$(grep -o '^\s*' <<< "$(sed -n "${start_line}p" "$file")")

    # Determine the end line of openlitespeed block
    # End is the line before next line with same or less indentation
    end_line=$(tail -n +"$((start_line+1))" "$file" | awk -v indent_len=${#indent} '
      function ltrim(s){sub(/^[ \t]+/, "", s); return s}
      {
        line_indent = length($0) - length(ltrim($0))
        if(line_indent <= indent_len && NF>0){print NR; exit}
      }
    ')
    if [[ -z "$end_line" ]]; then
        end_line=$(wc -l < "$file")
    else
        end_line=$((start_line + end_line -1))
    fi

    # Check if volumes block exists inside openlitespeed
    if sed -n "${start_line},${end_line}p" "$file" | grep -qE '^\s*volumes:'; then
        echo "Volumes block found, checking..."
        # Add missing volumes
        while IFS= read -r volume; do
            if [[ -n "$volume" ]] && ! sed -n "${start_line},${end_line}p" "$file" | grep -qF "$volume"; then
                sed -i "${start_line},${end_line} s|^\(\s*volumes:\)|\1\n$volume|" "$file"
                echo "  ➕ Added volume: $volume"
            fi
        done <<< "$REQUIRED_VOLUMES"
    else
        echo "No volumes block, creating it..."
        # Insert a new volumes block after working_dir:
        sed -i "${start_line},${end_line} s|^\(\s*working_dir:.*\)|\1\n    volumes:|" "$file"
        while IFS= read -r volume; do
            if [[ -n "$volume" ]]; then
                sed -i "${start_line},${end_line} s|^\(\s*volumes:\)|\1\n$volume|" "$file"
            fi
        done <<< "$REQUIRED_VOLUMES"
        echo "Volumes block created."
    fi

    echo "---------------------------------------------------------------"
done
