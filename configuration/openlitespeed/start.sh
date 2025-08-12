#!/bin/bash
set -e

CONF_DIR=/usr/local/lsws/conf
VHOST_DIR=$CONF_DIR/vhosts
OUTPUT_CONF=$CONF_DIR/httpd_config.conf

echo "Generating OpenLiteSpeed main config..."

# Start server block and listen section
cat > $OUTPUT_CONF <<EOL
server {
  listen 80 {
    address *:80
    secure 0
  }

  listener Default {
    address *:80
    secure 0
EOL

# Loop through vhost conf files to add domain -> vhost mappings
for f in $VHOST_DIR/*.conf; do
  # Extract domain line (strip spaces, commas)
  domain_line=$(grep -oP '(?<=domain\s)[^\n]+' "$f" | tr -d ' ')
  # Extract virtual host name from <virtualHost name>
  vhname=$(grep -oP '(?<=<virtualHost\s)[^>]+(?=>)' "$f")

  if [[ -n "$domain_line" && -n "$vhname" ]]; then
    # Split domains by comma and map each
    IFS=',' read -ra domains <<< "$domain_line"
    for d in "${domains[@]}"; do
      echo "    map $d $vhname" >> $OUTPUT_CONF
    done
  fi
done

# Close listener block
echo "  }" >> $OUTPUT_CONF

# Add errorLog and accessLog sections
cat >> $OUTPUT_CONF <<EOL

  errorLog {
    file \$SERVER_ROOT/logs/error.log
    rollingSize 10M
  }

  accessLog {
    file \$SERVER_ROOT/logs/access.log
    rollingSize 10M
  }

EOL

# Append all virtual host configs into main config
for f in $VHOST_DIR/*.conf; do
  echo "" >> $OUTPUT_CONF
  cat "$f" >> $OUTPUT_CONF
  echo "" >> $OUTPUT_CONF
done

# Close server block
echo "}" >> $OUTPUT_CONF

echo "Generated $OUTPUT_CONF successfully."

echo "Starting OpenLiteSpeed..."
exec /usr/local/lsws/bin/lswsctrl start && tail -f /usr/local/lsws/logs/error.log
