#!/bin/bash

# we split emails.py to 8 separate modules, so if emails was enabled we need to:
# - enable other 7 modules on update
# - add them in OpenAdmin > Settings > Modules
# - enable for all feature sets that had emails on OpenAdmin > Hosting Plans > Features Manager

readonly CONFIG_FILE_PATH='/etc/openpanel/openpanel/conf/openpanel.config'
readonly FEATURES_DIR='/etc/openpanel/openpanel/features'
readonly EMAIL_SUBMODULES=("email_deliverability" "email_filters" "email_aliases" "email_import" "email_export" "email_default" "webmail")

sync_config_file() {
    local enabled_modules new_modules module
    if [ ! -f "$CONFIG_FILE_PATH" ]; then
        echo "Config file not found: $CONFIG_FILE_PATH" >&2
        return 1
    fi
    enabled_modules=$(grep '^enabled_modules=' "$CONFIG_FILE_PATH" | cut -d'=' -f2 | tr -d '"')
    if ! echo "$enabled_modules" | grep -q '\bemails\b'; then
        return 0
    fi
    new_modules="$enabled_modules"
    for module in "${EMAIL_SUBMODULES[@]}"; do
        if ! echo "$new_modules" | grep -q "\b${module}\b"; then
            if [ -z "$new_modules" ]; then
                new_modules="$module"
            else
                new_modules="${new_modules},${module}"
            fi
        fi
    done
    if [ "$new_modules" != "$enabled_modules" ]; then
        sed -i "s|^enabled_modules=.*|enabled_modules=\"${new_modules}\"|" "$CONFIG_FILE_PATH"
        echo "Email sub-modules enabled: ${new_modules}"
    fi
}

sync_feature_files() {
    local file module added

    if [ ! -d "$FEATURES_DIR" ]; then
        echo "Features directory not found: $FEATURES_DIR" >&2
        return 0
    fi

    shopt -s nullglob
    for file in "$FEATURES_DIR"/*.txt; do
        if ! grep -qx 'emails' "$file"; then
            continue
        fi

        added=0
        for module in "${EMAIL_SUBMODULES[@]}"; do
            if ! grep -qx "$module" "$file"; then
                [ -s "$file" ] && [ -z "$(tail -c1 "$file")" ] || echo >> "$file"
                echo "$module" >> "$file"
                added=1
            fi
        done

        if [ "$added" -eq 1 ]; then
            echo "Email sub-modules added to: $file feature set."
        fi
    done
    shopt -u nullglob
}

main() {
    sync_config_file
    sync_feature_files
}

main "$@"
