#!/bin/bash

wget -O /etc/openpanel/apache/httpd.conf https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/apache/httpd.conf

INSERT_TEXT="Listen 443"

for dir in /home/*; do
    file="$dir/httpd.conf"
    user=$(basename "$dir")

    # ADDS :443 FOR APACHE HTTPD
    
    if [[ -f "$file" ]]; then
        if ! grep -q "$INSERT_TEXT" "$file"; then
            sed -i "/Listen 80/a \\
$INSERT_TEXT" "$file"
            echo "Updated: $file"
    
            if docker --context "$user" ps --format '{{.Names}}' | grep "apache"; then
                cd /home/"$user"
                docker --context "$user" compose down apache   
                docker --context "$user" compose up -d apache
                echo "Apache restarted for user: $user"
            fi
        fi
    fi


    # ADDS ROOT USER FOR PHP-FPM
    file="$dir/docker-compose.yml"
    if [[ -f "$file" ]]; then
        sed -i '/^\s*php-fpm\s*$/s/php-fpm/php-fpm --allow-to-run-as-root/' $file       
    fi


    
    BACKUP_FILE="${file}.bak"
    
    # Create a backup before modifying
    cp "$file" "$BACKUP_FILE"
    echo "Backup created at $BACKUP_FILE"
    
    # Remove duplicate old values
    sed -i '/- APP_USER=${CONTEXT:-www-data}/d' "$file"
    sed -i '/- APP_GROUP=${CONTEXT:-www-data}/d' "$file"
    sed -i '/- APP_UID=0 #${USER_ID:-0}/d' "$file"
    sed -i '/- APP_GID=0 #${USER_ID:-0}/d' "$file"
    
    # Replace existing values if found
    sed -i 's|- APP_USER=${CONTEXT:-www-data}|- APP_USER=${CONTEXT:-root}|' "$file"
    sed -i 's|- APP_GROUP=${CONTEXT:-www-data}|- APP_GROUP=${CONTEXT:-root}|' "$file"
    sed -i 's|- APP_UID=0 #${USER_ID:-0}|- APP_UID=${USER_ID:-0}|' "$file"
    sed -i 's|- APP_GID=0 #${USER_ID:-0}|- APP_GID=${USER_ID:-0}|' "$file"
    
    # Check if PHP_UPLOAD_MAX_FILESIZE exists
    if grep -q "PHP_UPLOAD_MAX_FILESIZE=" "$file"; then
        # Extract indentation of PHP_UPLOAD_MAX_FILESIZE=
        INDENTATION=$(grep -m1 "PHP_UPLOAD_MAX_FILESIZE=" "$file" | sed -E 's/(.*)- PHP_UPLOAD_MAX_FILESIZE=.*/\1/')
    
        # Construct formatted lines
INSERT_LINES="${INDENTATION}- APP_USER=\${CONTEXT:-root}\n\
${INDENTATION}- APP_GROUP=\${CONTEXT:-root}\n\
${INDENTATION}- APP_UID=\${USER_ID:-0}\n\
${INDENTATION}- APP_GID=\${USER_ID:-0}"
    
        # Ensure the lines are present after PHP_UPLOAD_MAX_FILESIZE=
        if ! grep -q "APP_USER=" "$file"; then
            sed -i "/PHP_UPLOAD_MAX_FILESIZE=/a\\
$INSERT_LINES" "$file"
            echo "Lines added after PHP_UPLOAD_MAX_FILESIZE."
        fi
    else
        echo "PHP_UPLOAD_MAX_FILESIZE not found in the file."
    fi



done
