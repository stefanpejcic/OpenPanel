# Leminősítési licenc

Az OpenPanel két kiadásban érhető el:

- **Közösség** - ingyenes hosting vezérlőpanel Debian és Ubuntu OS számára, alkalmas VPS-re és magáncélra.
- **Enterprise** - fejlett funkciókat kínál a felhasználók elkülönítéséhez és kezeléséhez, alkalmas webtárhely-szolgáltatók számára.

További információ: [OpenPanel Community VS Enterprise Edition](/enterprise)

## Leminősítés

:::info
Az Enterprise verzióról a közösségi kiadásra való visszaminősítés azonnal eltávolítja az összes vállalati funkciót és modult az OpenPanel felhasználói felületről. A meglévő vállalati szolgáltatások – például az FTP és az e-mail – továbbra is futnak, de Ön már nem tudja kezelni őket az OpenPanel felületéről.
:::

Az Enterprise kiadásról a Community-re való visszalépéshez:

- Az OpenAdminból:
Lépjen az **OpenAdmin > Licenc** elemre, és kattintson a „Leépítés” gombra:
  
![távolítsa el a licenckulcsot](/img/guides/downgrade_license.png)

- Terminálról:
``` bash
opencli licenc törlése
  ```
