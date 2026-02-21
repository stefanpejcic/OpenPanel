# Tiltsa le a terminált az OpenAdminban

Ha valamilyen oknál fogva le szeretné tiltani a parancssori *root* hozzáférést az OpenAdmin terminálszolgáltatásán keresztül, ezt megteheti a szerverén.

Az OpenAdminban nincs erre vonatkozó beállítás, de manuálisan létrehozhat egy fájlt, amely letiltja ezt a funkciót:

1. SSH-n keresztül rootként jelentkezzen be a szerverre.
2. Hajtsa végre a következő parancsot:
``` bash
érintse meg a /root/disable_openadmin_terminal_ui elemet
   ```
