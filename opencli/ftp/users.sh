#!/bin/sh

USERS=""

for dir in /etc/openpanel/ftp/users/*; do
    file="$dir/users.list"
    user=$(basename "$dir")
    if [[ -f "$file" ]]; then
        # Get the group ID of the user
        group_id=$(id -g "$user")
        
        # Replace /var/www/html/ with /home/user/ and append group_id with '|' separator
        while IFS= read -r line; do
            modified_line=$(echo "$line" | sed "s|/var/www/html/|/home/${user}/docker-data/volumes/${user}_html_data/_data/|g")
            USERS+="$modified_line|$group_id "
        done < "$file"
    fi
done

echo "$USERS" 

echo "USERS=\"$USERS\"" > /etc/openpanel/ftp/all.users
