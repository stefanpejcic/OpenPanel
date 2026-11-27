#!/bin/bash

# AI assistant: reasoning effort setting, empty leaves it out of requests so the model's default applies
OPENPANEL_CONFIG=/etc/openpanel/openpanel/conf/openpanel.config
if [ -f "$OPENPANEL_CONFIG" ] && grep -q "^\[AI\]" "$OPENPANEL_CONFIG" && ! grep -q "^ai_reasoning_effort=" "$OPENPANEL_CONFIG"; then
    echo "Adding ai_reasoning_effort to the [AI] section of $OPENPANEL_CONFIG..."
    sed -i "/^\[AI\]/a ai_reasoning_effort=" "$OPENPANEL_CONFIG"
fi
