#!/bin/bash

CONFIG_FILE="/etc/openpanel/openpanel/conf/openpanel.config"

logrotate_enable="yes"
logrotate_size_limit="100m"
logrotate_retention=10
logrotate_keep_days=30

get_config_value() {
    local key=$1
    local default_value=$2
    grep -E "^${key}=" "$CONFIG_FILE" | awk -F'=' '{print $2}' | tr -d ' ' || echo "$default_value"
}

# Check and install logrotate if missing
if ! command -v logrotate &> /dev/null; then
    echo "logrotate is not installed. Installing..."

    if command -v apt-get &> /dev/null; then
        sudo apt-get update > /dev/null 2>&1
        sudo apt-get install -y -qq logrotate > /dev/null 2>&1
    elif command -v yum &> /dev/null; then
        sudo yum install -y -q logrotate > /dev/null 2>&1
    elif command -v dnf &> /dev/null; then
        sudo dnf install -y -q logrotate > /dev/null 2>&1
    else
        echo "Error: No compatible package manager found. Please install logrotate manually."
        exit 1
    fi

    if ! command -v logrotate &> /dev/null; then
        echo "Error: logrotate installation failed. Please install manually."
        exit 1
    fi
fi

# Load config values or use defaults
logrotate_enable=$(get_config_value "logrotate_enable" "$logrotate_enable")
logrotate_size_limit=$(get_config_value "logrotate_size_limit" "$logrotate_size_limit")
logrotate_retention=$(get_config_value "logrotate_retention" "$logrotate_retention")
logrotate_keep_days=$(get_config_value "logrotate_keep_days" "$logrotate_keep_days")

if [ "$logrotate_enable" != "yes" ]; then
    echo "Log rotation is not enabled."
    exit 0
fi

LOGROTATE_CONF="/etc/logrotate.d/caddy-logs"

# Write logrotate config for caddy logs (including subdirs)
cat <<EOF > "$LOGROTATE_CONF"
/var/log/caddy/domlogs/coraza_waf/*.log
/var/log/caddy/domlogs/domlogs/*/access.log {
    size $logrotate_size_limit
    rotate $logrotate_retention
    daily
    missingok
    notifempty
    compress
    delaycompress
    copytruncate
    create 640 root adm
    postrotate
        docker --context=default exec caddy bash -c "caddy validate && caddy reload" >/dev/null 2>&1
    endscript
    maxage $logrotate_keep_days
}
EOF

# Force logrotate to apply the config now
/usr/sbin/logrotate --force "$LOGROTATE_CONF"

echo "Caddy log rotation configuration applied successfully."
