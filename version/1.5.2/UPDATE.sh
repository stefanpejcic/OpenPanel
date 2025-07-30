#!/bin/bash

#wget -O /etc/openpanel/ftp/start_vsftpd.sh https://raw.githubusercontent.com/stefanpejcic/OpenPanel-FTP/refs/heads/master/start_vsftpd.sh


touch /root/openpanel_restart_needed
wget -O /etc/openpanel/openpanel/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/service/service.config.py


