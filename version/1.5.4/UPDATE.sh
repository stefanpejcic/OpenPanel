#!/bin/bash

# https://github.com/shinsenter/php/discussions/301
sed -i '/command: >/{N;s/command: >\n *sh -c "php-fpm --allow-to-run-as-root"/command: php-fpm --allow-to-run-as-root/}' /home/*/docker-compose.yml

sed -i '/command: >/{N;s/command: >\n *sh -c "php-fpm --allow-to-run-as-root"/command: php-fpm --allow-to-run-as-root/}' /etc/openpanel/docker/compose/1.0/docker-compose.yml
