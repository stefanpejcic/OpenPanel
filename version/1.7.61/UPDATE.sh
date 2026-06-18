#!/bin/bash

# option to enforce 2fa for users
CONFIG="/etc/openpanel/openpanel/conf/openpanel.config"
if ! grep -q '^twofa_enforce=' "$CONFIG"; then
    sed -i '/^\[USERS\]/a twofa_enforce=no' "$CONFIG"
fi

# remove phpmyadmin per-user images and files
for env_file in /home/*/docker-compose.yml; do
    user_dir="$(dirname "$env_file")"
    rm -rf /home/$user_dir/pma.php
    timeout 5 docker --context="$user_dir" image rm phpmyadmin:latest > /dev/null 2>&1
done

# remove files created in 1.7.58 update
if [ -f "/etc/openpanel/docker/compose/1.0/docker-compose_backup_before_1.7.58_update.yml" ]; then  
  mv /etc/openpanel/docker/compose/1.0/docker-compose_backup_before_1.7.58_update.yml /tmp/docker-compose_backup_before_1.7.58_update.yml
fi

if [ -f "/usr/local/mail/openmail/docker-data/dms/config/postfix-accounts_backup_before_1.7.58_update.cf" ]; then  
  mv /usr/local/mail/openmail/docker-data/dms/config/postfix-accounts_backup_before_1.7.58_update.cf /tmp/postfix-accounts_backup_before_1.7.58_update.cf
fi

# add 2FA module for OpenAdmin
source /usr/local/admin/venv/bin/activate
pip install -r /usr/local/admin/requirements.txt
deactivate
