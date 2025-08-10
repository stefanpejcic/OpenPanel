#!/bin/bash

# datatracker.ietf.org/doc/html/rfc1035#autoid-31

for file in /etc/openpanel/bind9/zone_template.txt /etc/openpanel/bind9/zone_template_ipv6.txt; do
    sed -i 's/\(@[[:space:]]\+IN[[:space:]]\+SOA[[:space:]]\+{ns1}\.\)[[:space:]]\+{ns2}\./\1 {rpemail}./' "$file"
done
