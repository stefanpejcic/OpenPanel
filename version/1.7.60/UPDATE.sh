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





: '
#TODO:

https://github.com/stefanpejcic/openpanel-configuration/commit/cae9e63f4615993752e2755edac0fcb7b370a3ee

https://github.com/stefanpejcic/openpanel-configuration/commit/dd0378f1f72b11f499214b5ad6ceb787073b47e5



'






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
