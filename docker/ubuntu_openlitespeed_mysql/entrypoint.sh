#!/bin/bash
echo "Container is starting..."

: '
CONFIGURATION
On restart grant sudo if set, store random ip
'
CONTAINER_IP=$(hostname -i)
OLD_IP="tst"
SUDO="NO"


# Configuration files and directories
memcached_dir="/var/run/memcached/"




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
PHP82FPM_STATUS="on"
PHP83FPM_STATUS="off"
PHP84FPM_STATUS="off"
MYSQL_STATUS="off"
CRON_STATUS="off"
LITESPEED_STATUS="off"

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
start_service "$LITESPEED_STATUS" "lsws"




: '
CRON
fix for https://github.com/stefanpejcic/OpenPanel/issues/75
'
chown :crontab /var/spool/cron/crontabs/






: '
LITESPEED
Enable all user websites.
'

if [ -z "$(ls -A -- "/usr/local/lsws/conf/")" ]; then
	cp -R /usr/local/lsws/.conf/* /usr/local/lsws/conf/
fi
if [ -z "$(ls -A -- "/usr/local/lsws/admin/conf/")" ]; then
	cp -R /usr/local/lsws/admin/.conf/* /usr/local/lsws/admin/conf/
fi
chown 994:994 /usr/local/lsws/conf -R
chown 994:1001 /usr/local/lsws/admin/conf -R

sites_available_dir="/usr/local/lsws/conf"


# if there are any sites, start the service
if [ "$(ls -A $sites_available_dir | grep -v 'default.conf')" ]; then
    #service lsws start
    /usr/local/lsws/bin/lswsctrl start
    echo "Litespeed service started."
else
    echo "No websites found in $sites_available_dir. Litespeed service not started automatically."
fi


# pm2
if [ -s /root/.pm2/dump.pm2 ]; then
    echo "There are saved NodeJS/Python applications, starting them.."
    pm2 resurrect
else
    echo "No saved NodeJS/Python applications to run. Skipping 'pm2 resurrect'."
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
