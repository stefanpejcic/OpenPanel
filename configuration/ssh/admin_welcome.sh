#!/bin/bash
#  nano /etc/profile.d/welcome.sh && chmod +x /etc/profile.d/welcome.sh
[ "$(id -u)" -ne 0 ] && return

# skip banner when launched from the OpenAdmin > Advanced > Web Terminal
[ -n "$OPENPANEL_HIDE_WELCOME" ] && return 0 2>/dev/null || true

VERSION=$(/usr/local/bin/opencli version)
OPENADMIN_STATUS=$(/usr/local/bin/opencli admin)

echo -e  "================================================================"
echo -e  ""
echo -e "  🚀 OpenPanel ${VERSION}"
echo -e  ""
echo -e  "$OPENADMIN_STATUS"
echo -e  ""
echo -e "  📖 Resources:"
echo -e "     • Docs:     https://openpanel.com/docs/admin/intro/"
echo -e "     • Forums:   https://community.openpanel.org/"
echo -e "     • Discord:  https://discord.openpanel.org/"
echo -e ""
echo -e "================================================================"
