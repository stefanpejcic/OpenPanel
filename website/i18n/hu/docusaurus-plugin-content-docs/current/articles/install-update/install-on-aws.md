# Telepítés AWS EC2-re

Ez az útmutató végigvezeti Önt az **OpenPanel** telepítésén Ubuntu vagy Debian AWS EC2 példányon.

## Előfeltételek

- Aktív [AWS-fiók](https://aws.amazon.com/console/).
- Az EC2, SSH és terminálparancsok alapszintű ismerete.

---

## 1. lépés: Indítson el egy EC2 példányt

1. Jelentkezzen be az **[AWS Management Console](https://aws.amazon.com/console/)** szolgáltatásba.
2. Lépjen a **EC2** → **Példányok** → **Példányok indítása** lehetőségre.
3. Válasszon ki egy operációs rendszert (válasszon az alábbiak közül):
- Ubuntu 24.04
- Debian 11
*(x86 vagy ARM architektúra támogatott)*.
4. Válasszon egy példánytípust, amely megfelel a minimális követelményeknek:
-**1 CPU**
-**1 GB RAM**
*(Ajánlott: "t3.small" vagy azzal egyenértékű)*.

---

## 2. lépés: Az SSH-hozzáférés konfigurálása

1. Hozzon létre vagy válasszon egy meglévő **kulcspárt** az SSH-hozzáféréshez.
2. Töltse le a privát kulcsot (`.pem`), és tárolja biztonságosan.

---

## 3. lépés: Konfigurálja a hálózati beállításokat

1. Válassza ki a kívánt **VPC** és **alhálózatot**.
2. Hozzon létre egy **biztonsági csoportot**, amely engedélyezi az összes TCP-forgalmat:
> Az OpenPanel saját tűzfalát kezeli, így az összes TCP engedélyezése elfogadható.
3. Rendeljen ki és rendeljen hozzá egy **Elasztikus IP-címet** a példányához egy rögzített nyilvános IP-címhez.

---

## 4. lépés: Csatlakozzon a példányához

Használja az SSH-t az EC2-példányhoz való csatlakozáshoz:

```bash
ssh -i your_private_key.pem ubuntu@yourElasticIpAddress
```

> Cserélje ki a "your_private_key.pem"-et a kulcsfájljával, a "yourElasticIpAddress"-et pedig a hozzárendelt rugalmas IP-címmel.


---



## 5. lépés: Futtassa az OpenPanel telepítőt

Miután bejelentkezett, futtassa a telepítő szkriptet:

```bash
bash <(curl -sSL https://openpanel.org)
```

---

## 6. lépés: Nyissa meg az OpenPanel-t

Telepítés után:

1. Nyissa meg a böngészőt, és keresse meg az Elastic IP-címét és a „2087-es” portot.

---

## Megjegyzések

* Győződjön meg arról, hogy szervere megfelel az OpenPanel [minimális követelményeinek](https://openpanel.com/docs/admin/intro/#requirements).
* Éles környezetben a biztonság érdekében fontolja meg a biztonsági csoporthoz való hozzáférés korlátozását a telepítés után.
* A speciális beállításokhoz tekintse meg az [OpenPanel Install Command Generator] (https://openpanel.com/install) webhelyet.

---

**Gratulálunk!** Sikeresen telepítette az OpenPanel alkalmazást az AWS EC2 rendszeren.

