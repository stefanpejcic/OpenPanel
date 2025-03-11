#!/bin/bash

# replace path in template
sed -i 's|/opt/bitnami/php/etc/conf.d|/usr/local/etc/php/|g' /etc/openpanel/docker/compose/1.0/docker-compose.yml

# replace for existing users 
sed -i 's|/opt/bitnami/php/etc/conf.d|/usr/local/etc/php/|g' /home/*/docker-compose.yml
