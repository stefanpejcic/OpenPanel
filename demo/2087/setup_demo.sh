#!/bin/bash

# this script is used for creating demo.openpanel.org on every new version release
# it installs latest version of openpanel, creates dummy accounts with data, and finally adds them on the login pages.
#
# todo: edit existing restore task to use new snapshot and droplet id's.
#


setup_admin_panel() {
  echo "Creating demo admin user"
  wget -O /tmp/generate.sh https://gist.githubusercontent.com/stefanpejcic/905b7880d342438e9a2d2ffed799c8c6/raw/a1cdd0d2f7b28f4e9c3198e14539c4ebb9249910/random_username_generator_docker.sh > /dev/null 2>&1
  source /tmp/generate.sh
  new_username=($random_name)
  new_password=$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 16) 
  sqlite3 /etc/openpanel/openadmin/users.db "CREATE TABLE IF NOT EXISTS user (id INTEGER PRIMARY KEY, username TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL, role TEXT NOT NULL DEFAULT 'user', is_active BOOLEAN DEFAULT 1 NOT NULL);"  > /dev/null 2>&1 && 
  
  file_path="/usr/local/admin/templates/login.html"
  
  opencli admin new "$new_username" "$new_password"  > /dev/null 2>&1 && 
  echo " "
  echo "Createad admin user and set data on login form:"
  echo "Username: $new_username"
  echo "Password: $new_password"
  echo " "
  
    # set the data on login form
  sed -i -e "s/Type the Username and Password and click Signin./Type the Username <code>$new_username<\/code> and Password <code>$new_password<\/code> and click Signin./" \
         -e "s/<input type=\"text\" class=\"form-control\" name=\"username\" placeholder=\"admin\" autocomplete=\"off\" required=\"\" autofocus>/<input type=\"text\" class=\"form-control\" name=\"username\" placeholder=\"admin\" autocomplete=\"off\" required=\"\" autofocus value=\"$new_username\">/" \
         -e "s/<input type=\"password\" name=\"password\" class=\"form-control\" placeholder=\"\*\*\*\*\*\*\*\*\" autocomplete=\"off\">/<input type=\"password\" name=\"password\" class=\"form-control\" placeholder=\"\*\*\*\*\*\*\*\*\" autocomplete=\"off\" value=\"$new_password\">/" \
      $file_path

    echo "Restarting admin service for 2087"
    service admin restart
}


setup_user_panel(){
  generae_pass=$(opencli user-password stefan random)
  new_password=$(echo "$generae_pass" | grep "new generated password is:" | awk '{print $NF}')
  
  echo "Generated password: $new_password"


  escaped_password=$(printf '%s\n' "$new_password" | sed -e 's/[\/&]/\\&/g')
  echo "Escaped password: $escaped_password"

  file_path="/usr/local/panel/templates/user/login.html"
  
  # Prepare sed commands
  sed_command_username="s|<input type=\"text\" id=\"username\" name=\"username\" required class=\"form-control\" placeholder=\"{{ _('Enter your panel username') }}\" autofocus>|<input type=\"text\" id=\"username\" name=\"username\" required class=\"form-control\" placeholder=\"{{ _('Enter your panel username') }}\" autofocus value=\"stefan\">|"
  sed_command_password="s|<input type=\"password\" id=\"password\" name=\"password\" class=\"form-control\" required placeholder=\"{{ _('Enter your password') }}\">|<input type=\"password\" id=\"password\" name=\"password\" class=\"form-control\" required placeholder=\"{{ _('Enter your password') }}\" value=\"$escaped_password\">|"

  echo "Sed command for username: $sed_command_username"
  echo ""

  echo "Sed command for password: $sed_command_password"
  echo ""
  docker exec openpanel sed -i "$sed_command_username" "$file_path"
  docker exec openpanel sed -i "$sed_command_password" "$file_path"
      echo ""
      echo "Restarting docker container for 2083 panel.."
    docker restart openpanel

}
  


write_fake_data(){
  echo "Creating dummy data.."
  file_path="/etc/openpanel/openadmin/usage_stats.json"
  today=$(date +%Y-%m-%d)
  
  dates=()
  for i in {4..0}; do
      dates+=($(date -d "$today - $i day" +%Y-%m-%d))
  done
    
  printf '{"timestamp": "%s", "users": 0, "domains": 0, "websites": 0}\n' "${dates[0]}" > "$file_path"
  
  printf '{"timestamp": "%s", "users": 1, "domains": 1, "websites": 0}\n' "${dates[1]}" >> "$file_path"
  printf '{"timestamp": "%s", "users": 1, "domains": 2, "websites": 2}\n' "${dates[2]}" >> "$file_path"
  printf '{"timestamp": "%s", "users": 2, "domains": 3, "websites": 4}\n' "${dates[3]}" >> "$file_path"
  printf '{"timestamp": "%s", "users": 2, "domains": 3, "websites": 4}\n' "${dates[4]}" >> "$file_path"


  echo "Usage stats JSON data written to $file_path"

}



get_droplet_id() {
  droplet_id=$(curl http://169.254.169.254/metadata/v1/id)
  echo "droplet id: $droplet_id"
}



echo "Creating dummy accounts for demo.."

# real user for 2083 demo
opencli user-add stefan random stefan@pejcic.rs ubuntu_nginx_mysql
cd /root && docker compose up -d nginx
# ln -s /etc/nginx/sites-available/*.conf /etc/nginx/sites-enabled/
opencli domains-add openpanel.org stefan
opencli domains-add pejcic.rs stefan
opencli domains-add example.net stefan
opencli domains-add demo.openpanel.org stefan

# todo: install wp on one!


# fake suspended user
opencli user-add another random stefan@netops.rs ubuntu_apache_mysql
opencli user-suspend another

echo "Setting demo..."

##########################################


upload_wp_site_files() {
  local wp_site_path="/home/stefan/demo.openpanel.org"
  local wp_archive="https://wordpress.org/wordpress-latest.tar.gz"
  
  rm -rf /tmp/wp-site
  mkdir -p /tmp/wp-site
  cd /tmp/wp-site
  wget $wp_archive
  tar -xzvf wordpress-latest.tar.gz
  cp -r wordpress/. $wp_site_path
  rm -rf /tmp/wp-site
}


create_db_user_import_wpdb() {
  # step 1. start mysql
  docker exec stefan bash -c "service mysql start"

  # Define variables
  db_name="stefan_wp"
  db_user="stefan_wp"
  db_password="9823bdbds6732fdsw232rsd"
  
  # Function to run a command in the Docker container
  run_command_in_user_container() {
      command="$1"
      docker exec stefan bash -c "mysql -u root -e \"$command\""
  }
  
  # Define SQL commands
  create_db_command="CREATE DATABASE $db_name;"
  create_user_command="CREATE USER '$db_user'@'%' IDENTIFIED BY '$db_password';"
  privileges_command="GRANT ALL PRIVILEGES ON $db_name.* TO '$db_user'@'%';"
  phpmyadmin_also="GRANT ALL ON *.* TO 'phpmyadmin'@'localhost';"
  flush_command="FLUSH PRIVILEGES;"
  
  # Execute SQL commands
  run_command_in_user_container "$create_db_command"
  run_command_in_user_container "$create_user_command"
  run_command_in_user_container "$privileges_command"
  run_command_in_user_container "$phpmyadmin_also"
  run_command_in_user_container "$flush_command"


  # TODO REST FROM > https://git.devnet.rs/stefan/2083/-/blob/main/modules/wordpress.py
}


connect_wpdb_and_files() {
    # edit wpconfig
    wp_config_file="wp-config.php"
    domain="demo.openpanel.org"
    username="stefan"
    cd /home/stefan/demo.openpanel.org
    mv /home/stefan/demo.openpanel.org/wp-config-sample.php /home/stefan/demo.openpanel.org/$wp_config_file
    sed -i "s/database_name_here/$db_name/g" "$wp_config_file"
    sed -i "s/username_here/$db_user/g" "$wp_config_file"
    sed -i "s/password_here/$db_password/g" "$wp_config_file"
    
    # install
    docker exec stefan bash -c 'wp core install --url=https://${domain} --title="Demo Site" --admin_user=${username} --admin_password="ash732vfadsf" --admin_email=admin@${domain} --path=/home/${username}/${domain} --allow-root'

    # autologin
    docker exec stefan bash -c 'wp package install aaemnnosttv/wp-cli-login-command --path=/home/${username}/${domain} --allow-root'

    # prettylinks
    echo "
    # BEGIN WordPress
    RewriteEngine On
    RewriteRule .* - [E=HTTP_AUTHORIZATION:%{HTTP:Authorization}]
    RewriteBase /
    RewriteRule ^index\\.php$ - [L]
    RewriteCond %{REQUEST_FILENAME} !-f
    RewriteCond %{REQUEST_FILENAME} !-d
    RewriteRule . /index.php [L]
    # END WordPress
    " > /home/${username}/${domain}/.htaccess

    # permissions
    chown -R 1000:33 /home/${username}/${domain}/

    # salts
    docker exec ${username} bash -c "wp config shuffle-salts --path=/home/${username}/${domain}/ --allow-root"
}

cleanup() {
  rm -rf /root/demo.sh
  history -c
  history -w
  > ~/.bash_history
}


##########################################
echo "creating fake data for admin dashbord"
# todo: activity log and access, docker stats
write_fake_data

echo "configuring admin panel on port 2087"
setup_admin_panel

echo "configuring user panel on port 2083"
setup_user_panel

# TODO: also some helloworld py or node app
echo "download wp files"
upload_wp_site_files

echo "Creating db and user"
create_db_user_import_wpdb

echo "connect files to database"
connect_wpdb_and_files

echo "add site to wpmanager"
opencli websites-scan -all

echo "get droplet id"
get_droplet_id

echo "cleaning up.."
cleanup

echo "creating snapshot"


# todo: change snapshot id in the file for job!
# test in 1hr
# $snapshot_id




