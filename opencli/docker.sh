#!/bin/bash
################################################################################
# Script Name: docker.sh
# Description: Manage OpenPanel system or user containers with lazydocker.
# Usage: opencli docker [<user> [<container>]]
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 07.03.2025
# Last Modified: 16.05.2026
# Company: openpanel.com
# Copyright (c) openpanel.com
# 
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
# 
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
# 
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.
################################################################################

set -euo pipefail

die() { echo "ERROR: $*" >&2; exit 1; }

ensure_fzf() {
  command -v fzf &>/dev/null && return
   echo "Installing fzf - please wait.."
  if command -v apt-get &>/dev/null; then
    apt-get install -y -qq fzf &>/dev/null
  elif command -v dnf &>/dev/null; then
    dnf install -y -q fzf &>/dev/null
  elif command -v yum &>/dev/null; then
    yum install -y -q fzf &>/dev/null
  else
    die "Cannot install fzf: no supported package manager found (apt/dnf/yum)"
  fi
  command -v fzf &>/dev/null || die "fzf installation failed"
}

ensure_fzf

get_socket() {
  local username="$1"
  if [ -z "$username" ]; then
    echo "/var/run/docker.sock"
    return
  fi
  local context user_uid
  context=$(mysql -e "SELECT server FROM users WHERE username = '$username';" -sN)
  [ -n "$context" ] || die "No context found for username: $username"
  user_uid=$(id -u "$context" 2>/dev/null)
  [[ "$user_uid" =~ ^[0-9]+$ ]] || die "Invalid USER_UID: $user_uid"
  echo "/hostfs/run/user/${user_uid}/docker.sock"
}

check_socket() {
  local socket="$1" label="${2:-docker}"
  [ -e "$socket" ] || die "Docker socket not found: $socket ($label not running)"
}

# List running containers via a given socket
list_containers() {
  local socket="$1"
  DOCKER_HOST="unix://${socket}" docker ps --format '{{.Names}}\t{{.Image}}\t{{.Status}}' 2>/dev/null
}

# Pick a container with fzf; echoes the selected name
pick_container() {
  local socket="$1" prompt="${2:-container}"
  local selection
  selection=$(list_containers "$socket" | fzf --prompt="$prompt > " --header="Select a container (ESC to quit)" --layout=reverse --border) || return 1
  echo "$selection" | awk '{print $1}'
}

# Pick a user from the DB with fzf; echoes the selected username
pick_user() {
  local selection
  selection=$(mysql -e "SELECT username FROM users ORDER BY username;" -sN 2>/dev/null | fzf --prompt="user > " --header="Select a user (ESC to quit)" --layout=reverse --border) || return 1
  echo "$selection"
}

# Open an interactive shell in a container
exec_shell() {
  local socket="$1" container="$2"
  # Try bash first, fall back to sh
  DOCKER_HOST="unix://${socket}" docker exec -it "$container" bash 2>/dev/null || DOCKER_HOST="unix://${socket}" docker exec -it "$container" sh
}

# Main
case $# in

  # no args = pick user > pick container > shell
  0)
    USERNAME=$(pick_user) || die "No user selected"
    SOCKET=$(get_socket "$USERNAME")
    check_socket "$SOCKET" "rootless docker for $USERNAME"
    CONTAINER=$(pick_container "$SOCKET" "$USERNAME") || die "No container selected"
    exec_shell "$SOCKET" "$CONTAINER"
    ;;

  # <user> = pick container > shell
  1)
    USERNAME="$1"
    SOCKET=$(get_socket "$USERNAME")
    check_socket "$SOCKET" "rootless docker for $USERNAME"
    CONTAINER=$(pick_container "$SOCKET" "$USERNAME") || die "No container selected"
    exec_shell "$SOCKET" "$CONTAINER"
    ;;

  # <user> <container> = shell directly
  2)
    USERNAME="$1"
    CONTAINER="$2"
    SOCKET=$(get_socket "$USERNAME")
    check_socket "$SOCKET" "rootless docker for $USERNAME"
    exec_shell "$SOCKET" "$CONTAINER"
    ;;

  *)
    echo "Usage: opencli docker [<user> [<container>]]" >&2
    exit 1
    ;;

esac
