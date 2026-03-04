# Gyorsítótárazott adatok az OpenPanel UI-ban

Az **OpenPanel végfelhasználói felület** csökkenti a szerverlemez I/O-ját a Redis gyorsítótárazásának kihasználásával.

* **Rendszeradatok** (csak a szolgáltatás indításakor jön létre):
* `enabled_modules`
* "logó", "márkanév".
* `dev_mode`
* `talált_hibalinket`
* `panel_verzió`
* `avatar_type`, `gravatar_image_url`

* **Felhasználói adatok** (a felhasználó bejelentkezésekor keletkezik):
* "felhasználói_azonosító", "felhasználónév", "e-mail".
* `docker_context`
* `hosting_terv`
* `feature_set`

* **Webhely- és domainadatok** (csak a felhasználói felületen keresztül jön létre):
* Weboldalak listája
* Domain rekordok

Az adatok időnként frissülnek, de a rendszergazda azonnali frissítést is kényszeríthet az OpenPanel tároló újraindításával, amely minden adatot újragenerál.

**Az adatok gyorsítótárának időtartama az OpenPanel UI-ban**:

| Tétel / Leírás | Gyorsítótár (másodperc) |
|-------------------------------------------------------------------------|-----------------|
| Vállalati engedély érvényessége | 3600 |
| Fióktevékenység lista | 300 |
| IP-cím országkódja | 360 |
| Bejelentkezési előzmények | 600 |
| Értesítési beállítások | 300 |
| Aktív munkamenetek | 10 |
| Webhelyek száma az Auto Installerben | 30 |
| Lakk állapota | 360 |
| A szerver nyilvános IPv4-címe | 3600 |
| OpenPanel verzió | 7200 |
| SSL-tanúsítvány megléte a felhasználói panelhez | 360 |
| Domain beállítva az OpenPanel felhasználói felülethez | 3600 |
| OpenPanel konfiguráció | 6000 |
| Felhasználói webhelyek a keresési eredmények között | 600 |
| Felhasználói webhelyek az irányítópult statisztikai widgetjében | 300 |
| Lemez- és inode-használat a felhasználó számára (műszerfali widget) | 360 |
| Tárhelyterv korlátai (műszerfali widget) | 3600 |
| Domainhasználat (irányítópult-statisztika) | 360 |
| Adatbázisok száma (műszerfali statisztika) | 360 |
| FTP-fiókok száma (irányítópult-statisztika) | 360 |
| Docker statisztikák (CPU, memória, tárolók, PID-k) (irányítópult-statisztikák) | 360 |
| CNAME rekord létezése rekord hozzáadásakor a DNS-szerkesztőben | 10 |
| Domain átirányítás létezése | 10 |
| Átirányítási URL a domainhez | 10 |
| Tervenkénti domain korlát | 3600 |
| A csúcsok és az aldomainek száma | 30 |
| SSL állapot a domainhez | 30 |
| Mappa lemezhasználat (Fájlkezelő) | 5 |
| Webmail domain | 3600 |
| Fájltartalom (Fájlkezelő - Fájlnézegető) | 30 |
| Almappák és fájlok listája (Fájlkezelő) | 5 |
| Fájlstatisztikák (tulajdonos, méret, engedélyek) (Fájlkezelő) | 30 |
| Fájlok a kukában | 5 |
| Inode száma elérési úthoz (Inodes Explorer) | 10 |
| FTP felhasználónév ellenőrzése | 60 |
| Rosszindulatú programok vizsgálatának eredménye | 30 |
| Karanténban lévő fájlok (Malware Scanner) | 30 |
| GoAccess HTML jelentés a domainhez | 7200 |
| Szerver információs oldal adatai | 360 |
| Tervezze meg a korlátokat a műveletek végrehajtása során | 300 |
| Nyitott portok Redis, Memcached, phpMyAdmin konténerekhez | 300 |
| Felhasználói fiók UID | 7200 |
| Domain tulajdonjogának ellenőrzése | 60 |
| Webhely tulajdonjogának ellenőrzése | 60 |
| Utolsó bejelentkezési IP (műszerfali widget) | 600 |
| PHP verzió a domainhez | 30 |
| MySQL adatbázis mérete | 300 |
| Adatbázis-információk a wp-config.php-ről (WP Manager) | 600 |
| Adatbázis információ a local.php-ről (Mautic Manager) | 300 |
| A szerver üzemideje és terhelése (webhely információ) | 30 |
| Szerverinformáció (CPU, IP, üzemidő, processzor) (webhely információs oldala) | 600 |
| Tárhelyterv korlátai (webhely információs oldal) | 600 |
| Adatbázis információ: méret, felhasználók, támogatások | 600 |
| Szolgáltatási naplók | 300 |
| Képernyőkép az API-ból | 3600 |
| Képernyőkép hivatkozás a távoli API-ból | 60 |
| Adatbázis-korlát tervenként (új adatbázis létrehozása) | 3600 |
| Az összes adatbázis, felhasználó, támogatások listája (MySQL/Postgres) | 300 |
| Konfigurációs értékek (my.cnf / postgresql.conf) | 30 |
| PHP kiterjesztések listája | 30 |
| Elérhető PHP-verziók | felhasználó számára 3600 |
| PHP szolgáltatások futtatása felhasználók számára | 10 |
| NodeJS/Python tárolónaplók | 60 |
| NodeJS/Python Docker Hub címkék | 86400 |
| A felhasználó által használt webszerver | 300 |
| Tartománybeállítások ideiglenes hivatkozásokhoz API (OpenPanel) | 7200 |
| Ideiglenes link a webhelyhez (API) | 800 |
| WP verzió a wp-config.php webhelyről (WP Manager) | 60 |
| Adatbázis-információk a wp-config.php-ről (WP Manager) | 60 |
| PageSpeed ​​eredmények a webhelyre | 60 |
| MySQL/MariaDB verzió felhasználó számára (WP Manager) | 3600 |
| Docroot, DB info, PHP, WP, MySQL verziók (WP Manager) | 300 |
| WP adminisztrátori e-mail (WP Manager) | 300 |
| Biztonsági másolat létezésének ellenőrzése (WP Manager) | 300 |
| PageSpeed ​​API-kulcs ellenőrzése | 60 |
| Weboldal létezésének ellenőrzése | 30 |

















