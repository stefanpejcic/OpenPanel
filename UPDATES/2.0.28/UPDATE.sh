#!/bin/bash

# remove the backups 2.0.22 made: docker-compose.yml files (docker-proxy restart line) and the Caddyfile (reuse_private_keys)
rm -f /etc/openpanel/docker/compose/1.0/docker-compose.yml.bak-2.0.22 /home/*/docker-compose.yml.bak-2.0.22
rm -f /etc/openpanel/caddy/Caddyfile.bak-2.0.22
