#!/bin/bash

# should be run on clean droplet installation of ubuntu24
# it will:
#
# install openpanel
# delete admin account
# setup script to create new admin on first ssh login
# cleanup the image
# trigger api to convert droplet to iso
# add iso to existing marketplace item



# install latest panel!
bash <(curl -sSL https://openpanel.org) --hostname=demo.openpanel.org 

# Integrate consolidated build routine for Docker and docker-compose setup
bash /home/getsuper/OpenPanel/scripts/build_common.sh

# remove admin accounts
truncate -s 0 /etc/openpanel/openadmin/users.db

# cleanup logs
rm -rf /etc/openpanel/admin/*
rm -rf /etc/openpanel/user/*
rm -rf /root/openpanel_install.log

# do image cleanup
bash scripts/03-force-ssh-logout.sh
bash scripts/03-force-ssh-logout.sh
bash scripts/99-img-check.sh



# get droplet id
droplet_id=$(curl http://169.254.169.254/metadata/v1/id)
echo "DROPLET ID: $droplet_id"


# fun stuff
DO_API_TOKEN="TOKEN_HEREEEEEEEEEEEE"


# create snapshot
snapshot_name="openpanel-$(date +%Y-%m-%d-%H-%M-%S)"
response=$(curl -X POST -H "Authorization: Bearer $DO_API_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"type":"snapshot", "name":"'$snapshot_name'"}' \
    "https://api.digitalocean.com/v2/droplets/$droplet_id/actions")

action_id=$(echo "$response" | jq -r '.action.id')
echo "Action ID: $action_id"


status="in-progress"
while [ "$status" == "in-progress" ]; do
    sleep 10
    response=$(curl -X GET -H "Authorization: Bearer $DO_API_TOKEN" \
        "https://api.digitalocean.com/v2/actions/$action_id")
    status=$(echo "$response" | jq -r '.action.status')
    echo "Snapshot status: $status"
done


snapshot_id=$(curl -X GET -H "Authorization: Bearer $DO_API_TOKEN" \
    "https://api.digitalocean.com/v2/droplets/$droplet_id/snapshots" | jq -r '.snapshots[0].id')

echo "Snapshot ID: $snapshot_id"


