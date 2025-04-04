#!/bin/bash

wget -O /etc/openpanel/ofelia/users.ini https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/ofelia/users.ini

for dir in /home/*; do
    file="$dir/httpd.conf"
    user=$(basename "$dir")

    if [[ -f "$file" ]]; then
        cp /etc/openpanel/ofelia/users.ini $dir/crons.ini
    fi

    : '
    todo: edit compose.yml file and .env for use
    
    https://github.com/stefanpejcic/openpanel-configuration/commit/348eb898657d9d4443e37a79e79796e4d3a6306c

    https://github.com/stefanpejcic/openpanel-configuration/commit/5230e2e90056ca669f8c83f79bbb2fd018728140
    '

done
