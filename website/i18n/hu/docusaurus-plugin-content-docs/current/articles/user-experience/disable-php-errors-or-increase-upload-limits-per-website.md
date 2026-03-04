# user.ini

Ha több webhely van ugyanabban a tartomány mappájában, és csak egy bizonyos webhely php-beállításait szeretné engedélyezni vagy letiltani, hozzon létre egy `.user.ini` fájlt a webhely mappájában.

A fájlban állítsa be a kívánt php értékeket, például:

Állítsa be a bejegyzést és töltse fel a maximális fájlméretet:

```bash
upload_max_filesize = 200M
post_max_size = 200M
```

az összes php hiba megjelenítése:

```bash
error_reporting = E_ALL
display_errors = on
```

Szerkesztés után indítsa újra a php szolgáltatást. Keresse meg a PHP.INI szerkesztőt, és válassza ki a domain által használt php verziót. Egyszerűen kattintson a mentés gombra, és a szolgáltatás újraindul, és azonnal alkalmazza a beállításokat a .user.ini fájlból.
