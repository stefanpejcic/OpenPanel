#!/bin/bash

BASE_PATH="/var/www/html/domains"
LOG_FILE="/var/www/delete.log"
CURRENT_DATE=$(date '+%Y-%m-%d %H:%M:%S')

# Find and delete config.php files older than 15 minutes
find "$BASE_PATH"/*/config.php -type f -mmin +15 | while read -r file; do
    # Log the deletion
    echo "[$CURRENT_DATE] Deleted config.php on path '$file' - It was older than 15 minutes." >> "$LOG_FILE"
    rm -f "$file"
done

exit 0

