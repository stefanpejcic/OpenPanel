---
sidebar_position: 1
---

# Domain

Webhely létrehozásához az első lépés a [domainnév hozzáadása](/docs/panel/domains/#create-a-new-domain).

Ha a **Domains** modul engedélyezve van a szerveren, és felhasználói fiókja rendelkezik hozzáféréssel, akkor megjelenik egy táblázat, amely felsorolja az összes jelenlegi domaint, a tartományok teljes számát, egy keresősávot és egy új domain hozzáadásának lehetőségét.

Ezen a felületen megtekintheti:

* **Domain állapota**: Aktív vagy Felfüggesztett
* **Dokumentumgyökér** (mappa, ahol a webhely fájljait tárolják)
* **SSL beállítások**: Automatikus vagy Egyéni
* **Domainkezelési beállítások**: Az engedélyezett szolgáltatások alapján

## Elérhető műveletek

A szerveren engedélyezett szolgáltatásoktól függően a következő műveletek érhetők el tartományonként:

* **DNS zóna szerkesztése** — ha a DNS szolgáltatás engedélyezve van
* **WAF kezelése** – ha a WAF funkció engedélyezve van
* **Dokumentum gyökér módosítása**
* ** VirtualHosts szerkesztése**
* **Átirányítás** – ha az Átirányítás funkció engedélyezve van
* **Nagybetűs írás** – ha a nagybetűs írás funkció engedélyezve van
* **Domain felfüggesztése/felfüggesztésének visszavonása**
**Domain törlése**

## Hozzon létre egy új domaint

Új domain hozzáadása:

1. Kattintson a **"Domain hozzáadása"** gombra.
2. Adja meg a domain nevet.
3. Kattintson a **"Domain hozzáadása"** gombra a mentéshez.

Más panelekkel ellentétben az OpenPanel minden tartományt egyformán kezel. Ezen az egyetlen felületen **elsődleges domaineket**, **addon domaineket** vagy **aldomaineket** adhat hozzá.

A hozzáadást követően a rendszer automatikusan megpróbál ingyenes [Let’s Encrypt](https://letsencrypt.org/getting-started/) SSL-tanúsítványt kiállítani. Sikeres esetben a tanúsítványt azonnal alkalmazzák.

## Domain törlése

Domain törlése:

1. Kattintson a **"Domain törlése"** lehetőségre a domain legördülő menüjében.
2. Megjelenik egy megerősítő oldal. A folytatáshoz kattintson a **"Törlés"** gombra.

> Ha a tartomány aktív alkalmazásokhoz van kapcsolva (pl. [PM2](/docs/panel/applications/pm2) vagy [WP Manager](/docs/panel/applications/wordpress)), a törlés le lesz tiltva mindaddig, amíg el nem távolítják ezeket az alkalmazásokat.
> Ez megakadályozza a futó webhelyekhez kapcsolódó domainek véletlen eltávolítását.

Egy domain törlésével **véglegesen eltávolítja** a következőket:

1. Nginx konfigurációs fájl
2. DNS zónafájl
3. SSL tanúsítvány
4. IP-blokkoló szabályok a tartományhoz
5. A tartományhoz társított átirányítások

## Átirányítások

### Átirányítás hozzáadása

Átirányítás létrehozása:

1. Kattintson a domain melletti **"Átirányítás létrehozása"** gombra.
2. Írja be a teljes URL-t ("http://" vagy "https://" előtaggal kell kezdődnie).
3. Kattintson a **"Mentés"** gombra az alkalmazáshoz, vagy a **"Mégse"** gombra az elvetéshez.

### Átirányítás szerkesztése

Kattintson a **ceruza ikonra** egy meglévő átirányítási URL mellett a módosításához.

### Átirányítás törlése

Kattintson a **kereszt ikonra** az átirányítási URL mellett az eltávolításhoz.

## VirtualHosts fájl szerkesztése

A **VirtualHosts** fájl határozza meg a tartomány konfigurációját az Nginxen vagy az Apache-on belül. Olyan beállításokat tartalmaz, mint:

* Hozzáférés a naplókhoz
* PHP verzió
* Alkalmazás futók
* Átirányítási szabályok
* Egyéni utasítások

A fájl szerkesztése:

1. Kattintson a **"Edit VirtualHosts"** lehetőségre a domain legördülő menüjében.
2. Megnyílik egy új oldal a Vhost fájl tartalmával.
3. Végezze el a módosításokat, majd kattintson a **"Mentés"** gombra.

Mentés után az OpenPanel automatikusan újraindítja a webszervert a módosítások alkalmazásához.
