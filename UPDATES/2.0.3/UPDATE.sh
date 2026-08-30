#!/bin/bash

# pm2 and docker modules are split, replace them in the enabled modules lists:
panel_config="/etc/openpanel/openpanel/conf/openpanel.config"
grep -q 'pm2' "$panel_config" && ! grep -qE 'python|nodejs|ruby' "$panel_config" && sed -i 's/pm2/python,nodejs,ruby/g' "$panel_config"
grep -q 'docker' "$panel_config" && ! grep -qE 'terminal|change_image' "$panel_config" && sed -i 's/docker/docker,terminal,change_image/g' "$panel_config"

# and then in every feature set that uses them
features_dir="/etc/openpanel/openpanel/features"
for f in "$features_dir"/*.txt; do
  [ -e "$f" ] || continue

  if grep -qx 'pm2' "$f" && ! grep -qE '^(python|nodejs|ruby)$' "$f"; then
    sed -i '/^pm2$/c\python\nnodejs\nruby' "$f"
  fi

  if grep -qx 'docker' "$f" && ! grep -qE '^(terminal|change_image)$' "$f"; then
    printf 'terminal\nchange_image\n' >> "$f"
  fi
done
