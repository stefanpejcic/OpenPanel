#!/usr/bin/env bash
# fix-permissions.sh
#
# Description:
# This patch fixes the file permissions for OpenPanel configuration files
# to ensure they are not world-writable and are owned by the proper user.
# It is safe to run multiple times.
#
# Usage:
# This script is intended to be run via opencli patch:
#   opencli patch fix-permissions
#
# Warning
# This is an example patch.
#

set -euo pipefail

CONFIG_DIR="/etc/openpanel/openpanel"

echo "Fixing permissions in ${CONFIG_DIR}..."

find "${CONFIG_DIR}" -type f -exec chmod 640 {} \;
find "${CONFIG_DIR}" -type d -exec chmod 750 {} \;
chown -R root:root "${CONFIG_DIR}"

echo "Permissions and ownership fixed."
