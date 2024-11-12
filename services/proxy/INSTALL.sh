#!/bin/bash

# first replace preview.openpanel.org  with your domain where the api will be available
# then replace openpanel.org with the main domain where sub-domains will be created! example: my.com  to have links some123.my.com some124.my.com etc. 
# its best to use cf for the proxy, and enable force https there
# then generate ssl on it that covers both *.your.domain and your.domain 
# put certificate files in `caddy/fullchain.pem` and `caddy/privkey.pem` 
# then run this script on freshly installed OS RockyLinux 9.4

# SYS
dnf update -y
dnf install -y 'dnf-command(copr)'


# FILES
mv html /var/www/html
mv delete_cron.sh /var/www/delete_cron.sh
cd /var/www/html 

# CADDY
dnf copr enable @caddy/caddy
dnf install caddy -y
systemctl enable --now caddy

mv caddy/fullchain.pem  /etc/caddy/certs/fullchain.pem 
mv caddy/privkey.pem /etc/caddy/certs/privkey.pem
mv caddy/Caddyfile /etc/caddy/Caddyfile

service caddy restart
usermod -aG www-data caddy
usermod -aG caddy caddy

# PERMISSIONS
chown -R caddy:caddy /var/www/html/
chmod -R 755 /var/www/html/domains
chmod +x /var/www/delete_cron.sh

# PHP
dnf module reset php
dnf module enable php:8.1
dnf install php php-fpm php-cli php-common -y
systemctl start php-fpm
systemctl enable php-fpm

chown caddy:caddy  /run/php-fpm/www.sock
sed -i 's/^user = .*/user = caddy/' /etc/php-fpm.d/www.conf
sed -i 's/^group = .*/group = caddy/' /etc/php-fpm.d/www.conf
service php-fpm restart


# CRON

cron_job="*/5 * * * * bash /var/www/delete_cron.sh"
(crontab -l | grep -q "$cron_job") || (crontab -l; echo "$cron_job") | crontab -

