#!/bin/bash

DROPLET_ID="442891520"
DROPLET_IP="45.55.60.4"
DROPLET_IMAGE="ubuntu-24-04-x64"
SSH_KEY_PATH="~/.ssh/id_rsa"
PANEL_HOSTNAME="demo.openpanel.org"
DEMO_SCRIPT="https://raw.githubusercontent.com/stefanpejcic/OpenPanel/refs/heads/main/demo/2087/setup_demo.sh"


# Step 1: Rebuild the droplet
echo "Rebuilding the droplet..."
REBUILD_RESPONSE=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" \
  -d "{\"type\":\"rebuild\",\"image\":\"$DROPLET_IMAGE\"}" \
  "https://api.digitalocean.com/v2/droplets/$DROPLET_ID/actions")

ACTION_STATUS=$(echo "$REBUILD_RESPONSE" | jq -r '.action.status')

if [[ "$ACTION_STATUS" != "in-progress" ]]; then
  echo "Droplet rebuild failed. Response: $REBUILD_RESPONSE"
  exit 1
fi

echo "Droplet rebuild initiated. Waiting for completion..."
sleep 300

# Step 2: SSH to the droplet and run commands
echo "Connecting to the droplet via SSH to set up the panel..."
ssh -o StrictHostKeyChecking=no -i "$SSH_KEY_PATH" root@$DROPLET_IP <<EOF
  wget -O /root/demo.sh $DEMO_SCRIPT
  bash <(curl -sSL https://openpanel.org) --hostname=$PANEL_HOSTNAME --post_install=/root/demo.sh
EOF

if [[ $? -ne 0 ]]; then
  echo "Failed to connect or execute commands on the droplet."
  exit 1
fi


# Step 3. create a snapshot
version=(opencli version)

response=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" \
  -d "{\"type\":\"snapshot\",\"name\":\"${version} demo snapshot\"}" \
  "https://api.digitalocean.com/v2/droplets/$droplet_id/actions")
  
snapshot_id=$(echo "$response" | jq -r '.action.resource_id')
  
if [ "$snapshot_id" != "null" ] && [ -n "$snapshot_id" ]; then
  echo "Snapshot ID: $snapshot_id"
else
  echo "Failed to retrieve snapshot ID."
fi


# step 4. set the job to restore it
# TODO!!!!!


echo "Panel installation and demo setup completed successfully."
