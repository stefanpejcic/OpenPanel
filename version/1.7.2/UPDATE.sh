#!/bin/bash

echo 
echo "Updating service logs for OpenAdmin Log Viewer.."
wget -O /etc/openpanel/openadmin/config/log_paths.json https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/config/log_paths.json


mv /etc/openpanel/openadmin/config/terms_temporary_off /etc/openpanel/openadmin/config/terms.md
