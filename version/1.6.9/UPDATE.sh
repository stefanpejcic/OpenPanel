#!/bin/bash

echo
echo "Updating OpenPanel configuration to show errors regardless of dev_mode setting"
wget -O /etc/openpanel/openpanel/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/service/service.config.py

echo
echo "Updating WP-CLI"
wget -O /etc/openpanel/wordpress/wp-cli.phar https://raw.githubusercontent.com/wp-cli/builds/gh-pages/phar/wp-cli.phar

chmod +x /etc/openpanel/wordpress/wp-cli.phar
