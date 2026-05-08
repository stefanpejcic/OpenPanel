#!/bin/bash

# update caddy check to issue www. SSL
wget -O /etc/openpanel/caddy/check.conf https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/caddy/check.conf

# fixes bug on user-add for 4-8.May
wget -O /etc/openpanel/docker/dockerd-rootless-setuptool.sh https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/docker/dockerd-rootless-setuptool.sh ; chmod +x /etc/openpanel/docker/dockerd-rootless-setuptool.sh

# add cron for updating du info
QUOTA_LINE='*/5 * * * * root /usr/local/bin/opencli user-quota && echo "$(date) Collected disk and inodes usage for all users" >> /var/log/openpanel/admin/cron.log'

if ! grep -q "user-quota" "/etc/cron.d/openpanel"; then
    sed -i "/^# STATISTICS/a ${QUOTA_LINE}" "/etc/cron.d/openpanel"
fi
