---
sidebar_position: 3
---

# Opciók

A PHP-beállítások oldal lehetővé teszi a php.ini fájlban tárolt PHP-beállítások módosítását.

![php beállítások oldal](/img/panel/v2/php_options.png)

A haladóbb felhasználók számára a [**PHP.INI Editor**](/docs/panel/php/php_ini_editor/) használható - ha elérhető.

---

## A PHP-beállítások megváltoztatása

1. Nyissa meg az **OpenPanel** elemet, és lépjen a **PHP > PHP Options** elemre.
2. Válassza ki a kívánt **PHP verziót** a legördülő listából.
3. Módosítsa a módosítani kívánt beállításokat.
4. Kattintson a **Mentés** gombra a módosítások alkalmazásához.

A beállítások a kiválasztott PHP verzióhoz tartozó `php.ini` fájlba kerülnek, és a PHP-tároló újraindul, hogy azonnal alkalmazza a változtatásokat. Ezek a beállítások az adott PHP-verziót használó **minden domainre** vonatkoznak.

---

## Elérhető opciók

A konfigurálható PHP opciók listáját a rendszergazda határozza meg. Alapértelmezés szerint a következő lehetőségek állnak rendelkezésre:

- "allow_url_fopen".
- "dátum.időzóna".
- `függvények letiltása`
- "megjelenítési_hibák".
- "hibabejelentés".
- `file_uploads`
- "napló_hibák".
- "max. végrehajtási_idő".
- "max._bemeneti_idő".
- "max_input_vars".
- "memóriakorlát".
- "open_basedir".
- "kimeneti_pufferelés".
- "post_max_size".
- `short_open_tag`
- "feltöltési_maximális_fájlméret".
- "zlib.output_compression".

**Az elérhető opciók testreszabása**

Az adminisztrátorok személyre szabhatják a rendelkezésre álló lehetőségeket:

- **Minden új felhasználó számára**: Szerkessze az `/etc/openpanel/php/options.txt` fájlt
- **Egy adott felhasználóhoz**: Szerkessze a `/home/FELHASZNÁLÓNÉV/php.ini/options.txt` fájlt

---

## Fontos megjegyzések

Néhány PHP beállítás nem konfigurálható erről az oldalról. Ehelyett a [**PHP Limits** oldalon] (/docs/panel/php/options) kell beállítani őket:

- "max. végrehajtási_idő".
- "max._bemeneti_idő".
- "max_input_vars".
- "memóriakorlát".
- "post_max_size".
- "feltöltési_maximális_fájlméret".

Ezek a korlátozások a tároló szintjén érvényesülnek, és eltérő konfigurációs módszert igényelnek.

