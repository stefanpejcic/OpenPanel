#!/bin/bash

echo "Adding support for custom SSL on OpenPanel GUI.."
# https://github.com/stefanpejcic/OpenPanel/issues/671#issuecomment-3242952264
wget -O /etc/openpanel/openpanel/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/fc1b6503cc62e3d61abab6473e42175e63fb7436/openpanel/service/service.config.py
docker restart openpanel

