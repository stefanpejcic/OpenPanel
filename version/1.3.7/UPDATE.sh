#!/bin/bash


mkdir -p /etc/openpanel/openpanel/features

wget --inet4-only -O /etc/openpanel/openpanel/features/default.txt https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/features/default.txt
wget --inet4-only -O /etc/openpanel/openpanel/features/basic.txt https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/features/basic.txt
wget --inet4-only -O /etc/openpanel/openpanel/features/mysql_only.txt https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/features/mysql_only.txt



DB_CONFIG_FILE="/usr/local/opencli/db.sh"
. "$DB_CONFIG_FILE"

COLUMN_EXISTS=$(mysql --defaults-extra-file=/etc/my.cnf -D panel -sN -e "SHOW COLUMNS FROM plans LIKE 'feature_set';")

if [ -z "$COLUMN_EXISTS" ]; then
  mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "ALTER TABLE plans ADD COLUMN feature_set VARCHAR(255) DEFAULT 'default';"
fi

mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "UPDATE plans SET feature_set = 'default';"
