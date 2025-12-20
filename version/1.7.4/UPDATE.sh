#!/bin/bash

# https://github.com/stefanpejcic/OpenPanel/issues/806
sed -i 's/^reseller=no$/reseller=yes/' /etc/openpanel/openadmin/config/admin.ini
