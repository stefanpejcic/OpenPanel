---
sidebar_position: 3  
---

# PHP.INI szerkesztő

A PHP.INI szerkesztő lehetővé teszi az egyes telepített PHP-verziók konfigurációs fájljának megtekintését és módosítását.

> Megjegyzés: A PHP.INI Editor [csak akkor érhető el, ha a `php_ini` funkciót a rendszergazda engedélyezte] (/docs/admin/settings/modules#phpini-editor).

Szükség szerint növelheti a korlátokat, engedélyezheti az új alapértelmezett értékeket, vagy módosíthatja a beállításokat. A változtatások újraindítják a megfelelő PHP szolgáltatást.

![openpanel_php_ini_editor](/img/panel/v2/openpanel_php_ini_editor.gif)

> MEGJEGYZÉS: A következő beállításokat: "max_execution_time" "max_input_time" "max_input_vars" "memory_limit" "post_max_size" "upload_max_filesize" a [**PHP Limits** oldalon] (/docs/panel/php/limits) kell megadni.

---

A PHP-szolgáltatás (tároló) teljes CPU- vagy memória-erőforrás-korlátjának módosításához használja a [**Containers** oldalt](/docs/panel/containers/).
