#!/bin/bash

sed -i 's|/usr/bin/wp|/usr/local/bin/wp:ro|g' /etc/openpanel/docker/compose/1.0/docker-compose.yml

