---
sidebar_position: 8
---

# Engedélyek javítása
Az **Engedélyek javítása** eszközt úgy tervezték, hogy automatikusan javítsa a webhelyek fájl- és mappaengedélyeit, biztosítva azok biztonságosságát és megfelelő működését.

Nyissa meg a **Javítási engedélyek** oldalt, és állítsa be az elérési utat, vagy hagyja üresen (`/var/www/html/`), hogy áttekintse az összes fájlt.

## Hogyan működik
Az **Engedélyek javítása** gombra kattintva:

- Rekurzívan állítsa be a fájlok és mappák megfelelő **tulajdonjogát**.
- Alkalmazzon szabványos **engedélyeket**:
- Fájlok: "644".
- Mappák: "755".
- Javítsa ki az ismert CMS-könyvtárak, például a "wp-content", "storage", "cache" stb. gyakori engedélyezési problémáit.

## Vigyázat

- Az egyéni engedélyek vagy a speciális konfigurációk felülbírálhatók.
- Ha olyan egyéni szkripteket futtat, amelyek eltérő engedélyeket igényelnek (pl. futtatható `.sh` fájlok), akkor az eszköz használata után manuálisan újra kell alkalmaznia ezeket a beállításokat.

## Javasolt használat

Csak akkor használja a **Engedélyek javítása** opciót, ha:

- Webhelye engedélyekkel kapcsolatos hibákat mutat.
- Kézzel töltött fel vagy módosított fájlokat (például FTP-n vagy Fájlkezelőn keresztül).
- A CMS-ed (WordPress, Joomla stb.) nem tud írni bizonyos mappákba.
- Tulajdonjoggal vagy engedélyekkel kapcsolatos problémákra gyanakszik egy áttelepítés vagy kézi szerkesztés után.

Csak akkor használja ezt az eszközt, ha problémákat tapasztal. A gyakori, szükségtelen használat nem szükséges, és visszavonhatja a szándékos engedélymódosításokat.
