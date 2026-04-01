#!/bin/bash
echo 
echo "Updating service file to show timestamp in docker log for openpanel service.."
wget -O /etc/openpanel/openpanel/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/service/service.config.py
docker restart openpanel

echo 
echo "Disabling CoreRuleSet REQUEST-941-APPLICATION-ATTACK-XSS.conf"
mv /etc/openpanel/caddy/coreruleset/rules/REQUEST-941-APPLICATION-ATTACK-XSS.conf /etc/openpanel/caddy/coreruleset/rules/REQUEST-941-APPLICATION-ATTACK-XSS.conf.disabled

echo 
echo "Adding 'ssl' module for Community edition.."
wget -O /etc/openpanel/openadmin/config/features.json https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/config/features.json

echo 
echo "Increasing activity_items_per_page value from 25 to 100"
opencli config update activity_items_per_page 100

echo
echo "Setting limits for all existing users..."
for home_dir in /home/*; do
    [[ -d "$home_dir" ]] || continue
    env_file="$home_dir/.env"
    [[ -f "$env_file" ]] || continue

    # Windows \r characters
    env_data=$(tr -d '\r' < "$env_file")
    eval "$env_data"  # safe for trusted env files

    [[ -z "$USER_ID" ]] && continue

    # --- CPU Limit ---
    if [[ -n "$TOTAL_CPU" ]]; then
        if [[ "$TOTAL_CPU" == "0" ]]; then
            systemctl set-property "user-${USER_ID}.slice" CPUQuota=infinity
        else
            cpu_percent=$(( TOTAL_CPU * 100 ))
            systemctl set-property "user-${USER_ID}.slice" CPUQuota="${cpu_percent}%"
        fi
    fi

    # --- RAM Limit ---
    if [[ -n "$TOTAL_RAM" ]]; then
        if [[ "$TOTAL_RAM" == "0" ]]; then
            systemctl set-property "user-${USER_ID}.slice" MemoryMax=infinity
        else
            # Add G if numeric
            if [[ "$TOTAL_RAM" =~ ^[0-9]+$ ]]; then
                TOTAL_RAM="${TOTAL_RAM}G"
            fi
            # Uppercase the unit
            TOTAL_RAM="${TOTAL_RAM^^}"
            systemctl set-property "user-${USER_ID}.slice" MemoryMax="$TOTAL_RAM"
        fi
    fi

done
