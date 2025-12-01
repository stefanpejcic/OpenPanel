#!/bin/bash
################################################################################
# Script Name: files/purge_trash.sh
# Description: Auto-purge .Trash folders for users.
# Usage: opencli files-purge_trash --user [USERNAME]
# Author: Stefan Pejcic
# Created: 03.06.2025
# Last Modified: 30.11.2025
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

CONFIG_FILE="/etc/openpanel/openpanel/conf/openpanel.config"
FORCE_PURGE=false
DRY_RUN=false
TARGET_USER=""
TMP_DIR="/tmp/trash_purge_$$"
mkdir -p "$TMP_DIR"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --force) FORCE_PURGE=true; shift ;;
    --dry-run) DRY_RUN=true; shift ;;
    --user)
      if [[ -n "$2" ]]; then
        TARGET_USER="$2"
        shift 2
      else
        echo "[ERROR] --user requires a username argument"
        exit 1
      fi
      ;;
    *)
      echo "[ERROR] Unknown argument: $1"
      exit 1
      ;;
  esac
done

if $FORCE_PURGE; then
  echo "[INFO] Running in FORCE mode: All trash will be purged."
fi

if $DRY_RUN; then
  echo "[INFO] DRY-RUN mode: No files will be deleted."
fi

if ! $FORCE_PURGE; then
  if [[ ! -f "$CONFIG_FILE" ]]; then
    echo "[ERROR] Config file not found: $CONFIG_FILE"
    exit 1
  fi

  autopurge_trash=$(grep -E "^autopurge_trash=" "$CONFIG_FILE" | cut -d'=' -f2)
  if [[ -z "$autopurge_trash" ]]; then
    echo "[INFO] autopurge_trash not set. Skipping cleanup."
    exit 0
  fi

  max_age_seconds=$((autopurge_trash * 86400))
  now_epoch=$(date +%s)
fi

purge_user_trash() {
  local user_home="$1"
  local user_name
  user_name=$(basename "$user_home")
  local trash_dir="$user_home/.local/share/Trash"
  local restore_file="$trash_dir/.trash_restore"
  local freed=0

  [[ -d "$trash_dir" ]] || return 0
  [[ -f "$restore_file" ]] || touch "$restore_file"

# FORCE MODE: Remove everything & recreate .trash_restore empty
if $FORCE_PURGE; then
  if $DRY_RUN; then
    freed=$(du -sb "$trash_dir" 2>/dev/null | cut -f1)
    echo "[DRY-RUN] Would purge ALL trash for user: $user_name"
  else
    freed=$(du -sb "$trash_dir" 2>/dev/null | cut -f1)
    shopt -s dotglob nullglob
    rm -rf -- "${trash_dir:?}"/*
    shopt -u dotglob
    touch "$restore_file"
  fi
  echo "$freed" > "$TMP_DIR/$user_name"
  return
fi

  # NORMAL & DRY-RUN MODE
  while IFS= read -r line; do
    [[ -z "$line" ]] && continue
    local file_name file_path deletion_date deletion_epoch age file_size

    # Extract parts
    file_name="${line%%=*}"                                      # name inside trash dir
    file_path="${line#*=}"                                       # original file path
    file_path="${file_path%%|deletion_date=*}"
    deletion_date="${line#*deletion_date=}"
    deletion_epoch=$(date -d "$deletion_date" +%s 2>/dev/null || echo 0)

    [[ $deletion_epoch -eq 0 ]] && continue

    age=$((now_epoch - deletion_epoch))

    if (( age > max_age_seconds )); then
      local trash_file="$trash_dir/$file_name"
      if [[ -e "$trash_file" ]]; then
        file_size=$(du -sb "$trash_file" 2>/dev/null | cut -f1)
      else
        file_size=0
      fi
      freed=$((freed + file_size))

      if $DRY_RUN; then
        echo "[DRY-RUN] Would delete: $trash_file (user: $user_name)"
      else
        rm -rf -- "$trash_file"
        # Remove line from .trash_restore (exact match)
        tmpfile="${restore_file}.tmp"
        grep -Fxv -- "$line" "$restore_file" > "$tmpfile" || true
        mv "$restore_file.tmp" "$restore_file"
      fi
    fi
  done < "$restore_file"

  echo "$freed" > "$TMP_DIR/$user_name"
}

users_to_process=()

if [[ -n "$TARGET_USER" ]]; then
  user_home="/home/$TARGET_USER"
  if [[ ! -d "$user_home" ]]; then
    echo "[ERROR] User home directory does not exist: $user_home"
    exit 1
  fi
  users_to_process+=("$user_home")
else
  for user_home in /home/*; do
    [[ -d "$user_home" ]] || continue
    users_to_process+=("$user_home")
  done
fi

for user_home in "${users_to_process[@]}"; do
  purge_user_trash "$user_home" &
done

wait

total_freed=0
echo ""
echo "🧹 Trash Cleanup Report:"
for stat_file in "$TMP_DIR"/*; do
  user=$(basename "$stat_file")
  bytes=$(cat "$stat_file")
  total_freed=$((total_freed + bytes))
  hr=$(numfmt --to=iec "$bytes")B
  if $DRY_RUN; then
    echo "  - $user: $hr (would be) freed"
  else
    echo "  - $user: $hr freed"
  fi
done

hr_total=$(numfmt --to=iec "$total_freed")B
echo "----------------------------------------"
if $DRY_RUN; then
  echo "🔍 Total that *would* be freed: $hr_total"
else
  echo "✅ Total freed across all users: $hr_total"
fi

rm -rf "$TMP_DIR"
