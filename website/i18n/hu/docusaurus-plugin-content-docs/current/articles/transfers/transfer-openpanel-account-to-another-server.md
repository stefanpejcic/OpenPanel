# Számla átvitele

Ez a funkció lehetővé teszi az adminisztrátorok számára, hogy egyéni fiókokat vigyenek át (másolják) az egyik OpenPanel-kiszolgálóról a másikra.

Ha az összes felhasználót és fájlt egyszerre szeretné migrálni, tekintse meg a [Server Migration Documentation] (/docs/articles/transfers/migrate-openadmin-to-new-server/) című részt.

## Hogyan vigyünk át egy fiókot

### Az OpenAdmin használatával

1. Lépjen az **OpenAdmin > Felhasználók** oldalra.
2. Kattintson az átvinni kívánt felhasználóra.
3. Az **Átvitel** részben adja meg a távoli szerver adatait:
   
* **Szerver**: a távoli szerver IP-címe vagy tartományneve
* **Port**: SSH port (általában "22")
* **Felhasználónév**: SSH-felhasználónév a távoli kiszolgálóhoz (általában "root")
* **Password**: SSH-jelszó a megadott felhasználóhoz

[![2025-07-17-18-32.png](https://i.postimg.cc/BvbcRp2k/2025-07-17-18-32.png)](https://postimg.cc/Y4cWW1nz)

5. (Opcionális) Engedélyezze az **Élő átvitel** opciót.
Ha be van jelölve, a fiók az áttelepítés után felfüggesztésre kerül az aktuális szerveren, és a DNS továbbítódik az új szerverre.

6. Kattintson az **Átvitel indítása** gombra.

Az átvitel megkezdése után megkezdődik az átviteli folyamat. Létrejön egy naplófájl – kattintson a nevére az élő folyamat megtekintéséhez.

---

### Terminál használata

OpenPanel-fiók másik kiszolgálóra történő átviteléhez a terminálról használja a következő parancsot:

```bash
opencli user-transfer --account <OPENPANEL_USER> --host <DESTINATION_IP> --username <SSH_USERNAME> --password <DESTINATION_SSH_PASSWORD> [--live-transfer]
```

Ha a célszerver egyéni SSH-portot használ, adja meg a "--port" jelzőt, például:

```bash
--port 2222
```

Ha jelszó helyett SSH-kulcsos hitelesítést szeretne használni, egyszerűen hagyja ki a "--password" jelzőt. Győződjön meg arról, hogy az SSH-kulcs már konfigurálva van a célkiszolgálón, és ellenőrizze a kapcsolatot a következő futással:

```bash
ssh root@<DESTINATION_IP>
```
