# E-mail hibák

Az e-mailes kommunikáció létfontosságú minden szervezet számára, de olyan problémák merülhetnek fel, amelyek megzavarják ezt az alapvető szolgáltatást. Ez az útmutató gyakorlati lépéseket tartalmaz az [OpenPanel Enterprise Edition]-ban (/product/openpanel-premium-control-panel/) előforduló gyakori e-mail-hibák diagnosztizálására és megoldására, biztosítva ezzel, hogy levelezőrendszerei zökkenőmentesen és hatékonyan működjenek.

## Nem tud e-mailt küldeni

A felhőtárhely-szolgáltatók, például a Hetzner blokkolják az e-mailek küldését a „25” és „465” portok használatával. Győződjön meg arról, hogy a szolgáltató nem blokkolja ezeket a portokat.


## HIBAlista levelező felhasználó: '/tmp/docker-mailserver/postfix-accounts.cf' nem létezik

A Dovecot nem indul el az első e-mail fiók létrehozásáig, ezért az e-mail fiókok listázása az "opencli email-setup email list" paranccsal az e-mail fiók létrehozása előtt hibát eredményez:
```bash
root@stefi:/usr/local/mail/openmail# opencli email-setup email list
2024-08-27 07:55:15+00:00 ERROR listmailuser: '/tmp/docker-mailserver/postfix-accounts.cf' does not exist
2024-08-27 07:55:15+00:00 ERROR listmailuser: Aborting
```

**Megoldás**: Először hozzon létre egy e-mail fiókot az [opencli email-setup email add](https://dev.openpanel.com/cli/email.html#Create-email) segítségével.

----


## Hiba: auth-master: userdb lookup(user@example.net): connect(/run/dovecot/auth-userdb) sikertelen: Nincs ilyen fájl vagy könyvtár

Közvetlenül az első e-mail fiók létrehozása után a dovecot még mindig a háttérben hozza létre a szükséges fájlokat, így az "opencli email-setup e-mail lista" azonnali futtatása az első e-mail fiók létrehozása után hibát eredményez:

```
root@stefi:/usr/local/mail/openmail# opencli email-setup email list
doveadm(stefan@stefi.openpanel.site)<577><>: Error: auth-master: userdb lookup(stefan@stefi.openpanel.site): connect(/run/dovecot/auth-userdb) failed: No such file or directory
doveadm(stefan@stefi.openpanel.site): Error: User lookup failed: Internal error occurred. Refer to server log for more information.
2024-08-27 07:55:46+00:00 ERROR listmailuser: Supplied non-number argument '' to '_bytes_to_human_readable_size()'
2024-08-27 07:55:46+00:00 ERROR listmailuser: Aborting
2024-08-27 07:55:46+00:00 ERROR listmailuser: Supplied non-number argument '' to '_bytes_to_human_readable_size()'
2024-08-27 07:55:46+00:00 ERROR listmailuser: Aborting
* stefan@stefi.openpanel.site (  /  ) [%]
```

**Megoldás**: Ismételje meg az "openncli email-setup email list" parancsot

----
