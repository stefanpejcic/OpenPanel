#!/bin/bash

mkdir -p /etc/openpanel/openpanel/features

wget --inet4-only -O /etc/openpanel/openpanel/features/default.txt https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/features/default.txt
wget --inet4-only -O /etc/openpanel/openadmin/config/features.json https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/config/features.json

wget --inet4-only -O /etc/openpanel/caddy/templates/domain.conf_with_modsec https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/caddy/templates/domain.conf_with_modsec
wget --inet4-only -O /etc/openpanel/caddy/templates/domain.conf https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/caddy/templates/domain.conf

wget --inet4-only -O /etc/openpanel/docker/compose/1.0/.env https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/docker/compose/1.0/.env
wget --inet4-only -O /etc/openpanel/docker/compose/1.0/docker-compose.yml https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/docker/compose/1.0/docker-compose.yml
