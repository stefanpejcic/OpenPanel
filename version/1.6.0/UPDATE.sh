#!/bin/bash

wget -O /etc/openpanel/openpanel/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/fc1b6503cc62e3d61abab6473e42175e63fb7436/openpanel/service/service.config.py
docker restart openpanel

