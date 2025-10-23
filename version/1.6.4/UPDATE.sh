#!/bin/bash



wget -O /etc/openpanel/nginx/vhosts/1.1/nginx_proxy_headers.txt https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/nginx/vhosts/1.1/nginx_proxy_headers.txt


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

FEATURES_JSON_PATH="/etc/openpanel/openadmin/config/features.json"
FEATURES_URL="https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/config/features.json"
FEATURES_DIR="/etc/openpanel/openpanel/features"
FILES=("default.txt" "basic.txt")

echo "Downloading new features.json..."
if curl -fsSL "$FEATURES_URL" -o "$FEATURES_JSON_PATH"; then
    echo "features.json successfully replaced."
else
    echo "Error downloading features.json!"
    exit 1
fi

for FILE in "${FILES[@]}"; do
    FILE_PATH="$FEATURES_DIR/$FILE"

    if ! grep -qx "services" "$FILE_PATH"; then
        echo "services" >> "$FILE_PATH"
        echo "Added 'services' to $FILE_PATH"
    else
        echo "'services' already exists in $FILE_PATH"
    fi
done
