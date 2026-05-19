#!/bin/bash

[ -z "$1" ] && { echo "Error: pass license key to the script!"; exit 1; }

# defaults
LICENSE_KEY="$1"
PANEL_USERNAME="testinguser"
PANEL_PASSWORD="testingpassword"
PANEL_EMAIL="test@test.com"

# INCREASE LIMITS SO TESTS DONT GET BLOCKED
opencli config update login_ratelimit 100

# ENABLE ALL FEATURES
for f in basic.txt default.txt; do
  wget -qO- https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/config/features.json | jq -r '.[].name' > "/etc/openpanel/openpanel/features/$f"
done

# ENABLE ALL MODULES
sed -i "s/^enabled_modules=.*/enabled_modules=$(paste -sd, /etc/openpanel/openpanel/features/default.txt)/" /etc/openpanel/openpanel/conf/openpanel.config

# INSTALL ALL LOCALES
opencli locale $(curl -s "https://api.github.com/repos/stefanpejcic/openpanel-translations/contents" | jq -r '.[] | select(.type=="dir") | .name' | tr '\n' ' ')

# ADD LICENSE LICENSE
opencli license $LICENSE_KEY

# ENABLE EMAILS
opencli email-server install

# ENABLE FTP
cd /root && docker --context=default compose up -d openadmin_ftp

# RESTART USER-PANEL TO APPLY ALL CHANGES!
docker --context=default restart openpanel

# CREATE USER FOR TESTS
opencli user-add "$PANEL_USERNAME" "$PANEL_PASSWORD" "$PANEL_EMAIL" 'Developer plus'
BASE_URL=$(opencli user-login testinguser | sed -E 's#(https?://[^/]+).*#\1#')

# PRINT INFO for /root/playwright-test/openpanel/.env
echo "BASE_URL=$BASE_URL"
echo "PANEL_USERNAME=$PANEL_USERNAME"
echo "PANEL_PASSWORD=$PANEL_PASSWORD"
