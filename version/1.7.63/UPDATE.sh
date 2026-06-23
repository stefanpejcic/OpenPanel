#!/bin/bash

# https://github.com/stefanpejcic/OpenPanel/issues/998
source /usr/local/admin/venv/bin/activate
pip install -r /usr/local/admin/requirements.txt
deactivate
