#!/bin/bash

wget -O /etc/openpanel/openadmin/config/features.json https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/config/features.json

wget -O /etc/openpanel/openpanel/features/default.txt https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/features/default.txt

CADDYFILE="/etc/openpanel/caddy/Caddyfile"

if grep -q "^ *admin off" "$CADDYFILE"; then
    sed -i 's/^ *admin off/#admin off/' "$CADDYFILE"
    echo "'admin off' has been commented out."
else
    echo "No 'admin off' line found or it's already commented."
fi


# add views for domain count!

mysql -uroot <<'EOF'
USE panel;

-- Step 1: Sanitize existing data
UPDATE users SET user_domains = '0' WHERE user_domains = '';

-- Step 2: Modify the column to INT
ALTER TABLE users MODIFY COLUMN user_domains INT NOT NULL DEFAULT 0;

-- Step 3: Create trigger for insert
DELIMITER //

CREATE TRIGGER increment_user_domains
AFTER INSERT ON domains
FOR EACH ROW
BEGIN
  UPDATE users SET user_domains = user_domains + 1 WHERE id = NEW.user_id;
END;
//

-- Step 4: Create trigger for delete
CREATE TRIGGER decrement_user_domains
AFTER DELETE ON domains
FOR EACH ROW
BEGIN
  UPDATE users SET user_domains = user_domains - 1 WHERE id = OLD.user_id;
END;
//

DELIMITER ;
EOF
