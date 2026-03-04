# Hozzáférés az OpenAdminhoz

Futtassa az "opencli admin" parancsot, hogy megkeresse azt a címet, amelyen az adminisztrációs panel elérhető. Példa kimenet:

```bash
root@server:/home# opencli admin
● OpenAdmin is running and is available on: https://server.openpanel.org:2087/
```

Az adminisztrációs panelre való bejelentkezéshez felhasználónévre és jelszóra van szüksége.

![openadmin bejelentkezési oldal](/img/admin/openadmin_login_page.png)

Mind a felhasználónév, mind a jelszó véletlenszerűen generálódik a telepítés során.

Az adminisztrátori fiókok megtekintéséhez:

```bash
opencli admin list
```

Új jelszó beállításához az admin fiók futtatási parancsához: `opencli admin jelszó USER_HERE NEW_PASSWORD_HERE`

Példa:
```bash
root@server:/home# opencli admin password stefan ba63vfav7fq36vas
Password for user 'stefan' changed.

===============================================================
● OpenAdmin is running and is available on: https://server.openpanel.co:2087/

- username: stefan
- password: ba63vfav7fq36vas

===============================================================
```
