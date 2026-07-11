#!/bin/bash
export DEBIAN_FRONTEND=noninteractive
export NEEDRESTART_MODE=a
export CI=true
export npm_config_yes=true

APT_OPTS='-y -o Dpkg::Options::=--force-confold -o Dpkg::Options::=--force-confdef'

sudo apt update -y && sudo apt upgrade $APT_OPTS
sudo apt install $APT_OPTS ubuntu-desktop
sudo apt install $APT_OPTS xrdp
sudo systemctl enable xrdp
sudo systemctl restart xrdp
sudo adduser xrdp ssl-cert

sudo tee /etc/polkit-1/rules.d/45-allow-colord.rules >/dev/null <<EOF
polkit.addRule(function(action, subject) {
    if ((action.id == "org.freedesktop.color-manager.create-device" ||
         action.id == "org.freedesktop.color-manager.create-profile" ||
         action.id == "org.freedesktop.color-manager.delete-device" ||
         action.id == "org.freedesktop.color-manager.delete-profile" ||
         action.id == "org.freedesktop.color-manager.modify-device" ||
         action.id == "org.freedesktop.color-manager.modify-profile") &&
        subject.isInGroup("users")) {
        return polkit.Result.YES;
    }
});
EOF


sudo systemctl mask sleep.target suspend.target hibernate.target hybrid-sleep.target

gsettings set org.gnome.settings-daemon.plugins.power sleep-inactive-ac-type 'nothing'
gsettings set org.gnome.settings-daemon.plugins.power sleep-inactive-battery-type 'nothing'
gsettings set org.gnome.desktop.session idle-delay 0
gsettings set org.gnome.desktop.screensaver lock-enabled false
gsettings set org.gnome.desktop.screensaver idle-activation-enabled false


sudo tee /etc/polkit-1/localauthority/50-network-manager.d/xrdp-color-manager.pkla >/dev/null <<EOF
[Allow colord for all users]
Identity=unix-user:*
Action=org.freedesktop.color-manager.create-device;org.freedesktop.color-manager.create-profile;org.freedesktop.color-manager.delete-device;org.freedesktop.color-manager.delete-profile;org.freedesktop.color-manager.modify-device;org.freedesktop.color-manager.modify-profile
ResultAny=yes
ResultInactive=yes
ResultActive=yes
EOF

curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install $APT_OPTS nodejs git
mkdir -p ~/playwright-test
git clone https://github.com/stefanpejcic/openpanel-tests/ ~/playwright-test

cd ~/playwright-test

npm init -y
npm install -y @playwright/test

npx --yes playwright install --with-deps

npm install dotenv
npm install basic-ftp
npm install otplib

CRON_JOB="0 3 * * * bash /root/playwright-test/opencli/os_install.sh"
(crontab -l 2>/dev/null | grep -F "$CRON_JOB") || \
(crontab -l 2>/dev/null; echo "$CRON_JOB") | crontab -

if [ -d "node_modules/@playwright/test" ]; then
  echo "--- Setup Complete ---"
  echo "1. Reboot your machine: sudo reboot"
  echo "2. RDP into the server using your Ubuntu username/password."
  echo "3. Open a terminal inside the RDP session."
  echo "4. Create /root/playwright-test/openpanel/.env AND /root/playwright-test/openadmin/.env AND edit /root/playwright-test/opencli/os_install.sh"
  echo "5. Run tests as described in README.md files"
else
  echo "Install failed!"
fi
