#!/bin/bash

opencli config update screenshots local


wget -O /etc/openpanel/docker/daemon/rootless.json https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/docker/daemon/rootless.json
