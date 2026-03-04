# Tiltsa le az újraindítást az OpenAdminban

Ha valamilyen okból ki szeretné kapcsolni a *Server Reboot* funkciót az OpenAdminon keresztül, ezt megteheti a szerverén.

Az OpenAdminban nincs erre vonatkozó beállítás, de manuálisan létrehozhat egy fájlt, amely letiltja ezt a funkciót:

1. SSH-n keresztül rootként jelentkezzen be a szerverre.
2. Hajtsa végre a következő parancsot:
``` bash
érintse meg a /root/disable_openadmin_reboot_ui elemet
   ```
