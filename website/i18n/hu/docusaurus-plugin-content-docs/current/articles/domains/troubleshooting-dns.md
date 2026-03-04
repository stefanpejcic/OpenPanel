# DNS hibaelhárítás

Győződjön meg arról, hogy a gazdagépen lévő DNS-kiszolgáló válaszol a tartományzónára vonatkozó kérésekre:

`dig <domain> @<IP-cím> ANY +short`


## szerver nem érhető el
A "kiszolgálók nem érhetők el" válaszüzenet azt jelzi, hogy a bind9 szolgáltatás nem fut.

Próbáld meg újraindítani a szolgáltatást az **OpenAdmin > Szolgáltatások** menüpontból.

És ha a szolgáltatás újraindítása sikertelen, ellenőrizze a szolgáltatási naplókat.

A szokásos hibás egy domain érvénytelen DNS-zónafájlja, ezért szerkessze a hibában említett fájlt, mentse el a változtatásokat, majd indítsa újra a szolgáltatást.

----

## üres dig válasz

A dig parancs üres válasza azt jelzi, hogy a DNS-kiszolgáló nem rendelkezik információval a tartományról. Lehetséges, hogy nincs hozzáadva egyetlen fiókhoz sem, vagy hiányzik a DNS-zónafájlja.

Ha ellenőrizni szeretné, hogy a domain hozzá van-e adva valamelyik fiókhoz, használja a következő parancsot:
`openncli domains-whoowns <domain_name>`

---

## Nem lehet csatlakozni a szerverhez

Próbáljon meg Telneten keresztül csatlakozni a szerver 53-as portjához:

`telnet <A szerver IP-címe> 53`

Ha nem tud csatlakozni, ellenőrizze a tűzfal beállításait az adminisztrációs panelen.


---

## zóna nem töltődik be hibák miatt

Amikor manuálisan szerkeszti a zónafájlokat, vagy importál, majd egy másik szolgáltatótól, fontos, hogy ellenőrizze a fájlokat szintaktikai hibákra. Ha bármilyen hiba van, a zónafájl kimarad a szolgáltatás újratöltésekor, de a szolgáltatás újraindítása meghiúsul.

A tartományzóna fájl szintaxisának ellenőrzése:
```bash
named-checkzone <DOMAIN> /etc/bind/zones/<DOMAIN>.zone
```

Példa:
```bash
# named-checkzone example.com /etc/bind/zones/example.com.zone
zone example.com/IN: NS 'ns17.openpanel.example.com' has no address records (A or AAAA)
zone example.com/IN: not loaded due to errors.
```
