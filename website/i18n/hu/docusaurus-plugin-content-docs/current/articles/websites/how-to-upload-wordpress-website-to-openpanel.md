# WordPress®-telepítés áttelepítése OpenPanelre

Ez az útmutató végigvezeti Önt a WordPress webhely **OpenPanel**-ba való feltöltésének folyamatán, beleértve a domain beállítását, az adatbázis-konfigurációt, a fájlfeltöltéseket és a végső tesztelést.

Ha frissen kezd egy új webhelyen, kövesse a következőket: [A WordPress® telepítése OpenPanel segítségével](/docs/articles/websites/how-to-install-wordpress-with-openpanel/)

---

## 1. Domain

### 1.1 Domainnév hozzáadása

1. Jelentkezzen be az **OpenPanel** irányítópultjára.
2. Lépjen a **Domainek** → **Domain hozzáadása** elemre.
3. Adja meg a domain nevét (pl. "example.com").
4. Mentés.

![domain hozzáadása](https://i.postimg.cc/mkDT1Mhh/add-domain.png)

Ügyeljen arra, hogy domainje DNS-rekordjait (A rekord) az OpenPanel-kiszolgáló IP-címére irányítsa.

---

## 2. Adatbázis

### 2.1 Adatbázis és felhasználó létrehozása

1. Lépjen a **MySQL** → **Adatbázis varázsló** elemre.
2. Állítsa be az **Adatbázis neve**, **Adatbázis felhasználó** lehetőséget, és állítson be erős jelszót.
3. Kattintson a **DB létrehozása, felhasználói és jogosultságok megadása** gombra.

![mysql db varázsló](https://i.postimg.cc/rm4Pnwbg/db-wizard.png)

### 2.2 SQL fájl importálása az adatbázisba

Az adatbázis feltöltésének két módja van:

- Ha elérhető, használja az „Adatbázis importálása” opciót.
- phpMyAdmin felület használata.

**SQL-fájl importálása az importálási adatbázis használatával**:

1. Válassza a **MySQL** → **Adatbázis importálása** lehetőséget.
2. Válassza az **Adatbázis** lehetőséget, és válassza ki a fájlt az eszközről.
3. Kattintson a **Feltöltés és importálás** lehetőségre.

![sql fájl importálása](https://i.postimg.cc/Vfwnd6sR/db-import.png)


**2.3 SQL fájl importálása a phpMyAdmin segítségével**:

Az SQL-fájlok phpMyAdmin felülettel való feltöltéséhez kövesse ezt az útmutatót: [Az SQL-fájlok importálása](/docs/panel/mysql/phpmyadmin/#import-sql-files)

---

## 3. Fájlok

### 3.1 Fájlok feltöltése a Fájlkezelőn keresztül

Feltöltés előtt győződjön meg róla, hogy a biztonsági másolat archiválva van (`.zip`, `.tar` vagy `.tar.gz`). Ha nem, először tömörítse a fájlokat egyetlen archívumba.

1. Nyissa meg a **Fájlkezelőt** az OpenPanelben.
2. Keresse meg domainje gyökérkönyvtárát (pl. `/var/www/html/pejcic.rs/`).
3. Kattintson a **Feltöltés** gombra.
4. Válassza ki és töltse fel a biztonsági másolat archívumát.
5. A feltöltés után lépjen vissza a **Fájlkezelőbe**, válassza ki az archívumot, majd kattintson a **Kicsomagolás** → kibontás megerősítése lehetőségre.
6. A kicsomagolás befejezése után törölje a feltöltött archívumot. Ehhez válassza ki azt, kattintson a **Törlés** gombra, és erősítse meg.

---

### 3.2 A wp-config.php fájl módosítása

1. A Fájlkezelőben keresse meg a **`wp-config.php`** fájlt.
![wp konfigurációs fájl szerkesztése](https://i.postimg.cc/mkLgNk21/edit-wp-config.png)
2. Módosítsa az adatbázis beállításait:

``` php
define('DB_NAME', 'adatbázisod_neve');
define('DB_FELHASZNÁLÓ', 'az_adatbázis_felhasználója');
define('DB_JELSZÓ', 'adatbázisod_jelszava');
define('DB_HOST', 'mysql'); // mariadb vagy mysql - az aktuális beállítástól függően
   ```
![példa wp config ph](https://i.postimg.cc/3NtJLhdS/edit-wp-config-file.png)
3. Mentse el a változtatásokat.

---

## 4. Teszt

### 4.1 Teszt tartománynévvel

1. Keresse fel domainjét (pl. „https://example.com”).
2. Ha a DNS elterjedt, a WordPress webhelyének be kell töltenie.

### 4.2 Teszt az élő előnézeten keresztül

1. Ha a DNS még nincs megjelölve, használja az OpenPanel **Élő előnézet** funkcióját.
2. Nyissa meg a WordPress Manager alkalmazást az OpenPanelben → kattintson a **Meglévő telepítés keresése** lehetőségre, és várja meg, amíg a folyamat befejeződik.
3. Frissítse az oldalt, és kattintson a webhelyére.
4. Kattintson az **Élő előnézet** gombra a webhely ellenőrzéséhez.

![wp scan](https://i.postimg.cc/npmRW3S9/scan-wp.png)
