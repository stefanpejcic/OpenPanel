#!/bin/bash
# Fallback method to update OpenPanel to the latest version directly from GitHub

# STEP 1. get the latest OpenPanel version
version=$(curl -s https://raw.githubusercontent.com/stefanpejcic/OpenPanel/refs/heads/main/version/latest)

# STEP 2. download update script
wget -O /tmp/openpanel-update-$version https://raw.githubusercontent.com/stefanpejcic/OpenPanel/refs/heads/main/version/$version/UPDATE.sh

# STEP 3. run the update script
chmod +x /tmp/openpanel-update-$version && bash /tmp/openpanel-update-$version

# STEP 4. clean up by removing the update script
rm -rf /tmp/openpanel-update-$version
