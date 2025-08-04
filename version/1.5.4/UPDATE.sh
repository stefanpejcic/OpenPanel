#!/bin/bash

# https://github.com/shinsenter/php/discussions/301
sed -i '/command: >/{N;s/command: >\n *sh -c "php-fpm --allow-to-run-as-root"/command: php-fpm --allow-to-run-as-root/}' /home/*/docker-compose.yml
