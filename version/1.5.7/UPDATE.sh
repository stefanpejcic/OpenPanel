#!/bin/bash

echo "Adding Mautic, SSL and Docroot modules.."
wget -O /etc/openpanel/openadmin/config/features.json  https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/config/features.json

echo "Fix for '/root/openpanel_restart_needed'"
wget -O /etc/openpanel/openpanel/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/service/service.config.py



echo "Fix for crons page bug in OpenAdmin.."
FILE="/etc/cron.d/openpanel"
sed -i 's|root /bin/bash |root |' "$FILE"
sed -i 's|/usr/local/opencli|/usr/local/bin/opencli|g' "$FILE"


echo "OpenLitespeed.."
mkdir -p /etc/openpanel/openlitespeed
wget -O /etc/openpanel/nginx/vhosts/1.1/docker_openlitespeed_domain.conf https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/nginx/vhosts/1.1/docker_openlitespeed_domain.conf
wget -O /etc/openpanel/openlitespeed/httpd_config.conf https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openlitespeed/httpd_config.conf
wget -O /etc/openpanel/openlitespeed/start.sh https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openlitespeed/start.sh
chmod +x /etc/openpanel/openlitespeed/start.sh
