#!/bin/bash


wget -O /etc/openpanel/openadmin/config/features.json https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/config/features.json

wget -O /etc/openpanel/openpanel/features/default.txt https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/features/default.txt

CADDYFILE="/etc/openpanel/caddy/Caddyfile"

if grep -q "^ *admin off" "$CADDYFILE"; then
    sed -i 's/^ *admin off/#admin off/' "$CADDYFILE"
    echo "'admin off' has been commented out."
else
    echo "No 'admin off' line found or it's already commented."
fi




wget -O /tmp/tailwind.css http://185.119.90.220:2083/static/dist/output.css
docker cp /tmp/tailwind.css openpanel:/static/dist/output.css
rm -rf /tmp/tailwind.css





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
    if [[ ",$current_modules," == *",$module,"* ]]; then
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





# add views for domain count!

mysql -uroot <<'EOF'
USE panel;

-- Step 1: Sanitize existing data
UPDATE users SET user_domains = '0' WHERE user_domains = '';

-- Step 2: Modify the column to INT
ALTER TABLE users MODIFY COLUMN user_domains INT NOT NULL DEFAULT 0;

-- Step 3: Recalculate user_domains count
UPDATE users u
LEFT JOIN (
    SELECT user_id, COUNT(*) AS domain_count
    FROM domains
    GROUP BY user_id
) d ON u.id = d.user_id
SET u.user_domains = COALESCE(d.domain_count, 0);

-- Step 4: Create triggers
DELIMITER //

CREATE TRIGGER increment_user_domains
AFTER INSERT ON domains
FOR EACH ROW
BEGIN
  UPDATE users SET user_domains = user_domains + 1 WHERE id = NEW.user_id;
END;
//

CREATE TRIGGER decrement_user_domains
AFTER DELETE ON domains
FOR EACH ROW
BEGIN
  UPDATE users SET user_domains = user_domains - 1 WHERE id = OLD.user_id;
END;
//

DELIMITER ;
EOF
