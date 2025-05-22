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

# Check if the config file exists
if [[ ! -f "$CONFIG_FILE" ]]; then
    echo "Config file not found: $CONFIG_FILE"
    exit 1
fi

# Extract the current enabled_modules (strip quotes)
current_modules=$(grep "^enabled_modules=" "$CONFIG_FILE" | cut -d= -f2 | sed 's/^"\(.*\)"$/\1/')

# Convert to array
IFS=',' read -ra current_array <<< "$current_modules"

# Create a set for quick lookup
declare -A current_set
for mod in "${current_array[@]}"; do
    current_set["$mod"]=1
done

modules_modified=false

# Loop through each required module
for module in "${MODULES_TO_CHECK[@]}"; do
    if [[ -n "${current_set[$module]}" ]]; then
        echo "'$module' is already enabled."
    else
        echo "Adding '$module' to enabled_modules..."
        current_modules="${current_modules},$module"
        modules_modified=true
    fi
done

# Update the config file only if changed
if [ "$modules_modified" = true ]; then
    sed -i "s|^enabled_modules=.*|enabled_modules=\"${current_modules}\"|" "$CONFIG_FILE"
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
