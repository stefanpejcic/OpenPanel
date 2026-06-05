#!/bin/bash

# SHARED PHPMYADMIN
open_csf_port() {
    local type=$1 port=$2
    local conf="/etc/csf/csf.conf"
    for dir in "$type" "${type/4/6}"; do
        grep -q "${dir} = .*${port}" "$conf" || sed -i "s/${dir} = \"\(.*\)\"/${dir} = \"\1,${port}\"/" "$conf"
    done
}

open_csf_port TCP_IN 2053
open_csf_port TCP_IN 8888

CADDYFILE="/etc/openpanel/caddy/Caddyfile"
MARKER="# START PHPMYADMIN DOMAIN #"
INSERT_AFTER="# END WEBMAIL DOMAIN #"
BLOCK="\n# START PHPMYADMIN DOMAIN #\nexample.net:2053 {\n    reverse_proxy localhost:8888\n}\n# END PHPMYADMIN DOMAIN #"

if ! grep -qF "$MARKER" "$CADDYFILE"; then
    docker cp $CADDYFILE /tmp/Caddyfile_1.7.60
    sed -i "/$INSERT_AFTER/a $BLOCK" "$CADDYFILE"
    echo "phpMyAdmin domain block added, bekap is stored at /tmp/Caddyfile_1.7.60"
    docker --context=default restart caddy
fi

# TODO, update them!
curl -L https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/mysql/phpmyadmin/pma.php -o /etc/openpanel/mysql/phpmyadmin/pma.php
curl -L https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/mysql/phpmyadmin/config.inc.php -o /etc/openpanel/mysql/phpmyadmin/config.inc.php

for env_file in /home/*/docker-compose.yml; do
    user_dir="$(dirname "$env_file")"
    timeout 5 docker --context="$user_dir" compose down phpmyadmin > /dev/null 2>&1
done








# created in 1.7.58 update
if [ -f "/etc/openpanel/docker/compose/1.0/docker-compose_backup_before_1.7.58_update.yml" ]; then  
  mv /etc/openpanel/docker/compose/1.0/docker-compose_backup_before_1.7.58_update.yml /tmp/docker-compose_backup_before_1.7.58_update.yml
fi

if [ -f "/usr/local/mail/openmail/docker-data/dms/config/postfix-accounts_backup_before_1.7.58_update.cf" ]; then  
  mv /usr/local/mail/openmail/docker-data/dms/config/postfix-accounts_backup_before_1.7.58_update.cf /tmp/postfix-accounts_backup_before_1.7.58_update.cf
fi

# ENABLE GZIP
DIR="/etc/openpanel/caddy/domains"
mkdir -p /tmp/domain_backups_1.7.59

for file in "$DIR"/*.conf; do
  [ -f "$file" ] || continue

  cp "$file" /tmp/domain_backups_1.7.59/

  awk '
  function reset_block() {
    mode=""
    after_header=0
    encode_present=0
    inserted=0
  }

  BEGIN {
    reset_block()
  }

  # Detect HTTP block header
  /^http:\/\/.*\{$/ {
    print
    mode="http"
    after_header=1
    encode_present=0
    inserted=0
    next
  }

  # Detect HTTPS block header
  /^https:\/\/.*\{$/ {
    print
    mode="https"
    after_header=1
    encode_present=0
    inserted=0
    next
  }

  {
    print

    # detect existing encode line
    if ($0 ~ /encode zstd gzip/) {
      encode_present=1
    }

    # insert right after header if missing
    if (after_header && !inserted) {
      if (!encode_present) {
        print "  encode zstd gzip"
      }
      inserted=1
      after_header=0
    }

    # reset when block ends
    if ($0 ~ /^}/) {
      reset_block()
    }
  }
  ' "$file" > "$file.tmp" && mv "$file.tmp" "$file"

done
