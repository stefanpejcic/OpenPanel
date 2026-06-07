#!/bin/bash
################################################################################
# Script Name: docker.sh
# Description: Manage OpenPanel system or user containers with lazydocker.
# Usage: opencli docker [<user> [<container>]]
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 07.03.2025
# Last Modified: 06.06.2026
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

get_context() {
  local input="$1"
  [ -z "$input" ] && { echo "default"; return; }
  local username="${input##*_}"
  local esc_username="${username//\'/\'\'}"

  local context
  context=$(
    mysql -Nse "
      SELECT server FROM users WHERE username = '${username}'
      UNION ALL
      SELECT server FROM users WHERE username LIKE 'SUSPENDED#_%#_${username}' ESCAPE '#'
      LIMIT 1;
    " 2>/dev/null
  )

  [ -n "$context" ] || die "No docker context found for username: $username"
  echo "$context"
}

list_containers() {
  local context="$1"
  docker --context="$context" ps --format '{{.Names}}\t{{.Image}}\t{{.Status}}' 2>/dev/null
}

pick_container() {
  local context="$1" prompt="${2:-container}"
  local selection
  selection=$(list_containers "$context" | fzf --prompt="$prompt > " --header="Select a container (ESC to quit)" --layout=reverse --border) || return 1
  echo "$selection" | awk '{print $1}'
}

pick_user() {
  local selection
  selection=$(mysql -e "SELECT username FROM users ORDER BY username;" -sN 2>/dev/null | sed 's/.*_//' | fzf --prompt="user > " --header="Select a user (ESC to quit)" --layout=reverse --border) || return 1
  echo "$selection"
}

exec_shell() {
  local context="$1" container="$2"
  docker --context="$context" exec -it "$container" bash 2>/dev/null || docker --context="$context" exec -it "$container" sh
}

# Main
case $# in
  # no args = pick user > pick container > shell
  0)
    USERNAME=$(pick_user) || die "No user selected"
    CONTEXT=$(get_context "$USERNAME")
    CONTAINER=$(pick_container "$CONTEXT" "$USERNAME") || die "No container selected"
    exec_shell "$CONTEXT" "$CONTAINER"
    ;;
  # <user> = pick container > shell
  1)
    USERNAME="$1"
    CONTEXT=$(get_context "$USERNAME")
    CONTAINER=$(pick_container "$CONTEXT" "$USERNAME") || die "No container selected"
    exec_shell "$CONTEXT" "$CONTAINER"
    ;;
  # <user> <container> = shell directly
  2)
    USERNAME="$1"
    CONTAINER="$2"
    CONTEXT=$(get_context "$USERNAME")
    exec_shell "$CONTEXT" "$CONTAINER"
    ;;
  *)
    echo "Usage: opencli docker [<user> [<container>]]" >&2
    exit 1
    ;;
esac
