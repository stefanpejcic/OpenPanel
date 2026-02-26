---
sidebar_position: 2
---

# Irányítópult

Az irányítópult oldal az OpenAdmin felület központja, és áttekintést nyújt a szerver aktuális teljesítményéről és használatáról.

Az irányítópult oldal widgeteket tartalmaz:

- **Felhasználói tevékenység** widget: Valós idejű kombinált tevékenységnaplót jelenít meg az összes OpenPanel felhasználóról.
- **Legfrissebb hírek** widget: Az OpenPanel blog blogcikkeit jeleníti meg.
- **Rendszerinformációk** widget: Információkat jelenít meg a szerver konfigurációjáról: Gazdanév, operációs rendszer, OpenPanel verzió, Kernel, CPU típus, üzemidő, futó folyamatok száma és elérhető csomagfrissítések.

## Használat

Az **OpenAdmin** minden oldalának jobb felső sarkában a rendszergazdák valós idejű erőforrás-használatot követhetnek nyomon, beleértve a **Betöltés**, **Memória**, **CPU** és **Lemez** elemeket.

Ha az egérmutatót az egyes mutatók fölé viszi, akkor részletes információk jelennek meg:

* **Load** – A rendszer átlagos terhelése 1, 5 és 15 perc alatt
* **Memória** – Fizikai memória és SWAP használata
* **CPU** – Használat CPU magonként
* **Lemez** – Használat lemezpartíciónként

![SSE Widget](https://i.postimg.cc/9Q9DMPH0/openadmin-sse.gif)

## Felhasználói tevékenység

A **Felhasználói tevékenység** widget az összes OpenPanel-felhasználó legfrissebb műveleteinek kombinált listáját tartalmazza (tevékenységnaplójuk).

## Legfrissebb hírek

A **Legfrissebb hírek** modul az [OpenPanel Blog](https://openpanel.com/blog) híreit jeleníti meg.

## Rendszerinformációk

A **Rendszerinformációk** widget információkat jelenít meg a szerverről:

- Gazdanév
- OS
- OpenPanel verzió
- Szerveridő
- Kernel verzió
- CPU modell
- Üzemidő
- A futó folyamatok száma
- Elérhető csomagfrissítések

## Próbálja ki az Enterprise-t

A **Try Enterprise** widget csak a Community Edition verzióban jelenik meg, és megjeleníti az Enterprise kiadás funkcióit és a frissítési lehetőségeket.

## Hibát találtam

Alapértelmezés szerint mind az OpenPanel, mind az OpenAdmin felhasználói felületén minden oldal alján tartalmaz egy **"Hibát talált? Tudassa velünk"** hivatkozást. Ez a link lehetővé teszi a felhasználók számára, hogy közvetlenül a [GitHub-problémák](https://github.com/stefanpejcic/OpenPanel/issues) oldalunkon jelentsék be a problémákat, és alapvető információkat tartalmaz a probléma reprodukálásához.

Az OpenPanel felhasználói felületén az adminisztrátorok letilthatják ezt a hivatkozást az **OpenAdmin > Beállítások > OpenPanel** menüpontban, és kikapcsolják a **"Link megjelenítése a hibák bejelentéséhez"** lehetőséget.

## Sötét mód

A Sötét mód engedélyezéséhez kattintson a felhasználónevére a bal alsó sarokban, és válassza ki a Hold ikont. Ha vissza szeretne váltani Fény módba, kattintson a Nap ikonra.

![openadmin sötét mód](/img/admin/dashboard/openadmin_dark_mode_toggle.gif)

## Menü

Az OpenAdmin menü felsorolja az OpenAdmin felületén elérhető összes opciót. Egyszerűen kattintson a menüelemre a megnyitáshoz.

## Keresés

A keresés eredménye:

- OpenPanel felhasználók bejelentkezési hivatkozással az OpenPanelhez
- A felhasználók webhelye/domainjei
- Funkciók/oldalak az Admin felületen

## Billentyűparancsok

Az OpenAdmin felhasználói felülete billentyűkódokkal navigálható: [dokumentáció megtekintése](/docs/articles/dev-experience/openadmin-keyboard-shortcuts/).

## Kijelentkezés

Az OpenAdmin fiókból való kijelentkezéshez kattintson a felhasználónevére a bal alsó sarokban, és válassza a "Kijelentkezés" lehetőséget.
