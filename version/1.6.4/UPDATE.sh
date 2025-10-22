#!/bin/bash

echo "Updating default NodeJS template..."
default_node_file="/etc/openpanel/docker/compose/nodejs.yml"

if [ -f "$default_node_file" ]; then
    cp "$default_node_file" "$default_node_file.bak"
    sed -i 's/_NODE_INSTALL/_NODE_REQUIREMENTS/g' "$default_node_file"
    echo " - Updated $default_node_file (backup: $default_node_file.bak)"
else
    echo " - Default template not found at $default_node_file"
fi


echo "[*] Updating user docker-compose.yml files..."
for f in /home/*/docker-compose.yml; do
    [ -f "$f" ] || continue
    echo " - Processing $f"

    cp "$f" "$f.bak"
    sed -i 's/_NODE_INSTALL/_NODE_REQUIREMENTS/g' "$f"
    sed -i 's/_PY_INSTALL/_PY_REQUIREMENTS/g' "$f"
done

echo "All replacements completed."
