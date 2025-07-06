#!/bin/bash
################################################################################
# Script Name: version.sh
# Description: Displays the current (installed) version of OpenPanel docker image.
# Usage: opencli version 
# Author: Stefan Pejcic
# Created: 15.11.2023
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

# CHECK IMAGE AS A FALLBACK
check_images() {
    LOCAL_TAG=$(docker --context=default images --format "{{.Tag}}" "openpanel/openpanel-ui" | grep -E '^[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | tail -n 1)            
    if [ -n "$LOCAL_TAG" ]; then
        echo $LOCAL_TAG
    else
        echo '{"error": "OpenPanel UI docker image not detected."}' >&2
        exit 1
    fi
}

# CHECK ENV FILE FIRST
version_check() {
    if [ -f "/root/.env" ]; then
        image_version=$(grep "^VERSION=" /root/.env | sed -E 's/^VERSION="([^"]+)"$/\1/' | xargs)
        if [ -n "$image_version" ]; then
            echo $image_version
        else
            check_images
        fi
    else
        echo '{"error": "Docker image or .env files are missing."}' >&2
        exit 1
    fi
}

version_check
