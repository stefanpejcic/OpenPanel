---
sidebar_position: 4
---

# phpMyAdmin

A phpMyAdmin egy fejlett MySQL adatbázis-kezelő eszköz, és csak tapasztalt felhasználók számára ajánlott.

## Kezelés

A phpMyAdmin eléréséhez nyissa meg az **Adatbázisok > phpMyAdmin** oldalt az oldalsávon.

Ezen az oldalon engedélyezheti/letilthatja a phpMyAdmin szolgáltatást, ellenőrizheti az aktuális állapotot, módosíthatja a `php_max_execution_time`, `php_memory_limit`, `php_upload_limit` értékeket, és beállíthatja a `PMA_ABSOLUTE_URI`-t (a phpmyadmin számára használandó egyéni tartomány).

![databases_phpmyadmin.png](/img/panel/v2/phpmyadmin_page.png)


## Hozzáférés

Kattintson a „phpMyAdmin bejelentkezés megnyitása” gombra, ekkor automatikusan bejelentkezik, és megtekintheti az összes létező adatbázist és azok tábláit.

![databases_phpmyadmin.png](/img/panel/v1/databases/databases_phpmyadmin.png)


:::veszély
A táblázatok közvetlen szerkesztése vagy manipulálása a phpMyAdmin programban tönkreteheti webhelyét, ha helytelenül végzi el. Ha nem biztos benne, forduljon a fejlesztőhöz, mielőtt változtatásokat hajtana végre.
:::

Miután bejelentkezett, a következőket teheti:
- Adatbázis táblák megtekintése
- SQL lekérdezések futtatása
- Dobj asztalokat
- Adatok importálása
- Adatbázisok exportálása (pl. WordPress adatbázis)
- És még több

A phpMyAdmin használatával kapcsolatos további információkért tekintse meg [a phpMyAdmin hivatalos dokumentációját] (https://www.phpmyadmin.net/docs/).

---

## SQL-fájlok importálása

A phpMyAdmin lehetővé teszi az .sql fájlok importálását különböző forrásokból.

### 1. Importálás a készülékről
1. Nyissa meg a phpMyAdmin programot, és (opcionálisan) válassza ki azt az adatbázist, amelybe importálni szeretne.
2. Kattintson az **Importálás** lehetőségre a felső menüben.
3. Az **Importálandó fájl** alatt kattintson a **Fájl kiválasztása** lehetőségre, és válassza ki az `.sql` fájlt a számítógépéről.
4. Kattintson az **Importálás** elemre alul, és várja meg, amíg a folyamat befejeződik.

### 2. Importálás a kiszolgáló feltöltési könyvtárából
A `/var/www/html/` mappában a szerverére már feltöltött fájlokat is importálhatja.

1. Töltse fel az `.sql` fájlt a `/var/www/html/` mappába a Fájlkezelő segítségével.
2. Nyissa meg a phpMyAdmin programot, és (opcionálisan) válassza ki a céladatbázist.
3. Kattintson az **Importálás** lehetőségre a felső menüben.
4. A *Kiválasztás a webszerver feltöltési könyvtárából /html/:* alatt válassza ki a fájl nevét.
5. Kattintson az **Importálás** elemre alul, és várja meg, amíg a folyamat befejeződik.


### 3. Importálás terminálon keresztül (Docker)
Ha a **docker** funkció engedélyezve van, közvetlenül importálhat a `mysql` vagy `mariadb` tárolójában lévő terminálból.

1. Nyissa meg a **Containers > Terminal** elemet.
2. Válassza ki a "mysql" vagy a "mariadb" tárolót.
3. Töltse le vagy helyezze el az .sql fájlt a tárolóba.
4. Futtassa a megfelelő parancsot:
``` bash
mysql -u felhasználónév -p adatbázis_neve < fájl.sql
   ```
Cserélje ki a "felhasználónév", "adatbázis_neve" és "fájl.sql" paramétereket a tényleges adataival.

