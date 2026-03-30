#!/bin/bash

echo 
echo "Adding patch for FTP service"
yes yes | opencli patch 896

wget -O /etc/openpanel/openpanel/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/service/service.config.py

ln -s /usr/local/mail/openmail/mailserver.env /usr/local/mail/openmail/.env

echo
echo "Setting local screenshots.."
sed -i 's/^\(screenshots=\).*$/\1local/' /etc/openpanel/openpanel/conf/openpanel.config

docker restart openpanel
