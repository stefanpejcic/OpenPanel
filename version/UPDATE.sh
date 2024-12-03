#!/bin/bash
# fallback method to update OpenPanel to latest version directly from GitHub

# get latest openpanel version
version=($curl https://raw.githubusercontent.com/stefanpejcic/OpenPanel/refs/heads/main/version/latest)

# download update script
wget -O /tmp/openpanel-update-$version https://raw.githubusercontent.com/stefanpejcic/OpenPanel/refs/heads/main/version/$version/UPDATE.sh


# run it
chmod +x /tmp/openpanel-update-$version && bash /tmp/openpanel-update-$version
rm -rf /tmp/openpanel-update-$version
