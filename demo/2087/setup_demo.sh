#!/bin/bash

# Parse command line options
while getopts "yf" opt; do
  case $opt in
    y)
      confirm="yes"
      ;;
    f)
      confirm="force"
      ;;
    *)
      usage
      ;;
  esac
done



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
echo "Creating demo panel user"
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



# If no flags are set, ask for user confirmation
if [ -z "$confirm" ]; then
  read -p "This will convert your website to host public OpenPanel demo. Do you want to proceed? (y/n) " user_input
  case $user_input in
    [Yy]*)
      #echo "User confirmed to proceed."
      
      ;;
    *)
      exit 1
      ;;
  esac
fi

echo "Setting demo..."

write_fake_data

setup_admin_panel

setup_user_panel


#echo "DONE."

get_droplet_id


