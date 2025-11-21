#!/bin/badh

# https://sentinelfirewall.org/
echo
echo "Upgrading ConfigServer Firewall & Security to Sentinel Firewall"
echo "https://sentinelfirewall.org/"
bash <(curl -sSL https://raw.githubusercontent.com/sentinelfirewall/sentinel/refs/heads/main/public/upgrade.sh)

echo
echo "Downloading WordPress .htaccess templates for Apache and Litespeed webservers.."
mkdir -p /etc/openpanel/wordpress/htaccess/
wget -O /etc/openpanel/wordpress/htaccess/litespeed.htaccess https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/wordpress/htaccess/litespeed.htaccess
wget -O /etc/openpanel/wordpress/htaccess/apache.htaccess https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/wordpress/htaccess/apache.htaccess
