#!/bin/bash

# env
sed -i 's/^# LITESPEED (tag is <release>-lsphpXX, the php version comes with the image)$/# LITESPEED/' /etc/openpanel/docker/compose/1.0/.env /home/*/.env 2>/dev/null

#Update account.sh
cd /usr/local/mail/openmail/openpanel && git pull
