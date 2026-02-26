# Az OpenAdmin felhasználói felületi hibáinak elhárítása

## 500-as Hiba

Ha **500-as hiba** (500 Error) jelenik meg az OpenAdmin felületén, engedélyezd a `dev_mode` beállítást, és vizsgáld meg a naplófájlt, hogy lásd a futtatott pontos parancsot és a kapott választ.

Ha segítségre van szükséged, a hibaüzenetet bemásolhatod a [támogatási fórumunkra](https://community.openpanel.org/) vagy a [Discord csatornánkra](https://discord.openpanel.com/) a hibaelhárításhoz.

---

## A felület nem reagál

Ha egy funkció nem a várt módon működik (például ha egy gombra kattintva nem történik semmi), az valószínűleg egy **front-end probléma**.

A hibaelhárításhoz:

1. Nyisd meg a böngésződ **Fejlesztői eszközeit / Developer Tools** (általában `F12` vagy `Ctrl+Shift+I` / `Cmd+Option+I` billentyűkombináció).
2. Lépj a **Network** (Hálózat) fülre.
3. Ismételd meg a nem működő műveletet a weblapon.
4. Ellenőrizd a **Network** fülön megjelenő kéréseket, és nézd meg, hogy vannak-e hibák a **Console** (Konzol) naplóban.

* A hálózati fülön kiválasztva egy adott kérést megnézheted a háttérrendszer (backend) által visszaküldött választ.
* Ha a válasz nem tartalmaz elegendő információt a probléma feltárásához, engedélyezd a **dev_mode**-ot a szerveren, majd ellenőrizd a Docker naplókat.

---

## Dev Mode (Fejlesztői mód)

A **dev_mode** engedélyezése lehetővé teszi, hogy az OpenPanel és az OpenAdmin felületek minden egyes kérésről részletes hibakeresési (debugging) információkat naplózzanak. Ez az opció segít az adminisztrátoroknak abban, hogy lássák a vezérlőpult által futtatott pontos parancsokat és az azokra kapott válaszokat.

A `dev_mode` bekapcsolásához:

```bash
opencli config update dev_mode on
```

Ezután indítsd újra a panelt.

Amikor a fejlesztői mód aktív, a felhasználói pult részletes naplóit ezen a módon érheted el:

```bash
tail -f /var/log/openpanel/admin/error.log
```

Ezek a naplók rendkívül részletes információkat biztosítanak a hatékony hibaelhárításhoz.
