# Hibaelhárítás: Hiba az adatbázis-kapcsolat létrehozása során

Ha webhelye/alkalmazása figyelmezteti Önt, hogy hiba történt az adatbázis-kapcsolat létrehozása során, tegye meg a következő lépéseket a probléma elhárításához:

![dberror.png](https://i.postimg.cc/Kzdwy4HN/example.png)

## Ellenőrizze a konfigurációs fájlt, hogy vannak-e hibák

Az OpenPanel minden felhasználói szolgáltatást saját tárolójában futtat, és helyi hálózatokat használ az elkülönítésre. Ez azt jelenti, hogy az alkalmazások nem a "localhost" vagy a "127.0.0.1" segítségével csatlakoznak az adatbázishoz, hanem tároló gazdagépneveken keresztül.

Ellenőrizheti az adatbázis-kapcsolat paramétereit, például az állomásnevet és a portot az OpenPanel > MySQL/Databases oldalon, ha a tároló nevére viszi az egérmutatót az oldal tetején:

![hostnameport.png](https://i.postimg.cc/d1L3xNbm/hostnameport.png)

Hasonlítsa össze ezeket a paramétereket a webhelyek konfigurációs fájljában található paraméterekkel, ebben a példában egy WordPress-telepítés wp-config.php fájlját szerkesztjük a Fájlkezelő segítségével:

![configbad.png](https://i.postimg.cc/W3QCPtHW/configbad.png)

Mivel a localhost a fájlon belül volt beállítva, a kapcsolat meghiúsul, a probléma megoldásához lecseréljük a localhost-ot a megfelelő gazdagépnévre (esetünkben a mariadb), és a szerkesztő jobb felső sarkában kattintson a Mentés gombra:

![configfixed.png](https://i.postimg.cc/d1zpKDx5/configfixed.png)

A módosítás mentése után az új konfigurációt alkalmazzuk, és a WordPress telepítésünk sikeresen csatlakozik az adatbázishoz:

![examplefixed.png](https://i.postimg.cc/Znv9Csjm/examplefixed.png)

## Győződjön meg arról, hogy az adatbázis-szolgáltatás (tároló) fut

A legegyszerűbb módja annak, hogy megbizonyosodjon arról, hogy az adatbázis-szolgáltatás fut, az Openpanel > MySQL/Databases oldal megnyitásával, ha az adatbázisok szerepelnek ezen az oldalon, ez azt jelenti, hogy a tároló rendben van és fut.

Ha ehelyett egy üres oldalt vagy egy üres táblát kap, az azt jelenti, hogy az adatbázis-tároló inaktív, az oldal egyszerű elérése egy triggert küld a tároló automatikus elindításához.

Ha azonban néhány percen belül nem jelennek meg az adatok az oldalon, probléma van az adatbázis-tárolóval.

Ha a Docker szolgáltatás engedélyezve van a tervben, ellenőrizheti az adatbázis-tároló naplóit az OpenPanel > Docker/Logs oldalon, valamint leállíthatja és elindíthatja a tárolót az OpenPanel > Docker/Containers oldalon.

![containers.png](https://i.imgur.com/GGvHfXb.png)

Ha a Docker funkció nincs engedélyezve a tervhez, és az adatok nem jelennek meg az Openpanel > MySQL/Databases oldalon, lépjen kapcsolatba a kiszolgáló rendszergazdájával, hogy további ellenőrzéseket végezhessenek.

## Győződjön meg arról, hogy a webhely adatbázis-felhasználója megfelelően hozzá van rendelve az adatbázishoz

Bármikor hozzárendelhet egy felhasználót egy adatbázishoz az OpenPanel > MySQL/Felhasználó hozzárendelése a DB-hez menüpontban:

![felhasználó hozzárendelése az adatbázishoz](https://i.postimg.cc/prsWZfTz/assignuser.png)

Egyszerűen válassza ki a felhasználót és az adatbázist a legördülő menükből, majd kattintson a "Felhasználó hozzárendelése az adatbázishoz" gombra.

## Próbálja meg visszaállítani az adatbázis-felhasználó jelszavát

Az adatbázis felhasználói jelszavát az OpenPanel > MySQL/Felhasználók oldalon módosíthatja:

![password1.png](https://i.postimg.cc/1tQPpC9F/password1.png)

Kattintson az adatbázis felhasználója melletti "Jelszó módosítása" gombra, és a jelszó visszaállító oldalra kerül.

Ezen az oldalon másolja ki az új jelszót, mielőtt a „Jelszó módosítása” gombra kattintana a módosítás megerősítéséhez.

![password2.png](https://i.postimg.cc/PqGkj3vM/password2.png)

Szerkessze webhelye konfigurációs fájlját a Fájlkezelővel, és módosítsa a meglévő jelszót az imént beállított új jelszóra. Ha elkészült, kattintson a Mentés gombra a szerkesztő jobb felső sarkában.

![passconfig.png](https://i.postimg.cc/J7Q7QLDg/passconfig.png)

## Kézzel tesztelje a kapcsolatot

Ahhoz, hogy távoli helyről csatlakozhasson az adatbázisához, először meg kell nyitnia a következő oldalt: OpenPanel > MySQL/Remote MySQL .

![remotemysql.png](https://i.postimg.cc/43vjTJg9/remotemysql.png)

A távoli adatbázis-hozzáférés engedélyezéséhez kattintson a "Kattintson az engedélyezéshez" gombra, és várja meg a megerősítést, hogy a távoli hozzáférés funkció engedélyezve van.

Ezután nyissa meg az adatbázis-ügyfelet, példánkban a DBeaver for Windows alkalmazást fogjuk használni.

Hozzon létre egy új távoli kapcsolatot a kliensen belül, másolja ki a szerver- és portparamétereket az OpenPanel Remote MySQL oldalának „Távoli” részéből, és illessze be őket az ügyfélbe.

![dbclient.png](https://i.postimg.cc/Ssw3pjvZ/dbclient.png)

Töltse ki az adatbázist, a felhasználót és a jelszót a db paraméterekkel, és próbálja meg a kapcsolatot az ügyfélen keresztül ehhez a teszthez.





