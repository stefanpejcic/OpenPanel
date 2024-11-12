#!/bin/bash

# FILES
mv html /var/www/html 
cd /var/www/html 

# CADDY
mv caddy/fullchain.pem  /etc/caddy/certs/fullchain.pem 
mv caddy/privkey.pem /etc/caddy/certs/privkey.pem
mv caddy/Caddyfile /etc/caddy/Caddyfile


service caddy restart

# PHP
chown caddy:caddy  /run/php-fpm/www.sock
sed -i 's/^user = .*/user = caddy/' /etc/php-fpm.d/www.conf
sed -i 's/^group = .*/group = caddy/' /etc/php-fpm.d/www.conf
service php-fpm restart


# CRON
mv delete_cron.sh /var/www/delete_cron.sh
cron_job="*/5 * * * * bash /var/www/delete_cron.sh"
(crontab -l | grep -q "$cron_job") || (crontab -l; echo "$cron_job") | crontab -

