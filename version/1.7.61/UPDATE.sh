#!/bin/bash

for env_file in /home/*/docker-compose.yml; do
    user_dir="$(dirname "$env_file")"
    timeout 5 docker --context="$user_dir" image rm phpmyadmin:latest > /dev/null 2>&1
done
