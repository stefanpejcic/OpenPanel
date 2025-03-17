#!/bin/bash
# env added to pip and services files edited to remove lock files on startup

source /usr/local/admin/venv/bin/activate
pip install --default-timeout=3600 --force-reinstall --ignore-installed -r requirements.txt  > /dev/null 2>&1 || pip install --default-timeout=3600 --force-reinstall --ignore-installed -r requirements.txt --break-system-packages  > /dev/null 2>&1

deactivate

wget -O /etc/openpanel/openadmin/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/service/service.config.py
wget -O /etc/openpanel/openpanel/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/service/service.config.py

systemctl daemon-reload

systemctl restart admin
