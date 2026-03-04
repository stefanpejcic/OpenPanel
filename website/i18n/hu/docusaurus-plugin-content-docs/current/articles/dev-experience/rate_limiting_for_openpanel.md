# Limit OpenPanel

Mind az OpenPanel, mind az OpenAdmin interfész rendelkezik beépített sebességkorlátozással és IP-cím blokkolással, hogy megvédje a brute force támadásokat.

Ez a funkció alapértelmezés szerint minden telepítésen engedélyezve van, és az *OpenAdmin > Beállítások > OpenPanel* menüpontban vagy a terminálról is testreszabható.

Beállíthatja az IP-címenként megengedett sikertelen bejelentkezési kísérletek maximális számát (alapértelmezett `5` percenként) és a sikertelen kísérletek teljes számát (alapértelmezett `20`), ami után a tűzfal ideiglenesen blokkolja a sértő IP-címet egy órára.

Az OpenPanel korlátai a /etc/openpanel/openpanel/conf/openpanel.config fájlban állíthatók be:
```bash
[USERS]
login_ratelimit=5
login_blocklimit=20
```

![felhasználói sebességkorlát](/img/panel/v1/user_block.png)

Az OpenAdmin korlátai a /etc/openpanel/openadmin/config/admin.ini fájlban konfigurálhatók:
```bash
[PANEL]
login_ratelimit=5
login_blocklimit=20
```

![admin ratelimit](/img/admin/admin_block.png)

Ha egy felhasználó sikeresen bejelentkezik, a `login_blocklimit` számlálója visszaáll.

A sikertelen bejelentkezési kísérletek és a blokkolt IP-címek az OpenAdmin esetében a `/var/log/openpanel/admin/failed_login.log` fájlban, az OpenPanel esetében pedig a `/var/log/openpanel/user/failed_login.log` fájlban kerülnek naplózásra.
