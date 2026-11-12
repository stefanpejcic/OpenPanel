#!/bin/bash

# imunifyav module is new in 2.0.24, list it under OpenAdmin > Settings > Modules so admins can turn it on
FEATURES_JSON="/etc/openpanel/openadmin/config/features.json"
if [ -f "$FEATURES_JSON" ] && ! jq -e '.[] | select(.name == "imunifyav")' "$FEATURES_JSON" >/dev/null 2>&1; then
    echo "Adding the ImunifyAV module to $FEATURES_JSON..."
    tmp=$(mktemp)
    if jq '. + [{
        "name": "imunifyav",
        "title": "ImunifyAV",
        "description": "If enabled, users get ImunifyAV'"'"'s own interface for their account and a summary of infected files and scans. Needs ImunifyAV installed from OpenAdmin > Security > ImunifyAV.",
        "link": "/imunifyav",
        "type": "community",
        "help_link": "https://openpanel.com/docs/admin/settings/modules/#imunifyav"
    }]' "$FEATURES_JSON" > "$tmp"; then
        # write in place so the file keeps its owner and permissions
        cat "$tmp" > "$FEATURES_JSON"
    else
        echo "Failed to add the ImunifyAV module to $FEATURES_JSON"
    fi
    rm -f "$tmp"
fi

# openpanel mounts imunify's rpc socket dir once the module is on, without the dir podman won't start the container
install -d -m 755 /run/defence360agent
echo "d /run/defence360agent 0755 root root -" > /etc/tmpfiles.d/openpanel-imunify.conf

# servers that already have ImunifyAV get the module turned on, same as a fresh opencli imunify install
CONFIG_FILE="/etc/openpanel/openpanel/conf/openpanel.config"
if command -v imunify-antivirus >/dev/null 2>&1 && [ -d /etc/sysconfig/imunify360 ] && [ -f "$CONFIG_FILE" ]; then
    if ! grep -q "imunify-gui-panel" /usr/local/opencli/imunify.sh 2>/dev/null; then
        echo "opencli is older than 2.0.24, run 'opencli update --cli' and then 'opencli imunify start' to set up the ImunifyAV page for users"
    else
        newly_enabled=""
        enabled_modules=$(grep '^enabled_modules=' "$CONFIG_FILE" | cut -d'=' -f2- | tr -d '"')
        if ! echo ",$enabled_modules," | grep -q ',imunifyav,'; then
            echo "ImunifyAV is installed, enabling the imunifyav module..."
            newly_enabled=1
            if [ -n "$enabled_modules" ]; then
                sed -i "s|^enabled_modules=.*|enabled_modules=\"${enabled_modules},imunifyav\"|" "$CONFIG_FILE"
            else
                sed -i 's|^enabled_modules=.*|enabled_modules="imunifyav"|' "$CONFIG_FILE"
            fi
        fi
        had_mount=""
        grep -q "/run/defence360agent" /root/docker-compose.yml 2>/dev/null && had_mount=1
        # mounts the socket into openpanel (recreates it when the mount is new) and sets up the gui service for the user panel
        timeout 300 opencli imunify start || echo "opencli imunify start failed, run it by hand to set up the ImunifyAV page for users"
        # no recreate happened, so openpanel needs a restart to load the module
        if [ -n "$newly_enabled" ] && [ -n "$had_mount" ] && podman ps --format '{{.Names}}' | grep -qx openpanel; then
            podman restart -t 10 openpanel >/dev/null 2>&1
        fi
    fi
fi
