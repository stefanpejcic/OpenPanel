# Hogyan lehet beállítani vagy növelni a PHP INI memory_limit vagy más értékeket?

Az OpenPanel PHP-beállításait többféleképpen is kezelheti. Ez az útmutató végigvezeti a rendelkezésre álló módszereken, és bemutatja, hogy mikor érdemes mindegyiket használni.

## OpenPanel felhasználóknak

Attól függően, hogy mely modulok engedélyezettek, a felhasználók a PHP-beállításokat a következő interfészek közül egy vagy több használatával kezelhetik:

- **PHP Limits** – Akkor érhető el, ha a php modul engedélyezve van. Lehetővé teszi az alapvető PHP korlátok szerkesztését.
- **PHP.INI Editor** – Akkor érhető el, ha a `php_ini` modul engedélyezve van. Teljes hozzáférést biztosít a php.ini fájl szerkesztéséhez.
- **PHP Options** – A `php_options` modullal érhető el. Csak az előre konfigurált beállításokat teszi lehetővé a felhasználók számára az ini fájlból.

### Alapértelmezett PHP verzió

Az új tartományokhoz használt alapértelmezett PHP-verzió a következő módon konfigurálható:

**OpenPanel > PHP > Alapértelmezett verzió**

1. Válassza ki a kívánt PHP verziót.
2. Kattintson a "Módosítás" gombra az alkalmazáshoz.

![alapértelmezés módosítása](/img/panel/v2/openpanel_cahnge_default_php_version.gif)

### PHP korlátok

A **PHP Limits** felület lehetővé teszi a legfontosabb PHP direktívák konfigurálását. Itt ** kell** beállítani a következőket:

- "max. végrehajtási_idő".
- "max._bemeneti_idő".
- "max_input_vars".
- "memóriakorlát".
- "post_max_size".
- "feltöltési_maximális_fájlméret".

[PHP Limits Documentation] (/docs/panel/php/limits)

![openpanel_edit_php_limits](/img/panel/v2/openpanel_edit_php_limits.gif)

> MEGJEGYZÉS: A *PHP Limits* oldal beállításai felülírják a *PHP.INI Editor* vagy a *PHP Options* programban megadottakat.

---

### PHP opciók

Ez az interfész lehetővé teszi az általánosan használt PHP-beállítások módosítását – a rendszergazda által előre beállítottakra korlátozva.

[PHP-beállítások dokumentációja](/docs/panel/php/options)

![php beállítások oldal](/img/panel/v2/php_options.png)

---

### PHP.INI szerkesztő

A teljes irányítás érdekében a **PHP.INI Editor** lehetővé teszi a felhasználók számára, hogy közvetlenül módosítsák a php.ini fájlt a kiválasztott PHP verziójukhoz.

[PHP.INI szerkesztő dokumentáció](/docs/panel/php/php_ini_editor)

![openpanel_php_ini_editor](/img/panel/v2/openpanel_php_ini_editor.gif)

---

## OpenAdmin felhasználóknak

A rendszergazdák konfigurálhatják a rendszerszintű PHP alapértelmezéseket, és szabályozhatják, hogy a felhasználók milyen beállításokat módosíthatnak.

### Alapértelmezett PHP verzió

Állítsa be az alapértelmezett PHP-verziót az új felhasználók számára:
[**OpenAdmin > Beállítások > Felhasználói alapértelmezések**](/docs/admin/settings/defaults/)

1. Válassza ki az alapértelmezett PHP verziót.
2. A megerősítéshez kattintson a "Mentés" gombra.

![default.png](https://i.postimg.cc/cJNjVQPM/default.png)

---

### PHP korlátok

Az adminisztrátorok az alábbi módon határozhatják meg a felhasználók számára elérhető alapértelmezett PHP-korlátokat:
[**OpenAdmin > Beállítások > Felhasználói alapértelmezések**](/docs/admin/settings/defaults/)

![limits.png](https://i.postimg.cc/4df03XfY/limits.png)

---

### PHP opciók

A felhasználók csak azokat a beállításokat módosíthatják, amelyeket a rendszergazda kifejezetten engedélyezett. Ezek a kulcsok előre meghatározottak és a rendszer `.ini` fájljaiból kerülnek alkalmazásra.

A rendszergazdák kétféleképpen kezelhetik ezeket a beállításokat:

A panelről:
[**OpenAdmin > Beállítások > PHP beállítások**](/docs/admin/settings/php/) oldalon.

![options.png](https://i.postimg.cc/1zcTP8Qx/options.png)

Vagy terminálon keresztül:
- **Globális alapértelmezett (minden felhasználó)**: `/etc/openpanel/php/options.txt`
- **Felhasználó-specifikus beállítások**: `/home/FELHASZNÁLÓNÉV/php.ini/options.txt`

---

### PHP.INI szerkesztő

A rendszergazdák minden PHP-verzióhoz beállíthatják az alapértelmezett php.ini fájlokat.

A panelről:
[**OpenAdmin > Beállítások > PHP beállítások**](/docs/admin/settings/php/) oldalon.

![phpini.png](https://i.postimg.cc/dVCpHWNV/phpini.png)

Vagy terminálon keresztül:
- **Globális alapértelmezett (minden felhasználó)**: `/etc/openpanel/php/ini/`
- **Felhasználó-specifikus beállítások**: `/home/FELHASZNÁLÓNÉV/php.ini/`

