# DNS-fürtözés

A DNS-fürtözés lehetővé teszi a DNS-rekordok szinkronizálását több OpenPanel-kiszolgáló között, így redundanciát és méretezhetőséget biztosít a DNS-infrastruktúra számára.

A fürtben részt vevő összes kiszolgálón futnia kell a [BIND9](https://www.isc.org/bind/) programnak – vagy az OpenPanel által telepítve, vagy önálló szolgáltatásként vagy tárolóként.

## Ossza meg a DNS-t több OpenPanel-kiszolgálón

A redundáns DNS-fürt felépítésének legegyszerűbb módja az OpenPanel futtatása két vagy több kiszolgálón, amelyek mindegyike szinkronban kezeli a DNS-zónákat és a névszervereket. Az alábbiakban leírtakon kívül nincs szükség további konfigurációra.

### 1. lépés: Hozzon létre SSH-hozzáférést a kiszolgálók között

Tegyük fel, hogy két szervere van:

- Szerver #1 IP: "185.241.214.214".
- Szerver #2 IP: `95.217.216.36`

Először be kell állítania az SSH-kulcs alapú hitelesítést mindkét irányban (az 1. kiszolgálóról a 2. szerverre és fordítva), hogy a root SSH hozzáférés jelszókérés nélkül is lehetséges legyen.

SSH-kulcsok létrehozása minden szerveren (ha még nem hozta létre):

```bash
ssh-keygen -t rsa -b 4096
```

Ezután másolja át mindegyik szerver nyilvános kulcsát a másikra:

```bash
ssh-copy-id root@185.241.214.214
ssh-copy-id root@95.217.216.36
```

Ellenőrizze a jelszó nélküli SSH-kapcsolatok működését:

```bash
ssh root@185.241.214.214
ssh root@95.217.216.36
```


### 2. lépés: Hozzon létre névszervereket a DNS-zónában

Példaként a `sajatdomain.com` tartomány használatával határozzon meg két névszervert:

- `dns1.yourdomain.com` → a kiszolgáló #1 IP-címére mutat (A rekord)
- "dns2.yourdomain.com" → a szerver #2 IP-címére mutat (A rekord)

Adja hozzá ezeket az A rekordokat a domain DNS-szolgáltatójához a "sajatdomain.com" domainhez.


### 3. lépés: Regisztráljon névszervereket az OpenPanelben

A **főkiszolgálón** nyissa meg az OpenAdmin panelt:
* Lépjen a **Beállítások > OpenPanel > Névszerverek** elemre.
* Adja hozzá a „dns1.yourdomain.com” és a „dns2.yourdomain.com” címet is

[![2025-07-09-17-30.png](https://i.postimg.cc/kXnvzCwW/2025-07-09-17-30.png)](https://postimg.cc/jCFfnGpj)

### 4. lépés: Engedélyezze a DNS-fürtözést

A **főkiszolgálón**:

* Lépjen az **OpenAdmin > Domains > DNS Cluster** oldalra.
* Kattintson a **DNS-fürtözés engedélyezése** lehetőségre

[![2025-07-09-17-32.png](https://i.postimg.cc/FzG3NfG3/2025-07-09-17-32.png)](https://postimg.cc/2LbVxSsS)

* Kattintson a **Kiszolgáló hozzáadása** lehetőségre, és adja meg a szolgakiszolgáló IP-címét, majd az **Hozzáadás** gombot.

[![2025-07-09-17-33.png](https://i.postimg.cc/7PX2C2MT/2025-07-09-17-33.png)](https://postimg.cc/3W4RVWqK)

### Tesztelje klaszterét

Adjon hozzá új tartományt bármelyik kiszolgálón egy OpenPanel felhasználói fiókon keresztül.

Ezután ellenőrizze, hogy a DNS-zóna szinkronizálva van-e mindkét szerveren a "dig" paranccsal:

```bash
dig A +short yourdomain.com @185.241.214.214
dig A +short yourdomain.com @95.217.216.36
```

Cserélje ki a "sajatdomain.com" címet a hozzáadott domainre.

Ha mindkét szerver a megfelelő IP-t adja vissza, akkor a DNS-fürtbeállítás működik!
