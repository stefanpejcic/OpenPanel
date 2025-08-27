#!/bin/sh

USERS=""

for dir in /etc/openpanel/ftp/users/*; do
    file="$dir/users.list"
    user=$(basename "$dir")
    if [[ -f "$file" ]]; then
        while IFS= read -r line; do
            modified_line=$(echo "$line" | sed "s|/var/www/html/|/home/${user}/docker-data/volumes/${user}_html_data/_data/|g")
            USERS+="$modified_line"
        done < "$file"
    fi
done

echo "$USERS" 

echo "USERS=\"$USERS\"" > /etc/openpanel/ftp/all.users
