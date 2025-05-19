#!/bin/bash

echo ""
echo "📥 Updating openadmin service.."
wget -O /etc/openpanel/openadmin/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/service/service.config.py
systemctl restart admin  > /dev/null 2>&1

CONFIG_FILE="/etc/openpanel/openpanel/conf/openpanel.config"
MODULES_TO_CHECK=("mysql" "domains" "filemanager" "php")

# Check if the config file exists
if [[ ! -f "$CONFIG_FILE" ]]; then
    echo "Config file not found: $CONFIG_FILE"
    exit 1
fi

# Get the current enabled_modules line
current_modules=$(grep "^enabled_modules=" "$CONFIG_FILE" | cut -d= -f2)

# Keep track of changes
modules_modified=false

# Loop through each required module
for module in "${MODULES_TO_CHECK[@]}"; do
    if echo "$current_modules" | grep -qw "$module"; then
        echo "'$module' is already enabled."
    else
        echo "Adding '$module' to enabled_modules..."
        current_modules="${current_modules},$module"
        modules_modified=true
    fi
done

# Update the file only if we added new modules
if [ "$modules_modified" = true ]; then
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
