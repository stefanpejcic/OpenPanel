#!/bin/bash

wget -O /etc/openpanel/ftp/Dockerfile https://raw.githubusercontent.com/stefanpejcic/OpenPanel-FTP/refs/heads/master/Dockerfile


wget -O /etc/openpanel/openadmin/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/service/service.config.py


wget -O /etc/openpanel/openadmin/config/features.json https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/config/features.json
