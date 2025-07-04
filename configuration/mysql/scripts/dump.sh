#!/bin/bash

set -e

excluded_databases=("Database" "information_schema" "performance_schema" "mysql" "sys" "phpmyadmin")
mkdir -p /tmp/dumps

is_excluded() {
    local value="$1"
    for excluded in "${excluded_databases[@]}"; do
        if [[ "$excluded" == "$value" ]]; then
            return 0
        fi
    done
    return 1
}

dump_databases() {
    local db_type="$1"
    local dump_cmd="mysqldump"

    if [[ "$db_type" == "mariadb" ]]; then
        if command -v mariadb-dump >/dev/null 2>&1; then
            dump_cmd="mariadb-dump"
        else
            echo "mariadb-dump not found, falling back to mysqldump"
        fi
    fi

    databases=$($db_type -e "SHOW DATABASES;" | tail -n +2)

    for db in $databases; do
        if is_excluded "$db"; then
            echo "- Skipping excluded database: $db"
        else
            echo "- Dumping database: $db"
            $dump_cmd "$db" > "/tmp/dumps/$db.sql"
        fi
    done
}

if command -v mysql >/dev/null 2>&1; then
        echo "MySQL detected"
        dump_databases "mysql"
elif command -v mariadb >/dev/null 2>&1; then
        echo "MariaDB detected"
        dump_databases "mariadb"
else
    echo "MySQL/MariaDB not installed or not in PATH"
    exit 1
fi
