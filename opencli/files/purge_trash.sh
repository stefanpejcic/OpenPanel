#!/bin/bash
################################################################################
# Script Name: files/purge_trash.sh
# Description: Auto-purge .Trash folders for users.
# Usage: opencli files-purge_trash --user [USERNAME]
# Author: Stefan Pejcic
# Created: 03.06.2025
# Last Modified: 04.07.2025
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
  local trash_info_dir="$user_home/.local/share/Trash/info"
  local trash_files_dir="$user_home/.local/share/Trash/files"
  local freed=0

  [[ -d "$trash_info_dir" ]] || exit 0

  for info_file in "$trash_info_dir"/*.trashinfo; do
    [[ -f "$info_file" ]] || continue

    file_base=$(basename "$info_file" .trashinfo)
    file_path="$trash_files_dir/$file_base"

    if [[ -e "$file_path" ]]; then
      file_size=$(du -sb "$file_path" 2>/dev/null | cut -f1)
    else
      file_size=0
    fi

    should_delete=false

    if $FORCE_PURGE; then
      should_delete=true
    else
      deletion_date=$(grep "^DeletionDate=" "$info_file" | cut -d'=' -f2)
      deletion_epoch=$(date -d "$deletion_date" +%s 2>/dev/null)
      [[ -z "$deletion_epoch" ]] && continue
      age=$((now_epoch - deletion_epoch))
      if (( age > max_age_seconds )); then
        should_delete=true
      fi
    fi

    if $should_delete; then
      freed=$((freed + file_size))
      if $DRY_RUN; then
        echo "[DRY-RUN] Would delete: $file_base (user: $user_name)"
      else
        [[ "$file_path" == "$trash_files_dir/"* ]] && rm -rf -- "$file_path"
        rm -f -- "$info_file"
      fi
    fi
  done

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
  echo "  - $user: $hr ${DRY_RUN:+(would be) }freed"
done

hr_total=$(numfmt --to=iec "$total_freed")B
echo "----------------------------------------"
if $DRY_RUN; then
  echo "🔍 Total that *would* be freed: $hr_total"
else
  echo "✅ Total freed across all users: $hr_total"
fi

rm -rf "$TMP_DIR"
