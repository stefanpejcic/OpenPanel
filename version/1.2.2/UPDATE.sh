#!/bin/bash

echo "Updating template: /etc/openpanel/varnish/default.vcl"
wget -O /etc/openpanel/varnish/default.vcl https://github.com/stefanpejcic/openpanel-configuration/blob/main/varnish/default.vcl

for dir in /home/*; do
    file="$dir/.env"
    user=$(basename "$dir")

    if [[ -f "$file" ]]; then
        
        cp /etc/openpanel/varnish/default.vcl file="$dir/default.vcl"
        echo "- Updated Varnish default.vcl template for user: $user"
        
        if ! grep -q 'CRONJOBS' "$file"; then
            sed -i '/BUSYBOX_RAM="0.1G"/a \
# CRONJOBS\nCRON_CPU="0.1"\nCRON_RAM="0.25G"' "$file"

            echo "- Updated $file for user: $user to add CRON limits"
        fi
    fi
    
done
