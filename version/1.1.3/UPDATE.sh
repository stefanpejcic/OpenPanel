#!/bin/bash

#  force templates!
wget -O /etc/openpanel/docker/compose/1.0/docker-compose.yml https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/docker/compose/1.0/docker-compose.yml
wget -O /etc/openpanel/docker/compose/1.0/.env https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/docker/compose/1.0/.env

# backup first!
find /home/* -maxdepth 1 -name docker-compose.yml -exec cp {} {}.1.1.2_backup_before_update \;
find /home/* -maxdepth 1 -name .env -exec cp {} {}.1.1.2_backup_before_update \;

# then replace for existing users 
sed -i 's|/opt/bitnami/php/etc/conf.d|/usr/local/etc/php/|g' /home/*/docker-compose.yml
