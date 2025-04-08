#!/bin/bash

wget -O /etc/openpanel/wordpress/wp-cli.phar https://raw.githubusercontent.com/wp-cli/builds/gh-pages/phar/wp-cli.phar

chmod +x /etc/openpanel/wordpress/wp-cli.phar

INSERT_TEXT="user root;"

for dir in /home/*; do
    file="$dir/nginx.conf"
    user=$(basename "$dir")

    if [[ -f "$file" ]]; then
        if ! grep -q "$INSERT_TEXT" "$file"; then
            sed -i "/worker_rlimit_nofile 65535;/a \\
$INSERT_TEXT" "$file"
            echo "Updated: $file"
    
            if docker --context "$user" ps --format '{{.Names}}' | grep "nginx"; then
                cd /home/"$user"
                docker --context "$user" compose down nginx   
                docker --context "$user" compose up -d nginx
                echo "Nginx restarted for user: $user"
            fi
        fi

    # step 2. 
    BACKUP_FILE="${file}.120_bak"
    
    # Create a backup before modifying
    cp "$file" "$BACKUP_FILE"
    echo "Backup created at $BACKUP_FILE"
    # OVDE
    fi

done
