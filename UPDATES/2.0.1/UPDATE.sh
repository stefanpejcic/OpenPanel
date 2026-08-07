#!/bin/bash

: '
Symlinks each domains email storage directory into that users mail Docker volume.
This ensures mailbox data is included in backups of the users docker-data volumes.
'


# 0. ensure 'mysql' command exists and create symlink
if ! command -v mysql >/dev/null 2>&1; then
    if command -v mariadb >/dev/null 2>&1; then
        ln -sf "$(command -v mariadb)" /usr/local/bin/mysql
        echo "Created 'mysql' -> 'mariadb' symlink for compatibility."
    fi
fi

# 1. check if Enteprise
key_value=$(grep "^key=" "/etc/openpanel/openpanel/conf/openpanel.config" | cut -d'=' -f2-)
[ -z "$key_value" ] && { echo "enterprise license not present - nothing to do."; exit 0; }


# 2. check email server is installed
[ -f "/usr/local/mail/openmail/compose.yml" ] || { echo "email server not installed - nothing to do."; exit 0; }


# 3. check storage location
email_storage_path=$(grep -E '^email_storage_location=' /etc/openpanel/openadmin/config/admin.ini | cut -d'=' -f2- | xargs)
[[ "$email_storage_path" == /* ]] || { echo "email_storage_location is not an absolute path (value: '$email_storage_path') - nothing to link."; exit 0; }


# 4. list users
source /usr/local/opencli/db.sh
users=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -e "
SELECT user_id, server FROM users;
")

created=0
updated=0
skipped_not_symlink=0

while IFS=$'\t' read -r user_id server; do
    [[ -z "$user_id" ]] && continue
    [[ -z "$server" ]] && continue


    # 5. Get all domains for this user
    domain_urls=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -e "
    SELECT GROUP_CONCAT(domain_url) FROM domains WHERE user_id='$user_id';
    ")
    [[ -z "$domain_urls" ]] && continue
    IFS=',' read -ra domains_array <<< "$domain_urls"


    # 6. symlink existing domain docroot to mail volume
    for domain in "${domains_array[@]}"; do
        [[ -z "$domain" ]] && continue

        source_dir="$email_storage_path/$domain"
        target_link="/home/$server/docker-data/volumes/${server}_mail_data/_data/$domain"

        [[ -d "$source_dir" ]] || mkdir -p "$source_dir"

        mkdir -p "$(dirname "$target_link")"

        if [[ -L "$target_link" ]]; then
            current_target=$(readlink -f "$target_link")
            if [[ "$current_target" != "$(readlink -f "$source_dir")" ]]; then
                ln -sfn "$source_dir" "$target_link"; echo "Updated symlink: $target_link -> $source_dir"; ((updated++))
            fi
        elif [[ -e "$target_link" ]]; then
            echo "Target exists and is not a symlink, skipping: $target_link"; ((skipped_not_symlink++))
        else
            ln -s "$source_dir" "$target_link"; echo "Created symlink: $target_link -> $source_dir"; ((created++))
        fi
    done
done <<< "$users"

echo "----------------------------------------"
echo "Summary: created=$created updated=$updated skipped_not_symlink=$skipped_not_symlink"
