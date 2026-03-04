# OpenCLI Hooks

Az OpenPanel támogatja a **pre** és **post hook-okat** minden olyan parancshoz, amely az opencli parancsok futtatása előtt vagy után bash szkripteket futtat.

Hook létrehozásához először hozzon létre egy új könyvtárat: `/etc/openpanel/openpanel/hooks/`, és azon belül hozzon létre egy fájlt a kívánt parancs alapján:

- `pre_` előtag a parancsfájl **előtt** futtatásához.
- `post_` előtag a szkript futtatásához **az opencli parancs végrehajtása után**.

## Példák

Egyéni szkript futtatásához a felhasználó létrehozási folyamata előtt (opencli *user-add*) hozzon létre egy új fájlt:
```bash
/etc/openpanel/openpanel/hooks/pre_user-add.sh
````


Egy másik példa a **domain hozzáadása után** futtatásra:

Amikor egy felhasználó új tartományt ad hozzá a felhasználói felületen keresztül, a végrehajtandó mögöttes parancs a következő:

```bash
/etc/openpanel/openpanel/hooks/post_domains-add.sh
```

Ezek a szkriptek automatikusan végrehajtásra kerülnek, amikor a parancs fut, attól függően, hogy a végrehajtás előtti vagy utóbbi fázisba kapcsolódott be.

## Érvek átadása Hooksnak

Az "opencli" parancsnak átadott összes argumentum a hook-szkriptnek is továbbításra kerül.

Például, ha a következő parancs végrehajtásra kerül:

```bash
opencli domains-add example.com stefan --docroot /var/www/example --php_version 8.1 --skip_caddy --skip_vhost --skip_containers --skip_dns --debug
```

A hook szkriptje ugyanazokat az argumentumokat fogja kapni:

```bash
bash /etc/openpanel/openpanel/hooks/post_domains-add.sh example.com johndoe --docroot /var/www/example --php_version 8.1 --skip_caddy --skip_vhost --skip_containers --skip_dns --debug
```
