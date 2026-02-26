# OpenPanel telepítés a Microsoft Azure-ban

Ez az útmutató végigvezeti az **OpenPanel** telepítésén egy Azure virtuális gépen Ubuntuval.

---

## Előfeltételek

- Egy aktív [Microsoft Azure-fiók](https://portal.azure.com/).
- Az SSH, az Azure Virtual Machines és a hálózati/tűzfalbeállítások alapvető ismerete.

---

## 1. lépés: Hozzon létre egy virtuális gépet

1. Jelentkezzen be az **[Azure Portal](https://portal.azure.com/)** webhelyre.
2. Lépjen a **Virtuális gépek → Létrehozás → Virtuális gép** elemre.
3. Válasszon ki egy **képet**:
- Ubuntu 24.04
4. Válasszon egy **méretet**, amely megfelel a minimális követelményeknek:
- **1 CPU, 2 GB RAM** (ajánlott: "Standard_B1s" vagy azzal egyenértékű)
5. Válassza a **hitelesítési típust**:
- **Nyilvános SSH kulcs** (ajánlott)
- Adjon meg egy felhasználónevet, és töltse fel vagy illessze be nyilvános SSH-kulcsát
6. Hálózati és tűzfalbeállítások konfigurálása:
- Engedélyezze az összes TCP-forgalom **vagy**-specifikus portját:
- "22" (SSH)
- "80" (HTTP)
- "443" (HTTPS)
- "2083" (OpenPanel)
- "2087" (OpenAdmin)
- "443 UDP" (HTTP/3 támogatáshoz)

---

## 2. lépés: Csatlakozzon a virtuális géphez

Használja az SSH-t a csatlakozáshoz:

```bash
ssh -i your_private_key azure@yourIpAddress
````

> Cserélje ki a "saját_privát_kulcsot" a privát kulcs fájljára, az "a saját IP címét" pedig a virtuális gép nyilvános IP-címére.

---

## 3. lépés: Futtassa az OpenPanel telepítőt

Miután bejelentkezett, futtassa a telepítő szkriptet:

```bash
bash <(curl -sSL https://openpanel.org)
```

> Kövesse az utasításokat a kívánt adatbázismotor kiválasztásához és a telepítés befejezéséhez.

---

## 4. lépés: Nyissa meg az OpenPanel-t

1. Nyissa meg böngészőjét, és keresse meg a következőt:

```
https://yourIpAddress:2087
```

2. Jelentkezzen be a telepítés során létrehozott hitelesítő adatokkal.

---

## Megjegyzések

* Győződjön meg arról, hogy virtuális gépe megfelel az [OpenPanel minimális követelményeinek](https://openpanel.com/docs/admin/intro/#requirements).
* A telepítés után fontolja meg a tűzfal hozzáférésének korlátozását a nagyobb biztonság érdekében.
* Használja az [OpenPanel Install Command Generator](https://openpanel.com/install) alkalmazást a speciális konfigurációs lehetőségekért.

---

**Gratulálunk!** Sikeresen telepítette az OpenPanel alkalmazást a Microsoft Azure rendszeren.
