---
sidebar_position: 7
---

# MySQL típus váltása

A **Docker > MySQL-típus váltása** oldalon átkapcsolhatja az aktuális mysql-szolgáltatást a rendelkezésre álló lehetőségek között: **MySQL** és **MariaDB**.

Az aktuálisan aktív típus az oldal jobb felső sarkában jelenik meg.

## Követelmények

A funkció eléréséhez:

- A **Docker** modult a rendszergazdának engedélyeznie kell **szerveren**.
- Fiókjában engedélyezni kell a **Docker** funkciót.

## Használat

Mielőtt megváltoztatná az adatbázis típusát, ellenőrizze a következőket:

- **Az összes meglévő adatbázist és felhasználót el kell távolítani.**
- Az új mysql tárolót **le kell állítani**, mielőtt elindítaná az újat.

> ⚠️ Ha már vannak konfigurálva adatbázisok, **készítsen biztonsági másolatot az összes adatról**, távolítsa el az összes adatbázist és felhasználót, majd folytassa a mysql szerver váltásával.
> Az állásidő elkerülése érdekében a legjobb, ha ezt a módosítást **adatbázisok hozzáadása előtt** hajtja végre.

### A váltás lépései

1. Az OpenPanel menüben lépjen a **Docker > Switch MySQL Type** elemre.
2. A legördülő menüből válassza ki a használni kívánt új típust.
3. Kattintson a **Switch** gombra a folyamat elindításához.

Megerősítés után:

- A meglévő adatbázis-tároló leáll, és az adatai törlődnek.
- Elindul az új mysql típusú szerver.
- Ezután újra felveheti az adatbázist és a felhasználókat az új szerver alá.
