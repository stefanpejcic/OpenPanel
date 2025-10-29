#!/bin/bash

sed -i -e 's/^PGADMIN_MAIL="\([^"]*\)"/PGADMIN_MAIL=\1/' \
       -e 's/^PGADMIN_PASS="\([^"]*\)"/PGADMIN_PASS=\1/' \
       /etc/openpanel/docker/compose/1.0/.env

for file in /home/*/.env; do
    [ -f "$file" ] || continue
    sed -i 's/^PGADMIN_MAIL="\([^"]*\)"/PGADMIN_MAIL=\1/' "$file"
    sed -i 's/^PGADMIN_PASS="\([^"]*\)"/PGADMIN_PASS=\1/' "$file"
done
