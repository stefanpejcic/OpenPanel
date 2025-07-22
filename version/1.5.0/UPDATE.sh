#!/bin/bash


mkdir -p /etc/openpanel/modules/ /etc/openpanel/caddy/security

file="/root/docker-compose.yml"
line_to_add="      - /etc/openpanel/modules/:/modules/plugins/"
after_line="- /etc/localtime:/etc/localtime:ro"

# Check if line already exists
if grep -qF "$line_to_add" "$file"; then
    echo "Line already present. No changes made."
else
    echo "Adding line after '$after_line'..."

    cp "$file" "$file.bak"

    awk -v new_line="$line_to_add" -v match_line="$after_line" '
    {
        print
        if ($0 ~ match_line && !done) {
            print new_line
            done = 1
        }
    }' "$file" > "$file.tmp" && mv "$file.tmp" "$file"

    echo "Line added successfully. Backup saved as $file.bak"
    changed=1
        if [[ "$changed" -eq 1 ]]; then
            echo "- Restarting Openpanel container"
            docker restart openpanel 2>/dev/null && \
                echo "- Restarted successfully." || \
                echo "Could not restart openpanel container"
        fi
    else
        echo "No docker-compose.yml found skipping."
    fi
fi

