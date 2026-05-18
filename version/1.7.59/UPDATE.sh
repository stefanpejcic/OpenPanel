#!/bin/bash

# fixes bug with autologin
cd /usr/local/mail/openmail
docker --context=default compose down roundcube
rm -rf /usr/local/mail/openmail/roundcube-plugins/autologin.php
wget -O "/usr/local/mail/openmail/roundcube-plugins/autologin.php" https://raw.githubusercontent.com/stefanpejcic/OpenMail/refs/heads/main/roundcube-plugins/autologin.php
docker --context=default compose up -d roundcube



# created in 1.7.58 update
if [ -f "/etc/openpanel/docker/compose/1.0/docker-compose_backup_before_1.7.58_update.yml" ]; then  
  mv /etc/openpanel/docker/compose/1.0/docker-compose_backup_before_1.7.58_update.yml /tmp/docker-compose_backup_before_1.7.58_update.yml
fi

if [ -f "/usr/local/mail/openmail/docker-data/dms/config/postfix-accounts_backup_before_1.7.58_update.cf" ]; then  
  mv /usr/local/mail/openmail/docker-data/dms/config/postfix-accounts_backup_before_1.7.58_update.cf /tmp/postfix-accounts_backup_before_1.7.58_update.cf
fi


