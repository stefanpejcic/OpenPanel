#!/bin/bash
################################################################################
# Script Name: dev.sh
# Description: Overwrite OpenPanel container files and restart service to apply.
# Usage: opencli dev [path]
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 07.03.2025
# Last Modified: 11.02.2026
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

set -e

# ======================================================================
# Constants
readonly CONTAINER_NAME="openpanel"
readonly DOCKER_CONTEXT="default"
readonly LAST_PATH_FILE="/tmp/last_dev_path"
readonly SUPPORTED_EXTENSIONS=("py" "html")

readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly NC='\033[0m'

# ======================================================================
# Helpers
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

usage() {
    cat << EOF
Usage: opencli dev [path]

Description:
    This script is intended for development purposes only to overwrite system files 
    in the OpenPanel UI container. It will open the file in the nano editor, validate 
    its contents, and then copy the changes back to the container. Afterwards, it 
    restarts the OpenPanel container to pick up the new file and follow the logs.

Options:
    path    The path to a file inside the OpenPanel Docker container (either .py or .html).
            If no path is provided, the script will display available files for you to choose.

Examples:
    1. Edit a Python module by providing its path:
       opencli dev modules/ssh.py
    
    2. Edit an HTML template by providing its path:
       opencli dev templates/base.html
    
    3. Interactively select a file to edit:
       opencli dev

EOF
    exit 1
}

command_exists() {
    command -v "$1" &> /dev/null
}



install_tool() {
    local tool="$1"
    
    install_package() {
        local package="$1"
        log_info "Attempting to install $package..."
        
        if command_exists apt; then
            apt install -y "$package" > /dev/null 2>&1
        elif command_exists dnf; then
            dnf install -y "$package" > /dev/null 2>&1
        elif command_exists yum; then
            yum install -y "$package" > /dev/null 2>&1
        else
            log_error "No supported package manager found (apt/dnf/yum)"
            return 1
        fi
    }
    
    if ! command_exists "$tool"; then
        if ! install_package "$tool"; then
            log_error "Failed to install $tool. Please install it manually."
            exit 1
        fi
        log_info "$tool installed successfully"
    fi
}

is_container_running() {
    docker --context "$DOCKER_CONTEXT" ps --format "table {{.Names}}" | grep -q "^$CONTAINER_NAME$"
}

get_all_available_files() {
    local available_files
    
    if ! is_container_running; then
        log_error "OpenPanel container is not running!"
        exit 1
    fi
    
    available_files=$(docker --context "$DOCKER_CONTEXT" exec "$CONTAINER_NAME" \
        find / -maxdepth 3 \( -name "app.py" -o -path "/modules/*.py" -o -path "/templates/*.html" \) \
        2>/dev/null | sort | uniq)
    
    if [ -z "$available_files" ]; then
        log_error "No files found in the container!"
        exit 1
    fi

    local last_path=""
    if [ -f "$LAST_PATH_FILE" ]; then
        last_path=$(cat "$LAST_PATH_FILE")
    fi
    
    path=$(echo "$available_files" | fzf --prompt="Select a file: " --query="$last_path")
    
    if [ -z "$path" ]; then
        log_warn "No file selected."
        exit 1
    fi

    echo "$path" > "$LAST_PATH_FILE"
    log_info "Selected file: $path"
}


is_supported_extension() {
    local extension="$1"
    local supported_ext
    
    for supported_ext in "${SUPPORTED_EXTENSIONS[@]}"; do
        if [[ "$extension" == "$supported_ext" ]]; then
            return 0
        fi
    done
    return 1
}

validate_file() {
    local file="$1"
    local ext="$2"

    case "$ext" in
        py)
            grep -q "import" "$file" \
                || { log_error "Invalid Python file: missing import statements"; return 1; }
            ;;
        html)
            grep -qi "<div" "$file" \
                || { log_error "Invalid HTML file: missing HTML tags"; return 1; }
            ;;
        *)
            log_error "Unsupported file extension: $ext"
            return 1
            ;;
    esac
}


copy_to_container() {
    local source_file="$1"
    local target_path="$2"
    
    if ! docker --context "$DOCKER_CONTEXT" cp "$source_file" "$CONTAINER_NAME:/$target_path"; then
        log_error "Failed to copy file to container"
        return 1
    fi
    
    log_info "File copied to container successfully"
    return 0
}

restart_container_and_follow_logs() {
    log_info "Restarting OpenPanel container to pick up the new file..."
    
    if ! docker --context "$DOCKER_CONTEXT" restart "$CONTAINER_NAME"; then
        log_error "Failed to restart container"
        return 1
    fi
    
    log_info "Following new logs:"
    docker --context "$DOCKER_CONTEXT" logs --follow --since=0s "$CONTAINER_NAME"
}

process_file() {
    local file_path="$1"
    local extension
    local tmpfile

    extension="${file_path##*.}"
    
    if ! is_supported_extension "$extension"; then
        log_error "Unsupported file extension: $extension"
        log_error "Supported extensions: ${SUPPORTED_EXTENSIONS[*]}"
        exit 1
    fi
    
    tmpfile=$(mktemp)
    
    # cleanup tmpfile
    cleanup_tmpfile() {
        rm -f "$tmpfile"
    }
    trap cleanup_tmpfile EXIT
    
    # open nano editor
    nano "$tmpfile"
    
    # abort if empty
    if [ ! -s "$tmpfile" ]; then
        log_error "Aborting: The file is empty!"
        exit 1
    fi
    
    # validate HTML / PYTHON
    if ! validate_file "$tmpfile" "$extension"; then
        log_error "File validation failed"
        exit 1
    fi
    
    # copy to container
    if ! copy_to_container "$tmpfile" "$file_path"; then
        exit 1
    fi

    # enable dev_mode
    opencli config update dev_mode on  > /dev/null 2>&1

    # restart and tail
    restart_container_and_follow_logs
}


# ======================================================================
main() {
    local file_path="$1"
    
    if [ $# -gt 0 ]; then
        if [ -z "$file_path" ]; then
            log_error "Invalid path provided!"
            usage
        fi
    else
        install_tool fzf
        install_tool nano
        get_all_available_files
        file_path="$path"
    fi
    
    process_file "$file_path"
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
