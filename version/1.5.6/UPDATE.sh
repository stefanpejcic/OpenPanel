#!/bin/bash

# https://my.openpanel.com/adminad/supporttickets.php?action=view&id=3937
mysql --database=panel -e "
ALTER TABLE domains CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
ALTER TABLE sites CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
"


# datatracker.ietf.org/doc/html/rfc1035#autoid-31
for file in /etc/openpanel/bind9/zone_template.txt /etc/openpanel/bind9/zone_template_ipv6.txt; do
    sed -i 's/\(@[[:space:]]\+IN[[:space:]]\+SOA[[:space:]]\+{ns1}\.\)[[:space:]]\+{ns2}\./\1 {rpemail}./' "$file"
done
