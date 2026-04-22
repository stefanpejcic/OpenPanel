#!/bin/bash
#  nano /etc/profile.d/welcome.sh && chmod +x /etc/profile.d/welcome.sh
[ "$(id -u)" -ne 0 ] && return

VERSION=$(opencli version)
OPENADMIN_STATUS=$(opencli admin)

echo -e  "================================================================"
echo -e  ""
echo -e "  🚀 OpenPanel ${VERSION} is installed and ready!"
echo -e  ""
echo -e  "$OPENADMIN_STATUS"
echo -e  ""
echo -e "  📖 Resources:"
echo -e "     • Docs:     https://openpanel.com/docs/admin/intro/"
echo -e "     • Forums:   https://community.openpanel.org/"
echo -e "     • Discord:  https://discord.openpanel.org/"
echo -e ""
echo -e "================================================================"
