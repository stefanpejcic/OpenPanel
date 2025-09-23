#!/bin/bash

echo "Adding support for detailed debug messages when dev_mode=on for OpenPanel GUI.."
wget -4 -O /etc/openpanel/openpanel/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/fc1b6503cc62e3d61abab6473e42175e63fb7436/openpanel/service/service.config.py && docker restart openpanel

echo "Adding support for detailed debug messages when dev_mode=on for OpenAdmin GUI.."
wget -4 -O /etc/openpanel/openadmin/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/service/service.config.py
