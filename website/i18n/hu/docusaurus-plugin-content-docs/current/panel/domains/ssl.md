---
sidebar_position: 5
---

# SSL

Az OpenPanel automatikusan generál és megújít SSL-tanúsítványokat minden tartományhoz a **Titkosítsuk** vagy **ZeroSSL** használatával.

**Egyéni SSL-tanúsítványt** is beállíthat bármely domainhez az **OpenPanel > Domains > SSL** menüpontban.

## Egyéni SSL

Saját SSL-tanúsítvány használatához:

1. Töltse fel tanúsítványfájljait az **OpenPanel > FileManager** oldalon keresztül.
2. Lépjen az **OpenPanel > Domains** elemre, majd kattintson az **SSL** elemre a domain mellett.
3. Az **Egyéni SSL konfigurálása** részben adja meg a fájl elérési útját:
* Tanúsítványfájl (`.crt`)
* Privát kulcs fájl (`.key`)
4. Kattintson a **Konfigurálás** gombra.

A konfigurálás után az egyéni tanúsítvány részletei ugyanazon az oldalon jelennek meg.

![képernyőkép a domainről egyéni ssl-vel](/img/panel/v2/openpanel_customssl.png)


## AutoSSL

Az **AutoSSL** az OpenPanel alapértelmezett beállítása.
Ha új domaint ad hozzá, nincs teendője.

Az **egyéni tanúsítványról az AutoSSL-re** visszaváltáshoz:

1. Lépjen az **OpenPanel > Domains** elemre, és kattintson az **SSL** lehetőségre a domain mellett.
2. Kattintson a **Váltás vissza AutoSSL-re** lehetőségre.

A tanúsítványt a rendszer automatikusan újra kiállítja, miután a domainhez a „https://” hivatkozáson keresztül hozzáfértek. A generálás után ugyanazon az oldalon jelenik meg.

![képernyőkép az autossl-lel rendelkező domainről](/img/panel/v2/openpanel_autossl.png)


### Követelmények

Az SSL sikeres generálása érdekében:

* A domain **A rekordjának** a szerver **IPv4-címére** kell mutatnia.
* A DNS-t **teljesen terjeszteni kell**. Az ellenőrzéshez használjon olyan eszközöket, mint a [whatsmydns.net](https://www.whatsmydns.net/#A).
* A tanúsítvány létrehozásához a domaint legalább egyszer el kell érni a „https://” protokollon keresztül. Nyissa meg a domaint egy böngészőben „https” használatával.

Ha:

* A domain nemrég lett hozzáadva,
* A DNS még nem mutatott a szerverre, ill
* A domainhez nem fértek hozzá `https://` keresztül,

Ezután az SSL részben a **„Nem található tanúsítvány”** felirat jelenik meg.

![képernyőkép a domainről autossl-lel, de még nincs ssl](/img/panel/v2/openpanel_autossl_no_ssl.png)
