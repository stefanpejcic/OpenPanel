#!/bin/bash

wget -O /etc/openpanel/ssh/admin_welcome.sh https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/ssh/admin_welcome.sh && chmod +x /etc/profile.d/welcome.sh


wget -O /etc/openpanel/wordpress/wp-cli.phar https://github.com/wp-cli/wp-cli/releases/download/v2.12.0/wp-cli-2.12.0.phar && chmod +x /etc/openpanel/wordpress/wp-cli.phar
