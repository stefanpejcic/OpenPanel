#!/bin/bash

#apt install python3-pip
#pip install  /home/screenshot/requirements.txt

# DEV venv/bin/python -m flask run --host=0.0.0.0 -p 80


cd /home/screenshot/
python3 -m venv venv
apt install python3.12-venv
source venv/bin/
apt install python3-pip
apt install python3-playwright
pip install  /home/screenshot/requirements.txt

# limit access to proxy only!

: '
   62  sudo ufw reset
   63  ufw default deny incoming
   64  ufw default allow outgoing
   65  ufw allow from 82.117.216.242
   66  ufw allow from 185.119.89.240
   67  ufw enable
   96  ufw allow from 10.116.0.5
   97  ufw allow from 24.144.64.6
  113  ufw reload
'

# add cron
chmod +x /home/screenshot/start.sh
cronjob="@reboot /bin/bash /home/screenshot/start.sh"
(crontab -l 2>/dev/null | grep -v -F "$cronjob"; echo "$cronjob") | crontab -

# start
bash /home/screenshot/start.sh

