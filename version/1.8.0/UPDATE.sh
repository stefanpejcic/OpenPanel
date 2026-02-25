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


# SENTINEL
cd /usr/local/
git clone https://github.com/stefanpejcic/sentinel 
cd /usr/local/sentinel 
#TODO: https://github.com/stefanpejcic/opencli/edit/main/opencli

ARCH=$(uname -m)

case "$ARCH" in
    x86_64)
        echo "Architecture is AMD64 (x86_64) - Updating Sentinel"
        ln -s /usr/local/sentinel/sentinel-amd64 /usr/local/bin/sentinel
        ;;
    aarch64|arm64|armv7l)
        echo "Architecture is ARM ($ARCH) - Updating Sentinel"
        ln -s /usr/local/sentinel/sentinel-arm64 /usr/local/bin/sentinel
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac
