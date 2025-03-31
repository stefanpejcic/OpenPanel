#!/bin/bash

wget -O /etc/openpanel/apache/httpd.conf https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/apache/httpd.conf

INSERT_TEXT="Listen 443"

for dir in /home/*; do
    file="$dir/httpd.conf"
    user=$(basename "$dir")
    
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
done
