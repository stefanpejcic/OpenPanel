# Frissítés LF_ALERT_TO

Hogyan állítsuk be az e-mail címet a CSF/LFD riasztásokhoz

---

Ha frissíti az értesítő e-mailt az **OpenAdmin** segítségével, akkor a **CSF/LFD** által a figyelmeztetésekhez használt e-mail-cím is frissül.

Ehhez parancssoron keresztül egyszerűen futtassa:

```bash
opencli config update email youremail@yourdomain.com
```

Ez frissíti az OpenAdmin értesítéseit és a CSF/LFD `LF_ALERT_TO' beállítását.

---

## Másik e-mail használata a CSF/LFD-hez

Ha azt szeretné, hogy a **CSF/LFD** a riasztásokat az OpenAdminban megadottól eltérő címre küldje:

1. Nyissa meg a CSF konfigurációs fájlt a kívánt szövegszerkesztőben:

``` bash
nano /etc/csf/csf.conf
   ```

2. Keresse meg a következő sort:

   ```
LF_ALERT_TO = ""
   ```

3. Cserélje ki a kívánt e-mail-címre:

   ```
LF_ALERT_TO = "sajatmail@sajatdomain.com"
   ```

4. Mentse el a fájlt, és indítsa újra a CSF-et a módosítások alkalmazásához:

``` bash
csf -r
   ```

Ennyi! A CSF/LFD mostantól értesítéseket küld a megadott e-mail címére.

---

> **Megjegyzés:** A CSF/LFD harmadik féltől származó szoftver, amelyet nem az OpenPanel fejlesztett vagy karbantart.
> Támogatásért forduljon közvetlenül a Sentinel Firewallhoz:
> [https://sentinelfirewall.org/](https://sentinelfirewall.org/)
