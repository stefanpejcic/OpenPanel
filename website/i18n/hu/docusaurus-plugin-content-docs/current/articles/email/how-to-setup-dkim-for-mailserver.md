# DKIM beállítása

Győződjön meg arról, hogy az OpenPanel [**Enterprise kiadása**](https://openpanel.com/enterprise/) fut. Az e-mailes támogatás csak ebben a verzióban érhető el.

Kövesse [ezt az útmutatót] (/docs/articles/user-experience/how-to-setup-email-in-openpanel/), hogy engedélyezze az e-maileket az OpenPanelben.

---

Az OpenDKIM jelenleg alapértelmezés szerint engedélyezve van.

## 1. DKIM-kulcsok létrehozása

Egy tartományhoz tartozó DKIM-kulcsok létrehozásához futtassa:

```bash
opencli email-setup config dkim domain DOMAIN_NAME_HERE
```

---

## 2. DNS rekord

Amikor a DKIM-kulccsal aláírt leveleket küldi a levelezőszerver, a címzettnek ellenőriznie kell egy DNS TXT rekordot, hogy megbizonyosodjon a DKIM-aláírás megbízhatóságáról.

### OpenPanel DNS

Ha névszervereket használ ugyanazon az OpenPanel-kiszolgálón, amelyen a levelezőszerver fut, akkor futtassa:
```bash
docker cp openadmin_mailserver:/tmp/docker-mailserver/opendkim/keys/DOMAIN_NAME_HERE/mail.txt /tmp/mail.txt && cat /tmp/mail.txt >> /etc/bind/zones/DOMAIN_NAME_HERE.zone
```

Ügyeljen arra, hogy a `DOMAIN_NAME_HERE` helyett a *sajatdomain.com*-ot

Ez hozzáfűzi a DKIM rekordot a DNS-zónához.

---

### Külső DNS

Ha külső névszervereket, például [Cloudflare](https://www.cloudflare.com/) használ, futtassa ezt a parancsot a rekord megtekintéséhez:

```bash
docker cp openadmin_mailserver:/tmp/docker-mailserver/opendkim/keys/DOMAIN_NAME_HERE/mail.txt /tmp/mail.txt && cat /tmp/mail.txt
```

Adja hozzá a TXT rekordot a **távoli szerver DNS-zónájához**:

| **Mező** | **Érték** |
| --------- | ------------------------------------------------------------------------------ |
| **Típus** | "TXT" |
| **Név** | `mail._domainkey` |
| **TTL** | Alapértelmezett érték (vagy "3600" másodperc) |
| **Adatok** | A fájl tartalma zárójelben – az alábbi tanácsok szerint formázva |

**A TXT rekord értékének megfelelő formázása:**
A 255 karakternél hosszabb értékű "TXT" rekordokat több részre kell felosztani. Ez az oka annak, hogy az előállított mail.txt fájl (amely tartalmazza a DKIM-hez használható nyilvános kulcsot) több értékű részt tartalmaz a `(` és `)` közé dupla idézőjelbe foglalva.

A DNS webes felülete ehelyett belsőleg kezelheti ezt a szétválasztást, és [az érték egyetlen sorban várható] (https://serverfault.com/questions/763815/route-53-doesnt-allow-adding-dkim-keys-because-length-is-too-long) a felosztás helyett. Ha ez szükséges, kézzel kell formáznia az értéket az alábbiak szerint.

A létrehozott DNS-rekordfájlnak (mail.txt) ehhez hasonlóan kell kinéznie:

```
mail._domainkey IN TXT ( "v=DKIM1; k=rsa; "
"p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqQMMqhb1S52Rg7VFS3EC6JQIMxNDdiBmOKZvY5fiVtD3Z+yd9ZV+V8e4IARVoMXWcJWSR6xkloitzfrRtJRwOYvmrcgugOalkmM0V4Gy/2aXeamuiBuUc4esDQEI3egmtAsHcVY1XCoYfs+9VqoHEq3vdr3UQ8zP/l+FP5UfcaJFCK/ZllqcO2P1GjIDVSHLdPpRHbMP/tU1a9mNZ"
"5QMZBJ/JuJK/s+2bp8gpxKn8rh1akSQjlynlV9NI+7J3CC7CUf3bGvoXIrb37C/lpJehS39KNtcGdaRufKauSfqx/7SxA0zyZC+r13f7ASbMaQFzm+/RRusTqozY/p/MsWx8QIDAQAB"
) ;
```

Vegye ki a "( ... )" közötti tartalmat, kombinálja az idézőjelbe burkolt tartalmat, és távolítsa el a kettős idézőjeleket, beleértve a közöttük lévő szóközt. Ez a "TXT" rekord értéke, a fenti példa a következő lesz:

```
v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqQMMqhb1S52Rg7VFS3EC6JQIMxNDdiBmOKZvY5fiVtD3Z+yd9ZV+V8e4IARVoMXWcJWSR6xkloitzfrRtJRwOYvmrcgugOalkmM0V4Gy/2aXeamuiBuUc4esDQEI3egmtAsHcVY1XCoYfs+9VqoHEq3vdr3UQ8zP/l+FP5UfcaJFCK/ZllqcO2P1GjIDVSHLdPpRHbMP/tU1a9mNZ5QMZBJ/JuJK/s+2bp8gpxKn8rh1akSQjlynlV9NI+7J3CC7CUf3bGvoXIrb37C/lpJehS39KNtcGdaRufKauSfqx/7SxA0zyZC+r13f7ASbMaQFzm+/RRusTqozY/p/MsWx8QIDAQAB
```
---

### DNS tesztelése

Annak teszteléséhez, hogy az új DKIM rekord helyes-e, kérdezze le a "dig" paranccsal. A "TXT" értékre adott válasznak egyetlen sornak kell lennie, több részre bontva, dupla idézőjelbe foglalva:

```bash
$ dig +short TXT mail._domainkey.example.com
"v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqQMMqhb1S52Rg7VFS3EC6JQIMxNDdiBmO.."
```

---

## 3. Indítsa újra a Mailservert

Jelenleg a levelezőszervert le kell állítani, és újra kell indítani, hogy megkezdhesse a kimenő e-mailek aláírását a DKIM-mel. Lépjen az **OpenAdmin > Szolgáltatások > Állapot** elemre, és kattintson a „Leállítás” gombra a levelezőszerverhez, majd ismét a „Start” gombra.

[![2025-07-17-15-37.png](https://i.postimg.cc/d3hgY01F/2025-07-17-15-37.png)](https://postimg.cc/ctNF70wk)

---

## 4. Hibaelhárítás

Az [MxToolbox rendelkezik egy DKIM-ellenőrzővel](https://mxtoolbox.com/dkim.aspx), amellyel ellenőrizheti DKIM DNS-rekordjait.
