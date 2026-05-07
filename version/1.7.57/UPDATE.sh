#!/bin/bash


wget -O /etc/openpanel/caddy/check.conf https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/caddy/check.conf
wget -O /etc/openpanel/docker/dockerd-rootless-setuptool.sh https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/docker/dockerd-rootless-setuptool.sh ; chmod +x /etc/openpanel/docker/dockerd-rootless-setuptool.sh
