# OpenPanel telepítés a Vultr-on

Ez az útmutató végigvezeti Önt az **OpenPanel** telepítésén egy Vultr-példányon Ubuntuval.

---

## Előfeltételek

- Aktív [Vultr-fiók] (https://www.vultr.com/).
- Az SSH és a terminálparancsok alapvető ismerete.

---

## 1. lépés: Hozzon létre egy Vultr-példányt

1. Jelentkezzen be **Vultr-fiókjába**.
2. Kattintson a **Termékek → Új példány telepítése** lehetőségre.
3. Válasszon egy **szerver helyét** a felhasználók közelében.
4. Válasszon ki egy **OS-képet**:
- Ubuntu 24.04
5. Válasszon egy **szerverméretet**, amely megfelel a minimális követelményeknek:
- **1 CPU, 2 GB RAM** (ajánlott: "Cloud Compute" vagy azzal egyenértékű)
6. Konfigurálja az **SSH-kulcs hitelesítést** (ajánlott), és adja hozzá nyilvános kulcsát.
7. Fejezze be a példány létrehozását.

---

## 2. lépés: Csatlakozzon a példányához

Használja az SSH-t a csatlakozáshoz:

```bash
ssh root@yourInstanceIp
````

> Cserélje ki a "yourInstanceIp" értéket a Vultr-példány nyilvános IP-címére.

---

## 3. lépés: Futtassa az OpenPanel telepítőt

Futtassa a telepítő szkriptet:

```bash
bash <(curl -sSL https://openpanel.org)
```

> Kövesse az utasításokat a kívánt adatbázismotor kiválasztásához és a telepítés befejezéséhez.

---

## 4. lépés: Nyissa meg az OpenPanel-t

1. Nyissa meg böngészőjét, és keresse meg a következőt:

```
https://yourInstanceIp:2087
```

2. Jelentkezzen be a telepítés során létrehozott hitelesítő adatokkal.

---

## Megjegyzések

* Győződjön meg arról, hogy a Vultr példány megfelel az [OpenPanel minimális követelményeinek](https://openpanel.com/docs/admin/intro/#requirements).
* Éles környezetben fontolja meg a tűzfal hozzáférésének korlátozását a telepítés után.
* Használja az [OpenPanel Install Command Generator](https://openpanel.com/install) alkalmazást a speciális konfigurációs lehetőségekért.

---

**Gratulálunk!** Sikeresen telepítette az OpenPanel alkalmazást a Vultr rendszeren.
