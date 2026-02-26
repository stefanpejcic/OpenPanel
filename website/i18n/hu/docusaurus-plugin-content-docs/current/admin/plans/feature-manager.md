---
sidebar_position: 2
---

# Funkciókezelő

A Feature Manager lehetővé teszi a rendszergazdák számára, hogy bizonyos funkciókat (oldalakat) engedélyezzenek vagy tiltsanak le az OpenPanel felhasználói felületén. Ez hasznos a vezérlőpult felhasználói szerepkörök, biztonsági szabályzatok vagy tárhelytervek alapján történő testreszabásához.

Minden funkció külön-külön átkapcsolható a felületen keresztül. Az aktiválás után a funkció minden felhasználó számára láthatóvá és elérhetővé válik. A funkció kikapcsolása elrejti azt a panelen, és letiltja a funkcióit.

![openadmin features](/img/admin/tremor/features.png)

![openadmin features](/img/admin/tremor/features_edit.png)

---

## Elérhető funkciók
Íme az átírt információ táblázat formátumban a kért oszlopokkal:

| **Név** | **Link** | **Leírás** | **Megjegyzés** |
| --------------------------- | ------------------------- | --------------------------------------------------------- | ----------------------------------------------- |
| E-mail értesítések | "/fiók/értesítések" | Kezelje az e-mail értesítési beállításokat.                   | Az e-maileket a kiválasztott beállítások alapján küldjük el. |
| Területi helyek (nyelvváltás) | "/fiók/nyelvek" | Módosítsa a panel felületének nyelvét.                     |                                                |
| Kedvencek (Könyvjelzők) | "/fiók/kedvencek" | A gyakran használt oldalakat könyvjelzők közé helyezheti.                          |                                                |
| Lakk gyorsítótár | "/gyorsítótár/lakk" | Kezelje a Lakk gyorsítótárazást domainenként.                       |                                                |
| Docker (konténerek) | "/konténerek" | Erőforrások lefoglalása és a konténerek életciklusainak kezelése.      |                                                |
| FTP-fiókok | "/ftp" | FTP-fiókok létrehozása és kezelése.                          | Külön FTP szerver konfigurációt igényel.    |
| E-mail fiókok | "/e-mailek" | E-mail fiókok kezelése.                                   | Külön levelezőszerver konfigurációt igényel.   |
| Távoli MySQL | `/mysql/remote-mysql` | Távoli MySQL-kapcsolatok engedélyezése vagy letiltása.                 |                                                |
| MySQL konfiguráció | `/mysql/configuration` | Módosítsa a MySQL beállításait a panelről.                    |                                                |
| PHP beállítások | "/php/options" | Szerkessze a PHP direktívákat egy felhasználóbarát oldalon.            |                                                |
| PHP.INI szerkesztő | "/php/ini" | Szerkessze közvetlenül a php.ini fájlt.                        | Bármely konfigurált PHP verzióra vonatkozik.         |
| phpMyAdmin | `/mysql/phpmyadmin` | Adatbázisok kezelése a phpMyAdmin segítségével.                        |                                                |
| Cronjobs | "/cronjobs" | Cron-feladatok létrehozása, szerkesztése és ütemezése.                    |                                                |
| WordPress | `/wordpress` | WordPress webhelyek telepítése és kezelése.                      | A WP Manageren keresztül kezelhető.                        |
| Lemezhasználat Explorer | "/lemezhasználat" | Vizuálisan fedezze fel a lemezhasználatot a könyvtárak között.          |                                                |
| Inodes Explorer | "/inodes-explorer" | Tekintse meg az inode használatát könyvtáronként.                          |                                                |
| Erőforrások felhasználása | "/használat" | Tekintse meg a Docker-tároló erőforrás-használatát.                    |                                                |
| Szerver Info | "/server/info" | Tekintse meg a tárhelykorlátokat és a szerver részleteit.                  |                                                |
| Apache/Nginx konfiguráció | `/server/webserver_conf` | A webszerver (Apache/Nginx) tároló konfigurációjának módosítása. |                                                |
| Időzóna módosítása | "/szerver/időzóna" | Frissítse a rendszer időzóna-beállításait a tárolókhoz.          |                                                |
| Coraza WAF | "/waf" | Coraza WAF kezelése domainenként.                            | Alapértelmezés szerint engedélyezve van az új domaineknél.            |
| Engedélyek javítása | `/fix-permissions` | Javítsa ki a webhelyek fájlok tulajdonjogát és engedélyeit.         |                                                |
| DNS | "/dns" | DNS-rekordok kezelése zónaszerkesztővel.                   | BIND9 szerver szükséges.                         |
| Domain átirányítások | "/domains/redirects" | Domainszintű átirányítások kezelése.                           |                                                |
| Malware Scanner | `/malware-scanner` | Keressen rosszindulatú programokat a ClamAV segítségével.                           | A címtárkizárások konfigurálhatók.        |
| GoAccess | "/domains/logs" | Tekintse meg a GoAccess által generált naplójelentéseket.                     |                                                |
| Process Manager | "/folyamatkezelő" | Rendszerfolyamatok megtekintése és leállítása.                     |                                                |
| Redis | `/cache/redis` | A Redis konfigurálása felhasználónként.                                |                                                |
| Memcached | `/cache/memcached` | A Memcached konfigurálása felhasználónként.                            |                                                |
| Elasticsearch | `/cache/elasticsearch` | Konfigurálja az Elasticsearch-ot a panelről.                  |                                                |
| Opensearch | `/cache/opensearch` | Az Opensearch konfigurálása a panelről.                     |                                                |
| Ideiglenes linkek | "/webhelyek" | Tesztelje a webhelyeket ideiglenes OpenPanel aldomainekkel.      | A linkek 15 perc múlva lejárnak.                 |
| Bejelentkezési előzmények | `/account/loginlog` | Az utolsó 20 IP-bejelentkezés előzményeinek megtekintése.                   |                                                |
| 2FA | "/account/2fa" | Kéttényezős hitelesítés engedélyezése.                        |                                                |
| Tevékenységnapló | "/fiók/tevékenység" | Tekintse át az összes rögzített fiókműveletet.                     |                                                |

## Használati esetek

A **Funkciókészletek** segítségével szabályozható, hogy a felhasználók mely felhasználói felületi funkciókhoz férhetnek hozzá a hozzárendelt tárhelycsomagjuk alapján. Ez lehetővé teszi a felhasználói típusok és a szolgáltatási szintek egyértelmű elkülönítését.

### 1. példa: Csak adatbázis-csomagok

Hozzon létre egy **"Csak MySQL"** nevű szolgáltatáskészletet, és csak a MySQL-hez kapcsolódó funkciókat engedélyezze benne.
Rendelje hozzá ezt a szolgáltatáskészletet az összes adatbázis-központú tárhelycsomaghoz. Például:

* Egy csomag legfeljebb **10 adatbázist** engedélyez.
* Egy másik csomag **korlátlan számú adatbázist** engedélyez ("0" korlátozás nélkül).

A korlátok közötti különbség ellenére az ezen előfizetésekben szereplő összes felhasználó **csak a MySQL-lel kapcsolatos oldalakat** fogja látni a felhasználói felületen.

### 2. példa: Kezdő vs. haladó felhasználók

Hozzon létre két külön funkciókészletet:

* **Haladó felhasználók készlet**:
Engedélyezze az olyan funkciókat, mint a **Docker** és a **PHP.INI Editor**, hogy a tapasztalt felhasználók teljes irányítást biztosíthassanak – például egyéni erőforráskorlátok beállítása, szolgáltatások újraindítása stb.

* **Kezdő felhasználói készlet**:
**ne** engedélyezze a speciális funkciókat. Ehelyett engedélyezze a hozzáférést egy **PHP választóhoz** korlátozott lehetőségekkel. Ezáltal a felhasználói felület egyszerű és biztonságos a minimális technikai tapasztalattal rendelkező felhasználók számára.


## A funkció nem jelenik meg?

A szolgáltatások csak akkor érhetők el a felhasználók számára, ha a megfelelő **Modul** aktív](/docs/admin/settings/modules/). A modulok szabályozzák, hogy mely OpenPanel-szolgáltatások érhetők el, míg a **Funkciókészletek** a hozzáférést a felhasználó tárhelycsomagja alapján határozzák meg.

Például, ha a „Docker” funkciót hozzáadja egy tervhez, **nem** ad hozzáférést a Docker (tárolók) oldalaihoz a felhasználói felületen, kivéve, ha a **Docker modul** szintén aktiválva van az **OpenAdmin > Beállítások > Modulok** alatt.

