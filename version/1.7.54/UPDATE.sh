#!/bin/bash

FILE="/etc/openpanel/openadmin/config/notifications.ini"
if ! grep -qE '^[[:space:]]*max_total_conn=' "$FILE"; then
    cat >> "$FILE" <<'EOF'

[HTTP]
max_total_conn=5000
max_conn_per_ip=200
EOF
    echo "HTTP config added to OpenAdmin > Notifications."
fi
