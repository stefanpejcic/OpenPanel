#!/usr/bin/env bash
# Description:
# This patch fixes the 500 error on user pages when Services module is enabled.
# It is safe to run multiple times.
#
# Usage:
# This script is intended to be run via opencli patch:
#   opencli patch services-js
#
wget -O /tmp/file.html https://gist.githubusercontent.com/stefanpejcic/33b8c802d8aa91eeecfe39c5c0519124/raw/aeb59f090d8f61c7d18a47fb1534f9878d0a4bba/file.html

docker cp /tmp/file.html openpaanel:/templates/partials/_service_js.html

docker restart openpanel
