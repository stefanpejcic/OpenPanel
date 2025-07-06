#!/bin/bash
################################################################################
# Script Name: dev.sh
# Description: Overwrite OpenPanel contianer files and restart service to apply.
# Usage: opencli dev [path]
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 07.03.2025
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


usage() {
    echo "Usage: opencli dev [path]"
    echo ""
    echo "Description:"
    echo "This script is intended for development purposes only to overwrite system files in the OpenPanel UI container."
    echo "It will open the file in the nano editor, validate its contents, and then copy the changes back to the container."
    echo "Afterwards, it restarts the OpenPanel container to pick up the new file and follow the logs."
    echo ""
    echo "Options:"
    echo "  path  The path to a file inside the OpenPanel Docker container (either .py or .html)."
    echo "        If no path is provided, the script will display available files for you to choose."
    echo ""
    echo "Examples:"
    echo "  1. Edit a Python module by providing its path:"
    echo "     opencli dev modules/ssh.py"
    echo "  2. Edit an HTML template by providing its path:"
    echo "     opencli dev templates/base.html"
    echo "  2. Interactively select a file to edit:"
    echo "     opencli dev"
    echo ""
    exit 1
}



install_fzf() {
    if ! command -v fzf &> /dev/null; then
        echo "Attempting to install fzf..."
        apt install -y fzf > /dev/null 2>&1 || dnf install -y fzf
        if ! command -v fzf &> /dev/null; then
            echo "Failed to install fzf. Please install it manually."
            exit 1
        fi
    fi   
}

install_nano() {
    if ! command -v nano &> /dev/null; then
        echo "Attempting to install nano editor..."
        apt install -y nano > /dev/null 2>&1 || dnf install -y nano
        if ! command -v nano &> /dev/null; then
            echo "Failed to install nano. Please install it manually."
            exit 1
        fi
    fi   
}


get_all_available_files(){
    available_files=$(docker --context default exec openpanel find / -maxdepth 3 \( -name "app.py" -o -path "/modules/*.py" -o -path "/templates/*.html" \) 2>/dev/null | sort | uniq)
    if [ -z "$available_files" ]; then
      echo "No files found in the container! is openpanel running?"
      usage
      exit 1
    fi

    # Check if last used path exists
    last_path=""
    if [ -f /tmp/last_dev_path ]; then
        last_path=$(cat /tmp/last_dev_path)
    fi
    
    path=$(echo "$available_files" | fzf --prompt="Select a file: " --query="$last_path")
    
    if [ -z "$path" ]; then
        echo "No file selected."
        usage
        exit 1
    fi

    echo "$path" > /tmp/last_dev_path
}



if [ $# -gt 0 ]; then
    path="$1"
    if [ -z "$path" ]; then
        echo "Invalid path!"
        usage
        exit 1
    fi
else
    install_fzf
    install_nano
    get_all_available_files
fi


TMPFILE=$(mktemp)


if [[ "$path" =~ \.py$ || "$path" =~ \.html$ ]]; then

# Validate the path
  nano $TMPFILE

  # Check if the file is empty
  if [ ! -s "$TMPFILE" ]; then
    echo "Aborting: The file is empty!"
    rm "$TMPFILE"
    exit 1
  fi


  # Validate Python files
  if [[ "$path" == *.py ]]; then
    if ! grep -q "import" "$TMPFILE"; then
      echo "Aborting: you are trying to insert HTML code inside a Python file: $path"
      rm "$TMPFILE"
      exit 1
    fi
  fi

  # Validate HTML files
  if [[ "$path" == *.html ]]; then
    if ! grep -qi "<div" "$TMPFILE"; then
      echo "Aborting: you are trying to insert python code inside an HTML file: $path"
      rm "$TMPFILE"
      exit 1
    fi
  fi


  docker --context default cp $TMPFILE openpanel:/"$path" #> /dev/null 2>&1
  rm $TMPFILE
  echo "Restarting OpenPanel container to pickup the new file.."
  docker --context default restart openpanel
  echo "Following new logs:"
  docker --context default logs --follow --since=0s openpanel
else

    echo "invalid file!"
    usage
fi

exit 0
