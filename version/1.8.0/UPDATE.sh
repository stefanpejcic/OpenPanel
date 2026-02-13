#!/bin/bash


# MAJOR

echo
echo "Updating public_suffix_list file for new TLDs.."
wget -o /etc/openpanel/openpanel/conf/public_suffix_list.dat https://publicsuffix.org/list/public_suffix_list.dat


echo
echo "Updating admin panel service to lower resource consumption.."
wget -O /etc/openpanel/openadmin/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/service/service.config.py
wget -O /etc/openpanel/openadmin/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/service/service.config.py
systemctl daemon-reload
systemctl restart  admin
