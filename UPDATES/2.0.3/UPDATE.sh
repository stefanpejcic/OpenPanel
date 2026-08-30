#!/bin/bash

# pm2 module is split
panel_config="/etc/openpanel/openpanel/conf/openpanel.config"
grep -q 'pm2' "$panel_config" && ! grep -qE 'python|nodejs|ruby' "$panel_config" && sed -i 's/pm2/python,nodejs,ruby/g' "$panel_config"

# docker module is split
grep -q 'docker' "$panel_config" && ! grep -qE 'terminal|change_image' "$panel_config" && sed -i 's/docker/docker,terminal,change_image/g' "$panel_config"
