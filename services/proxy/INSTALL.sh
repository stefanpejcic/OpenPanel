#!/bin/bash

# FILES


# CADDY
mv Caddyfile /etc/caddy/Caddyfile
service caddy restart

# PHP

chown caddy:caddy  /run/php-fpm/www.sock
service php-fpm restart

