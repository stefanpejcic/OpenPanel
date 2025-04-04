#!/bin/bash

wget -O /etc/openpanel/ofelia/users.ini https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/ofelia/users.ini

for dir in /home/*; do
    file="$dir/httpd.conf"
    user=$(basename "$dir")

    if [[ -f "$file" ]]; then
        cp /etc/openpanel/ofelia/users.ini $dir/crons.ini
    fi

    # todo: edit compose.yml file and .env for use
    #
    #
    #  




done
