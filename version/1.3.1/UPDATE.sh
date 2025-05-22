#!/bin/bash



opencli config update session_duration 100

echo ""
echo "📥 Updating openadmin service.."
wget -O /etc/openpanel/openadmin/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/service/service.config.py

echo ""
echo "📥 Installing features for openadmin.."
wget -O /etc/openpanel/openadmin/config/features.json https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/config/features.json

systemctl restart admin > /dev/null 2>&1

CONFIG_FILE="/etc/openpanel/openpanel/conf/openpanel.config"
MODULES_TO_CHECK=("mysql" "domains" "autoinstaller" "filemanager" "php")

current_modules=$(grep "^enabled_modules=" "$CONFIG_FILE" | sed -E 's/^enabled_modules="(.*)"/\1/')
modules_modified=0

IFS=',' read -r -a enabled_array <<< "$current_modules"

for module in "${MODULES_TO_CHECK[@]}"; do
    if printf '%s\n' "${enabled_array[@]}" | grep -qx "$module"; then
        echo "'$module' is already enabled."
    else
        echo "Adding '$module' to enabled_modules..."
        enabled_array+=("$module")
        modules_modified=1
    fi
done

if [[ $modules_modified -eq 1 ]]; then
    updated_modules=$(IFS=','; echo "${enabled_array[*]}")
    # Use double quotes around the value when writing back
    sed -i "s/^enabled_modules=\".*\"/enabled_modules=\"${updated_modules}\"/" "$CONFIG_FILE"
    sed -i "s/^enabled_modules=\(.*\)$/enabled_modules=\"\1\"/" "$CONFIG_FILE"
    echo "Updated enabled_modules in config file."
fi

# Update the file only if we added new modules
if [ "$modules_modified" -eq 1 ]; then
    sed -i "s/^enabled_modules=.*/enabled_modules=${current_modules}/" "$CONFIG_FILE"
    echo "Updated enabled_modules in config file."
fi

# Function to check if a Docker container is running
is_container_running() {
    docker ps --format '{{.Names}}' | grep -q "^$1$"
}

# Flush Redis if openpanel_redis is running
if is_container_running "openpanel_redis"; then
    echo "Flushing Redis cache in 'openpanel_redis'..."
    docker exec -it openpanel_redis bash -c "redis-cli FLUSHALL"
else
    echo "Container 'openpanel_redis' is not running. Skipping Redis flush."
fi

# Restart OpenPanel if running
if is_container_running "openpanel"; then
    echo "Restarting 'openpanel' container..."
    docker restart openpanel
else
    echo "Container 'openpanel' is not running. Skipping restart."
fi




echo ""
echo "📥 Changing branch from '1.1' to 'main' for OpenCLI scripts.."
rm -rf /usr/local/opencli
cd /usr/local/ && git clone https://github.com/stefanpejcic/opencli
