#!/bin/bash

#########################################################################
############################### DB LOGIN ################################ 
#########################################################################

# This file is included in all scripts that interact with the MySQL database.
# Modify this file only if:
#   1. You have moved it to a different location, or
#   2. The database name has changed.

config_file="/etc/my.cnf"

[[ -f "$config_file" ]] || {
    echo "FATAL ERROR: MySQL config file not available: $config_file"
    exit 1
}

mysql_database="panel"

#########################################################################
