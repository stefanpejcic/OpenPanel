#!/bin/bash

source /usr/local/admin/venv/bin/activate && pip install flask-sock && deactivate
service admin restart
