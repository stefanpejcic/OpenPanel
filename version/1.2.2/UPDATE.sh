#!/bin/bash

for dir in /home/*; do
    file="$dir/.env"
    user=$(basename "$dir")

    if [[ -f "$file" ]]; then
        # Check if lines already exist (loosely)
        if ! grep -q 'CRONJOBS' "$file"; then
            sed -i '/BUSYBOX_RAM="0.1G"/a \
# CRONJOBS\nCRON_CPU="0.1"\nCRON_RAM="0.25G"' "$file"
            echo "Updated: $file"
        fi
    fi
done
