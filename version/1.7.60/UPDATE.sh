#!/bin/bash

# SHARED PHPMYADMIN


grep -q "phpmyadmin" /root/docker-compose.yml || sed -i '/^volumes:/i\
\
  # PHPMYADMIN\
  phpmyadmin:\
    container_name: phpmyadmin\
    restart: unless-stopped\
    image: phpmyadmin:${PHPMYADMIN_VERSION:-latest}\
    volumes:\
      - /etc/openpanel/mysql/phpmyadmin/pma.php:/var/www/html/pma.php\
      - /etc/openpanel/mysql/phpmyadmin/config.inc.php:/etc/phpmyadmin/config.inc.php:ro\
      - /home:/home:ro\
    ports:\
      - "8888:80"\
    networks:\
      - openpanel_network\
    deploy:\
      resources:\
        limits:\
          cpus: "${PHPMYADMIN_CPU:-1.0}"\
          memory: "${PHPMYADMIN_RAM:-1.0G}"\
          pids: 5000\
    environment:\
      MAX_EXECUTION_TIME: ${PHPMYADMIN_MAX_EXECUTION_TIME:-600}\
      MEMORY_LIMIT: ${PHPMYADMIN_MEMORY_LIMIT:-512M}\
      UPLOAD_LIMIT: ${PHPMYADMIN_UPLOAD_LIMIT:-256M}\
' /root/docker-compose.yml


grep -q "PHPMYADMIN" /root/.env || cat >> /root/.env << 'EOF'

# phpmyadmin
PHPMYADMIN_VERSION="latest"
PHPMYADMIN_CPU="1.0"
PHPMYADMIN_RAM="1.0G"
PHPMYADMIN_MAX_EXECUTION_TIME="900"
PHPMYADMIN_MEMORY_LIMIT="2048M"
PHPMYADMIN_UPLOAD_LIMIT="1024"
EOF

open_csf_port() {
    local type=$1 port=$2
    local conf="/etc/csf/csf.conf"
    for dir in "$type" "${type/4/6}"; do
        grep -q "${dir} = .*${port}" "$conf" || sed -i "s/${dir} = \"\(.*\)\"/${dir} = \"\1,${port}\"/" "$conf"
    done
}

open_csf_port TCP_IN 2053
open_csf_port TCP_IN 8888

CADDYFILE="/etc/openpanel/caddy/Caddyfile"
MARKER="# START PHPMYADMIN DOMAIN #"
INSERT_AFTER="# END WEBMAIL DOMAIN #"
BLOCK="\n# START PHPMYADMIN DOMAIN #\nexample.net:2053 {\n    reverse_proxy localhost:8888\n}\n# END PHPMYADMIN DOMAIN #"

if ! grep -qF "$MARKER" "$CADDYFILE"; then
    docker cp $CADDYFILE /tmp/Caddyfile_1.7.60
    sed -i "/$INSERT_AFTER/a $BLOCK" "$CADDYFILE"
    echo "phpMyAdmin domain block added, bekap is stored at /tmp/Caddyfile_1.7.60"
    docker --context=default restart caddy
fi

curl -L https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/mysql/phpmyadmin/pma.php -o /etc/openpanel/mysql/phpmyadmin/pma.php
curl -L https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/mysql/phpmyadmin/config.inc.php -o /etc/openpanel/mysql/phpmyadmin/config.inc.php

for env_file in /home/*/docker-compose.yml; do
    user_dir="$(dirname "$env_file")"
    timeout 5 docker --context="$user_dir" compose down phpmyadmin > /dev/null 2>&1
done

cd /root && timeout 60 docker compose up -d phpmyadmin
