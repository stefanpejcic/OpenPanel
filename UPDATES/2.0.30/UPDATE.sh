#!/bin/bash

# secrets that used to ship world-readable: the MySQL root password, OpenAdmin's users and 2FA secrets, the panel config with API keys
for f in /etc/openpanel/mysql/host_my.cnf /etc/openpanel/mysql/container_my.cnf /etc/openpanel/openadmin/users.db /etc/openpanel/openpanel/conf/openpanel.config; do
    if [ -f "$f" ] && [ ! -L "$f" ]; then
        echo "Restricting $f to root..."
        chown root:root "$f"
        chmod 600 "$f"
    fi
done

# one-time OpenAdmin autologin tokens, one per user dir
find /etc/openpanel/openpanel/core/users -mindepth 2 -maxdepth 2 -name logintoken.txt -type f -exec chmod 600 {} + 2>/dev/null
