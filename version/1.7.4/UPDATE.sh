#!/bin/bash

# https://github.com/stefanpejcic/OpenPanel/issues/806
sed -i 's/^reseller=no$/reseller=yes/' /etc/openpanel/openadmin/config/admin.ini


sudo curl -sSL -o /etc/openpanel/caddy/check.conf https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/caddy/check.conf
