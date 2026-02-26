# FOSSBilling

Az OpenPanel Enterprise Edition számlázási integrációkat tartalmaz a [WHMCS](/docs/articles/extensions/openpanel-and-whmcs/) és a FOSSBilling szolgáltatással.

## OpenPanel

### FOSSBilling engedélyezési listája

Mielőtt beállítaná az OpenPanel szerverkezelőt a FOSSBillingben, először engedélyeznie kell a FOSSBilling szerver IP-címét az OpenAdmin felületen belül, és engedélyeznie kell az API-hozzáférést.

A FOSSBilling szerver IP-címéhez való hozzáférés engedélyezéséhez először ellenőrizze az adott szerver IP-címét, a terminálról futtathatja:

```bash
curl ip.openpanel.com
```

Jelentkezzen be az OpenAdminba, és a **Beállítások > Tűzfal** alatt adja meg a FOSSBilling szerver IP-címét az **IP-cím engedélyezése** alatt:

![whitelist ip](https://i.postimg.cc/433M6LBr/2024-08-04-16-10.png)

### API hozzáférés engedélyezése

Az API hozzáférés engedélyezéséhez az OpenPanel felületén lépjen a **Beállítások > API-hozzáférés** elemre az OpenAdmin felületen, és kattintson az „API hozzáférés engedélyezése” gombra:

![api engedélyezése](https://i.postimg.cc/VsthWbWL/2024-08-04-16-14.png)

## FOSSBilling

### Az OpenPanel Server Manager letöltése

A FOSSBilling szerveren keresse meg azt a könyvtárat, ahol a FOSSBilling telepítve van, és futtassa ezt a parancsot a legújabb OpenPanel Server Manager letöltéséhez:

```bash
wget -O library/Server/Manager/OpenPanel.php https://raw.githubusercontent.com/stefanpejcic/FOSSBilling-OpenPanel/main/OpenPanel.php
```

### OpenPanel Server hozzáadása

Jelentkezzen be FOSSBilling adminisztrációs paneljére, és lépjen a **Rendszer -> Tárhelytervek és szerverek** elemre a navigációs sávon belül, majd kattintson az "Új szerver" elemre:

![új szerver](https://i.postimg.cc/bYV8DngC/2024-08-04-15-19.png)

az új formában be kell állítanunk:

![nyílt panel szerver hozzáadása](https://i.postimg.cc/jKcjYwHJ/2024-08-04-15-21.png)


4. Név: bármi, amivel azonosítani szeretné a szervert
5. Gazdanév: ha tartományt használ az OpenPanel eléréséhez, adja hozzá ide, ellenkező esetben adja meg az IP-címet.
6. IP: az OpenPanel szerver IP címének beállítása.
7. Hozzárendelt IP-címek: az OpenPanel szerveren hozzáadott IP-címek beállítása.
8. Névszerver 1: Állítsa be a tartományokhoz használt névszervert
9. 2. névszerver: Állítsa be a tartományokhoz használt névszervert
10. A szerver kezeli: **Válassza ki az OpenPanel-t**
11. Felhasználónév: Állítsa be az OpenAdmin panel felhasználónevét
12. Jelszó: Állítsa be az OpenAdmin panel jelszavát
13. Csatlakozási port: Állítsa `2087-re
14. Biztonságos kapcsolat használata: *Igen* ha domain nevet használ a paneleléréshez, ellenkező esetben *Nem*

és kattintson a "Szerver hozzáadása" gombra.

### Tárhelyterv hozzáadása az OpenPanelhez

A FOSSBilling adminisztrációs paneléről lépjen a **Rendszer -> Tárhelytervek és szerverek** elemre a navigációs sávon belül, majd kattintson az "Új tárhelyterv" elemre.

Adja meg a csomag nevét **ugyanaz, mint az OpenPanel tárhelytervén**, majd kattintson a "Hostolási terv létrehozása" gombra.

![terv hozzáadása](https://i.postimg.cc/02LsZqL7/2024-08-04-15-23.png)


A mysql típust és a webszervert is beállíthatja további paraméterként:

| Paraméter | Értékek |
| -------- | ------- |
| `mysql_type` | `apache` `nginx` `openresty` `lakk+apache` `varnish+nginx` `lakk+openresty` |
| "webszerver" | `mysql` `mariadb` |

![további paraméterek](https://i.postimg.cc/S4DGnS4S/2025-05-01-10-33.png)


### OpenPanel Server hozzárendelése a termékhez

A FOSSBilling adminisztrációs paneléről lépjen a **Termékek -> Termékek és szolgáltatások** elemre a navigációs sávon belül, majd kattintson a terv szerkesztési ikonjára:

![terv szerkesztése](https://i.postimg.cc/N0twqkGM/2024-08-04-15-24.png)

Kattintson a "Konfiguráció" elemre, és a Szervernél válassza ki az OpenPanel kiszolgálót, és tárhelytervhez állítsa be úgy, hogy megegyezzen az OpenPanel szerveren lévő terv nevével:

![fontos](https://i.postimg.cc/GmG155CV/2024-08-04-15-26.png)

### Kapcsolat tesztelése

Hozzon létre egy új klienst, és rendelje meg azt a terméket, amely OpenPanel-fiók létrehozására van konfigurálva.

A felhasználónak be kell tudnia jelentkezni OpenPanel-fiókjába:

![bejelentkezés](https://i.postimg.cc/x882pjf3/2024-08-04-15-17.png)

és a jelszó visszaállításához:

![jelszó visszaállítása](https://i.postimg.cc/PJ7kgGNs/2024-08-04-15-17-1.png)
