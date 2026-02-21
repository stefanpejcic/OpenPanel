---
sidebar_position: 1
---

# Konténerek

Az OpenPanel **Containers** oldala lehetővé teszi az adminisztrátor által a Docker Compose fájlokkal meghatározott Docker-szolgáltatások kezelését.

Ez a szakasz világos áttekintést nyújt a konténeres szolgáltatásokról és azok erőforrás-használatáról.

:::info
A funkció eléréséhez:
- A **Docker** modult a rendszergazdának engedélyeznie kell **szerveren**.
- Fiókjában engedélyezni kell a **Docker** funkciót.
:::

## Áttekintés

Az oldal tetején a következőket fogja látni:

- **Futó tárolók / Összes tárolók** - Azt jelzi, hogy hány szolgáltatás van jelenleg aktív.
- **Total CPU** – A tárhelytervhez rendelt CPU magok száma.
- **Teljes memória** – A tárhelytervhez rendelt RAM mennyisége (GB-ban).

A teljes erőforrások egy részét az egyes szolgáltatásokhoz rendelheti.

## Container Table

A táblázat minden sora egy konténeres szolgáltatást jelent, és a következőket jeleníti meg:

- **Név** – A Docker-szolgáltatás neve, a használt képpel és címkéjével (verziójával) együtt.
- **CPU használat**
- **Graph** – Valós idejű használat az allokált CPU százalékában.
- **Használat** – Mennyi CPU-t használ a tároló a hozzárendelt mennyiségből.
- **Kiosztva** – A szolgáltatáshoz kiosztott CPU magok száma.
- **memóriahasználat**
- **Graph** – Valós idejű használat a lefoglalt memória százalékában.
- **Felhasználás** – A szolgáltatás által a kiosztott mennyiségből felhasznált RAM.
- **Lefoglalt** – A szolgáltatáshoz lefoglalt memória (GB-ban).
- **Műveletek**
- Megmutatja, hogy a szolgáltatás **Engedélyezve** vagy **Letiltva**.
- Az állapotra kattintva váltható.
- Ha a szolgáltatás fut, megjelenik egy **Terminal** hivatkozás – kattintson rá egy webterminál ("docker exec") megnyitásához az adott tárolóhoz.

## Erőforrások szerkesztése

Egy szolgáltatás CPU- vagy memóriakorlátjának módosítása:

1. Mutasson az egérrel a **Kiosztva** érték fölé a táblázatban.
2. Kattintson a **ceruza** ikonra.
3. Állítsa be az értéket a beviteli mezőben.
4. Kattintson a **Mentés** gombra.

> Az OpenPanel ellenőrzi, hogy vannak-e rendelkezésre álló erőforrásai a módosítás végrehajtásához. Ha érvényes, a tároló automatikusan újraindul az új korlátokkal.

## Új szolgáltatások hozzáadása

Új Docker-szolgáltatás (tároló) hozzáadásához töltse ki a **Szolgáltatás hozzáadása** űrlapot a szükséges adatokkal.

- **Szolgáltatás neve** – A tároló egyedi neve.
- Betűvel kell kezdődnie, csak kisbetűket és számjegyeket tartalmazhat, és legalább 3 karakter hosszúnak kell lennie.
- Példa: "webapp", "redis1".

- **Kép** – Docker-kép használható.
- Példa: "nginx:latest", "redis:7.2".

- **Környezeti változók** - Nem kötelező. Adjon meg változókat "KEY=value" formátumban, soronként egyet.
- Példa:
    ```
REDIS_PASSWORD=titkos
DEBUG=igaz
    ```

- **CPU-korlát** – Maximális CPU-allokáció a tárolóhoz. Pozitív számnak kell lennie.
- Példa: "0,5", "1".

- **RAM Limit** – Maximális memóriafoglalás. Egy számnak kell lennie, amelyet "M" vagy "G" követ.
- Példa: "512M", "1,5G".

- **Hálózat** – Docker hálózat a tároló csatlakoztatásához.
- Példa: `backend_network`

- **Healthcheck (opcionális)** - YAML blokk, amely meghatározza a tároló állapotának ellenőrzését.
- Példa:
``` yaml
teszt: ["CMD", "curl", "-f", "http://localhost"]
intervallum: 30 mp
időtúllépés: 10 mp
újrapróbálkozás: 3
    ```  

**Érvényesítési szabályok**:

- **Szolgáltatás neve** – Egyedinek kell lennie, és követnie kell a formátumszabályokat.
- **CPU-korlát** – pozitív számnak kell lennie.
- **RAM-korlát** – „M” vagy „G” karakterrel kell végződnie.
- **Környezeti változók** – `KEY=value` formátumban kell lennie.
- **Healthcheck** - Érvényes YAML-nek kell lennie.

Minden szolgáltatás automatikusan **nagybetűs előtagot** használ a környezeti változókulcsokhoz.
A CPU és a RAM értékek környezeti változóként is tárolódnak az egyes szolgáltatásokhoz.

Például egy "nginx" nevű szolgáltatás a következő környezeti kulcsokkal rendelkezik:

  ```
NGINX_CPU
NGINX_RAM
  ``` 

> Az űrlap elküldése és érvényesítése után a szolgáltatás hozzáadódik a Docker Compose konfigurációjához, és a környezeti változók automatikusan frissülnek.

