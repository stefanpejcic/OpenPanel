#!/bin/bash
# build_common.sh – Shared build routines for OpenPanel installations/updates

# Logs a message with timestamp
log() {
    echo "[ $(date '+%Y-%m-%d %H:%M:%S') ] $*"
}

# Check if docker command is available
check_docker() {
    if ! command -v docker >/dev/null 2>&1; then
        log "ERROR: Docker command is not available."
        exit 1
    fi
    log "Docker command found."
}

# Restart Docker service and verify running status
start_docker() {
    systemctl daemon-reload
    systemctl start docker
    check_docker
    log "[ OK ] Docker is configured."
}

# Pull a list of Docker images concurrently
pull_docker_images() {
    local images=("$@")
    log "Pulling Docker images: ${images[*]}"
    nohup sh -c "echo ${images[*]} | xargs -P4 -n1 docker pull" </dev/null >nohup.out 2>nohup.err &
    log "Docker image pull initiated."
}

# Start docker-compose with current configuration
docker_compose_up() {
    log "Setting up docker-compose..."
    DOCKER_CONFIG=${DOCKER_CONFIG:-$HOME/.docker}
    mkdir -p "$DOCKER_CONFIG/cli-plugins"
    curl -SL https://github.com/docker/compose/releases/download/v2.27.1/docker-compose-linux-x86_64 \
         -o "$DOCKER_CONFIG/cli-plugins/docker-compose" > /dev/null 2>&1
    chmod +x "$DOCKER_CONFIG/cli-plugins/docker-compose"
    log "docker-compose installed at $DOCKER_CONFIG/cli-plugins/docker-compose"
    # ...existing docker-compose configuration steps...
}

# Rollback routine in case of build failure
rollback() {
    log "Rollback initiated: restoring previous state..."
    # ...existing rollback steps...
    log "Rollback completed."
}

# Main build routine
main_build() {
    log "Starting consolidated build process..."
    check_docker
    start_docker
    pull_docker_images "openpanel/nginx:latest" "openpanel/apache:latest"
    docker_compose_up
    log "Build process completed."
}

# If script is executed directly, run the main build routine.
[[ "${BASH_SOURCE[0]}" == "$0" ]] && main_build
