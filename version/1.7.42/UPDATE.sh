#!/bin/bash

: '

example file:
# PHP 8.4
PHP_FPM_8_4_RAM="0.25G"
PHP_FPM_8_4_CPU="0.125"
PHP_FPM_8_4_ENABLE_OPCACHE=true
PHP_FPM_8_4_PHP_MAX_EXECUTION_TIME="600


for file in /home/*/.env
    for key with _RAM= if value != 0 and has not g or G then add it

same for /etc/openpanel/docker/compose/1.0/.env


BUT EXCLUDE IN FILE< this key:
TOTAL_RAM=""


'












file="/etc/openpanel/openpanel/service/pagespeed.api"
key="AIzaSyDow0GLE7N5gcZXa72tpqIvIaJtn1bDtsk"

touch "$file"
if ! grep -Fxq "$key" "$file"; then
    echo "$key" >> "$file"
fi



CONFIG="/etc/openpanel/openpanel/conf/openpanel.config"

cp -a "$CONFIG" /tmp/openpanel.config_1.7.42.bak

if ! grep -q "^filemanager_buttons_style=" "$CONFIG"; then
  awk '
  BEGIN{added=0}
  /^\[FILES\]/ {print; in_files=1; next}
  in_files && /^\[/ {
    if (!added) {
      print "filemanager_buttons_style=classic"
      added=1
    }
    in_files=0
  }
  {print}
  END{
    if (in_files && !added) {
      print "filemanager_buttons_style=classic"
    }
  }
  ' "$CONFIG" > /tmp/openpanel.config.new

  mv /tmp/openpanel.config.new "$CONFIG"
  echo "Added filemanager_buttons_style=classic to [FILES]"
else
  echo "filemanager_buttons_style=classic already exists — skipping"
fi

if ! grep -q "^\[DATABASES\]" "$CONFIG"; then
  cat >> "$CONFIG" << 'EOF'

[DATABASES]
mysql_startup_time=10
mysql_import_max_size_gb=1
mysql_restricted_usernames="mysql.sys mysql sys mariadb.sys phpmyadmin mysql.session mysql.infoschema root debian-sys-maint healthcheck"
mysql_restricted_databases="information_schema performance_schema mysql phpmyadmin sys mariadb.sys"
EOF

  echo "Added [DATABASES] section"
else
  echo "[DATABASES] section already exists — skipping"
fi
