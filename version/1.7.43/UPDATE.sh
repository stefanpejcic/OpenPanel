#!/bin/bash

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
