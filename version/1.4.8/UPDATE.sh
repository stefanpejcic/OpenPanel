#!/bin/bash

DB_CONFIG_FILE="/usr/local/opencli/db.sh"
. "$DB_CONFIG_FILE"


timeout 3s mysql --defaults-extra-file=/etc/my.cnf -D panel -sN -e "ALTER TABLE sites MODIFY COLUMN ports VARCHAR(255);"

if [ $? -eq 0 ]; then
    echo "Database column modification succeeded."
else
    echo "Database column modification failed or timed out."
    echo ""
    echo "Run manually:"
    echo ""
    echo "mysql -e 'ALTER TABLE sites MODIFY COLUMN ports VARCHAR(255);'"
    echo ""
fi
