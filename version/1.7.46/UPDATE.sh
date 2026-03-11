#!/bin/bash

echo "Adding patch to sites table to prvenet duplicate site_name"
timeout 3 mysql -e "ALTER TABLE sites ADD UNIQUE KEY uniq_site_name (site_name);"

echo "Applying patch for WAF error starting Caddy.."
sed -i '/^SecRequestBodyJsonDepthLimit/ s/^/#/' /etc/openpanel/caddy/coraza_rules.conf
docker --context=default restart caddy


echo "Updating MOTD for admins to support custom OpenAdmin port.."
wget -O /etc/openpanel/ssh/admin_welcome.sh https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/ssh/admin_welcome.sh


echo "Adding support for Webhooks on OpenAdmin Notifications.."
file="/etc/openpanel/openadmin/config/notifications.ini"

if ! grep -q "webhook_url" "$file"; then
    echo "webhook_url=" >> "$file"
    echo "webhook_url option appended to $file."
else
    echo "webhook_url option already exists in $file."
fi

echo "Adding support for user-actions on OpenAdmin Notifications.."

if ! grep -q "^\[ACTIONS\]" "$file"; then
    cat <<EOL >> "$file"

[ACTIONS]
admin_status=yes
admin_api=yes
admin_create=yes
reseller_create=no
admin_password=yes
admin_rename=yes
admin_suspend=yes
admin_unsuspend=no

waf_domain=no
waf_status=yes

user_status=yes
user_create=yes
user_delete=yes
user_email=no
user_ip=no
user_password=yes
user_rename=yes

ftp_create=no
ftp_delete=no
ftp_password=no

domains_add=no
domains_delete=no
domains_status=no
domains_ssl=no
domains_hsts=no

EOL
    echo "Appended [ACTIONS] block to $file."
else
    echo "[ACTIONS] block already exists in $file."
fi


echo "Patching bug with rsyslog.."
rm -rf /etc/logrotate.d/syslog
service logrotate restart
opencli server-logrotate
