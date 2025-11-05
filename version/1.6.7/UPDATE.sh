#!/bin/bash


# Default template
sed -i -e 's/^PGADMIN_MAIL="\([^"]*\)"/PGADMIN_MAIL=\1/' \
       -e 's/^PGADMIN_PASS="\([^"]*\)"/PGADMIN_PW=\1/' \
       -e 's/^PGADMIN_PW="\([^"]*\)"/PGADMIN_PW=\1/' \
       /etc/openpanel/docker/compose/1.0/.env

# Existing users
for file in /home/*/.env; do
    [ -f "$file" ] || continue
    sed -i -e 's/^PGADMIN_MAIL="\([^"]*\)"/PGADMIN_MAIL=\1/' \
           -e 's/^PGADMIN_PASS="\([^"]*\)"/PGADMIN_PW=\1/' \
           -e 's/^PGADMIN_PW="\([^"]*\)"/PGADMIN_PW=\1/' \
           "$file"
done

CRON_FILE="/etc/cron.d/openpanel"
BACKUP_FILE="/etc/cron.d/openpanel.bak_$(date +%Y%m%d%H%M%S)"

if [[ ! -f "$CRON_FILE" ]]; then
    echo "Error: $CRON_FILE not found."
    exit 1
fi

cp "$CRON_FILE" "$BACKUP_FILE"
echo "Backup created at: $BACKUP_FILE"
sed -i '/server-motd/d' "$CRON_FILE"


