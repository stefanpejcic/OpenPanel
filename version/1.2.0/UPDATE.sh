#!/bin/bash

wget -O /etc/openpanel/docker/compose/1.0/docker-compose.yml https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/docker/compose/1.0/docker-compose.yml

wget -O /etc/openpanel/docker/compose/1.0/.env https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/docker/compose/1.0/.env

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
    fi
    
    file="$dir/docker-compose.yml"
    user=$(basename "$dir")
    if [[ -f "$file" ]]; then
        BACKUP_FILE="${file}.120_bak"
        cp "$file" "$BACKUP_FILE"
        echo "Backup created at $BACKUP_FILE"
        
        
        for dir in /home/*; do
            # Extract the user name from the directory path
            user=$(basename "$dir")
        
            # Change to the user's directory
            cd /home/$user || continue
        
            # List the services for the user's Docker context
            services=$(docker --context $user compose ps --services)

            cp /etc/openpanel/docker/compose/1.0/docker-compose.yml $file
        
            # If services are found, proceed with stopping and restarting them
            if [ -n "$services" ]; then
                # Stop the services
                docker --context $user compose down $services
        
                # Start the services in detached mode
                docker --context $user compose up -d $services
            else
                echo "No services found for user $user"
            fi
        done
    fi

done
