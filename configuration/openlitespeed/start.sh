#!/bin/bash

if [ -d "/vhosts" ] && [ "$(ls -A /vhosts)" ]; then
  cp -R /vhosts/* /usr/local/lsws/conf/vhosts/
fi

# Ensure permissions
chown -R 994:994 /usr/local/lsws/conf
chown -R 994:1001 /usr/local/lsws/admin/conf

# Ensure main config includes vhosts
grep -q "include conf/vhosts/*.conf" /usr/local/lsws/conf/httpd_config.conf || \
  echo "include conf/vhosts/*.conf" >> /usr/local/lsws/conf/httpd_config.conf

# Start the server
/usr/local/lsws/bin/lswsctrl start
$@

# Keep container running and monitor
while true; do
  if ! /usr/local/lsws/bin/lswsctrl status | /usr/bin/grep 'litespeed is running with PID *' > /dev/null; then
    break
  fi
  sleep 60
done
