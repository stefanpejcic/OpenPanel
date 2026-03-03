#!/bin/bash

echo "Applying patch for WP-CLI incorrect path on OpenLitespeed.."
sed -i 's|/usr/bin/wp|/usr/local/bin/wp:ro|g' /etc/openpanel/docker/compose/1.0/docker-compose.yml

for file in /home/*/docker-compose.yml; do
    if [ -f "$file" ]; then
        sed -i 's|/usr/bin/wp|/usr/local/bin/wp:ro|g' "$file"
        echo "Updated: $file"
    fi
done
