#!/bin/bash

echo
echo "Installing flask-sock for OpenAdmin > Advanced > Web Terminal"
source /usr/local/admin/venv/bin/activate && pip install flask-sock && deactivate
service admin restart


echo
echo ""
sed -i.bak 's|/run/user/:/hostfs/run/user/:ro|/run/user/:/hostfs/run/user/:shared,rslave|g' /root/docker-compose.yml
cd /root && docker --context=default compose down openpanel && docker --context=default compose up -d openpanel
