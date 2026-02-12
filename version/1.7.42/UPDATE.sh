#!/bin/bash

echo
echo "Updating existing plans in database to add a new column for 'max_email_quota'.."

timeout 10 mysql panel <<'EOF'

SELECT COUNT(*) INTO @column_exists
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = 'panel'
  AND TABLE_NAME = 'plans'
  AND COLUMN_NAME = 'max_email_quota';

-- Add column if it does not exist
SET @add_sql = IF(@column_exists = 0,
    'ALTER TABLE plans ADD COLUMN max_email_quota TEXT NOT NULL',
    'SELECT "Column already exists"'
);

PREPARE stmt FROM @add_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Ensure all rows have value '0' if empty or NULL
UPDATE plans
SET max_email_quota = '0'
WHERE max_email_quota IS NULL OR max_email_quota = '';

EOF




echo
echo "Checking default template and .env files for users for missing G in _RAM keys.."
(
  find /home -maxdepth 2 -type f -name ".env" -path "/home/*/.env"
  echo "/etc/openpanel/docker/compose/1.0/.env"
) | sort -u | while read -r file; do
  [ -f "$file" ] || continue

  perl -i_1.7.42.bak -pe '
    next if /^\s*#/;
    next if /^\s*TOTAL_RAM\s*=/;

    if (/^(\s*)([A-Z0-9_]+_RAM)\s*=\s*"?([0-9]+(?:\.[0-9]+)?)"?\s*$/) {
      my ($indent,$key,$val) = ($1,$2,$3);

      next if $val =~ /^0(?:\.0+)?$/;

      $_ = sprintf("%s%s=\"%sG\"\n", $indent, $key, $val);
    }
  ' "$file"

done






echo
echo "Adding PageSpeed API key for WP Manager.."
file="/etc/openpanel/openpanel/service/pagespeed.api"
key="AIzaSyDow0GLE7N5gcZXa72tpqIvIaJtn1bDtsk"

touch "$file"
if ! grep -Fxq "$key" "$file"; then
    echo "$key" >> "$file"
fi



echo
echo "Updating openpanel.config file to add 'filemanager_buttons_style' setting.."
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


if ! grep -q "^dir=" "$CONFIG"; then
  awk '
  BEGIN{added=0}
  /^\[DEFAULT\]/ {print; in_files=1; next}
  in_files && /^\[/ {
    if (!added) {
      print "dir=ltr"
      added=1
    }
    in_files=0
  }
  {print}
  END{
    if (in_files && !added) {
      print "dir=ltr"
    }
  }
  ' "$CONFIG" > /tmp/openpanel.config.new

  mv /tmp/openpanel.config.new "$CONFIG"
  echo "Added dir=ltr to [DEFAULT]"
else
  echo "dir=ltr already exists — skipping"
fi

echo
echo "Updating openpanel.config file to add 'mysql_startup_time' 'mysql_import_max_size_gb' 'mysql_restricted_usernames' 'mysql_restricted_databases' settings.."
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


