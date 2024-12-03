#!/bin/bash

set -e

if ! command -v curl &> /dev/null || ! command -v jq &> /dev/null; then
  echo "Error: curl and jq are required but not installed."
  exit 1
fi

if [ -z "$DIGITALOCEAN_TOKEN" ]; then
  echo "Error: DIGITALOCEAN_TOKEN is not set."
  exit 1
fi

if [ -z "$DROPLET_ID" ]; then
  echo "Error: DROPLET_ID is not set."
  exit 1
fi


if [ -z "$SNAPSHOT_ID" ]; then
  echo "Error: Snapshot '$SNAPSHOT_ID' not found."
  exit 1
fi

# Step 2: Restore the droplet from the snapshot
echo "Restoring droplet ID $DROPLET_ID from snapshot ID $SNAPSHOT_ID..."
RESTORE_RESPONSE=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" \
  -d "{\"type\":\"restore\",\"image\":\"$SNAPSHOT_ID\"}" \
  "https://api.digitalocean.com/v2/droplets/$DROPLET_ID/actions")

ACTION_ID=$(echo "$RESTORE_RESPONSE" | jq -r '.action.id')

if [ "$ACTION_ID" == "null" ] || [ -z "$ACTION_ID" ]; then
  echo "Error: Failed to initiate droplet restore."
  echo "Response: $RESTORE_RESPONSE"
  exit 1
fi

echo "Restore initiated with Action ID: $ACTION_ID"

# Step 3: Monitor the restore process
echo "Waiting for the restore process to complete..."
while true; do
  STATUS=$(curl -s -X GET \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" \
    "https://api.digitalocean.com/v2/actions/$ACTION_ID" | \
    jq -r '.action.status')

  echo "Current status: $STATUS"

  if [ "$STATUS" == "completed" ]; then
    echo "Restore completed successfully."
    break
  elif [ "$STATUS" == "errored" ]; then
    echo "Error: Restore failed."
    exit 1
  fi

  sleep 10
done

echo "Droplet restored successfully from snapshot."
