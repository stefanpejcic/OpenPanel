---
sidebar_position: 7
---

# Tevékenységnapló

A Fióktevékenység oldal naplóbejegyzést ad az OpenPanel felületen végrehajtott összes műveletéről, az időbélyeggel és az IP-címmel együtt, amelyről a műveletet végrehajtották. Ennek az oldalnak az elsődleges célja, hogy betekintést nyújtson abba, hogy ki hajtott végre bizonyos műveleteket, mint például egy fájl törlése, tartományok hozzáadása, a WordPress rendszergazdai jelszavak visszaállítása stb.

![account_activity_log.png](/img/panel/v2/activity.png)


## Rögzített műveletek:

Az OpenPanel felület a következő fióktevékenységeket rögzíti:

* Elem hozzáadva a kedvencekhez
* Eltávolítva az elemet a kedvencekből
* Jelszó megváltozott
* E-mail cím megváltozott
* Felhasználónév megváltozott
* Elfelejtett jelszó kérése (e-mailben)
* Jelszó visszaállítása (e-mailben)
* Jelszóval bejelentkezve
* 2FA kóddal jelentkezett be
* API-híváson keresztül bejelentkezve
* Kijelentkezett
* Az értesítési beállítások megváltoztak
* Megszakított aktív munkamenet
* Engedélyezett vagy letiltott kéttényezős hitelesítés (2FA)
* Engedélyezve vagy letiltva **ElasticSearch**
* Engedélyezve vagy letiltva **Memcached**
* Engedélyezve vagy letiltva **OpenSearch**
* Engedélyezve vagy letiltva **Redis**
* Engedélyezve vagy letiltva **Lakk**
* Engedélyezett vagy letiltott **Lakk gyorsítótár** egy domainhez
* Új DNS rekord hozzáadva egy domainhez
* Frissített DNS rekord egy domainhez
* Egy domain DNS-rekordja törölve
* Domain DNS-zóna visszaállítása
* Exportált DNS-zóna egy tartományhoz
* Egyéni Docker-szolgáltatás hozzáadása, szerkesztése vagy törlése
* Váltott MySQL típus
* Váltott webszerver típus
* Megváltozott képcímke a Docker szolgáltatáshoz
* Megváltozott a CPU vagy a memória korlátja a Docker szolgáltatáshoz
* Engedélyezett vagy letiltott Docker szolgáltatás
* Kényszerített kép a Docker szolgáltatáshoz
* Parancs végrehajtása webterminálon keresztül
* Egy domainhez szerkesztett VirtualHosts fájl
* Létrehozott vagy törölt átirányítási hivatkozás egy domainhez
* Domain hozzáadva
* Törölt domain
* Felfüggesztett vagy fel nem függesztett domain
* Engedélyezett vagy letiltott WAF egy domainhez
* Frissített WAF-szabályok egy domainhez
* Hozzáadott e-mail cím
* Módosított e-mail cím (jelszó, kvóta, bejövő vagy kimenő forgalom felfüggesztése/felfüggesztése)
* Letöltött e-mail konfiguráció az Outlook/Thunderbird számára
* Törölt e-mail cím
* Hozzáférés a webmailhez e-mailhez
* Szerkesztett szitaszűrő e-mailhez
* Létrehozott fájl vagy mappa
* Feltöltött fájlok
* Letöltött fájlok az URL-ről
* Archívum létrehozva
* Kivont archívum
* Átnevezett fájl vagy mappa
* Törölt fájlok a kukába
* Visszaállított fájlok a kukából
* Kiürített kuka
* Véglegesen törölt fájlok
* Megváltozott fájl/mappa engedélyek
* Másolt vagy áthelyezett fájlok/mappák
* Fájl szerkesztve a Fájlszerkesztővel
* Létrehozott vagy törölt FTP-fiók
* Megváltozott az FTP-fiók jelszava
* Megváltozott az FTP-fiók elérési útja
* Letöltött FTP konfiguráció (FileZilla/Cyberduck)
* Elindított ClamAV-keresés a mappához
* MySQL adatbázis létrehozása
* Létrehozott MySQL adatbázis-felhasználó
* Kiosztott vagy visszavont minden jogosultságot egy felhasználó számára egy MySQL adatbázisban
* Megváltozott MySQL root felhasználói jelszó
* Törölt MySQL adatbázis vagy felhasználó
* Megváltozott a MySQL adatbázis-felhasználó jelszava
* Szerkesztett MySQL konfiguráció
* Engedélyezett vagy letiltott távoli MySQL hozzáférés
* PostgreSQL adatbázis létrehozása
* PostgreSQL adatbázis-felhasználó létrehozva
* Kiosztott vagy visszavont minden jogosultságot egy felhasználó számára egy PostgreSQL adatbázisban
* Törölt PostgreSQL adatbázis vagy felhasználó
* Megváltozott a PostgreSQL adatbázis-felhasználó jelszava
* Szerkesztett PostgreSQL konfiguráció
* Engedélyezett vagy letiltott távoli PostgreSQL hozzáférés
* Engedélyezve vagy letiltva **pgAdmin**
* A **pgAdmin** szerkesztett verziója, jelszava vagy e-mail címe
* Engedélyezett vagy letiltott **phpMyAdmin**
* A **phpMyAdmin** maximális végrehajtási ideje, memóriakorlátja, feltöltési korlátja, verziója vagy abszolút URI-je szerkesztve
* Megváltozott az alapértelmezett PHP verzió az új domainekhez
* Megváltozott egy domain PHP verziója
* A PHP-verziók korlátai módosítva
* Szerkesztett PHP konfiguráció a PHP Selector segítségével
* Új Python / NodeJS alkalmazás készült
* Elindított, leállított vagy újraindított alkalmazás
* Szerkesztett alkalmazás
* Törölt alkalmazás
* A "pip", "npm" vagy "pnpm install" parancs végrehajtása
* Telepített, leválasztott vagy eltávolított WordPress
* Biztonsági másolatból visszaállított WordPress fájlok vagy adatbázis
* Teljes/fájlok/adatbázisok WordPress biztonsági mentése generált
* WordPress-telepítések keresése kezdeményezett
* Generált automatikus bejelentkezési link a wp-admin számára
* WP gyorsítótár kiürült
* Végrehajtott `wp-cli` parancsokat
* Elindította a PageSpeed ​​adatfrissítést
* Frissített WordPress hibakeresési lehetőségek
* Frissített WordPress webhelyinformációk
* Szerkesztette a WordPress automatikus frissítési beállításait
* Telepített, leválasztott vagy eltávolított Website Builder
* A folyamat megszakadt a Process Manager segítségével
* Szolgáltatás engedélyezése vagy letiltása
* Frissített erőforrás-használat a fiókhoz


Ezekre a műveletekre kereshet IP, felhasználónév, dátum vagy konkrét művelet alapján.

Alapértelmezés szerint oldalanként csak 25 művelet jelenik meg, navigálhat az oldalakon a lapozási hivatkozások segítségével, vagy kattintson az "Összes tevékenység megjelenítése" gombra az összes művelet megjelenítéséhez.
