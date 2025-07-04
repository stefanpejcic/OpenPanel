#!/bin/bash

# opencli server-motd generates this every 10hr
cat /etc/openpanel/skeleton/motd

USER_UID=$(id -u)

if [ "$USER_UID" -ne 0 ]; then
    if ! command -v docker &> /dev/null; then
        echo "Error: Docker is not accessible - contact Administrator!"
        exit 1
    fi

    if [ ! -e "/hostfs/run/user/$USER_UID/docker.sock" ]; then
        echo "Error: Docker context not accessible for the user $(whoami)  - contact Administrator!"
        exit 1
    fi

    exec docker run --rm -it \
        --cpus="0.1" \
        --memory="100m" \
        --pids-limit="10" \
        --security-opt no-new-privileges \
        -v /hostfs/run/user/$USER_UID/docker.sock:/var/run/docker.sock \
        lazyteam/lazydocker
else
    exec "$@"
fi
