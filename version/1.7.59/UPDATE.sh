#!/bin/bash

# fixes bug with autologin
cd /usr/local/mail/openmail
docker --context=default compose down roundcube
rm -rf /usr/local/mail/openmail/roundcube-plugins/autologin.php
mkdir -p /usr/local/mail/openmail/roundcube-plugins/dovecot_impersonate/
wget -O "/usr/local/mail/openmail/roundcube-plugins/autologin.php" https://raw.githubusercontent.com/stefanpejcic/OpenMail/refs/heads/main/roundcube-plugins/autologin.php
wget -O "/usr/local/mail/openmail/roundcube-plugins/dovecot_impersonate/dovecot_impersonate.php" https://raw.githubusercontent.com/stefanpejcic/OpenMail/refs/heads/main/roundcube-plugins/dovecot_impersonate/dovecot_impersonate.php
docker --context=default compose up -d roundcube

# fixes bug with ftp connections
VSFTPD_FILE="/etc/openpanel/ftp/vsftpd.conf"
LINE="setproctitle_enable=yes|"
if [ -f "$VSFTPD_FILE" ]; then
  if ! grep -q "^setproctitle_enable=" "$VSFTPD_FILE"; then
    echo "$LINE" >> "$VSFTPD_FILE"
    timeout 3 docker restart openadmin_ftp >/dev/null 2>&1
  fi
fi
