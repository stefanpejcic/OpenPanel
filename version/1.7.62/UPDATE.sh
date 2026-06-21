#!/bin/bash

# we split emails.py to 7 separate modules, so if emails was enabled lets enable other 6 on update:
readonly CONFIG_FILE_PATH='/etc/openpanel/openpanel/conf/openpanel.config'

readonly EMAIL_SUBMODULES=(
    "email_deliverability"
    "email_filters"
    "email_aliases"
    "email_import"
    "email_export"
    "email_default"
    "webmail"
)

main() {
    local enabled_modules new_modules module

    if [ ! -f "$CONFIG_FILE_PATH" ]; then
        echo "Config file not found: $CONFIG_FILE_PATH" >&2
        exit 1
    fi

    enabled_modules=$(grep '^enabled_modules=' "$CONFIG_FILE_PATH" | cut -d'=' -f2 | tr -d '"')

    if ! echo "$enabled_modules" | grep -q '\bemails\b'; then
        exit 0
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
        echo "Email sub-modules synced: ${new_modules}"
    else
        echo "Email sub-modules already in sync."
    fi
}

main "$@"
