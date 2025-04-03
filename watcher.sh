#!/bin/bash

SERVICE="OpenAdmin"
PORT=2087

# Check if the service is running
if ! nc -z localhost $PORT; then
    echo "$(date): $SERVICE is not running. Attempting to restart..."
    # Restart the service
    systemctl restart $SERVICE
    if [ $? -eq 0 ]; then
        echo "$(date): $SERVICE restarted successfully."
    else
        echo "$(date): Failed to restart $SERVICE."
    fi
else
    echo "$(date): $SERVICE is running on port $PORT."
fi
