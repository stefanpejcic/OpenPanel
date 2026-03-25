#!/bin/bash

echo
echo "Patching :ro bug for /run/user/ inside OpenPanel UI container.."
sed -i.bak 's|/run/user/:/hostfs/run/user/:ro|/run/user/:/hostfs/run/user/:shared,rslave|g' /root/docker-compose.yml
cd /root && docker --context=default compose down openpanel && docker --context=default compose up -d openpanel


echo 
echo "Adding patch for FTP service"
yes yes | opencli patch 896
