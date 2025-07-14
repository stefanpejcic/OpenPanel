#!/bin/bash

################################################################################
# Script Name: commands.sh
# Description: Lists all available OpenCLI commands.
# Usage: opencli commands
# Author: Stefan Pejcic
# Created: 15.11.2023
# Last Modified: 13.07.2025
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

# Colors
readonly GREEN='\033[0;32m'
readonly RESET='\033[0m'

# Files to exclude from command listing
readonly EXCLUDE_PATTERNS=(
    ".git/*"
    ".github/*"
    "error.py"
    "db.sh"
    "aliases.txt"
    "send_mail.sh"
    "LICENSE.md"
    "README.md"
    "*README.md"
    "*TODO*"
)

# Functions
check_scripts_directory() {
    if [[ ! -d "$SCRIPTS_DIR" ]]; then
        echo "Error: Scripts directory '$SCRIPTS_DIR' not found." >&2
        exit 1
    fi
}

initialize_alias_file() {
    # Clear existing aliases file
    > "$ALIAS_FILE"
    
    if [[ ! -w "$ALIAS_FILE" ]]; then
        echo "Error: Cannot write to alias file '$ALIAS_FILE'." >&2
        exit 1
    fi
}

build_find_excludes() {
    local exclude_args=()
    
    for pattern in "${EXCLUDE_PATTERNS[@]}"; do
        exclude_args+=("!" "-path" "${SCRIPTS_DIR}/${pattern}")
    done
    
    echo "${exclude_args[@]}"
}

extract_script_info() {
    local script="$1"
    local -n info_ref=$2
    
    # Extract description and usage from script comments
    info_ref[description]=$(grep -E "^# Description:" "$script" 2>/dev/null | sed 's/^# Description: //' || true)
    info_ref[usage]=$(grep -E "^# Usage:" "$script" 2>/dev/null | sed 's/^# Usage: //' || true)
}

generate_alias_name() {
    local script="$1"
    local script_name dir_name alias_name
    
    # Remove file extensions
    script_name=$(basename "$script" | sed 's/\(\.sh\|\.py\)$//')
    
    # Get directory name without full path
    dir_name=$(dirname "$script" | sed 's:.*/::')
    
    # Handle root opencli directory
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
    
    echo -e "${GREEN}${full_alias}${RESET} # for ${script}"
    
    if [[ -n "${script_info[description]:-}" ]]; then
        echo "Description: ${script_info[description]}"
    fi
    
    if [[ -n "${script_info[usage]:-}" ]]; then
        echo "Usage: ${script_info[usage]}"
    fi
    
    echo "------------------------"
}

process_scripts() {
    local exclude_args
    declare -a commands_list=()
    
    exclude_args=$(build_find_excludes)
    
    # Use process substitution to avoid subshell issues
    while IFS= read -r -d '' script; do
        if [[ ! -r "$script" ]]; then
            echo "Warning: Cannot read script '$script'" >&2
            continue
        fi
        
        local alias_name full_alias
        alias_name=$(generate_alias_name "$script")
        full_alias="opencli $alias_name"
        
        # Display command information
        display_command_info "$full_alias" "$script"
        
        # Collect commands for sorting
        commands_list+=("$full_alias")
        
    done < <(find "$SCRIPTS_DIR" -type f ${exclude_args[@]} -print0)
    
    # Write sorted commands to alias file
    if [[ ${#commands_list[@]} -gt 0 ]]; then
        printf '%s\n' "${commands_list[@]}" | sort > "$ALIAS_FILE"
    fi
}

show_summary() {
    local command_count
    
    if [[ -f "$ALIAS_FILE" ]]; then
        command_count=$(wc -l < "$ALIAS_FILE")
        echo ""
        echo "Total commands found: $command_count"
        echo "Aliases saved to: $ALIAS_FILE"
    fi
}

main() {
    check_scripts_directory
    initialize_alias_file
    process_scripts
    show_summary
}

# Script entry point
main "$@"
