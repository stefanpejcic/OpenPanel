#!/bin/bash

: '
CONFIG_FILE="/etc/openpanel/openadmin/config/admin.ini"

BLOCK="[EMAIL]
# available options: 
# user_dir   - /home/USERNAME/mail/DOMAIN/)
# /PARTITION - any partition, like /var/mail/ /email /storage etc.
email_storage_location=/var/mail/"

if grep -Fq "email_storage_location" "$CONFIG_FILE"; then
    echo "Option already exists in $CONFIG_FILE. Skipping."
else
    echo -e "\n$BLOCK" >> "$CONFIG_FILE"
    echo "Block added to $CONFIG_FILE."
fi
'


mkdir -p /etc/openpanel/wordpress/htaccess/
wget -O /etc/openpanel/wordpress/htaccess/apache.htaccess https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/wordpress/htaccess/apache.htaccess

# mailserver

ln -s /usr/local/mail/openmail/mailserver.env /usr/local/mail/openmail/.env
