#!/bin/bash

opencli config update screenshots local


wget -O /etc/openpanel/docker/daemon/rootless.json https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/docker/daemon/rootless.json

sed -i '/^SecRxPreFilter/s/^/#/' /etc/openpanel/caddy/coraza_rules.conf && docker --context=default restart caddy
