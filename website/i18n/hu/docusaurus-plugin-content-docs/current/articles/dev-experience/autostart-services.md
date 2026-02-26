# Szolgáltatások automatikus indítása

Az erőforrások pazarlásának elkerülése érdekében az OpenPanel szolgáltatásai csak akkor indulnak el, amikor valóban szükség van rájuk.

## Szolgáltatások automatikus indítása az OpenAdminban

Az OpenPanel telepítése után csak a következő szolgáltatások indulnak el:

- **OpenAdmin** – A teljes szerver és a felhasználók kezeléséhez.
- **Docker** – Minden egyéb konténeres szolgáltatáshoz és felhasználói fiókhoz szükséges.
- **Adatbázis** – A MySQL adatbázis létrejön és inicializálódik. Ez az adatbázis terveket, webhelyeket, domaineket és felhasználókat tartalmaz.
- **Tűzfal** – A [Sentinel Firewall](https://sentinelfirewall.org/) telepítve van és elindult.

Más szolgáltatások telepítése és elindítása csak szükség esetén történik.

| Szolgáltatás | Telepítve | Automatikus indítás |
|------------------------|----------|-----------------------------|
| OpenAdmin | ✔ | Telepítéskor |
| Docker | ✔ | Telepítéskor |
| Adatbázis | ✔ | Telepítéskor |
| Sentinel Firewall | ✔ | Telepítéskor |
| OpenPanel | ✘ | Az első felhasználói fiók hozzáadása után |
| BIND9 | ✘ | Az első domain név hozzáadása után |
| Certbot | ✘ | Az első domain név hozzáadása után |
| ClamAV | ✘ | Ha engedélyezte a rendszergazda |
| Dovecot & Postfix | ✘ | Ha engedélyezte a rendszergazda |
| FTP | ✘ | Ha a rendszergazda engedélyezi, az első FTP-fiók létrehozása után |

## A szolgáltatások automatikus indítása az OpenPanelben

Az OpenAdminhoz hasonlóan az OpenPanel szolgáltatásai is csak akkor indulnak el, ha szükség van rájuk. Ez jobb erőforrás-gazdálkodást tesz lehetővé.

Minden felhasználó számára automatikusan elinduló szolgáltatások:

| Szolgáltatás | Telepítve | Automatikus indítás |
|--------------------|------------|----------------------------------------------------|
| Apache / Nginx / OpenLitespeed | ✔ | Miután a felhasználó hozzáadta az első domaint |       |
| REDIS | ✘ | Miután a felhasználó aktiválta |
| Memcached | ✘ | Miután a felhasználó aktiválta |
| Elasticsearch | ✘ | Miután a felhasználó aktiválta |
| MySQL / MariaDB | ✔ | Miután a felhasználó hozzáadott legalább 1 adatbázist |
| Cron | ✔ | Miután a felhasználó hozzáadott legalább 1 cron feladatot |
| PHP verziók | ✘ | Miután a felhasználó beállította legalább 1 tartományhoz |
