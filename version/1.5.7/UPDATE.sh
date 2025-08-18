#!/bin/bash

echo "Adding Mautic, SSL and Docroot modules.."
wget -O /etc/openpanel/openadmin/config/features.json  https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/config/features.json



echo "Fix for crons page..
"
FILE="/etc/cron.d/openpanel"
sed -i 's|root /bin/bash |root |' "$FILE"
sed -i 's|/usr/local/opencli|/usr/local/bin/opencli|g' "$FILE"
