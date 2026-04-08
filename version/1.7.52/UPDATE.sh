#!/bin/bash

opencli config update screenshots local

wget -O /etc/openpanel/docker/daemon/rootless.json https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/docker/daemon/rootless.json

sed -i '/^SecRxPreFilter/s/^/#/' /etc/openpanel/caddy/coraza_rules.conf && docker --context=default restart caddy

timeout 5  mysql -uroot -e "SET GLOBAL event_scheduler = ON;"

timeout 5 mysql -uroot <<EOF
CREATE EVENT IF NOT EXISTS cleanup_expired_sessions
ON SCHEDULE EVERY 15 MINUTE
DO
  DELETE FROM active_sessions
  WHERE expires_at < NOW();
EOF
