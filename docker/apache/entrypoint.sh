#!/bin/bash
echo "Container is starting..."

# Get the current container's IP address
CONTAINER_IP=$(hostname -i)

# Old IP address to be replaced on docker run
OLD_IP="tst"

# Configuration files and directories
memcached_dir="/var/run/memcached/"
apache_default_site="/etc/apache2/sites-available/000-default.conf"
service cron start



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
APACHE_STATUS="off"

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
start_service "$MYSQ_STATUS" "mysql"
start_service "$CRON_STATUS" "cron"
start_service "$APACHE_STATUS" "apache2"




: '
CRON
fix for https://github.com/stefanpejcic/OpenPanel/issues/75
'
chown :crontab /var/spool/cron/crontabs/






: '
APACHE
Enable all user websites and create directory for access logs.
'


#Make domlogs
mkdir -p /var/log/apache2/domlogs

# Replace the old IP address with the containers new IP in nginx config files and enable website
replace_ip_in_sites() {
    if [ -n "$OLD_IP" ]; then
        local site_files="/etc/apache2/sites-available"
        for file in $site_files; do
            if [ -f "$file" ]; then
                echo "Updating IP in $file..."
                sed -i "s/$OLD_IP/$CONTAINER_IP/g" "$file"
                echo "Added $file to apache2 sites-enabled..."
                a2ensite "$file"
            fi
        done
    fi
}

# (OLD_IP Has value only on re-run)
replace_ip_in_sites

# to prevent users from editing the files..
chmod 700 /etc/apache2/sites-available
chmod 700 /etc/apache2/sites-enabled


sites_available_dir="/etc/apache2/sites-available"

if [ "$(grep 'OLD_IP="tst"' /etc/entrypoint.sh)" ]; then
    echo "Apache is not started, since there are no websites yet."
else
  # if there are any sites, start the service
  if [ "$(ls -A $sites_available_dir | grep -v 'default')" ]; then
    service apache2 start
    echo "Apache service started."
  else
    echo "No websites found in $sites_available_dir. Apache service not started automatically."
  fi
fi



# Save the current IP for reuse
sed -i "s/$OLD_IP/$CONTAINER_IP/g" "$0"
