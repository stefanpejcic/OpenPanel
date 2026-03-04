---
sidebar_position: 1
---

# Fájlkezelő

A **Fájlkezelő** felület lehetővé teszi a `/var/www/html/` könyvtárban található összes domain fájljainak kezelését.

A táblázatban a következő információkat láthatja:

**Név**
* **Méret** – Megjeleníti a fájlok méretét. Mappák esetén kattintson a **Számítás** lehetőségre a méretük lekéréséhez.
* **Last Modified** – Az utolsó módosítás időbélyegét mutatja.
* **Engedélyek** – Szimbolikus fájlengedélyeket jelenít meg (pl. `drwxr-xr-x`). Mutasson rájuk az oktális (numerikus) megfelelő (pl. "0755") megtekintéséhez.

A **Váltás** ikonra kattintva további részletek jelennek meg:

* **Tulajdonos** – A fájl vagy mappa tulajdonosának felhasználói azonosítója.
* **Csoport** – A fájlt vagy mappát birtokló csoport GID-je.
* **Hivatkozások** – A fájlra vagy könyvtárra mutató merev hivatkozások száma.
* **Link Target** – Ha szimbolikus hivatkozásról van szó, akkor ez azt a célt mutatja, amelyre mutat.
* **Típus** – Azt jelzi, hogy könyvtárról, fájlról vagy szimbolikus hivatkozásról van-e szó.


## Fájl létrehozása

Új fájl létrehozásához navigáljon a kívánt könyvtárba, és kattintson az **Új fájl** gombra. A modális mezőben adja meg a fájlnevet, és kattintson a **Létrehozás** gombra.

Opcionálisan, ha közvetlenül a létrehozás után szeretné megnyitni a fájlt a Szerkesztőben, jelölje be a **Megnyitás a Fájlszerkesztőben létrehozás után** lehetőséget.

## Mappa létrehozása

Új mappa létrehozásához keresse meg a kívánt könyvtárat, és kattintson az **Új mappa** gombra. A modálisban adja meg a mappa nevét, és kattintson a **Létrehozás** gombra.

## Fájlok feltöltése

A Fájlkezelő lehetővé teszi több fájl egyidejű feltöltését. Fájlokat az alábbi módszerek egyikével tölthet fel:

- **Drag&Drop in File Manager**: Navigáljon a kívánt mappához, és húzza át a fájlokat a fájllista táblázatába.
- **Feltöltés eszközről**: Kattintson a „Feltöltés” ​​gombra, majd az új oldalon húzza át vagy válasszon fájlokat eszközéről.
- **Letöltés URL-ről**: Kattintson a "Letöltés inkább az URL-ről" lehetőségre, majd adja meg a letölteni kívánt fájl linkjét.

A feltöltési méretkorlátokat a rendszergazda állíthatja be.


## Az összes kijelölése

Használja az egérkurzort több fájl és mappa kiválasztásához. Kattintson és húzza, majd kattintson újra a kijelölés megszüntetéséhez.

Több fájl egyenkénti kiválasztásához kattintson és tartsa lenyomva a Ctrl billentyűt, miközben a sorokra kattint.

A könyvtárban található összes fájl egyszerre történő kijelöléséhez kattintson az „Összes kiválasztása” gombra. Az összes fájl kijelölésének törléséhez kattintson a „Kijelölés törlése” gombra.


## Törlés

Fájlok vagy mappák törléséhez kattintson a „Törlés” gombra. Ha több fájl vagy mappa van kijelölve, a lista megjelenik a modálban, és kattintson a „Törlés” gombra a kiválasztott fájlok végleges törléséhez.

## Fájl letöltése

Fájl letöltéséhez kattintson a **Letöltés** gombra a fájl kiválasztása után.

:::info
Több fájl egyidejű letöltése nem támogatott. Ehelyett a „Tömörítés” opciót kell használni a fájlok archívumának létrehozásához, majd az archívum letöltéséhez.
:::

## Fájl megtekintése

Egy szövegszerkesztőben szerkeszthető vagy képként megnyitható fájl tartalmának megtekintéséhez kattintson a fájlra, majd a **Nézet** gombra.

## Fájl szerkesztése

Egy szövegszerkesztőben szerkeszthető fájl tartalmának szerkesztéséhez kattintson a fájlra, majd a **Szerkesztés** gombra. Megnyílik egy új oldal egy szövegszerkesztővel, ahol szerkesztheti a fájl tartalmát.

## Fájl átnevezése

Fájl vagy mappa átnevezéséhez kattintson rá, majd kattintson az **Átnevezés** gombra, és állítsa be az új nevet.

## Fájlok másolása

Fájlok egyik mappából a másikba másolásához először válassza ki a kívánt fájlokat, majd kattintson a **Másolás** gombra. Az új mód megjeleníti az összes kiválasztott fájlnév listáját, és lehetővé teszi annak a mappának a célnevének beállítását, ahová a fájlokat másolni kívánja.

A másolási folyamat elindításához kattintson a 'Másolás' gombra a modálban. Megjelenik egy folyamatjelző sáv, amely jelzi az előrehaladást, és a „Másolás kész” üzenet jelenik meg, amikor a folyamat befejeződött.


## Fájlok áthelyezése

A fájlok egyik mappából a másikba való áthelyezéséhez először válassza ki a kívánt fájlokat, majd kattintson az **Áthelyezés** gombra. Az új modális megjeleníti az összes kiválasztott fájlnév listáját, és lehetővé teszi annak a mappának a célnevének beállítását, ahová a fájlokat áthelyezi.

Az áthelyezési folyamat elindításához kattintson a modális 'Áthelyezés' gombra. Megjelenik egy folyamatjelző sáv, amely jelzi az elért előrehaladást, és a „Complete” üzenet jelenik meg, amikor a folyamat befejeződött.

## Archívum kibontása

Egy feltöltött archívum (`.zip`, `.tar`, `.tar.gz`) kicsomagolásához válassza ki a fájlt, majd kattintson a 'Kicsomagolás' gombra a menüben.

A modálisban állítsa be a mappa nevét, ahová a fájlokat kicsomagolja, majd kattintson a **Megerősítés** gombra a folyamat elindításához.

## Tömörítés az archívumba

A fájlok archívumának létrehozásához először válassza ki a kívánt fájlokat vagy mappákat, majd kattintson a **Tömörítés** gombra. Az új modális megjeleníti az összes kiválasztott fájlnév listáját, és lehetővé teszi az archívum nevének és kiterjesztésének beállítását (`.zip`, `.tar` vagy `.tar.gz`).

## Üres mappa

Ha egy mappa üres, a „Nem található elemek” üzenet jelenik meg. üzenetet, és a fájllehetőségeket tartalmazó menü el lesz rejtve. Csak az új fájl, mappa létrehozásának vagy fájlok feltöltésének lehetőségei lesznek elérhetők.

## Fájlok és mappák keresése

A keresőmező aktiválásához kattintson a jobb felső sarokban található Keresés ikonra. A Váltó ikonra kattintva megjelennek a csak fájlok vagy csak mappák keresésének lehetőségei, valamint a keresés elérési útja.

Teljesítmény okokból a keresési eredmények legfeljebb 10 találatra korlátozódnak a fájlok és 10 a mappák esetében.
