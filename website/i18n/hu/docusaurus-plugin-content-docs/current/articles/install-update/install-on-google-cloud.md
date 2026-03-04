# OpenPanel telepítés a Google Cloudon

Ez az útmutató végigvezeti az **OpenPanel** telepítésén egy Google Cloud VM-példányon Ubuntuval.

---

## Előfeltételek

- Egy aktív [Google Cloud-fiók](https://console.cloud.google.com/).
- Az SSH, a Compute Engine és a tűzfalbeállítások alapvető ismerete.

---

## 1. lépés: Hozzon létre egy virtuális gép-példányt

1. Jelentkezzen be a **Google Cloud Console** szolgáltatásba.
2. Lépjen a **Compute Engine → Virtuálisgép-példányok → Példány létrehozása** elemre.
3. Válasszon ki egy **indító lemezképet**:
- Ubuntu 24.04
4. Válasszon egy **géptípust**, amely megfelel a minimális követelményeknek:
- **1 CPU, 2 GB RAM** (ajánlott: "e2-small" vagy azzal egyenértékű)
5. Adjon hozzá egy **hálózati címkét** (pl. "OPENPANEL") a tűzfalszabályokhoz.
6. A **Hálózati interfészek** részben rendeljen hozzá egy **statikus külső IP-címet**.
7. Kattintson a **Létrehozás** gombra a virtuális gép elindításához.

---

## 2. lépés: Konfigurálja a tűzfalszabályokat

1. Lépjen a **VPC hálózat → Tűzfalszabályok → Tűzfalszabály létrehozása** elemre.
2. Állítsa be a szabályt:
- **Célok:** Meghatározott célcímkék → `OPENPANEL`
- **Forrás IP-tartományok:** `0.0.0.0/0`
- **Protokollok és portok:** `tcp:2087` (felügyeleti panel portja)
3. Mentse el a tűzfalszabályt.

> Az OpenPanel saját belső tűzfalát kezeli; a fenti szabály külső hozzáférést tesz lehetővé az adminisztrációs panelhez.

---

## 3. lépés: Csatlakozzon a virtuális géphez

Használja az SSH-t a csatlakozáshoz:

```bash
ssh username@yourStaticIpAddress
````

> Cserélje ki a "felhasználónév" szót a virtuális gép felhasználónevére (pl. "ubuntu" vagy "debian"), a "yourStaticIpAddress" pedig a statikus külső IP-címére.

---

## 4. lépés: Futtassa az OpenPanel telepítőt

Csatlakozás után futtassa a telepítő szkriptet:

```bash
bash <(curl -sSL https://openpanel.org)
```

> Kövesse az utasításokat a kívánt adatbázismotor kiválasztásához és a telepítés befejezéséhez.

---

## 5. lépés: Nyissa meg az OpenPanel-t

1. Nyissa meg böngészőjét, és keresse meg a következőt:

```
https://yourStaticIpAddress:2087
```

2. Jelentkezzen be a telepítés során létrehozott hitelesítő adatokkal.

---

## Megjegyzések

* Győződjön meg arról, hogy virtuális gépe megfelel az [OpenPanel minimális követelményeinek](https://openpanel.com/docs/admin/intro/#requirements).
* A telepítés után fontolja meg a tűzfal hozzáférésének korlátozását a nagyobb biztonság érdekében.
* Használja az [OpenPanel Install Command Generator](https://openpanel.com/install) alkalmazást a speciális konfigurációs lehetőségekért.

---

**Gratulálunk!** Sikeresen telepítette az OpenPanel alkalmazást a Google Cloud szolgáltatásban.
