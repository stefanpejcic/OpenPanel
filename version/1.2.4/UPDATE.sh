#!/bin/bash

echo ""
echo "Adding fix for custom files not loading.. - issue #444"
sed -i 's#/usr/local/panel/#/#g' /root/docker-compose.yml 
cd /root
docker compose down openpanel && docker compose up -d openpanel

