#!/bin/bash


todo: remove

```
	local pma_file="${ETC_DIR}mysql/phpmyadmin/pma.php"	
	local secret_key=$(cat "${ETC_DIR}openpanel/secret.key")
	sed -i "s/\(\$fileToken = \"\)[^\"]*\"/\1${secret_key}\"/" "$pma_file"
```
from install.sh

# SHARED PHPMYADMIN
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

# TODO, update them!
curl -L https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/mysql/phpmyadmin/pma.php -o /etc/openpanel/mysql/phpmyadmin/pma.php
curl -L https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/mysql/phpmyadmin/config.inc.php -o /etc/openpanel/mysql/phpmyadmin/config.inc.php

for env_file in /home/*/docker-compose.yml; do
    user_dir="$(dirname "$env_file")"
    timeout 5 docker --context="$user_dir" compose down phpmyadmin > /dev/null 2>&1
done

