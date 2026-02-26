# Frissítés Enterprise-ra

Az OpenPanel két kiadásban érhető el:

- **Közösség** - ingyenes hosting vezérlőpanel Debian és Ubuntu OS számára, alkalmas VPS-re és magáncélra.
- **Enterprise** - fejlett funkciókat kínál a felhasználók elkülönítéséhez és kezeléséhez, alkalmas webtárhely-szolgáltatók számára.

További információ: [OpenPanel Community VS Enterprise Edition](/enterprise)

## Vásárlási engedély

Az OpenPanel Enterprise kiadás licencének vásárlásához [kattintson ide](https://my.openpanel.com/cart.php?a=add&pid=1).

Fizetés után a licenckulcs automatikusan kiadásra kerül, és a Szolgáltatások oldalon látható:

![licenckulcs](/img/guides/add_license.png)

## Licenckulcs hozzáadása

### Meglévő telepítéshez:

- Az OpenAdminból:
Ha már telepítve van az OpenPanel Community Edition, nyissa meg az **OpenAdmin > Try Enterprise** elemet, és adja hozzá a kulcsot az űrlaphoz:
  
![licenckulcs hozzáadása](/img/guides/add_key.png)

- Terminálról:
``` bash
opencli licenc vállalat-XXX
  ```
Cserélje le a "enterprise-XXX" kifejezést a licenckulcsával.
  
![openncli licenc](/img/guides/key_command.png)

### Új telepítéshez:

Ha az OpenPanel Enterprise-t egy másik kiszolgálóra szeretné telepíteni, másolja ki az install parancsot a szolgáltatási oldalról, és illessze be az új kiszolgálóra:

![telepítési parancs](/img/guides/key_code.png)
