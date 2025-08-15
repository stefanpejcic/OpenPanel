#!/bin/bash
echo "[*] Initializing.."

# Ensure permissions
chown -R 994:994 /usr/local/lsws/conf
chown -R 994:1001 /usr/local/lsws/admin/conf

# https://github.com/litespeedtech/ols-dockerfiles/issues/13
usermod -aG nogroup root
usermod -aG root nobody

HTTPD_CONF="/usr/local/lsws/conf/httpd_config.conf"

echo "[*] Checking include sections for all VHosts files.."

# Collect all missing vhTemplate blocks
new_blocks=""
for vhfile in /usr/local/lsws/conf/vhosts/*.conf; do
  [ -e "$vhfile" ] || continue
  domain=$(basename "$vhfile" .conf)

  # Skip if already exists
  if grep -q "vhTemplate $domain {" "$HTTPD_CONF"; then
    echo "[✓] $domain already exists"
    continue
  fi
    "echo [!] Creating include section for domain: $domain"

  # Build block with real newlines
  new_blocks+=$'vhTemplate '"$domain"' {\n'
  new_blocks+=$'  templateFile            conf/vhosts/'"$domain"'.conf\n'
  new_blocks+=$'  listeners               HTTP, HTTPS\n'
  new_blocks+=$'  note                    '"$domain"$'\n'
  new_blocks+=$'\n'
  new_blocks+=$'  member localhost {\n'
  new_blocks+=$'    vhDomain              '"$domain"$'\n'
  new_blocks+=$'  }\n'
  new_blocks+=$'}\n\n'
done

# Insert before 'vhTemplate docker {' without rename()
if [ -n "$new_blocks" ]; then
  tmpfile=$(mktemp)
  awk -v block="$new_blocks" '
    /vhTemplate docker {/ { print block }
    { print }
  ' "$HTTPD_CONF" > "$tmpfile"
  cat "$tmpfile" > "$HTTPD_CONF"
  rm "$tmpfile"
fi

echo "[*] Starting LSWS process.."

# Start the server
/usr/local/lsws/bin/lswsctrl start
"$@"

# Keep container running and monitor
while true; do
  if ! /usr/local/lsws/bin/lswsctrl status | grep 'litespeed is running with PID' > /dev/null; then
    break
  fi
  sleep 60
done
