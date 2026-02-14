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
echo "Adding fix for #860"
wget -O /etc/openpanel/openadmin/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/service/service.config.py
wget -O /etc/openpanel/openpanel/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/service/service.config.py

# https://github.com/stefanpejcic/openpanel-configuration/commit/30f977d412d5ba0c4768a91a2c51177687f82d66
CONFIG="/etc/openpanel/openpanel/conf/openpanel.config"
cp -a "$CONFIG" /tmp/openpanel.config_1.7.43.bak

echo
echo "Updating openpanel.config file to add  [CAPTCHA] settings.."
if ! grep -q "^\[CAPTCHA\]" "$CONFIG"; then
  cat >> "$CONFIG" << 'EOF'

[CAPTCHA]
captcha_provider=google
recaptcha_site_key=
recaptcha_secret_key=
turnstile_site_key=
turnstile_secret_key=
custom_captcha_site_key=
EOF

  echo "Added [CAPTCHA] section"
fi

echo "Updating openpanel.config file to add 'cron_max_file_size_kb' setting.."
if ! grep -q "^cron_max_file_size_kb=" "$CONFIG"; then
  awk '
  BEGIN{added=0}
  /^\[FILES\]/ {print; in_files=1; next}
  in_files && /^\[/ {
    if (!added) {
      print "cron_max_file_size_kb=100"
      added=1
    }
    in_files=0
  }
  {print}
  END{
    if (in_files && !added) {
      print "cron_max_file_size_kb=100"
    }
  }
  ' "$CONFIG" > /tmp/openpanel.config.new

  mv /tmp/openpanel.config.new "$CONFIG"
  echo "Added cron_max_file_size_kb=100 to [FILES]"
fi
