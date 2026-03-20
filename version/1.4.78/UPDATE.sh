#!/bin/bash

echo
echo "Installing interactive Web Terminal on OpenAdmin.."
source /usr/local/admin/venv/bin/activate
timeout 30 pip install flask-sock
deactivate
systemctl restart admin
