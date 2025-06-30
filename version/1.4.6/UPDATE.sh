#!/bin/bash

wget -O /etc/openpanel/openadmin/config/features.json https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/config/features.json

# add [FILES]

config_file="/etc/openpanel/openpanel/conf/openpanel.config"
backup_file="${config_file}.1.4.6"
cp "$config_file" "$backup_file"

if grep -q "^activity_items_per_page=" "$config_file" && ! grep -q "^activity_lines_retention=" "$config_file"; then
    sed -i '/^activity_items_per_page=/a activity_lines_retention=1500' "$config_file"
    echo "Inserted activity_lines_retention=1500 after activity_items_per_page"
else
    echo "Either activity_lines_retention already exists or activity_items_per_page not found. No changes made."
fi


# Default [FILES] block content as associative array
declare -A files_defaults=(
  [autopurge_trash]=7
  [filemanager_edit_size]=5
  [filemanager_view_size]=5
  [filemanager_download_size]=2000
  [filemanager_upload_size]=2000
  [filemanager_compress_max_time]=5
  [filemanager_extract_max_time]=5
  [filemanager_download_max_time]=60
  [filemanager_edit_extensions]="\".txt .md error_log .log env gitconfig cfg htaccess .ini .php .sh .html .json .htm .html5 .xml .py .php5 .php7 .php8 .sql .css .js .conf\""
  [filemanager_image_extensions]="\".jpg .jpeg .png .gif .webp .avif\""
  [filemanager_archives_extensions]="\".zip .tar .gz .tar.gz\""
)

# Check if [FILES] section exists
if ! grep -q '^\[FILES\]' "$config_file"; then
  echo "[FILES]" >> "$config_file"
  for key in "${!files_defaults[@]}"; do
    echo "$key=${files_defaults[$key]}" >> "$config_file"
  done
  echo "[FILES] section added with default values."
  exit 0
fi

# Extract line numbers of [FILES] section start and next section start
start_line=$(grep -n '^\[FILES\]' "$config_file" | cut -d: -f1)
next_section_line=$(tail -n +$((start_line + 1)) "$config_file" | grep -n '^\[' | head -n1 | cut -d: -f1)

if [[ -z "$next_section_line" ]]; then
  end_line=$(wc -l < "$config_file")
else
  end_line=$((start_line + next_section_line - 1))
fi

# Get the block content lines (excluding section header)
block_lines=$(sed -n "$((start_line + 1)),$((end_line))p" "$config_file")

# For each key check if it exists inside the block
missing_keys=()
for key in "${!files_defaults[@]}"; do
  if ! echo "$block_lines" | grep -q "^$key="; then
    missing_keys+=("$key")
  fi
done

if [[ ${#missing_keys[@]} -eq 0 ]]; then
  echo "All keys already present in [FILES] section."
  exit 0
fi

# Insert missing keys after the [FILES] block end line, line by line
for key in "${missing_keys[@]}"; do
  # Escape slashes, ampersands, and backslashes for sed
  safe_value=$(printf '%s' "${files_defaults[$key]}" | sed -e 's/[\/&\\]/\\&/g')
  sed -i "${end_line}a\\
$key=${safe_value}
" "$config_file"
  ((end_line++)) # increment end_line because file grows
done

echo "Added missing keys to [FILES] section:"
for key in "${missing_keys[@]}"; do
  echo " - $key"
done

