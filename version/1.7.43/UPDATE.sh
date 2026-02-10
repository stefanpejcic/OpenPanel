#!/bin/bash


echo
echo "Adding fix for #860"
wget -O /etc/openpanel/openadmin/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/service/service.config.py


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
