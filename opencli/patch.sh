#!/bin/bash

################################################################################
# Script Name: patch.sh
# Description: Download and install a patch
# Usage: opencli patch <NAME>
# Author: Stefan Pejcic
# Created: 05.11.2025
# Last Modified: 18.02.2026
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

REPO_RAW_BASE="https://raw.githubusercontent.com/stefanpejcic/OpenPanel/main/patches"

usage() {
  echo "Usage: opencli patch <patch_name>   # patch_name may be given with or without .sh"
  exit 2
}

if [ $# -lt 1 ]; then
  usage
fi

PATCH_RAW="${1}"
if [[ "${PATCH_RAW}" != *.sh ]]; then
  PATCH_RAW="${PATCH_RAW}.sh"
fi

URL="${REPO_RAW_BASE}/${PATCH_RAW}"

TMPFILE="$(mktemp --suffix=.sh /tmp/patch.XXXXXX)" || TMPFILE="/tmp/patch.$$"
cleanup() {
  rm -f "${TMPFILE}" || true
}
trap cleanup EXIT

if ! curl -4sSfL --max-redirs 5 -o "${TMPFILE}" "${URL}"; then
  echo "Error: patch not found at ${URL}" >&2
  exit 3
fi

DESCRIPTION="$(awk '
  {
    # Match lines starting with "#" but NOT "#!"
    if ($0 ~ /^[[:space:]]*#[[:space:]]*[^!]/) {
      sub(/^[[:space:]]*#[[:space:]]?/, "")
      print
    } else {
      exit
    }
  }
' "${TMPFILE}")"

if [ -z "${DESCRIPTION}" ]; then
  echo
  echo "No description found in the patch file."
else
  echo
  echo "----- PATCH DESCRIPTION -----"
  echo "${DESCRIPTION}"
  echo "-----------------------------"
fi

echo
cat <<'WARN'
WARNING: This will download and execute the patch, 
it may cause downtime for the OpenPanel and OpenAdmin UI,
but will NOT affect user websites or data.

To proceed, you must type exactly: yes
WARN

read -r -p "Run this patch now? Type 'yes' to continue: " CONFIRM
if [ "${CONFIRM}" != "yes" ]; then
  echo "Aborted by user."
  exit 0
fi

chmod +x "${TMPFILE}"
echo "Running patch..."
if bash "${TMPFILE}"; then
  echo "Patch applied successfully."
  exit 0
else
  rc=$?
  echo "Patch exited with status ${rc}." >&2
  exit "${rc}"
fi

