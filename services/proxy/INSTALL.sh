#!/bin/bash

# FILES


# CADDY
mv Caddyfile /etc/caddy/Caddyfile
service caddy restart

# PHP
chown caddy:caddy  /run/php-fpm/www.sock
sed -i 's/^user = .*/user = caddy/' /etc/php-fpm.d/www.conf
sed -i 's/^group = .*/group = caddy/' /etc/php-fpm.d/www.conf
service php-fpm restart

