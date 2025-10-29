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
