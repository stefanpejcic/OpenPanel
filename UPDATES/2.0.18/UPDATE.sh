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
