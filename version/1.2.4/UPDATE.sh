#!/bin/bash

echo ""
echo "Adding fix for custom files not loading.. - issue #444"
sed -i 's#/usr/local/panel/#/#g' /root/docker-compose.yml 
cd /root
docker compose down openpanel && docker compose up -d openpanel

for dir in /home/*; do
    file="$dir/docker-compose.yml"
    user=$(basename "$dir")

    if [[ -f "$file" ]]; then
        echo ""
        echo "---------------------------------------------------------------"
        echo "user: $user"
        cp $file $dir/024-docker-compose.yml
        temp_file=$(mktemp)
        while IFS= read -r line; do
            if [[ "$line" =~ memory:\ \" ]]; then
                echo "$line" >> "$temp_file"
                indent=$(echo "$line" | sed 's/^\([[:space:]]*\).*/\1/')
                # pids: ${OS_PIDS:-100} # https://github.com/docker/cli/issues/5009
                echo "${indent}pids: 40" >> "$temp_file"
            else
                echo "$line" >> "$temp_file"
            fi
        done < "$file"
        mv "$temp_file" "$file"
        echo "Docker Compose file has been updated to limit PIDs per service to 40."
    fi
done
