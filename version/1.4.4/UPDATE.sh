#!/bin/bash


echo ""
echo "📥 Downloading files to allow auto-login to phpMyAdmin for users.."
mkdir -p /etc/openpanel/mysql/phpmyadmin/
wget -q -O "/etc/openpanel/mysql/phpmyadmin/pma.php" "https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/mysql/phpmyadmin/pma.php" && \
    echo "✔ downloaded /etc/openpanel/mysql/phpmyadmin/pma.php"
wget -q -O "/etc/openpanel/mysql/phpmyadmin/config.inc.php" "https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/mysql/phpmyadmin/config.inc.php" && \
    echo "✔ downloaded /etc/openpanel/mysql/phpmyadmin/config.inc.php"

