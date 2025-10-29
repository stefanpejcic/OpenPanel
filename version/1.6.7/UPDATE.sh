#!/bin/bash

for file in /home/*/.env; do
    [ -f "$file" ] || continue
    sed -i 's/^PGADMIN_MAIL="\([^"]*\)"/PGADMIN_MAIL=\1/' "$file"
    sed -i 's/^PGADMIN_PASS="\([^"]*\)"/PGADMIN_PASS=\1/' "$file"
done
