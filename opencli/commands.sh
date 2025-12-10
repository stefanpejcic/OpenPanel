#!/bin/bash
################################################################################
# Script Name: commands.sh
# Description: Lists all available OpenCLI commands.
# Usage: opencli commands
# Author: Stefan Pejcic
# Created: 15.11.2023
# Last Modified: 09.12.2025
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

# Constants
readonly SCRIPTS_DIR="/usr/local/opencli"
readonly ALIAS_FILE="${SCRIPTS_DIR}/aliases.txt"
readonly GREEN='\033[0;32m'
readonly RESET='\033[0m'

# Exclude
readonly EXCLUDE_PATTERNS=(
    ".git/*"
    ".github/*"
    "ftp/users.sh"
    "db.sh"
    "enterprise.sh"
    "ip_servers.sh"
    "progress_bar.sh"
)

# Functions
check_scripts_directory() {
    if [[ ! -d "$SCRIPTS_DIR" ]]; then
        echo "Error: Scripts directory '$SCRIPTS_DIR' not found." >&2
        exit 1
    fi
}

initialize_alias_file() {
    > "$ALIAS_FILE"

    if [[ ! -w "$ALIAS_FILE" ]]; then
        echo "Error: Cannot write to alias file '$ALIAS_FILE'." >&2
        exit 1
    fi
}

build_find_excludes() {
    exclude_args=()
    for pattern in "${EXCLUDE_PATTERNS[@]}"; do
        exclude_args+=("!" "-path" "${SCRIPTS_DIR}/${pattern}")
    done
}

extract_script_info() {
    local script="$1"
    local -n info_ref=$2
    info_ref[description]=$(grep -E "^# Description:" "$script" 2>/dev/null | sed 's/^# Description: //' || true)
    info_ref[usage]=$(grep -E "^# Usage:" "$script" 2>/dev/null | sed 's/^# Usage: //' || true)
}

generate_alias_name() {
    local script="$1"
    local script_name dir_name alias_name
    
    script_name=$(basename "$script" | sed 's/\(\.sh\|\.py\)$//')
    dir_name=$(dirname "$script" | sed 's:.*/::')
    
    if [[ "$dir_name" == "opencli" ]]; then
        alias_name="$script_name"
    else
        alias_name="${dir_name}-${script_name}"
    fi
    
    echo "$alias_name"
}

display_command_info() {
    local full_alias="$1"
    local script="$2"
    declare -A script_info
    extract_script_info "$script" script_info
    
    echo -e "${GREEN}${full_alias}${RESET}"
    if [[ -n "${script_info[description]:-}" ]]; then
        echo "Description: ${script_info[description]}"
    fi
    if [[ -n "${script_info[usage]:-}" ]]; then
        echo "Usage: ${script_info[usage]}"
    fi
    echo "------------------------"
}

process_scripts() {
    declare -a exclude_args
    declare -a commands_list=()
    build_find_excludes

    while IFS= read -r -d '' script; do
        if [[ ! -r "$script" ]]; then
            echo "Warning: Cannot read script '$script'" >&2
            continue
        fi

        local alias_name full_alias
        alias_name=$(generate_alias_name "$script")
        full_alias="opencli $alias_name"
        display_command_info "$full_alias" "$script"
        commands_list+=("$full_alias")
        
    done < <(find "$SCRIPTS_DIR" -type f -name "*.sh" "${exclude_args[@]}" -print0)

    # special cases
    echo -e "${GREEN}opencli error${RESET}"
    echo "Description: Displays information for specific error ID received in OpenPanel UI."
    echo "Usage: opencli error <ID_HERE>"
    echo "------------------------"
    commands_list+=("opencli error")

    echo -e "${GREEN}opencli locale${RESET}"
    echo "Description: Install locales (Languages) for OpenPanel UI."
    echo "Usage: opencli locale <CODE>"
    echo "------------------------"
    commands_list+=("opencli locale")

    : '
    echo -e "${GREEN}opencli terms${RESET}"
    echo "Description: View OpenPanel Terms and Conditions."
    echo "Usage: opencli terms"
    echo "------------------------"
    commands_list+=("opencli terms")

    echo -e "${GREEN}opencli eula${RESET}"
    echo "Description: End User License Agreement (EULA)."
    echo "Usage: opencli eula"
    echo "------------------------"
    commands_list+=("opencli eula")
    '
    
    if [[ ${#commands_list[@]} -gt 1 ]]; then
        printf '%s\n' "${commands_list[@]}" | sort > "$ALIAS_FILE"
    fi
}

main() {
    check_scripts_directory
    initialize_alias_file
    process_scripts
}

main "$@"
