#!/bin/bash

sed -i 's|/usr/local/admin/service/notifications.sh|/usr/local/opencli sentinel|g' /etc/cron.d/openpanel
