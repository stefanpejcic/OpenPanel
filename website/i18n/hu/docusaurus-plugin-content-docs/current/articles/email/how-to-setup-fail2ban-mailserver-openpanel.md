# Fail2ban beállítása

Győződjön meg arról, hogy az OpenPanel [**Enterprise kiadása**](https://openpanel.com/enterprise/) fut. Az e-mailes támogatás csak ebben a verzióban érhető el.

Kövesse [ezt az útmutatót] (/docs/articles/user-experience/how-to-setup-email-in-openpanel/), hogy engedélyezze az e-maileket az OpenPanelben.

---

A Fail2ban egy behatolásgátló szoftver keretrendszer. Python programozási nyelven íródott, és a brute force támadások megelőzésére szolgál. Képes olyan POSIX rendszereken futni, amelyek interfésszel rendelkeznek egy csomagvezérlő rendszerhez vagy helyileg telepített tűzfalhoz, például NFTables vagy TCP Wrapper. [További információ](https://en.wikipedia.org/wiki/Fail2ban)

A Fail2Ban támogatás engedélyezéséhez legalább a `NET_ADMIN' képességet meg kell adni a levelezőszervernek, hogy interakcióba lépjen a kernellel és tiltsa le az IP-címeket.


## 1. Engedélyezze a Fail2ban funkciót

A Fail2ban engedélyezéséhez lépjen az **OpenAdmin > E-mailek > Beállítások** elemre, és a „Szolgáltatások engedélyezése” alatt válassza a **fail2ban** lehetőséget, majd kattintson a „Verem mentése” gombra.

[![2025-07-17-14-33.png](https://i.postimg.cc/G3q4ByVh/2025-07-17-14-33.png)](https://postimg.cc/qg6JSzp9)

---

## 2. Új levelezőszerver létrehozása

A Fail2ban megköveteli, hogy a levelezőszerver-tároló leálljon, és újra induljon az aktiváláshoz. Lépjen az **OpenAdmin > Szolgáltatások > Állapot** elemre, és kattintson a „Leállítás” gombra a levelezőszerverhez, majd ismét a „Start” gombra.

[![2025-07-17-15-37.png](https://i.postimg.cc/d3hgY01F/2025-07-17-15-37.png)](https://postimg.cc/ctNF70wk)

---

## Konfiguráció szerkesztése

A Mailserver automatikusan letiltja azon gazdagépek IP-címeit, amelyek **6** sikertelen kísérletet generáltak az elmúlt hét során. Maguk a tilalmak egy hétig tartanak. A Postfix jail úgy van beállítva, hogy a `mode = extra`-t használja.

Ha létrejön, a következő konfigurációs fájlok másolásra kerülnek a tárolóba az indítás során. A konfiguráció testreszabásához egyszerűen hozzon létre egy új fájlt, és szabja testre a példákat:

- `/usr/local/mail/openmail/docker-data/dms/config/fail2ban-jail.cf` - az egyes börtönök konfigurációjának és alapértelmezett értékeinek módosítása, [példafájl](https://github.com/docker-mailserver/docker-mailserver/blob/master/config-examples/.cf)
- `/usr/local/mail/openmail/docker-data/dms/config/fail2ban-fail2ban.cf` - az F2B viselkedésének általános beállítása, [példafájl](https://github.com/docker-mailserver/docker-mailserver/blob/master/config-examples/fail2ban.cf)

**Példa:**

Az IP blokkolása előtti sikertelen bejelentkezések számának módosításához (alapértelmezett: 6), vagy a bantime módosításához (alapértelmezett: 1w):

- hozzon létre egy fájlt `/usr/local/mail/openmail/docker-data/dms/config/fail2ban-jail.cf`
- illessze be [ezt a példát](https://github.com/docker-mailserver/docker-mailserver/blob/master/config-examples/fail2ban-jail.cf)
- módosítsa a "maxretry" és a "bantime" értékeket:

```yaml
[DEFAULT]

# "bantime" is the number of seconds that a host is banned.
bantime = 1w

# A host is banned if it has generated "maxretry" during the last "findtime"
# seconds.
findtime = 1w

# "maxretry" is the number of failures before a host get banned.
maxretry = 6

# "ignoreip" can be a list of IP addresses, CIDR masks or DNS hosts. Fail2ban
# will not ban a host which matches an address in this list. Several addresses
# can be defined using space (and/or comma) separator.
ignoreip = 127.0.0.1/8

# default ban action
# nftables-multiport: block IP only on affected port
# nftables-allports:  block IP on all ports
banaction = nftables-allports

[dovecot]
enabled = true

[postfix]
enabled = true
# For a reference on why this mode was chose, see
# https://github.com/docker-mailserver/docker-mailserver/issues/3256#issuecomment-1511188760
mode = extra

[postfix-sasl]
enabled = true

# This jail is used for manual bans.
# To ban an IP address use: opencli email-setup fail2ban ban <IP>
[custom]
enabled = true
bantime = 180d
port = smtp,pop3,pop3s,imap,imaps,submission,submissions,sieve
```

---

## A kitiltások megtekintése és kezelése

- `opencli email-setup fail2ban` - megjeleníti az összes tiltott IP-címet
- `opencli email-setup fail2ban status` - részletesebb állapotot mutat
- `opencli email-setup fail2ban ban <IP>` - IP-cím tiltása
- `opencli email-setup fail2ban unban <IP>` - IP-cím tiltásának feloldása
- "Openncli email-setup fail2ban log" - A fail2ban naplófájl megtekintése
