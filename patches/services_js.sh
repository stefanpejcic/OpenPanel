#!/bin/bash
wget -O /tmp/file.html https://gist.githubusercontent.com/stefanpejcic/33b8c802d8aa91eeecfe39c5c0519124/raw/aeb59f090d8f61c7d18a47fb1534f9878d0a4bba/file.html
docker cp /tmp/file.html openpaanel:/templates/partials/_service_js.html
docker restart openpanel
