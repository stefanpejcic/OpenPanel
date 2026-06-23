#!/bin/bash

# defaults
LICENSE_KEY="enterprise-2a5da40ecd2f60" # ip restricted!
PANEL_USERNAME="testinguser"
PANEL_PASSWORD="testingpassword"
PANEL_EMAIL="test@test.com"

# INCREASE LIMITS SO TESTS DONT GET BLOCKED
opencli config update login_ratelimit 100
opencli plan-edit id=2 name="Developer plus" description="A professional plan" emails=500 max_email_quota=2G ftp=100 domains=10 websites=10 disk=50 inodes=1000000 databases=20 cpu=4 ram=6 bandwidth=500 max_hourly_email=6000
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

# CHANGE A RECORD ON CLOUDFLARE TO THIS SERVER (its ip restricted!)
IP=$(curl -s -4 https://ip.openpanel.com) && curl -s -o /dev/null -w "" -X PUT "https://api.cloudflare.com/client/v4/zones/576d997a15f8e381f18b8b39363b4023/dns_records/1a13b4f40e8751e8d848b4a49ba460ed" -H "Authorization: Bearer cfat_NmJr954LPMZYkJ9TTTQS2loWh9tIA2z1NvgpjyRPd4aa4c06" -H "Content-Type: application/json" --data "{\"type\":\"A\",\"name\":\"*.tests.openpanel.org\",\"content\":\"$IP\",\"ttl\":120,\"proxied\":true}" >/dev/null 2>&1

# PRINT INFO for /root/playwright-test/openpanel/.env
echo "BASE_URL=$BASE_URL"
echo "PANEL_USERNAME=$PANEL_USERNAME"
echo "PANEL_PASSWORD=$PANEL_PASSWORD"
