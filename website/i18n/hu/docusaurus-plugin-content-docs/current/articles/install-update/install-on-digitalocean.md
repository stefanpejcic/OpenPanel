# OpenPanel telepítés DigitalOcean-en

Ez az útmutató végigvezeti az **OpenPanel** telepítésén egy DigitalOcean Droplet Ubuntuval.

---

## Előfeltételek

- Aktív [DigitalOcean-fiók](https://www.digitalocean.com/).
- Az SSH és a terminálparancsok alapvető ismerete.

---

## 1. lépés: Hozzon létre egy cseppet

1. Jelentkezzen be **DigitalOcean fiókjába**.
2. Kattintson a **Létrehozás → Csepp** lehetőségre.
3. Válasszon operációs rendszert:
- Ubuntu 24.04
4. Válasszon ki egy **cseppméretet**:
- Ajánlott: **1 CPU, 2 GB RAM** (Prémium AMD NVMe SSD-vel a teljesítmény érdekében)
5. Válassza a **hitelesítés** lehetőséget:
- **SSH-kulcs** (ajánlott) vagy jelszó-hitelesítés
6. Rendeljen **Fenntartott IP-címet** egy statikus IP-címhez.
7. Fejezze be a létrehozást, és jegyezze fel Droplet IP-címét.

---

## 2. lépés: Csatlakozzon a Droplethez

Használja az SSH-t a csatlakozáshoz:
```bash
ssh root@yourDropletIpAddress
````

> Cserélje ki a `saját cseppcím-címét' a cseppecskéhez rendelt statikus IP-címre.

---

## 3. lépés: Futtassa az OpenPanel telepítőt

Miután bejelentkezett, futtassa a telepítő szkriptet:
```bash
bash <(curl -sSL https://openpanel.org)
```

---

## 4. lépés: Nyissa meg az OpenPanel-t

1. Nyissa meg böngészőjét, és keresse meg a következőt:

```
https://yourDropletIpAddress:2087
```

2. Jelentkezzen be a telepítés során létrehozott hitelesítő adatokkal.

---

## Megjegyzések

* Győződjön meg arról, hogy a Droplet megfelel az [OpenPanel minimális követelményeinek](https://openpanel.com/docs/admin/intro/#requirements).
* Gyártás esetén korlátozza a tűzfalhoz való hozzáférést a telepítés utáni expozíció korlátozása érdekében.
* Használja az [OpenPanel Install Command Generator](https://openpanel.com/install) alkalmazást a speciális konfigurációs lehetőségekért.

---

**Gratulálunk!** Sikeresen telepítette az OpenPanel szoftvert a DigitalOcean rendszeren.
