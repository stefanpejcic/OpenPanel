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

# Escape backslashes and single quotes so a value can be safely placed
# inside a single-quoted MySQL string literal, e.g.: "'$(mysql_escape "$value")'"
mysql_escape() {
    local s="$1"
    s="${s//\\/\\\\}"
    s="${s//\'/\\\'}"
    printf '%s' "$s"
}
