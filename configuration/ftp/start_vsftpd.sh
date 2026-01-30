#!/bin/sh

print_banner() {
  echo "-------------------------------------------"
  echo "            Starting FTP server            "
  echo "-------------------------------------------"
}

print_banner

echo "[*] Collecting lists of FTP users from all OpenPanel accounts..."

USERS=""

for dir in /etc/openpanel/ftp/users/*; do
    file="$dir/users.list"
    user=$(basename "$dir")
    if [ -f "$file" ]; then
        while IFS= read -r line; do
            # Replace /var/www/html/ with /home/... docker path
            modified_line=$(echo "$line" | sed "s|/var/www/html/|/home/${user}/docker-data/volumes/${user}_html_data/_data/|g")
            USERS="$USERS$modified_line "
        done < "$file"
    fi
done

echo "[*] Updating cached list in '/etc/openpanel/ftp/all.users' file..."
echo "USERS=\"$USERS\"" > /etc/openpanel/ftp/all.users




echo "[*] Removing all existing FTP users..."
grep '/ftp/' /etc/passwd | cut -d':' -f1 | xargs -r -n1 deluser

# Function to determine if a hostname is a FQDN
is_fqdn() {
  if [[ $1 =~ ^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
    return 0
  else
    return 1
  fi
}

# Function to get the server's IP address
get_ip_address() {
  hostname -i
}

# Function to read users from users.list files and create them
create_users() {
  echo "[*] Checking for existing users..."
  USER_LIST_FILES=$(find /etc/openpanel/ftp/users/ -name 'users.list')
  TOTAL_USER_COUNT=0
  for USER_LIST_FILE in $USER_LIST_FILES; do
    while IFS='|' read -r NAME _; do
      [ -z "$NAME" ] && continue  # empty
      TOTAL_USER_COUNT=$((TOTAL_USER_COUNT + 1))
    done < "$USER_LIST_FILE"
  done

  if [ "$TOTAL_USER_COUNT" -gt 0 ]; then
    echo "[*] Total users that will be created: $TOTAL_USER_COUNT"
  else
    echo "[*] No users found to create."
    return
  fi

  USER_COUNT=0
  echo "[*] Creating users from users.list files..."
  
  for USER_LIST_FILE in $USER_LIST_FILES; do
    BASE_DIR=$(dirname "$USER_LIST_FILE")
    OPENPANEL_USER=$(basename "$BASE_DIR")
    echo "[!] Processing users for OpenPanel account: $OPENPANEL_USER"
    while IFS='|' read -r NAME HASHED_PASS FOLDER UID GID; do
      USER_COUNT=$((USER_COUNT + 1))
      [ -z "$NAME" ] && continue  # Skip empty lines
      GROUP="${NAME#*.}"
      echo "[+] Creating user ${NAME} [${USER_COUNT}/${TOTAL_USER_COUNT}]"
      FAKE_FOLDER="$FOLDER" # used for display only!
      if [ -z "$FOLDER" ]; then
        FOLDER="/var/www/html/"
        echo "    - No folder specified, using default: $FOLDER"
      fi

      # Replace legacy path if needed
      FOLDER=$(echo "$FOLDER" | sed "s|/var/www/html/|/home/${GROUP}/docker-data/volumes/${GROUP}_html_data/_data/|g")

      # Validate folder starts with /home
      case "$FOLDER" in
        /home/*) ;;
        *)
          echo "    - Skipping user $NAME: folder $FOLDER is invalid"
          continue
          ;;
      esac

      UID_OPT=""
      GROUP_OPT=""

      if [ -n "$UID" ]; then
        UID_OPT="-u $UID"
      fi

      if [ -n "$GID" ]; then
        EXISTING_GROUP=$(getent group "$GID" | cut -d: -f1)
        if [ -n "$EXISTING_GROUP" ]; then
          echo "    - Group $EXISTING_GROUP already exists with GID $GID"
          GROUP_OPT="-G $EXISTING_GROUP"
        else
          echo "    - Creating group $GROUP with GID $GID"
          addgroup -g "$GID" "$GROUP"
          GROUP_OPT="-G $GROUP"
        fi
      fi

      echo "    - Adding user with home directory: $FAKE_FOLDER"
      adduser -h "$FOLDER" -s /sbin/nologin $UID_OPT $GROUP_OPT --disabled-password --gecos "" "$NAME"

      echo "    - Setting encrypted password '$HASHED_PASS'"
      if usermod -p "$HASHED_PASS" "$NAME"; then
          echo "    - Password set successfully"
      else
          echo "    - Failed to set password, removing user $NAME"
          userdel "$NAME"
          continue
      fi

      echo "    - Ensuring folder exists and ownership is correct (${UID}:${GID})"
      mkdir -p "$FOLDER"
      chown "${UID}:${GID}" "$FOLDER"

      echo ""
      unset NAME HASHED_PASS FOLDER UID GID GROUP UID_OPT GROUP_OPT
    done < "$USER_LIST_FILE"
  done

  if [ "$USER_COUNT" -gt 0 ]; then
    echo "[*] User creation complete: $USER_COUNT users created."
  else
    echo "[*] No users found to create."
  fi
}

# Call the function to create users
create_users

# Set default passive mode port range if not specified
if [ -z "$MIN_PORT" ]; then
  MIN_PORT=21000
fi

if [ -z "$MAX_PORT" ]; then
  MAX_PORT=21010
fi
echo "[*] Passive mode port range: $MIN_PORT-$MAX_PORT"

# Determine the address if not provided
if [ -z "$ADDRESS" ]; then
  HOSTNAME=$(hostname)
  echo "[*] Hostname detected: $HOSTNAME"
  if is_fqdn "$HOSTNAME"; then
    ADDRESS="$HOSTNAME"
    echo "[*] Using hostname as passive address: $ADDRESS"
  else
    ADDRESS=$(get_ip_address)
    echo "[*] Using IP address as passive address: $ADDRESS"
  fi
else
  echo "[*] Using provided passive address: $ADDRESS"
fi

if [ -n "$ADDRESS" ]; then
  ADDR_OPT="-opasv_address=$ADDRESS"
fi


# Used to run custom commands inside container
if [ -n "$1" ]; then
  echo "[*] Executing custom command: $*"
  exec "$@"
else
  echo "[*] Starting vsftpd and accepting user logins..."
  vsftpd -opasv_min_port=$MIN_PORT -opasv_max_port=$MAX_PORT $ADDR_OPT /etc/vsftpd/vsftpd.conf
  [ -d /var/run/vsftpd ] || mkdir /var/run/vsftpd
  pgrep vsftpd | tail -n 1 > /var/run/vsftpd/vsftpd.pid
  echo "-------------------------------------------"
  exec pidproxy /var/run/vsftpd/vsftpd.pid true
fi
