#!/bin/bash

CONFIG_FILE="/etc/openpanel/openpanel/conf/openpanel.config"
MODULE_NAME="mysql"

# Check if the config file exists
if [[ ! -f "$CONFIG_FILE" ]]; then
    echo "Config file not found: $CONFIG_FILE"
    exit 1
fi

# Extract current enabled_modules value
current_modules=$(grep "^enabled_modules=" "$CONFIG_FILE" | cut -d= -f2)

# Check if 'mysql' is already in the list
if echo "$current_modules" | grep -qw "$MODULE_NAME"; then
    echo "'$MODULE_NAME' is already enabled."
else
    # Add 'mysql' to the comma-separated list
    new_modules="${current_modules},$MODULE_NAME"

    # Update the file in place
    sed -i "s/^enabled_modules=.*/enabled_modules=${new_modules}/" "$CONFIG_FILE"
    echo "'$MODULE_NAME' added to enabled_modules."
fi

# Function to check if a Docker container exists and is running
is_container_running() {
    docker ps --format '{{.Names}}' | grep -q "^$1$"
}

# Check and flush Redis if openpanel_redis is running
if is_container_running "openpanel_redis"; then
    echo "Flushing Redis cache in 'openpanel_redis'..."
    docker exec -it openpanel_redis bash -c "redis-cli FLUSHALL"
else
    echo "Container 'openpanel_redis' is not running. Skipping Redis flush."
fi

# Restart OpenPanel if container is running
if is_container_running "openpanel"; then
    echo "Restarting 'openpanel' container..."
    docker restart openpanel
else
    echo "Container 'openpanel' is not running. Skipping restart."
fi
