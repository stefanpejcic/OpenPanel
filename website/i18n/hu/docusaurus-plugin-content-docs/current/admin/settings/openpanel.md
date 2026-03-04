---
sidebar_position: 2
---

# OpenPanel

Konfigurálja a névszervereket, a márkajelzést és a felhasználói felület megjelenítési beállításait az OpenPanel felületen.

---

## Márkaépítés

Szabja testre az OpenPanel megjelenését a márkához:

- **Márkanév**
A **"Márkanév"** mezőbe beírva állítson be egy egyéni nevet az OpenPanel oldalsávon és a bejelentkezési oldalakon.

- **logó**
A márkanév helyett emblémát jelenítsen meg úgy, hogy a **"Logókép"** mezőben megadja a kép URL-jét.
Támogatott formátumok: ".png" vagy ".svg".
**Javasolt méret:** `200px × 36px`

- **Kijelentkezési URL**
Adja meg azt az URL-címet, ahová a felhasználók átirányítják a panelből való kijelentkezést követően (általában a fő webhelyén).

## Névszerverek

- ns1
- ns2
- ns3
- ns4

[Útmutató a névszerverek megfelelő konfigurálásához](/docs/articles/domains/how-to-configure-nameservers-in-openpanel)

## Kijelző

További megjelenítési beállítások a következők:

- **Avatar típusa:** Válasszon a Gravatar, Letter vagy Ikon közül a felhasználói avatarokhoz.
- **Charts Mode for Resource Usage:** Válassza ki, hogy az Erőforrás-használat oldalon 1 diagram, 2 diagram vagy diagramok nélkül jelenjen meg.
- **Ellenőrizze a jelszavakat a Weakpass.com webhelyen:** Ha engedélyezve van, a felhasználói jelszavakat a fiók létrehozása és visszaállítása során a rendszer ellenőrzi a Weakpass.com feltört jelszavak listáján.
- **Jelszó-visszaállítás engedélyezése:** Lehetővé teszi a felhasználók számára a jelszavak visszaállítását a bejelentkezési űrlapon keresztül (biztonsági okokból nem ajánlott).
- **Display 2FA Widget:** Jelenítsen meg egy üzenetet a felhasználók irányítópultjain, és arra ösztönzi őket, hogy engedélyezzék a kéttényezős hitelesítést a fokozott biztonság érdekében.
- **Az útmutatók widget megjelenítése:** Hasznos útmutató cikkek megjelenítése a felhasználók irányítópult-oldalain.
- **Link megjelenítése a hibák bejelentéséhez:** Mutassa meg a „Hibát talált? Tudassa velünk” linket az összes felhasználói oldal alján a hibajelentés egyszerűsítése érdekében.
- **Országzászló-ikonok megjelenítése:** Országjelzők megjelenítése az utolsó bejelentkezési IP-cím mellett az OpenPanel irányítópultján.

## Fájlkezelő

Adja meg a következő beállításokat a Fájlkezelőhöz:

- **Maximális fájlméret a Viewerben:** A Viewerben megnyitható maximális fájlméret (MB-ban). Az ajánlott maximum 20 MB.
- **Maximális fájlméret a szerkesztőben:** A Kódszerkesztőben megnyitható maximális fájlméret (MB-ban). Az ajánlott maximum 10 MB.
- **Maximális feltöltési fájlméret:** A Fájlkezelőn keresztül feltölthető maximális fájlméret (MB-ban). Az ajánlott maximum 2000 MB.
- **Maximális letöltési fájlméret:** A Fájlkezelőn keresztül letölthető maximális fájlméret (MB-ban). Az ajánlott maximum 2000 MB.
- **A kuka automatikus törlése után:** A fájlok száma, ahány napig a felhasználó kukájában maradnak az automatikus törlés előtt. A 0-ra állítás letiltja az automatikus tisztítást. Az ajánlott beállítás 30 nap.
- **Maximális tömörítési idő:** Az archívumtömörítési folyamatok megszakítása előtti percek maximális száma. Javasoljuk, hogy ezt néhány percig tartsa.
- **Maximális kibontási idő:** Maximum percek száma az archívum kibontása előtt. Javasoljuk, hogy ezt néhány percig tartsa.
- **A (szöveges) kiterjesztések megtekintési és szerkesztési opcióinak engedélyezése:** Adja meg a Fájlkezelőben megnyitható és szerkeszthető fájlkiterjesztéseket (szöveges fájltípusoknak kell lenniük).
- **Nézet opció engedélyezése (base64 kép) kiterjesztésekhez:** Adja meg a képfájl kiterjesztéseket, amelyek megjeleníthetők base64 kódolással a Viewerben.
- **Kibontási és archiválási opciók engedélyezése (archívum) bővítményekhez:** Adja meg a fájlkezelővel kibontható archív fájlkiterjesztéseket.

## Statisztika

Konfigurálja a következő beállításokat a felhasználói bejelentkezési kísérletekkel, a munkamenet-kezeléssel és az adatmegőrzéssel kapcsolatban:

- **Sikertelen bejelentkezés percenként, mielőtt a felhasználó sebességkorlátozásra kerülne:** Egyetlen IP-ről percenként engedélyezett sikertelen bejelentkezési kísérletek száma, mielőtt az adott IP-cím átmenetileg korlátozva lenne a brute force támadások megelőzése érdekében.
- **Sikertelen bejelentkezés percenként a felhasználó 1 órás blokkolása előtt:** A sikertelen bejelentkezési kísérletek percenkénti küszöbértéke, amely 1 órás blokkolást vált ki a sértő IP-címre vonatkozóan.
- **Munkamenet időtartama (percben):** Az az időtartam, ameddig egy felhasználói munkamenet aktív marad, mielőtt újbóli hitelesítést igényelne.
- **Munkamenet élettartama (percben):** Egy munkamenet teljes maximális élettartama, amely után a felhasználó tevékenységétől függetlenül kijelentkezett.
- **Felhasználónként megőrizendő bejelentkezési rekordok:** A legutóbbi bejelentkezési kísérletek száma felhasználónként, auditálási és biztonsági nyomon követés céljából.
- **Felhasználónként tárolandó tevékenységrekordok:** A panelen belüli múltbeli műveletek áttekintésére megőrzött felhasználói tevékenységnaplók maximális száma.
- **Tevékenységelemek oldalanként:** A felhasználói felületen oldalanként megjelenített tevékenységnapló-bejegyzések száma.
- **Megjelenítendő erőforrás-használati tételek oldalanként:** Az oldalanként megjelenített erőforrás-használati rekordok száma a felhasználói vagy rendszererőforrás-statisztikák megtekintésekor.
- **Felhasználónként naplózandó erőforrás-használati tételek:** Az előzményelemzés céljából felhasználónként rögzített és tárolt erőforrás-használati bejegyzések száma.
- **Domainek oldalanként:** A domainkezelési listákon oldalanként megjelenő domain bejegyzések száma.
