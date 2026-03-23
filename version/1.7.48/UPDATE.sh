#!/bin/bash

echo
echo "Installing slask-sock for OpenAdmin > Advanced > Web Terminal"
source /usr/local/admin/venv/bin/activate && pip install flask-sock && deactivate
service admin restart
