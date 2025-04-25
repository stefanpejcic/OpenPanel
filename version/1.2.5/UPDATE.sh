#!/bin/bash

# Loop through all directories in /home/
for dir in /home/*; do
    username=$(basename "$dir")
    CONF_PATH="$dir/crons.ini"
    USER_CONF="/etc/openpanel/$username/users.ini"

    # 1. Handle crons.ini directory -> convert to file
    if [ -d "$CONF_PATH" ]; then
        echo "Found directory: $CONF_PATH - removing and creating a file."
        rm -rf "$CONF_PATH" && touch "$CONF_PATH"
    fi

    # 2. Check if users.ini exists
    if [ ! -f "$USER_CONF" ]; then
        echo "users.ini not found for $username, creating it..."
        mkdir -p "/etc/openpanel/$username/"
        wget -q -O "$USER_CONF" "https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/ofelia/users.ini"
    else
        echo "users.ini exists for $username, skipping."
    fi
done

