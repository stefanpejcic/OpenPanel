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
domains=()
for vhfile in /usr/local/lsws/conf/vhosts/*.conf; do
  [ -e "$vhfile" ] || continue
  domain=$(basename "$vhfile" .conf)
  domains+=("$domain")

  if grep -q "virtualhost $domain {" "$HTTPD_CONF"; then
    echo "[✓] $domain already exists"
    continue
  fi

  echo "[!] Creating include section for domain: $domain"
  new_blocks+=$'virtualhost '"$domain"' {\n'
  new_blocks+=$'  vhRoot            /var/www/html/\n'
  new_blocks+=$'  configFile            /usr/local/lsws/conf/vhosts/'"$domain"'.conf\n'
  new_blocks+=$'  allowSymbolLink                    1\n'
  new_blocks+=$'  enableScript                    1\n'
  new_blocks+=$'  restrained                    1\n'
  new_blocks+=$'  setUIDMode                    2\n'
  new_blocks+=$'}\n\n'
done

# Insert new vhost blocks
if [ -n "$new_blocks" ]; then
  tmpfile=$(mktemp)
  awk -v block="$new_blocks" '
    /vhTemplate docker {/ { print block }
    { print }
  ' "$HTTPD_CONF" > "$tmpfile"
  cat "$tmpfile" > "$HTTPD_CONF"
  rm "$tmpfile"
fi

echo "[*] Updating listener mappings.."

update_listener_maps() {
  listener=$1
  tmpfile=$(mktemp)

  awk -v lst="$listener" -v domains="${domains[*]}" '
    BEGIN {
      split(domains, d_arr, " ")
      in_listener = 0
    }
    {
      if ($1 == "listener" && $2 == lst && $3 == "{") {
        in_listener = 1
        map_lines = ""
      }
      
      if (in_listener) {
        if ($1 == "map") {
          existing_map[$2] = 1
        }
        if ($1 == "}") {
          # Before closing brace, add missing maps
          for (i in d_arr) {
            if (!(d_arr[i] in existing_map)) {
              print "  map                     " d_arr[i] " " d_arr[i]
            }
          }
          in_listener = 0
        }
      }

      print
    }
  ' "$HTTPD_CONF" > "$tmpfile"

  cat "$tmpfile" > "$HTTPD_CONF"
  rm "$tmpfile"
}

update_listener_maps "HTTP"
update_listener_maps "HTTPS"

echo "[*] Starting LSWS process.."
/usr/local/lsws/bin/lswsctrl start
"$@"

while true; do
  if ! /usr/local/lsws/bin/lswsctrl status | grep 'litespeed is running with PID' > /dev/null; then
    break
  fi
  sleep 60
done
