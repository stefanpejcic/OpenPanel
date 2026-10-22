#!/bin/bash

# email_migrate is new in 2.0.18, list it under OpenAdmin > Settings > Modules so admins can turn it on
FEATURES_JSON="/etc/openpanel/openadmin/config/features.json"
if [ -f "$FEATURES_JSON" ] && ! jq -e '.[] | select(.name == "email_migrate")' "$FEATURES_JSON" >/dev/null 2>&1; then
    echo "Adding the Email Migration module to $FEATURES_JSON..."
    tmp=$(mktemp)
    if jq '. + [{
        "name": "email_migrate",
        "title": "Email Migration",
        "description": "If enabled, users can copy emails from another IMAP server into their email accounts.",
        "link": "/emails/migrate",
        "type": "enterprise",
        "help_link": "https://openpanel.com/docs/admin/settings/modules/#email-migration"
    }]' "$FEATURES_JSON" > "$tmp"; then
        # write in place so the file keeps its owner and permissions
        cat "$tmp" > "$FEATURES_JSON"
    else
        echo "Failed to add the Email Migration module to $FEATURES_JSON"
    fi
    rm -f "$tmp"
fi

# servers that have the email modules get it enabled too, like the other email submodules
CONFIG_FILE="/etc/openpanel/openpanel/conf/openpanel.config"
if [ -f "$CONFIG_FILE" ]; then
    enabled_modules=$(grep '^enabled_modules=' "$CONFIG_FILE" | cut -d'=' -f2- | tr -d '"')
    if echo ",$enabled_modules," | grep -q ',emails,' && ! echo ",$enabled_modules," | grep -q ',email_migrate,'; then
        echo "Enabling the email_migrate module..."
        sed -i "s|^enabled_modules=.*|enabled_modules=\"${enabled_modules},email_migrate\"|" "$CONFIG_FILE"
    fi
fi

# plans that have the address importer get the migration page as well
for feature_set in /etc/openpanel/openpanel/features/*.txt; do
    [ -f "$feature_set" ] || continue
    if grep -qx 'email_import' "$feature_set" && ! grep -qx 'email_migrate' "$feature_set"; then
        echo "Adding email_migrate to feature set $(basename "$feature_set" .txt)..."
        sed -i '/^email_import$/a email_migrate' "$feature_set"
    fi
done

# postgre_custom.conf never reached postgres: its mount was commented out because inside the data dir it stops initdb on a new account ("directory exists but is not empty")
# it's now mounted outside the data dir and passed as config_file, an account picks it up the next time its postgres container is recreated
PG_CONF_LINE="./postgre_custom.conf:/etc/postgresql/postgresql.conf:ro"
PG_CONF_COMMAND='    command: ["postgres", "-c", "config_file=/etc/postgresql/postgresql.conf"]'
pg_conf_compose() {
    local f="$1" tmp
    grep -q "postgre_custom.conf:/var/lib/postgresql/data/postgresql.conf" "$f" || return 0
    grep -q "config_file=/etc/postgresql/postgresql.conf" "$f" && return 0
    tmp=$(mktemp)
    sed -e "s|^\(\s*\)#\?\s*- ./postgre_custom.conf:/var/lib/postgresql/data/postgresql.conf\s*$|\1- $PG_CONF_LINE|" \
        -e "/^    container_name: postgres\$/a\\$PG_CONF_COMMAND" "$f" > "$tmp"
    if grep -qF -- "- $PG_CONF_LINE" "$tmp" && grep -qF "$PG_CONF_COMMAND" "$tmp"; then
        [ -f "$f.bak-2.0.18" ] || cp -a "$f" "$f.bak-2.0.18"
        # written in place so the file keeps its owner and mode
        cat "$tmp" > "$f"
    else
        echo "Could not update the postgres service in $f, left as it is."
    fi
    rm -f "$tmp"
}
pg_conf_compose /etc/openpanel/docker/compose/1.0/docker-compose.yml
for f in /home/*/docker-compose.yml; do
    [ -f "$f" ] || continue
    home_dir=$(dirname "$f")
    # a missing file would be mounted as an empty directory and postgres would not start
    if [ ! -f "$home_dir/postgre_custom.conf" ] && [ -f /etc/openpanel/postgres/postgresql.conf ]; then
        cp /etc/openpanel/postgres/postgresql.conf "$home_dir/postgre_custom.conf"
        chown "$(stat -c '%U:%G' "$home_dir")" "$home_dir/postgre_custom.conf"
    fi
    pg_conf_compose "$f"
done

# the OpenAdmin template editor saved the caddy domain template with the browser's CRLF line endings, every domain made from it inherited them and domains-ssl couldn't find their tls block
for f in /etc/openpanel/caddy/templates/domain.conf /etc/openpanel/caddy/domains/*.conf; do
    [ -f "$f" ] && grep -q $'\r' "$f" || continue
    # written in place, caddy and the user containers see these through bind mounts
    sed 's/\r$//' "$f" > /tmp/op_2018_crlf && cat /tmp/op_2018_crlf > "$f"
done
rm -f /tmp/op_2018_crlf
podman exec caddy caddy reload --config /etc/caddy/Caddyfile >/dev/null 2>&1 </dev/null || true
