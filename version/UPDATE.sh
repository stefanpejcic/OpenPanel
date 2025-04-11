#!/bin/bash
# Fallback method to update OpenPanel to the latest version directly from GitHub

# STEP 1. use user-provided version or check latest OpenPanel version
if [[ "$1" == "--version" && -n "$2" ]]; then
    version=$2
else
    version=$(curl -s https://raw.githubusercontent.com/stefanpejcic/OpenPanel/refs/heads/main/version/latest)
fi

# STEP 2. download update script
wget -O /tmp/openpanel-update-$version https://raw.githubusercontent.com/stefanpejcic/OpenPanel/refs/heads/main/version/$version/UPDATE.sh

# STEP 3. run the update script
chmod +x /tmp/openpanel-update-$version && bash /tmp/openpanel-update-$version

# STEP 4. clean up by removing the update script
rm -rf /tmp/openpanel-update-$version

# Fix for https://github.com/stefanpejcic/OpenPanel/issues/411
chmod a+x /etc/openpanel/wordpress/wp-cli.phar

