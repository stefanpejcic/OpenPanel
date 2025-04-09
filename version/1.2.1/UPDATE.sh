#!/bin/bash

wget -O /etc/openpanel/mysql/user.cnf https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/mysql/user.cnf

for dir in /home/*; do
    file_or_folder="$dir/custom.cnf"
    user=$(basename "$dir")

    if [[ -d "$file_or_folder" ]]; then
      rm $file_or_folder
      cp /etc/openpanel/mysql/user.cnf $file_or_folder
    fi
done
