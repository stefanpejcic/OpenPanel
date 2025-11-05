#!/bin/bash

echo
echo "Applying fix for #746"

sed -i -e 's/^PGADMIN_MAIL="\([^"]*\)"/PGADMIN_MAIL=\1/' \
       -e 's/^PGADMIN_PASS="\([^"]*\)"/PGADMIN_PW=\1/' \
       -e 's/^PGADMIN_PW="\([^"]*\)"/PGADMIN_PW=\1/' \
       /etc/openpanel/docker/compose/1.0/.env

for file in /home/*/.env; do
    [ -f "$file" ] || continue
    sed -i -e 's/^PGADMIN_MAIL="\([^"]*\)"/PGADMIN_MAIL=\1/' \
           -e 's/^PGADMIN_PASS="\([^"]*\)"/PGADMIN_PW=\1/' \
           -e 's/^PGADMIN_PW="\([^"]*\)"/PGADMIN_PW=\1/' \
           "$file"
done

echo
echo "Removing server-motd command from crons"
CRON_FILE="/etc/cron.d/openpanel"

if [[ -f "$CRON_FILE" ]]; then
    sed -i '/server-motd/d' "$CRON_FILE"
fi


