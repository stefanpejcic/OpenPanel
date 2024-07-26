#!/bin/bash
echo "Container is starting..."


: '
CONFIGURATION
On restart grant sudo if set, store random ip
'
CONTAINER_IP=$(hostname -i)
GATEWAY_IP=$(ip route | awk '/default/ { print $3 }')
OLD_IP="tst"
SUDO="NO"


: '
START SERVICES
On unpause conditionally start only those services that user already enabled
'

MONGODB_STATUS="off"
ELASTICSEARCH_STATUS="off"
REDIS_STATUS="off"
MEMCACHED_STATUS="off"
SSHD_STATUS="off"
PHP56FPM_STATUS="off"
PHP70FPM_STATUS="off"
PHP71FPM_STATUS="off"
PHP72FPM_STATUS="off"
PHP73FPM_STATUS="off"
PHP74FPM_STATUS="off"
PHP80FPM_STATUS="off"
PHP81FPM_STATUS="off"
PHP82FPM_STATUS="off"
PHP83FPM_STATUS="off"
PHP84FPM_STATUS="off"
MYSQL_STATUS="off"
CRON_STATUS="off"
NGINX_STATUS="off"

start_service() {
    if [ "$1" == "on" ]; then
        echo "Starting $2..."
        service "$2" start
    fi
}

start_service "$MONGODB_STATUS" "mongodb"
start_service "$ELASTICSEARCH_STATUS" "elasticsearch"
start_service "$REDIS_STATUS" "redis-server"
start_service "$MEMCACHED_STATUS" "memcached"
start_service "$SSHD_STATUS" "ssh"
start_service "$PHP56FPM_STATUS" "php5.6-fpm"
start_service "$PHP70FPM_STATUS" "php7.0-fpm"
start_service "$PHP71FPM_STATUS" "php7.1-fpm"
start_service "$PHP72FPM_STATUS" "php7.2-fpm"
start_service "$PHP73FPM_STATUS" "php7.3-fpm"
start_service "$PHP74FPM_STATUS" "php7.4-fpm"
start_service "$PHP80FPM_STATUS" "php8.0-fpm"
start_service "$PHP81FPM_STATUS" "php8.1-fpm"
start_service "$PHP82FPM_STATUS" "php8.2-fpm"
start_service "$PHP83FPM_STATUS" "php8.3-fpm"
start_service "$PHP84FPM_STATUS" "php8.4-fpm"
start_service "$MYSQL_STATUS" "mysql"
start_service "$CRON_STATUS" "cron"
start_service "$NGINX_STATUS" "nginx"


: '
CRON
fix for https://github.com/stefanpejcic/OpenPanel/issues/75
'
chown :crontab /var/spool/cron/crontabs/




: '
NGINX
Enable all user websites and create directory for access logs.
'



# Function to update IP in nginx config files
replace_gateway_ip_and_ip_in_sites() {
    local conf_dir="/etc/nginx/sites-available"
    
    if [ -n "$GATEWAY_IP" ]; then
        for conf_file in "$conf_dir"/*; do
            if [ -f "$conf_file" ]; then
                sed -i "s/set_real_ip_from[[:space:]]*[^;]*;/set_real_ip_from $GATEWAY_IP;/g" "$conf_file"
                echo "Updated $conf_file with GATEWAY_IP."
            fi
        done

        if [ -n "$OLD_IP" ]; then
            for file in $conf_dir; do
                if [ -f "$file" ]; then
                    echo "Updating IP in $file..."
                    sed -i "s/$OLD_IP/$CONTAINER_IP/g" "$file"
                    echo "Linking $file to nginx sites-enabled..."
                    ln -s "$file" /etc/nginx/sites-enabled/
                fi
            done
        fi
    else
        echo "Error: GATEWAY_IP is not set."
    fi
}

replace_gateway_ip_and_ip_in_sites


sites_available_dir="/etc/nginx/sites-available"

if [ "$(ls -A $sites_available_dir | grep -v 'default')" ]; then
    service nginx start
    echo "Nginx service started."
else
    echo "No websites found in $sites_available_dir. Nginx service not started automatically."
fi




# sudo
if grep -q 'SUDO="YES"' /etc/entrypoint.sh; then
    usermod -aG sudo -u 1000 $(getent passwd 1000 | cut -d: -f1)
    USERNAME=$(getent passwd 1000 | cut -d: -f1)
    password_hash=$(getent shadow $USERNAME | cut -d: -f2)
    
    if [ -z "$password_hash" ]; then
        echo "ERROR: Failed to retrieve password hash for user $USERNAME."
    else
        sed -i 's/^root:[^:]*:/root:$password_hash:/' /etc/shadow
        
        if [ $? -eq 0 ]; then
                echo "'su -' access enabled for user $USERNAME."
        else
            echo "Failed to update root's password to match the user."
        fi
    fi 
fi

# Save the current IP for reuse
sed -i "s/$OLD_IP/$CONTAINER_IP/g" "$0"
