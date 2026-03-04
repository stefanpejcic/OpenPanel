---
sidebar_position: 4
---

# Modulok

A modulok új funkciók és oldalak hozzáadásával bővítik az OpenPanel felhasználói felületet. Ahhoz, hogy egy szolgáltatást elérhetővé tegyen egy felhasználó vagy terv számára, először modulként kell aktiválni.

- A modulok **alapvető funkciók**, amelyek már telepítéskor elérhetők, és az OpenPanel fejlesztése.
- A beépülő modulok egyéni szolgáltatások, amelyeket telepíteni kell, és amelyeket külső fejlesztők fejlesztettek ki.

Elérhető modulok:

## Értesítések

Az **`értesítések`** modul szükséges e-mailes értesítések küldéséhez a felhasználóknak.

Ha engedélyezve van:
* Az e-maileket az egyes felhasználók értesítési beállításai szerint küldjük el.
* A felhasználók az OpenPanel UI-n keresztül kezelhetik beállításaikat: [**Fiókok > E-mail értesítések**](/docs/panel/account/notifications/).

Kikapcsolt állapotban:
* A felhasználói beállításoktól függetlenül nem küldünk e-maileket.

Az e-mail értesítések testreszabása:
* Az alapértelmezett beállítások **új felhasználók számára történő beállításához** szerkessze a [`/etc/openpanel/skeleton/notifications.yaml`](https://github.com/stefanpejcic/openpanel-configuration/blob/main/skeleton/notifications.yaml) fájlt.
* Az **e-mail sablonok testreszabásához** tekintse meg az [OpenPanel e-mail sablonok testreszabása] (https://community.openpanel.org/d/214-customizing-openpanel-email-templates) című részt.
* Az **egyéni SMTP konfigurálásához** használja az [OpenAdmin > Beállítások > Értesítések oldalt] (/docs/admin/settings/notifications/).



## Fiók

A **`fiók`** modul szükséges ahhoz, hogy a felhasználók módosíthassák e-mail címüket, jelszavukat vagy felhasználónevüket.

Ha engedélyezve van:
* A felhasználók megváltoztathatják e-mail címüket, jelszavukat és felhasználónevüket az OpenPanel felhasználói felületén keresztül: [**Fiókok > Beállítások**](/docs/panel/account/).

Kikapcsolt állapotban:
* A felhasználók nem változtathatják meg jelszavaikat az OpenPanel UI-ból, csak a bejelentkezési űrlap „Jelszó visszaállítása” menüpontjából, ha ez az opció be van kapcsolva.

Jelszó- és felhasználónévmódosítások testreszabása:
* A jelszó-visszaállítás engedélyezése vagy letiltása a bejelentkezési űrlapokon** módosítsa a „Jelszó-visszaállítás engedélyezése bejelentkezéskor” beállítást az [OpenAdmin > Beállítások > OpenPanel] (/docs/admin/settings/openpanel/) oldalon.
* Ha meg szeretné akadályozni, hogy a felhasználók módosítsák felhasználónevüket**, szerkessze a „Felhasználónév megváltoztatásának engedélyezése” beállítást az [OpenAdmin > Beállítások > OpenPanel] (/docs/admin/settings/openpanel/) részben.


## Munkamenetek

A **`sessions`** modul lehetővé teszi a felhasználók számára az aktív munkamenetek megtekintését és kezelését.

Ha engedélyezve van:
* A felhasználók megtekinthetik összes aktív munkamenetüket, naplójukat, és bármely munkamenetet megszakíthatnak az OpenPanel UI-n keresztül: [**Fiókok > Aktív munkamenetek**](/docs/panel/account/active_sessions/).

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *Fiókok > Aktív munkamenetek* oldalt.

A munkamenetek időtartamának testreszabása:
* A **munkamenet időtartamának szabályozásához** módosítsa a „Munkamenet időtartama” beállítást az [OpenAdmin > Beállítások > OpenPanel] oldalon (/docs/admin/settings/openpanel/#Statistics).
* A **munkamenet élettartamának szabályozásához** módosítsa a „Munkamenet élettartama” beállítást az [OpenAdmin > Beállítások > OpenPanel] oldalon (/docs/admin/settings/openpanel/#Statistics).

## Nyelv

A **`locale`** (Nyelvek) modul lehetővé teszi a felhasználók számára a panel nyelvének megváltoztatását.

Ha engedélyezve van:
* A felhasználók a bejelentkezési oldalon és a [**Fiókok > Nyelv módosítása** oldalon] (/docs/panel/account/language/) módosíthatják az OpenPanel UI preferált nyelvét.

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *Fiókok > Nyelv módosítása* oldalt a területi beállítás megváltoztatásához.
* A felhasználókat az adminisztrátor által meghatározott alapértelmezett területi beállításokra kényszerítik.

A nyelvek testreszabása:
* Az alapértelmezett nyelv beállításához** használja az [OpenAdmin > Beállítások > Nyelvi beállítások] (/docs/articles/accounts/default-user-locales/) parancsot.
* Ha **új nyelvi beállításokat szeretne telepíteni a felhasználók számára**, használja az [OpenAdmin > Settings > Locales](/docs/admin/settings/locales/#install-locale) menüpontot.
* **Új fordítás létrehozásához** olvassa el a [Hogyan hozzunk létre új nyelvi beállítást] (/docs/admin/settings/locales/#edit-locale) című részt.


## Kedvencek

A **`kedvencek`** modul lehetővé teszi a felhasználók számára, hogy *rögzítsenek* elemeket az oldalsávi menüjükben a gyors navigáció érdekében.

Ha engedélyezve van:
* A felhasználók az oldal jobb felső sarkában található ⭐ ikonra **bal kattintással** hozzáadhatnak oldalakat a kedvencekhez.
* A felhasználók **jobb gombbal** az oldal jobb felső sarkában található ⭐ ikonra kattintva eltávolíthatnak oldalakat a kedvencekből.
* A felhasználók az oldalsáv menüjéből érhetik el kedvenceiket.
* A felhasználók elérhetik a [**Fiókok > Kedvencek** oldalt](/docs/panel/account/favorites/).

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *Fiókok > Kedvencek* oldalt a kedvencek kezeléséhez.
* A felhasználók nem látják kedvenceiket sem az oldalsávban, sem a ⭐ ikont az oldalak jobb felső sarkában.

Kedvencek testreszabása:
* A **felhasználó kedvenceinek teljes számának szabályozásához** (alapértelmezett 10) használja a [`favorites-items` config](https://dev.openpanel.com/cli/config.html#favorites-items) alkalmazást.
* A **felhasználók kedvenceinek szerkesztéséhez a terminálról** szerkessze a következő fájlt: `/etc/openpanel/openpanel/core/users/{current_username}/favorites.json`.


## Lakk

A **`lakk`** modul lehetővé teszi a felhasználók számára, hogy szabályozzák a lakk-gyorsítótárat a tartományukhoz.

Ha engedélyezve van:
* A Varnish szerver elindul a felhasználók számára, és visszaadja a forgalmat a webszerverükre.
* A felhasználók elérhetik a [**Caching > Lakk** oldalt](/docs/panel/caching/lakk/).
* A felhasználók engedélyezhetik/letilthatják a Lakk szolgáltatást.
* A felhasználók domainenként engedélyezhetik/letilthatják a Varnish gyorsítótárazást.
* A felhasználók megtekinthetik a Lakk szolgáltatás naplóit.

Kikapcsolt állapotban:
* A felhasználók nem férhetnek hozzá a *Gyorsítótár > Lakk* oldalhoz.
* A Lakk csak akkor használható, ha a rendszergazda engedélyezte a felhasználó számára a fiók létrehozásakor.

Beállítások testreszabása:
* A Varnish engedélyezéséhez/letiltásához minden új felhasználó számára** használja az [*OpenAdmin > Beállítások > Felhasználói alapértelmezések* oldalt és az *Enable Varnish Proxy* opciót] (/docs/admin/settings/defaults/).
* A Varnish engedélyezése/letiltása egyetlen felhasználó számára** a fiók létrehozásakor használja az [**Varnish Cache engedélyezése** lehetőséget] (/docs/articles/docker/how-to-set-nginx-apache-varnish-per-user-in-openpanel/).
* Az alapértelmezett CPU/RAM módosításához** használja az [*OpenAdmin > Settings > User Defaults* oldalt](/docs/admin/settings/defaults/).
* A Varnish alapértelmezett.vcl fájljának **szerkesztéséhez** használja az [*OpenAdmin > Domains > Edit Domain Templates* oldalt](/docs/admin/settings/defaults/), vagy szerkessze a fájlt: [`/etc/openpanel/varnish/default.vcl`](https://github.com/stefanpejcic/openpanel-configuration/blob/main/varnish/default.vcl).
* A **Lakk-gyorsítótár tisztításához** olvassa el a [How-to Guides > Purging Lakk Cache] (/docs/articles/websites/purge-varnish-cache-from-terminal/) részt.
* Ha ellenőrizni szeretné, hogy a Varnish engedélyezve van-e a domainhez**, olvassa el a [Hogyan ellenőrizhető, hogy a Varnish Caching engedélyezve van-e egy domainhez az OpenPanelben?](https://community.openpanel.org/d/207-how-to-check-if-varnish-caching-is-enabled-for-a-domain-in-openpanel)


## Docker

A **`docker`** modul lehetővé teszi a felhasználók számára, hogy új docker-tárolókat kezeljenek és adjunk hozzá.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Docker > Tárolók**](/docs/panel/containers/) oldalt a szolgáltatások megtekintéséhez és kezeléséhez.
* A felhasználók hozzáférhetnek a [**Docker > Tárolók > Új**](/docs/panel/containers/#adding-new-services) oldalhoz új szolgáltatások hozzáadásához.
* A felhasználók elérhetik a [**Docker > Terminal**](/docs/panel/containers/terminal/) oldalt a docker exec parancsok futtatásához.
* A felhasználók elérhetik a [**Docker > Képfrissítések**](/docs/panel/containers/image/) oldalt az elérhető képfrissítések ellenőrzéséhez.
* A felhasználók elérhetik a [**Docker > Naplók**](/docs/panel/containers/logs/) oldalt a szolgáltatásnaplók megtekintéséhez.
* A felhasználók elérhetik a [**Docker > Képcímke módosítása**](/docs/panel/containers/change/) oldalt a képcímke módosításához.
* A felhasználók elérhetik a [**Docker > Webszerver váltása**](/docs/panel/containers/webserver/) oldalt a webszerverek közötti váltáshoz.
* A felhasználók elérhetik a [**Docker > Switch MySQL Type**](/docs/panel/containers/mysql/) oldalt a mysql/mariadb közötti váltáshoz.

Kikapcsolt állapotban:
* A felhasználók nem férhetnek hozzá a *Docker* oldalak egyikéhez sem.

Beállítások testreszabása:
* Egyik sem


## Engedélyek javítása

A **`fix_permissions`** modul lehetővé teszi a felhasználók számára a fájlok/mappák engedélyeinek visszaállítását.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Fájlok > Engedélyek javítása** oldalt] (/docs/panel/files/fix_permissions/).

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *Fájlok > Engedélyek javítása* oldalt.


## FTP

Az **`ftp`** modul lehetővé teszi a felhasználók számára FTP-alfiókok létrehozását és kezelését.

Ha engedélyezve van:
* A felhasználók hozzáférhetnek a [**Fájlok > FTP** oldalhoz](/docs/panel/files/FTP/) az FTP-fiókok kezeléséhez.

Kikapcsolt állapotban:
* A felhasználók nem hozhatnak létre és nem kezelhetnek FTP-fiókokat.

Beállítások testreszabása:
* Az **FTP-kiszolgáló konfigurálásához** tekintse meg a [*How-to Guides > Setup FTP](/docs/articles/user-experience/how-to-setup-ftp-in-openpanel/) részt.
* A **VSFTPD konfiguráció szerkesztéséhez** szerkessze a [`/etc/openpanel/ftp/vsftpd.conf` fájlt](https://github.com/stefanpejcic/openpanel-configuration/blob/main/ftp/vsftpd.conf).
* **Az összes ftp-fiók megtekintéséhez** használja az [*OpenAdmin > Szolgáltatások > FTP* oldalt](/docs/admin/services/ftp/).
* A **felhasználónkénti ftp fiókok számának korlátozásához** módosítsa az ftp fiókok korlátját a tárhelycsomagok létrehozásakor/szerkesztésekor.

## E-mailek

Az **`emails`** modul lehetővé teszi a felhasználók számára e-mail fiókok létrehozását és kezelését.

Ha engedélyezve van:
* A felhasználók hozzáférhetnek az [**E-mailek** oldalhoz](/docs/panel/emails/) az e-mail fiókok kezeléséhez.
* A felhasználók hozzáférhetnek a [**Webmail** oldalhoz](/docs/panel/emails/webmail/).

Kikapcsolt állapotban:
* A felhasználók nem hozhatnak létre és nem kezelhetnek e-mail fiókokat.

Beállítások testreszabása:
* Az **e-mail szerver konfigurálásához** tekintse meg a [*How-to Guides > Configure Email Server](/docs/articles/user-experience/how-to-setup-email-in-openpanel/) részt.
* Az **e-mail kliens konfigurálásához** tekintse meg a [*How-to Guides > How to setup your email client](/docs/articles/email/how-to-setup-your-email-client/).
* **Az összes e-mail fiók megtekintéséhez** használja az [*OpenAdmin > E-mailek > E-mail fiókok* oldalt](/docs/admin/emails/).
* A **webmail domain vagy közvetítő gazdagép beállításához** használja az [*OpenAdmin > E-mailek > E-mail beállítások* oldalt](/docs/admin/emails/settings/).
* A **fail2ban beállításához** tekintse meg a [*How-to Guides > Setup Fail2ban](/docs/articles/email/how-to-setup-fail2ban-mailserver-openpanel/) részt.
* Az **Rspamd** beállításához olvassa el a [*How-to Guides > RSPAMD GUI](/docs/articles/email/rspamd-gui-port-11334/) részt.
* **A DKIM domainhez való beállításához** tekintse meg a [*How-to Guides > Setup DKIM](/docs/articles/email/how-to-setup-dkim-for-mailserver/) részt.
* A **felhasználónkénti e-mail fiókok számának korlátozásához** módosítsa az e-mail fiókok korlátját a tárhelycsomagok létrehozásakor/szerkesztésekor.


## MySQL

A **`mysql`** modul lehetővé teszi a felhasználók számára, hogy mysql adatbázisokat hozzanak létre és kezeljenek.

Ha engedélyezve van:
* A MySQL/MariaDB automatikusan elindul, amikor a felhasználó eléri az Adatbázisok részt, megnyitja a phpMyAdmin programot vagy telepíti a WordPress-t.
* A felhasználók elérhetik a [**MySQL > Adatbázisok** oldalt] (/docs/panel/mysql/databases/) adatbázisok kezeléséhez.
* A felhasználók elérhetik a [**MySQL > New Database** oldalt] (/docs/panel/mysql/new_db/) adatbázisok létrehozásához.
* A felhasználók elérhetik a [**MySQL > Database Wizard** oldalt](/docs/panel/mysql/wizard/) adatbázisok, felhasználók létrehozásához és jogosultságok hozzárendeléséhez.
* A felhasználók elérhetik a [**MySQL > Root Password** oldalt](#) a root felhasználói jelszó megváltoztatásához.
* A felhasználók elérhetik a [**MySQL > Process List** oldalt] (/docs/panel/mysql/processlist/) az összes aktív folyamat megtekintéséhez.
* A felhasználók hozzáférhetnek a [**MySQL > Felhasználók** oldalhoz](/docs/panel/mysql/users/) a felhasználók kezeléséhez.
* A felhasználók hozzáférhetnek a [**MySQL > Új felhasználó** oldalhoz](/docs/panel/mysql/new_user/) felhasználók létrehozásához.
* A felhasználók elérhetik a [**MySQL > Jelszó módosítása** oldalt](#) a felhasználó jelszavának megváltoztatásához.
* A felhasználók hozzáférhetnek a [**MySQL > Felhasználó hozzárendelése a DB-hez** oldalhoz] (/docs/panel/mysql/assign/), hogy minden jogosultságot hozzárendeljenek a felhasználóhoz egy adatbázis felett.
* A felhasználók elérhetik a [**MySQL > Felhasználó eltávolítása a DB-ből** oldalt] (/docs/panel/mysql/remove/), hogy visszavonják a felhasználó összes jogosultságát az adatbázis felett.

Kikapcsolt állapotban:
* A felhasználók nem férhetnek hozzá a *MySQL* részhez.

Beállítások testreszabása:
* A mysql vagy a mariadb beállításához az összes új felhasználóhoz** használja az [*OpenAdmin > Settings > User Defaults* oldalt és a *MySQL type* opciót] (/docs/admin/settings/defaults/).
* **A mysql, a percona vagy a mariadb beállításához egyetlen felhasználóhoz** használja a [**MySQL Type** beállítást] (/docs/articles/docker/how-to-set-mysql-mariadb-per-user-in-openpanel/).
* Az alapértelmezett CPU/RAM módosításához** használja az [*OpenAdmin > Settings > User Defaults* oldalt](/docs/admin/settings/defaults/).
* A **rendszerfelhasználók hozzáférésének korlátozásához** módosítsa a [`mysql_restricted_usernames`](https://dev.openpanel.com/cli/config.html#mysql-restricted-usernames) beállítást.
* A **rendszeradatbázisokhoz való hozzáférés korlátozásához** módosítsa a [`mysql_restricted_databases`](https://dev.openpanel.com/cli/config.html#mysql-restricted-databases) beállítást.
* A **a MySQL inicializálására várakozó indítási idő növelése** növelje a [`mysql_startup_time`](https://dev.openpanel.com/cli/config.html#mysql-startup-time) tartományt.


Útmutatók:
* Az adatbázishoz való csatlakozáshoz** tekintse meg a [*How-to Guides > Connecting to MySQL Server from Applications in OpenPanel](/docs/articles/databases/how-to-connect-to-mysql-from-php-applications-in-openpanel/) című részt.
* A **hibaelhárításhoz** tekintse meg a [*Útmutatók > Hibaelhárítás: Hiba az adatbázis-kapcsolat létrehozása során] (/docs/articles/databases/how-to-troubleshoot-error-establishing-a-database-connection/) című részt.

## Távoli MySQL

A **`remote_mysql`** modul lehetővé teszi a felhasználók számára, hogy engedélyezzék/letiltsák a távoli hozzáférést a mysql-hez.

Ha engedélyezve van:
* A távoli hozzáférés alapértelmezés szerint le van tiltva.
* Felhasználónként véletlenszerű port van kijelölve a mysql-példányaikhoz.
* A felhasználók elérhetik a [**MySQL > Remote Access** oldalt] (/docs/panel/mysql/remote/) a távoli hozzáférés engedélyezéséhez/letiltásához.
* A felhasználók bármely adatbázishoz csatlakozhatnak távoli helyről, ha az opció engedélyezve van.

Kikapcsolt állapotban:
* A távoli hozzáférés le van tiltva.

Beállítások testreszabása:
* Egyik sem


## phpMyAdmin

A **`phpmyadmin`** modul lehetővé teszi a felhasználók számára a phpMyAdmin szolgáltatás kezelését.

Ha engedélyezve van:
* A phpMyAdmin a felhasználó által kezelhető.
* A phpMyAdmin egyéni, felhasználónkénti porton érhető el.

Kikapcsolt állapotban:
* A felhasználók nem férhetnek hozzá a *phpMyAdmin* szakaszhoz.

Beállítások testreszabása:
* A **a php_max_execution_time, php_memory_limit, php_upload_limit** módosításához használja az [*OpenAdmin > MySQL > phpMyAdmin](/docs/panel/mysql/phpmyadmin) parancsot.
* **A phpMyAdmin alapértelmezett CPU/RAM-jának megváltoztatásához** használja a felső sarokban található „kezelés” gombot.

Útmutatók:
* A **táblázatok adatbázisba történő importálásához** tekintse meg a [**dokumentációt**](/docs/panel/mysql/phpmyadmin/#import-sql-files).
* **Egyéni domain beállításához a phpMyAdmin számára** tekintse meg a következőt: [**Útmutatók > Egyéni tartomány a phpMyAdmin számára**](/docs/articles/databases/phpmyadmin-domain/).

## MySQL importálás

A **`mysql_import`** modul lehetővé teszi a felhasználók számára, hogy fájlokat importáljanak adatbázisaikba.

Ha engedélyezve van:
* A felhasználók elérhetik a [**MySQL > Adatbázis importálása** oldalt] (/docs/panel/mysql/import/), hogy fájlokat importáljanak egy adatbázisba.

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *MySQL > Adatbázis importálása* oldalt.

Beállítások testreszabása:
* Az importáláshoz engedélyezett maximális fájlméret beállításához** növelje a [`mysql_import_max_size_gb`](https://dev.openpanel.com/cli/config.html#mysql-import-max-size-gb) értéket.

Útmutatók:
* Az adatbázisba történő **importáláshoz** tekintse meg a [*Útmutatók > Adatbázis importálása] (/docs/articles/docker/import-database/) részt.


## MySQL Conf

A **`mysql_conf`** modul lehetővé teszi a felhasználók számára a mysql szerver konfigurációjának szerkesztését.

Ha engedélyezve van:
* A felhasználók elérhetik a [**MySQL > Konfiguráció szerkesztése** oldalt](#) a szolgáltatás .cnf fájljának szerkesztéséhez.

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *MySQL > Konfiguráció szerkesztése* oldalt.

Beállítások testreszabása:
* A **konfiguráció elérhető beállításainak megadása** szerkesztheti a fájlt: [`/etc/openpanel/mysql/keys.txt`](https://github.com/stefanpejcic/openpanel-configuration/blob/main/mysql/keys.txt).
* A mysql.cnf fájl **szerkesztéséhez egyetlen felhasználó számára** szerkesztheti a fájlt: `/home/${felhasználónév}/custom.cnf`.
* A mysql.cnf fájl **szerkesztéséhez minden új felhasználó számára** szerkesztheti a fájlt: [`/etc/openpanel/mysql/user.cnf`](https://github.com/stefanpejcic/openpanel-configuration/blob/main/mysql/user.cnf).




## Távoli PostgreSQL

A **`remote_postgresql`** modul lehetővé teszi a felhasználók számára a PostgreSQL távoli elérésének engedélyezését/letiltását.

Ha engedélyezve van:
* A távoli hozzáférés alapértelmezés szerint le van tiltva.
* A rendszer felhasználónként véletlenszerű portot oszt ki a PostgreSQL-példányokhoz.
* A felhasználók elérhetik a [**PostgreSQL > Távoli elérés** oldalt](#) a távoli hozzáférés engedélyezéséhez/letiltásához.
* A felhasználók bármely adatbázishoz csatlakozhatnak távoli helyről, ha az opció engedélyezve van.

Kikapcsolt állapotban:
* A távoli hozzáférés le van tiltva.

Beállítások testreszabása:
* Egyik sem


## pgAdmin

A **`pgadmin`** modul lehetővé teszi a felhasználók számára a pgAdmin szolgáltatás kezelését.

Ha engedélyezve van:
* A pgAdmin-t a felhasználó kezelheti.
* A felhasználók hozzáférhetnek a *pgAdmin* részhez.
* A pgAdmin egyéni, felhasználónkénti porton érhető el.

Kikapcsolt állapotban:
* A felhasználók nem férhetnek hozzá a *pgAdmin* szakaszhoz.

Beállítások testreszabása:
* **A pgAdmin alapértelmezett CPU/RAM-jának megváltoztatásához** használja a „kezelés” gombot a felső sarokban.


## PostgreSQL importálás

A **`postgresql_import`** modul lehetővé teszi a felhasználók számára, hogy fájlokat importáljanak adatbázisaikba.

Ha engedélyezve van:
* A felhasználók elérhetik a [**PostgreSQL > Adatbázis importálása** oldalt](#), hogy fájlokat importálhassanak egy adatbázisba.

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *PostgreSQL > Adatbázis importálása* oldalt.

Beállítások testreszabása:
* Egyik sem


## PostgreSQL Conf

A **`postgresql_conf`** modul lehetővé teszi a felhasználók számára a PostgreSQL szerver konfigurációjának szerkesztését.

Ha engedélyezve van:
* A felhasználók elérhetik a [**PostgreSQL > Konfiguráció szerkesztése** oldalt](#) a szolgáltatás .cnf fájljának szerkesztéséhez.

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *PostgreSQL > Konfiguráció szerkesztése* oldalt.


## Crons

A **`crons`** modul lehetővé teszi a felhasználók számára, hogy ütemezzék az [Ofelia](https://hub.docker.com/r/mcuadros/ofelia) cron-feladatokat.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Speciális > Cron Jobs** oldalt] (/docs/panel/advanced/cronjobs/).
* A felhasználók [add cronjobs](/docs/panel/advanced/cronjobs/#add)
* A felhasználók [szerkeszthetik a cronjobs-okat](/docs/panel/advanced/cronjobs/#edit)
* A felhasználók [megtekinthetik a cronjobs naplóit](/docs/panel/advanced/cronjobs/#logs)
* A felhasználók [szerkeszthetik a crons fájlt](/docs/panel/advanced/cronjobs/#file-editor)
* A felhasználók [importálhatják és exportálhatják a cronjobokat](/docs/panel/advanced/cronjobs/#import--export)

Kikapcsolt állapotban:
* A felhasználók nem érhetik el az *Advanced > Cron Jobs* oldalt, és nem módosíthatják a cronokat.

Beállítások testreszabása:
* Az **előre beállított cronjobok új felhasználók számára** szerkesztéséhez szerkessze az `/etc/openpanel/ofelia/users.ini` fájlt.
* **A cron fájl maximális fájlméretének beállításához, hogy az OpenPanel UI-n keresztül szerkeszthető legyen**, állítsa be a [`cron_max_file_size_kb`](https://dev.openpanel.com/cli/config.html#cron-max-file-size-kb) értéket.



## Folyamatkezelő

A **`process_manager`** modul lehetővé teszi a felhasználók számára az összes futó szolgáltatás folyamatainak megtekintését és leállítását.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Advanced > Process Manager** oldalt](/docs/panel/advanced/process_manager/).

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *Speciális > Folyamatkezelő* oldalt.

Beállítások testreszabása:
* Egyik sem


## Szerver Info

Az **`info`** modul segítségével a felhasználók megtekinthetik a szerverinformációkat, a tárhelyterv-információkat és az OpenPanel-információkat.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Speciális > Szerverinformáció** oldalt] (/docs/panel/advanced/server_info/).

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *Speciális > Szerverinformáció* oldalt.

Beállítások testreszabása:
* Egyik sem



## Ideiglenes linkek

Az **`temporary_links`** modul lehetővé teszi a felhasználók számára, hogy ideiglenes aldomainekkel teszteljék webhelyeiket (a linkek 15 percig érvényesek).

Ha engedélyezve van:
* A felhasználók elérhetik a [**Élő előnézet** gombot a Webhelykezelőben] (/docs/panel/applications/wordpress/#temporary-link).

Kikapcsolt állapotban:
* A felhasználók nem érhetik el az *Élő előnézet* gombot a Site Manager oldalon.

Beállítások testreszabása:
* **Proxyszolgáltatás saját üzemeltetéséhez** – tekintse meg a [How-to Guides > Temporary Links API](/docs/articles/dev-experience/selfhosted-temporary-links-api/) részt.
* **Egyéni domain konfigurálásához** frissítse az [`temporary_links` opciót](https://dev.openpanel.com/cli/config.html#temporary-links).

## Bejelentkezési előzmények

A **`login_history`** modul lehetővé teszi a felhasználók számára, hogy megtekintsék fiókjuk bejelentkezési előzményeit.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Fiók > Bejelentkezési előzmények** oldalt](/docs/panel/account/login_history/).

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *Fiók > Bejelentkezési előzmények* oldalt.

Beállítások testreszabása:
* A **felhasználónként tárolt bejelentkezések számának szabályozásához** módosítsa a „Felhasználónkénti bejelentkezési rekordok megtartása” beállítást az [OpenAdmin > Beállítások > OpenPanel] oldalon (/docs/admin/settings/openpanel/#Statistics).


## 2FA

A **`twofa`** modul lehetővé teszi a felhasználók számára, hogy engedélyezzék a kétfaktoros hitelesítést a fiókjukhoz.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Fiók > Kéttényezős hitelesítés** oldalt](/docs/panel/account/2fa).

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *Speciális > Kéttényezős hitelesítés* oldalt, és nem kezelhetik a 2FA-t.

Beállítások testreszabása:
* A **2FA widget engedélyezéséhez** használja az [*OpenAdmin > Beállítások > OpenPanel* oldalt és a *Display 2FA widget* opciót] (/docs/admin/settings/openpanel/).
* A **egy felhasználó 2FA állapotának ellenőrzéséhez** tekintse meg a [Hogyan ellenőrizhető, hogy a 2FA aktív-e az OpenPanel felhasználói fiókhoz?](https://community.openpanel.org/d/220-how-to-check-if-2fa-is-active-for-openpanel-user-account) című részt.

## Tevékenység

Az **`activity`** modul lehetővé teszi a felhasználók számára, hogy megtekintsék tevékenységi naplóikat.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Fiók > Tevékenységnapló** oldalt](/docs/panel/account/account_activity).

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *Fiók > Tevékenységnapló* oldalt.

Beállítások testreszabása:
* A **tevékenységnapló szerkesztéséhez a terminálról** nyissa meg a következő fájlt: `/etc/openpanel/openpanel/core/users/{felhasználónév}/activity.log`.
* A **felhasználónkénti sorok teljes számának beállításához** módosítsa az "activity_lines_retention" beállítást.
* A **felhasználónkénti napló teljes méretének beállításához** módosítsa az `activity_max_size_bytes` beállítást.
* A **műveletek naplózásához harmadik féltől származó beépülő modulról** olvassa el a következőt: [*Műveletek naplózása egyéni beépülő modulokból a felhasználói tevékenységnaplóban*](https://community.openpanel.org/d/218-how-to-log-actions-from-custom-plugins-in-user-activity-log)


## Biztonsági mentések

A **`backups`** modul lehetővé teszi a felhasználók számára, hogy m=konfigurálják saját biztonsági másolataikat: miről kell biztonsági másolatot készíteni, cél, megőrzés, ütemezés stb.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Fájlok > Biztonsági másolatok** oldalt](/docs/panel/files/backups/).
* A felhasználók beállíthatják a biztonsági mentés ütemezését, a titkosítást, a megőrzést és a célállomást.

Kikapcsolt állapotban:
* A felhasználók nem férhetnek hozzá a *Fájlok > Biztonsági másolatok* oldalhoz.
* [A rendszergazdáknak be kell állítaniuk a biztonsági mentéseket a felhasználó számára](/docs/articles/backups/configuring-backups/#1-admin-configured).



## Szolgáltatások

A **`services`** modul lehetővé teszi a felhasználók számára a szolgáltatások engedélyezését/letiltását a Docker modul nélkül.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Speciális > Szolgáltatások** oldalt](/docs/panel/advanced/services/).
* A felhasználók engedélyezhetik/letilthatják a szolgáltatásokat.
* A felhasználó megtekintheti a szolgáltatás aktuális állapotát, az erőforrás-használatot (CPU, memória, lemez I/O, PID-k...), tároló nevét (más tárolókból a szolgáltatáshoz való csatlakozáshoz használandó).
* A felhasználók megtekinthetik a szolgáltatások naplóit.

Kikapcsolt állapotban:
* A felhasználók nem férhetnek hozzá a *Speciális > Szolgáltatások* oldalhoz.


## Memcached

A **`memcached`** modul lehetővé teszi a felhasználók számára a Memcached szolgáltatás engedélyezését/letiltását.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Caching > Memcached** oldalt](/docs/panel/caching/memcached/).
* A felhasználók engedélyezhetik/letilthatják a Memcached szolgáltatást.
* A felhasználó más tárolókból csatlakozhat a példányhoz a következő használatával: `elasticsearch:11211`
* A felhasználók megtekinthetik a Memcached szolgáltatás naplóit.

Kikapcsolt állapotban:
* A felhasználók nem férhetnek hozzá a *Gyorsítótár > Memcached* oldalhoz.

## Redis

A **`redis`** modul lehetővé teszi a felhasználók számára a Redis szolgáltatás engedélyezését/letiltását.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Caching > Redis** oldalt](/docs/panel/caching/redis/).
* A felhasználók engedélyezhetik/letilthatják a Memcached szolgáltatást.
* A felhasználó más tárolókból is csatlakozhat a példányhoz a `redis:6379` használatával
* A felhasználók megtekinthetik a Redis szolgáltatás naplóit.

Kikapcsolt állapotban:
* A felhasználók nem férhetnek hozzá a *Caching > Redis* oldalhoz.


## ElasticSearch

Az **`elasticsearch`** modul lehetővé teszi a felhasználók számára az ElasticSearch szolgáltatás engedélyezését/letiltását.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Caching > ElasticSearch** oldalt](/docs/panel/caching/elasticsearch/).
* A felhasználók engedélyezhetik/letilthatják az ElasticSearch szolgáltatást.
* A felhasználó más tárolókból csatlakozhat a példányhoz a következő használatával: `elasticsearch:9200`
* A felhasználók megtekinthetik az ElasticSearch szolgáltatás naplóit.

Kikapcsolt állapotban:
* A felhasználók nem férhetnek hozzá a *Gyorsítótár > ElasticSearch* oldalhoz.



## OpenSearch

Az **`opensearch`** modul lehetővé teszi a felhasználók számára az OpenSearch szolgáltatás engedélyezését/letiltását.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Caching > OpenSearch** oldalt](/docs/panel/caching/opensearch/).
* A felhasználók engedélyezhetik/letilthatják az OpenSearch szolgáltatást.
* A felhasználó más tárolókból is csatlakozhat a példányhoz az "opensearch:9200" használatával
* A felhasználók megtekinthetik az OpenSearch szolgáltatás naplóit.

Kikapcsolt állapotban:
* A felhasználók nem férhetnek hozzá a *Gyorsítótár > OpenSearch* oldalhoz.


## Lemezhasználat Explorer

A **`disk_usage`** modul lehetővé teszi a felhasználók számára a lemezhasználat könyvtáronkénti megtekintését.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Fájlok > Lemezhasználat** oldalt](/docs/panel/files/disk_usage/).

Kikapcsolt állapotban:
* A felhasználók nem férhetnek hozzá a *Fájlok > Lemezhasználat* oldalhoz.


## Inodes Explorer

A **`disk_usage`** modul lehetővé teszi a felhasználók számára a lemezhasználat könyvtáronkénti megtekintését.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Fájlok > Inodes Explorer** oldalt] (/docs/panel/files/inodes/).

Kikapcsolt állapotban:
* A felhasználók nem férhetnek hozzá a *Fájlok > Inodes Explorer* oldalhoz.





## Automatikus telepítő
Az **`autoinstaller`** modul lehetővé teszi a felhasználók számára a WordPress, a website Builder, a Mautic, a Python/NodeJS alkalmazások stb. automatikus telepítését.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Webhelyek > Automatikus telepítő** oldalt] (/docs/panel/applications/autoinstaller/).

Kikapcsolt állapotban:
* A felhasználók nem férhetnek hozzá a *Webhelyek > Automatikus telepítő* oldalhoz.


## PHP.INI szerkesztő
A **`php_ini`** modul lehetővé teszi a felhasználók számára a PNP.INI fájlok szövegszerkesztővel történő szerkesztését.

Ha engedélyezve van:
* A felhasználók elérhetik a [**PHP > PHP.INI Editor** oldalt](/docs/panel/php/php_ini_editor/).

Kikapcsolt állapotban:
* A felhasználók nem férhetnek hozzá a *PHP > PHP.INI Editor* oldalhoz.


## WordPress

A **`wordpress`** modul lehetővé teszi a felhasználók számára WordPress webhelyek telepítését és kezelését.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Webhelyek > WP Manager** oldalt](/docs/panel/applications/wordpress/).
* A felhasználók [a WP Manager segítségével kezelhetik a WordPress webhelyeket] (/docs/panel/applications/wordpress/#site-manager).
* A WordPress elérhető az Automatikus telepítő oldalon.
* A felhasználók [telepíthetik a WordPress-t az Auto Installer segítségével] (/docs/panel/applications/wordpress/#install-wordpress).
* A felhasználók [ellenőrizhetik és importálhatják a meglévő telepítéseket] (/docs/panel/applications/wordpress/#scanning-importing-installations).
* A felhasználók [beállíthatják a témákat és a beépülő modulokat automatikus telepítésre] (/docs/panel/applications/wordpress/#themes-and-plugins-sets).

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *Webhelyek > WP Manager* oldalt.
* A WordPress nem érhető el az Autoinstallerben.
* A WordPress webhelyek nem kezelhetők Openpanelen keresztül.

Beállítások testreszabása:
* A **témák vagy beépülő modulok új telepítéseknél történő automatikus telepítéséhez** olvassa el a következőt: [*WordPress Themes and Plugins Sets*](/docs/articles/websites/wordpress-plugins-themes-sets-in-openpanel/)
* **Egyéni Google PageSpeed ​​Insights API-kulcs hozzáadásához** tekintse meg: [*Útmutatók > Google PageSpeed ​​Insights API-kulcs*](/docs/articles/websites/google-pagespeed-insights-api-key/)
* Egy mu-plugin beállításához minden új webhelyen** szerkessze az `/etc/openpanel/wordpress/mu-plugin.php` fájlt.
* Ha **egyéni WP-CLI-t szeretne beállítani az összes webhelyhez**, cserélje ki az `/etc/openpanel/wordpress/wp-cli.phar` fájlt.
* **Az új webhelyekhez használt .htaccess fájlok testreszabása** a `/etc/openpanel/wordpress/htaccess/` mappában lévő fájlok szerkesztése.

## Webhelykészítő

A **`website_builder`** modul lehetővé teszi a felhasználók számára, hogy egyszerű webhelyeket hozzanak létre a HTML Drag & Drop Website Builder segítségével.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Webhelyek > Webhelykészítő** oldalt](/docs/panel/applications/builder/).
* A felhasználók [kezelhetik a statikus webhelyeket a Site Manager segítségével](/docs/panel/applications/builder/#edit-website).
* A Website Builder az Automatikus telepítő oldalon érhető el.
* A felhasználók [statikus webhelyeket hozhatnak létre az Auto Installer segítségével] (/docs/panel/applications/builder/#create-a-website).

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *Webhelyek > Webhelykészítő* oldalt.
* A Website Builder nem érhető el az Autoinstallerben.
* A statikus webhelyek nem kezelhetők Openpanelen keresztül.

## Mautic

A **`mautic`** modul lehetővé teszi a felhasználók számára a Mautic telepítését és kezelését az OpenPanelről.

> **MEGJEGYZÉS:** Ezt a modult már nem karbantartják aktívan, és nem szabad élesben használni (*BETA* címke).

Ha engedélyezve van:
* A felhasználók [kezelhetik a Mautic webhelyeket a Site Manager segítségével] (/docs/panel/applications/).
* A Mautic elérhető az AutoInstaller oldalon.
* A felhasználók [telepíthetik a Mautic-ot az AutoInstaller segítségével] (/docs/articles/websites/how-to-install-mautic-with-openpanel/).

Kikapcsolt állapotban:
* A Mautic nem érhető el az Autoinstallerben.

## ClamAV

A **`malware_scanner`** modul elindít egy ClamAV szolgáltatást, és lehetővé teszi a felhasználók számára a fájlok vizsgálatát.

> **MEGJEGYZÉS:** Ezt a modult már nem karbantartják aktívan, és nem szabad élesben használni (*DEPRECATED* címke).

Ha engedélyezve van:
* A felhasználók elérhetik a [**Fájlok > Malware Scanner** oldalt](/docs/panel/files/malware-scanner/).
* A ClamAV szolgáltatás elindul a szerveren.

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *Fájlok > Malware Scanner* oldalt.
* A ClamAV szolgáltatás nem indul el a szerveren.

Beállítások testreszabása:
* **A ClamAV szolgáltatás cpu/memóriakorlátainak testreszabásához** tekintse meg a következőt: [*OpenAdmin > Services > Service Limits*](/docs/admin/services/limits/).



## Fájlok

A **`fájlok`** modul lehetővé teszi a felhasználók számára a fájlok és mappák kezelését a Fájlkezelő segítségével.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Fájlok > Fájlkezelő** oldalt](/docs/panel/files/).
* A Fájlkezelő hivatkozások más oldalakon is elérhetők: Domains, WP Manager stb.

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *Fájlok > Fájlkezelő* oldalt.
* Más oldalakon nem jelennek meg hivatkozások a fájlok kezelésére.




## Domainek

A **`domains`** modul lehetővé teszi a felhasználók számára tartományok hozzáadását és kezelését.

Ha engedélyezve van:
* A felhasználók hozzáférhetnek a [**Domains** oldalhoz](/docs/panel/domains/).
* A felhasználók kezelhetik a domaineket.
* A felhasználók a menüben érhetik el a „Domains” aloldalakat.

Kikapcsolt állapotban:
* A felhasználók nem férhetnek hozzá a *Domains* oldalhoz.
* A felhasználók nem kezelhetnek domaineket.

Beállítások testreszabása:
* **A HSTS engedélyezése egy domain számára**: [*How-to Guides > How to Enable HSTS on a Domain in OpenPanel*](/docs/articles/domains/how-to-enable-hsts-on-a-domain-in-openpanel/)
* Az **alapértelmezett oldalak testreszabásához** lásd: [*OpenAdmin > Domains > Edit Domain Templates*](/docs/admin/domains/file_templates/)


## Nyers hozzáférési naplók

A **`domain_logs`** modul lehetővé teszi a felhasználók számára, hogy megtekintsék a tartományaik nyers hozzáférési naplóját.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Domainek > Nyers hozzáférési naplók** oldalt] (/docs/panel/domains/docroot/).

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *Domainek > Nyers hozzáférési naplók* oldalt.



## GoAccess

A **`goaccess`** modul ütemezetten futtatja a GoAccess szolgáltatást a nyers tartománynaplók feldolgozásához és az OpenPanel UI-n keresztül elérhető HTML-jelentések elkészítéséhez.

Ha engedélyezve van:
* A GoAccess szolgáltatás fut a szerveren.
* A felhasználók elérhetik a [**Domainek > GoAccess** oldalt](/docs/panel/domains/goaccess/).

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *Domains > GoAccess* oldalt.

Beállítások testreszabása:
* A **GoACcess jelentéskészítés letiltásához** frissítse a [*`goaccess_enable` értéket*](https://dev.openpanel.com/cli/config.html#goaccess-enable)
* **A jelentések generálási gyakoriságának módosításához (alapértelmezett = @napi)** módosítsa a [`domains-stats` cron] (https://dev.openpanel.com/crons.html#domains-stats) ütemezését és a [`goaccess_schedule` értékét](https://dev.openpanel.com/cli/scheduleccess.html-schedule).
* **Az adatok manuális generálásához** futtassa a [`domains-stats` cron](https://dev.openpanel.com/crons.html#domains-stats) parancsot.
* A jelentések **újragenerálásának kényszerítéséhez* tekintse meg: [*OpenCLI-dokumentáció > Domain-hozzáférési naplók elemzése*](https://dev.openpanel.com/cli/domains.html#Parse-domain-access-logs).


## Docroot

A **`docroot`** modul lehetővé teszi a felhasználók számára, hogy egyéni docroot-ot (mappát) állítsanak be tartományok hozzáadásakor, majd később módosítsák az elérési utat.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Domains > Change Docroot** oldalt](/docs/panel/domains/docroot/).
* A felhasználók egyéni docroot-ot állíthatnak be a domain hozzáadásakor.

Kikapcsolt állapotban:
* A felhasználók nem állíthatnak be egyéni docroot-ot egy domain hozzáadásakor, és később sem módosíthatják a docroot-ot.

## Átirányítások

Az **`átirányítások`** modul lehetővé teszi a felhasználók számára, hogy átirányításokat hozzanak létre tartományokhoz.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Domainek > Átirányítások** oldalt](/docs/panel/domains/redirects/).

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *Domainek > Átirányítások* oldalt.



## Domainek nagybetűvel

A **`capitalize_domains`** modul lehetővé teszi a felhasználók számára, hogy a tartomány nagybetűs verzióját állítsák be az OpenPanelben való megjelenítéshez.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Domainek > Domainek nagybetűs írása** oldalt] (/docs/panel/domains/capitalize/).

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *Domainek > Domainek nagybetűs írása* oldalt.


## VirtualHosts szerkesztése

Az **`edit_vhost`** modul lehetővé teszi a felhasználók számára, hogy szerkesszék a VirtualHosts fájlokat a tartományukhoz.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Domains > Edit VHosts File** oldalt](/docs/panel/domains/vhosts/).

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *Domains > Edit VHosts File* oldalt.

Beállítások testreszabása:
* Az Apache/Nginx/OpenLiteSpeed** vhost-fájlok testreszabásához** tekintse meg: [*OpenAdmin > Domains > Edit Domain Templates*](/docs/admin/domains/file_templates/#apache-virtualhost)


## Webszerver

A **`webserver_conf`** modul lehetővé teszi a felhasználók számára, hogy szerkesszék a webszervereik fő konfigurációs fájljait.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Speciális > Webszerver-beállítások** oldalt] (/docs/panel/advanced/webserver_settings/).
* A felhasználók szerkeszthetik az Apache `httpd.conf` fájlját.
* A felhasználók szerkeszthetik az Nginx/OpenResty `nginx.conf` fájlját.
* A felhasználók szerkeszthetik az OpenLiteSpeed ​​`openlitespeed.conf` fájlját.

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *Speciális > Webszerver beállításai* oldalt.

Beállítások testreszabása:
* Az Apache alapértelmezett `httpd.conf` fájljának testreszabásához** módosítsa az `/etc/openpanel/apache/httpd.conf` fájlt.
* Az alapértelmezett `nginx.conf` fájl testreszabásához** módosítsa az `/etc/openpanel/nginx/nginx.conf` fájlt.
* Az OpenLiteSpeed ​​alapértelmezett `openlitespeed.conf` fájljának testreszabásához** módosítsa az `/etc/openpanel/openlitespeed/httpd_config.conf` fájlt.
* Az alapértelmezett `nginx.conf` fájl testreszabásához** az OpenResty számára módosítsa az `/etc/openpanel/openresty/nginx.conf` fájlt.

## DNS

A **`dns`** modul helyi BIND9 szolgáltatást futtat, zónafájlokat hoz létre a tartományokhoz, és lehetővé teszi a felhasználók számára a DNS-rekordok kezelését.

Ha engedélyezve van:
* A BIND9 szolgáltatás fut a szerveren.
* A felhasználók elérhetik a [**Domains > DNS Zone Editor** oldalt](/docs/panel/domains/dns/).
* A DNS-zónafájlok új tartományokhoz jönnek létre.
* A felhasználók kezelhetik a DNS rekordokat.
* A 'Zóna szerkesztése' linkek a domainekhez az *OpenPanel > Domains* oldalon érhetők el.
* A rendszergazdák elérhetik az [**OpenAdmin > Domains > DNS Cluster** oldalt](/docs/admin/domains/dns-cluster/).
* A rendszergazdák elérhetik az [**OpenAdmin > Domains > Edit Zone Templates** oldalt](/docs/admin/domains/dns_templates/).
* A rendszergazdák elérhetik az [**OpenAdmin > Domains > DNS Zone Editor** oldalt](/docs/admin/domains/dns/).

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *Domains > DNS Zone Editor* oldalt.
* A rendszergazdák nem férhetnek hozzá az OpenAdmin *DNS Zone Editor*, *Edit Zone Templates* és *DNS Cluster* oldalaihoz.

Beállítások testreszabása:
* A **névszerverek konfigurálásához** tekintse meg a következőt: [*Útmutatók > Névszerverek konfigurálása*](/docs/articles/domains/how-to-configure-nameservers-in-openpanel/)
* A **DNS-zóna sablonok testreszabásához** tekintse meg a következőt: [OpenAdmin > Domains > Edit Zone Templates](/docs/admin/domains/dns_templates/)
* **DNS-fürt konfigurálásához** tekintse meg: [*Útmutatók > DNS-fürtözés*](/docs/articles/domains/how-to-setup-dns-cluster-in-openpanel/)




## WAF

A **`waf`** modul egyéni Caddy-képet futtat a CorazaWAF-fel, és lehetővé teszi a felhasználók számára, hogy tartományonként kezeljék a WAF-szabályokat és a be-/kikapcsolásvédelmet.

Ha engedélyezve van:
* `openpanel/caddy-coraza` dokkolóképet használnak.
* Az `/etc/openpanel/caddy/templates/domain.conf_with_modsec` sablon az új tartományokhoz használatos.
* A felhasználók elérhetik a [**Speciális > WAF** oldalt](/docs/panel/advanced/waf/).
* Az [OWASP CRS](https://github.com/coreruleset/coreruleset) telepítéskor be van állítva.
* A felhasználók tartományonként szerkeszthetik a WAF-szabályokat, és engedélyezhetik/letilthatják a védelmet.
* [A „Tűzfal” widget megjelenik a Site Managerben](/docs/panel/applications/wordpress/#firewall).

Kikapcsolt állapotban:
* `caddy:latest` dokkolókép használatos.
* Az `/etc/openpanel/caddy/templates/domain.conf` sablon az új tartományokhoz használatos.
* A felhasználók nem érhetik el a *Speciális > WAF* oldalt.
* A „Tűzfal” widget nem jelenik meg a Site Managerben.

Beállítások testreszabása:
* A **névszerverek konfigurálásához** tekintse meg a következőt: [*Útmutatók > Névszerverek konfigurálása*](/docs/articles/domains/how-to-configure-nameservers-in-openpanel/)
* A **DNS-zóna sablonok testreszabásához** tekintse meg a következőt: [OpenAdmin > Domains > Edit Zone Templates](/docs/admin/domains/dns_templates/)
* **DNS-fürt konfigurálásához** tekintse meg: [*Útmutatók > DNS-fürtözés*](/docs/articles/domains/how-to-setup-dns-cluster-in-openpanel/)



## PHP

A **`php`** modul lehetővé teszi a felhasználók számára a PHP verziók és beállítások kezelését.

Ha engedélyezve van:
* A felhasználók elérhetik a [**PHP > PHP verzió kiválasztása** oldalt](/docs/panel/php/domains/).
* A felhasználók elérhetik a [**PHP > Default Version** oldalt] (/docs/panel/php/default/).
* A felhasználók elérhetik a [**PHP > Extensions** oldalt](/docs/panel/php/extensions/).
* A felhasználók beállíthatják a PHP verzióját domainenként, beállíthatják az alapértelmezett verziót az új tartományokhoz, szerkeszthetik a beállításokat és megtekinthetik a telepített bővítményeket.

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *Select PHP Version*, *Default Version*, *Options*, *Extensions* oldalakat.
* A felhasználók nem állíthatnak be PHP-verziót domainenként, nem állíthatnak be alapértelmezett verziót az új tartományokhoz, nem szerkeszthetik a beállításokat és nem tekinthetik meg a telepített bővítményeket.

Beállítások testreszabása:
* Az új felhasználók számára használandó alapértelmezett PHP-verzió beállításához** tekintse meg: [*OpenAdmin > Beállítások > Felhasználói alapbeállítások szerkesztése > Alapértelmezett PHP-verzió*](/docs/panel/php/options/#available-options)
* **Az alapértelmezett cpu/memóriakorlátok beállításához a PHP verziókhoz és a további PHP-beállításokhoz** tekintse meg a következőt: [*OpenAdmin > Beállítások > Felhasználói alapbeállítások szerkesztése > Szolgáltatások*](/docs/panel/php/options/#available-options)
* **PHP-bővítmény telepítéséhez** olvassa el: [*How-to Guides > How to install a PHP-bővítmény az OpenPanelben*](/docs/articles/websites/how-to-install-php-extensions-in-openpanel/).
* A **PHP INI memóriakorlát növeléséhez** olvassa el a következőt: [*How-to Guides > Hogyan állítsunk be vagy növeljünk PHP INI memóriakorlátot vagy egyéb értékeket?*](/docs/articles/websites/how-to-set-or-increase-PHP-INI-memory-limit-or-other-values/).
* **A PHP-beállítások webhelyenkénti beállításához** tekintse meg: [*Útmutatók > PHP-beállítások webhelyenként (mappánként)*](/docs/articles/websites/php-user-ini-files/).
* Az **alapértelmezett .INI fájlok szerkesztéséhez** tekintse meg az **OpenAdmin > Beállítások > PHP beállítások > Alapértelmezett PHP.INI fájlok** részt, vagy szerkessze a fájlokat az `/etc/openpanel/php/ini` mappában.

## PHP opciók

A **`php_options`** modul lehetővé teszi a felhasználók számára, hogy kezeljék a PHP verzióik opcióit (korlátait).

Ha engedélyezve van:
* A felhasználók elérhetik a [**PHP > Options** oldalt](/docs/panel/php/options/).

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *PHP Options* oldalt.

Beállítások testreszabása:
* **A felhasználók számára elérhető PHP-beállítások testreszabásához** tekintse meg az **OpenAdmin > Beállítások > PHP-beállítások > Elérhető beállítások** oldalt, vagy szerkessze az */etc/openpanel/php/options.txt* fájlt.



## Alkalmazások

A **`pm2`** modul lehetővé teszi a felhasználók számára a konténeres Python és NodeJS alkalmazások beállítását és kezelését.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Webhelyek > WP Manager** oldalt](/docs/panel/applications/wordpress/).
* A felhasználók [kezelhetik a Python és NodeJS alkalmazásokat a Site Manager segítségével] (/docs/panel/applications/pm2/#manage-applications).
* A NodeJS és a Python elérhető az Autoinstaller oldalon.
* A felhasználók [NodeJS és Python alkalmazásokat állíthatnak be az Auto Installer segítségével] (/docs/panel/applications/pm2/#create-an-application).

Kikapcsolt állapotban:
* A NodeJS és a Python nem érhető el az Autoinstaller oldalon.
* A NodeJS és Python alkalmazások nem kezelhetők Openpanelen keresztül.

Beállítások testreszabása:
* **A Docker szolgáltatássablon testreszabásához az új Python-alkalmazásokhoz** szerkessze az `/etc/openpanel/docker/compose/python.yml` fájlt.
* A **docker szolgáltatássablon testreszabásához az új Node.JS alkalmazásokhoz** szerkessze az `/etc/openpanel/docker/compose/nodejs.yml` fájlt.
* **Az új python/node alkalmazás Nginx-proxyjának fejléceinek testreszabásához** szerkessze a `/etc/openpanel/nginx/vhosts/1.1/nginx_proxy_headers.txt` fájlt.
* **Egyéni Google PageSpeed ​​Insights API-kulcs hozzáadásához** tekintse meg: [*Útmutatók > Google PageSpeed ​​Insights API-kulcs*](/docs/articles/websites/google-pagespeed-insights-api-key/)




## Erőforrások használata

A **`usage`** modul lehetővé teszi a felhasználók számára, hogy megtekintsék szolgáltatásaik erőforrás-használatát.

Ha engedélyezve van:
* A felhasználók elérhetik a [**Speciális > Erőforráshasználat** oldalt](/docs/panel/advanced/resource_usage/).
* A felhasználók [kezelhetik a Python és NodeJS alkalmazásokat a Site Manager segítségével] (/docs/panel/applications/pm2/#manage-applications).
* A NodeJS és a Python elérhető az Autoinstaller oldalon.
* A felhasználók [NodeJS és Python alkalmazásokat állíthatnak be az Auto Installer segítségével] (/docs/panel/applications/pm2/#create-an-application).

Kikapcsolt állapotban:
* A felhasználók nem érhetik el a *Speciális > Erőforráshasználat* oldalt.

Beállítások testreszabása:
* Az **oldalbeállítások szerkesztéséhez** tekintse meg: [**OpenAdmin > Beállítások > OpenPanel > Statisztika** oldal](/docs/admin/settings/openpanel/#statistics).
* A **statisztikagyűjtés gyakoriságának módosításához (alapértelmezett = @óránként)** módosítsa a [`docker-collect_stats --all` cron] ütemezését (https://dev.openpanel.com/crons.html#docker-collect-stats-all).
* **Egyetlen kombinált vagy külön diagram megjelenítéséhez a cpu/ram számára** módosítsa a [`resource_usage_charts_mode` értéket](https://dev.openpanel.com/cli/config.html#resource-usage-charts-mode).
* Az oldalankénti elemek számának módosításához** módosítsa a [`resource_usage_items_per_page` értékét](https://dev.openpanel.com/cli/config.html#resource-usage-items-per-page).
* A** **elforgatásához szerkessze a [`resource_usage_retention` értékét](https://dev.openpanel.com/cli/config.html#resource-usage-retention).


