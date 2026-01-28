#!/bin/bash
# ======================================================================
# This file is included in all scripts that interact with MySQL.
# Only update this file if:
#   1. Path has changed (i.e., you are not using '/etc/my.cnf'), or
#   2. The database name has been changed (i.e. no longer 'panel').
# ======================================================================

config_file="/etc/my.cnf"
mysql_database="panel"

[[ -f "$config_file" ]] || {
    echo "FATAL ERROR: MySQL config file not available: $config_file"
    exit 1
}
