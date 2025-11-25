#!/bin/badh

# https://sentinelfirewall.org/
#echo
#echo "Upgrading ConfigServer Firewall & Security to Sentinel Firewall"
#echo "https://sentinelfirewall.org/"
#bash <(curl -sSL https://raw.githubusercontent.com/sentinelfirewall/sentinel/refs/heads/main/public/upgrade.sh)

echo
echo "Downloading WordPress .htaccess templates for Apache and Litespeed webservers.."
mkdir -p /etc/openpanel/wordpress/htaccess/
wget -O /etc/openpanel/wordpress/htaccess/litespeed.htaccess https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/wordpress/htaccess/litespeed.htaccess
wget -O /etc/openpanel/wordpress/htaccess/apache.htaccess https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/wordpress/htaccess/apache.htaccess



echo
echo "Generating secret key for OpenAdmin.."
SECRET_FILE="/etc/openpanel/openadmin/secret.key"
if [ ! -f "$SECRET_FILE" ]; then
    openssl rand -hex 32 > "$SECRET_FILE"
    chmod 600 "$SECRET_FILE"
fi


echo
echo "Generating secret key for OpenPanel.."
SECRET_FILE="/etc/openpanel/openpanel/secret.key"
if [ ! -f "$SECRET_FILE" ]; then
    openssl rand -hex 32 > "$SECRET_FILE"
    chmod 600 "$SECRET_FILE"
fi


CONFIG_FILE="/etc/openpanel/openpanel/conf/openpanel.config"
if grep -q "^terminal_command_timeout" "$CONFIG_FILE"; then
    sed -i 's/^terminal_command_timeout/terminal_timeout/' "$CONFIG_FILE"
    echo "Replaced terminal_command_timeout with terminal_timeout."
elif grep -q "^terminal_timeout" "$CONFIG_FILE"; then
    :
else
    echo "Inserting terminal_timeout=10 into [PANEL]..."
    sed -i '/^\[PANEL\]/a terminal_timeout=10' "$CONFIG_FILE"
    echo "terminal_timeout=10 added successfully."
fi

