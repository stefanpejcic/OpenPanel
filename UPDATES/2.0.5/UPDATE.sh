#!/bin/bash

panel_config="/etc/openpanel/openpanel/conf/openpanel.config"

if ! grep -qE '^[[:space:]]*onboarding=' "$panel_config"; then
    echo "Added onboarding option - https://github.com/stefanpejcic/OpenPanel/discussions/1105"
    sed -i '/^how_to_guides/a onboarding=yes' "$panel_config"
fi
