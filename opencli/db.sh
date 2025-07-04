#!/bin/bash

#########################################################################
############################### DB LOGIN ################################ 
#########################################################################

config_files=("/etc/my.cnf")

check_config_file() {
    for config_file in "${config_files[@]}"; do
        if [ -f "$config_file" ]; then
            return 0
        fi
    done
    return 1
}

if ! check_config_file; then
    echo "Mysql config file: $config_files is not available!"
    exit 1
fi

mysql_database="panel"

#########################################################################
