#!/bin/bash
################################################################################
# Script Name: imunify.sh
# Description: Install and manage ImunifyAV service.
# Usage: opencli imunify [status|start|stop|install|update|uninstall]
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 04.08.2025
# Last Modified: 14.11.2025
# Company: openpanel.com
# Copyright (c) openpanel.com
# 
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
# 
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
# 
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.
################################################################################

readonly SERVICE_NAME="ImunifyAV"



update_version() {
  local PANEL_INFO_SH="/etc/sysconfig/imunify360/get-panel-info.sh"

  cat <<'EOF' > "$PANEL_INFO_SH"
#!/usr/bin/env bash
opencli_version="$(opencli version 2>/dev/null || echo unknown)"
cat <<JSON
{
  "data": {
    "name": "OpenPanel",
    "version": "$opencli_version"
  },
  "metadata": {
    "result": "ok"
  }
}
JSON
EOF

  chmod 0755 "$PANEL_INFO_SH"
  chmod +x "$PANEL_INFO_SH"
}

status_av() {
  if pgrep -u _imunify -f "php -S 127.0.0.1:9000" >/dev/null; then
    echo "Imunify GUI is running."
    ps -u _imunify -f | grep "php -S 127.0.0.1:9000"
  else
    echo "Imunify GUI is not running."
  fi
}




configure_av_limits_and_email() {

echo "Configuring ImunifyAV notifications to use 'OpenAdmin > Settings > Notifications'..."
wget --timeout=5 --tries=3 --inet4-only -O /etc/sysconfig/imunify360/iav_hook.sh https://gist.githubusercontent.com/stefanpejcic/2318eae67c6833bb313eae7476aaa22f/raw/04bb0b6b4af7ff4515d17abaed891c50ff4f36d4/imav_email.sh
chmod +x /etc/sysconfig/imunify360/iav_hook.sh
imunify-antivirus notifications-config update '{"rules": {"USER_SCAN_MALWARE_FOUND": {"SCRIPT": {"scripts": ["/etc/sysconfig/imunify360/iav_hook.sh"], "enabled": true}}}}'
imunify-antivirus notifications-config update '{"rules": {"CUSTOM_SCAN_MALWARE_FOUND": {"SCRIPT": {"scripts": ["/etc/sysconfig/imunify360/iav_hook.sh"], "enabled": true}}}}'


# https://docs.imunifyav.com/config_file_description/
imunify-antivirus config update '{"MALWARE_SCANNING": {"hyperscan": true}}'

# ionice
echo "Setting 2CPU and 1GB Memory limits for ImunifyAV service.."
imunify-antivirus config update '{"MALWARE_SCAN_INTENSITY": {"cpu": 2}}'
imunify-antivirus config update '{"MALWARE_SCAN_INTENSITY": {"io": 2}}'
imunify-antivirus config update '{"MALWARE_SCAN_INTENSITY": {"ram": 1024}}'
imunify-antivirus config update '{"MALWARE_SCAN_INTENSITY": {"user_scan_cpu": 2}}'
imunify-antivirus config update '{"MALWARE_SCAN_INTENSITY": {"user_scan_io": 2}}'
imunify-antivirus config update '{"MALWARE_SCAN_INTENSITY": {"user_scan_ram": 1024}}'

imunify-antivirus config update '{"RESOURCE_MANAGEMENT": {"cpu_limit": 1}}'
imunify-antivirus config update '{"RESOURCE_MANAGEMENT": {"io_limit": 1}}'
imunify-antivirus config update '{"RESOURCE_MANAGEMENT": {"ram_limit": 500}}'

# cron
#imunify-antivirus config update '{"MALWARE_SCAN_SCHEDULE": {"day_of_month": 1}}'
#imunify-antivirus config update '{"MALWARE_SCAN_SCHEDULE": {"hour": 3}}'
#imunify-antivirus config update '{"MALWARE_SCAN_SCHEDULE": {"interval": "none"}}'
#imunify-antivirus config update '{"PERMISSIONS": {"allow_malware_scan": true}}'


# exclude paths!
echo "Adding Containers, Images and writable filesystems to ignored files list.."

cat <<\EOT >> /etc/sysconfig/imunify360/malware-filters-admin-conf/ignored.txt
^/home/(.*)/docker-data/containers/(.*)
^/home/(.*)/docker-data/image/(.*)
^/home/(.*)/docker-data/overlay2/(.*)
^/home/(.*)/bin/(.*)
EOT


imunify360-agent malware rebuild patterns
}


install_av() {

echo "Creating directories..."
mkdir -p /etc/sysconfig/imunify360/

PAM_DENY_FILE="/etc/pam.d/imunify360-deny"
if [ ! -f "$PAM_DENY_FILE" ]; then
  echo "Creating pam_deny.so file..."
  cat <<EOF > "$PAM_DENY_FILE"
auth required pam_deny.so
account required pam_deny.so
EOF
else
  echo "$PAM_DENY_FILE already exists, skipping..."
fi

INTEGRATION_CONF="/etc/sysconfig/imunify360/integration.conf"
if [ ! -f "$INTEGRATION_CONF" ]; then
  echo "Creating integration.conf file..."
  cat <<EOF > "$INTEGRATION_CONF"
[paths]
ui_path = /etc/sysconfig/imunify360/imav
ui_path_owner = _imunify:_imunify 

[pam]
service_name = imunify360-deny

[integration_scripts]
panel_info = /etc/sysconfig/imunify360/get-panel-info.sh
users = /usr/local/opencli/user/list.sh --json
domains = /usr/local/opencli/domains/all.sh --json

[malware]
basedir = /home
pattern_to_watch = ^/home/[^/]+/docker-data/volumes/[^/]+_html_data/_data(/.*)?$
EOF
else
  echo "$INTEGRATION_CONF already exists, skipping..."
fi

PANEL_INFO_JSON="/etc/sysconfig/imunify360/get-panel-info.json"
update_version

DEPLOY_SCRIPT="imav-deploy.sh"
if [ ! -f "$DEPLOY_SCRIPT" ]; then
  echo "Downloading deploy script..."
  wget --timeout=5 --tries=3 --inet4-only https://repo.imunify360.cloudlinux.com/defence360/imav-deploy.sh -O "$DEPLOY_SCRIPT"
else
  echo "$DEPLOY_SCRIPT already downloaded, skipping..."
fi

if ! grep -q "# Deployed by imav-deploy" "$DEPLOY_SCRIPT"; then
  echo "Running deploy script..."
  if ! bash "$DEPLOY_SCRIPT"; then
    echo
    echo "[ERROR] Installing ImunifyAV failed - please check the above output." >&2
    echo
    exit 1
  fi
else
  echo "Deploy script already executed or invalid, skipping..."
fi



configure_av_limits_and_email

echo "Allowing users to initiate a scan.."
imunify360-agent config update '{"PERMISSIONS": {"allow_malware_scan": true}}'
imunify-antivirus config update '{"PERMISSIONS": {"allow_malware_scan": true}}'


echo "Installing PHP if not present..."
if ! command -v php >/dev/null 2>&1; then
  if command -v apt-get >/dev/null 2>&1; then
    apt-get update
    apt-get install -y php-cli
  elif command -v yum >/dev/null 2>&1; then
    yum update
    yum install -y php-cli
  else
      echo "Unsupported package manager. Only apt-get and yum are supported."
      exit 1
  fi
else
  echo "PHP already installed."
fi








# add it to openadmin > services > status
FILE="/etc/openpanel/openadmin/config/services.json"

NEW_SERVICE='    {
        "name": "ImunifyAV",
        "type": "system",
        "real_name": "imunify-antivirus"
    }'

# Check if file ends with a closing array bracket
if tail -n 1 "$FILE" | grep -q '\]'; then
    # Remove the last line (closing bracket)
    head -n -1 "$FILE" > "$FILE.tmp"
    
    # Add comma to the last existing object if needed
    if tail -n 1 "$FILE.tmp" | grep -vq '},'; then
        sed -i '$s/}/},/' "$FILE.tmp"
    fi

    # Append the new service and closing bracket
    echo "$NEW_SERVICE" >> "$FILE.tmp"
    echo "]" >> "$FILE.tmp"

    # Replace the original file
    mv "$FILE.tmp" "$FILE"

    echo "New service added successfully."
else
    echo "Invalid JSON format in $FILE"
fi



echo "Install completed!"
}



uninstall_av() {
  echo "Removing files and directories..."
  rm -rf /etc/sysconfig/imunify360/
  rm -f /etc/pam.d/imunify360-deny
  rm -f imav-deploy.sh
  rm -f /var/log/imunify-php-server.log

  echo "Checking for _imunify user..."
  if id "_imunify" >/dev/null 2>&1; then
    echo "User '_imunify' exists. Not removing automatically to avoid breaking dependencies."
    echo "→ If you're sure, run: userdel _imunify"
  fi

  echo "Removing ImunifyAV from OpenAdmin > Services Status..."
  FILE="/etc/openpanel/openadmin/config/services.json"
  TMP_FILE="services.tmp.json"
  if jq 'map(select(.name != "ImunifyAV"))' "$FILE" > "$TMP_FILE"; then
      mv "$TMP_FILE" "$FILE"
      echo "Service 'ImunifyAV' removed successfully."
  else
      echo "Failed to parse or update $FILE. Check JSON format."
      rm -f "$TMP_FILE"
  fi

  echo "Uninstall complete."
}


start_av() {
  pkill -u _imunify -f "php -S 127.0.0.1:9000" 2>/dev/null || true

  chown -R _imunify /etc/sysconfig/imunify360/
  if ! pgrep -f "php -S 127.0.0.1:9000" >/dev/null; then
    nohup sudo -u _imunify php -S 127.0.0.1:9000 -t /etc/sysconfig/imunify360/ > /var/log/imunify-php-server.log 2>&1 &
    echo "ImunifyAV GUI started."
  else
    echo "ImunifyAV GUI already running."
  fi
}

stop_av() {
  pkill -f "php -S 127.0.0.1:9000"
  if ! pgrep -f "php -S 127.0.0.1:9000" >/dev/null; then
    echo "ImunifyAV GUI disabled."
  else
    echo "Failed to disable ImunifyAV GUI."
  fi
}



update_av() {
  # https://docs.imunify360.com/imunifyav/#update-instructions
  echo "Updating ImunifyAV..."
  if command -v apt-get >/dev/null 2>&1; then
      apt-get update
      apt-get install --only-upgrade -y imunify-antivirus
  elif command -v yum >/dev/null 2>&1; then
      yum update -y imunify-antivirus
  else
      echo "Unsupported package manager. Only apt-get and yum are supported."
      exit 1
  fi
}

# MAIN
case "$1" in
    status)
        status_av
        exit 0
        ;;
    install)
        echo "Installing $SERVICE_NAME..."
        install_av
        exit 0
        ;;
    uninstall)
        echo "Uninstalling $SERVICE_NAME..."
        stop_av
        uninstall_av
        exit 0
        ;;
    start)
        echo "Starting $SERVICE_NAME..."
        update_version
        start_av
        exit 0
        ;;
    stop)
        echo "Stopping $SERVICE_NAME..."
        stop_av
        exit 0
        ;;
    update)
        update_av
        exit 0
        ;;
    *)
        echo "Usage: opencli imunify {install|update|uninstall|start|stop}"
        exit 1
        ;;
esac
exit 0
