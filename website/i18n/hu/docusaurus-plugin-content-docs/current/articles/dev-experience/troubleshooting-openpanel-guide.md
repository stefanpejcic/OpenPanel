# Az OpenPanel hibaelhárítása

Az OpenPanel hibaelhárítása során az első lépés a probléma **reprodukciója**. Ha meg tudja győződni arról, hogy ez egy hiba (és mint minden szoftvernél, hibák előfordulnak), próbáljon meg következetes módon kiváltani.

* Ha a probléma **reprodukálható**, kérjük, [jelentse a GitHub Issues oldalon] (https://github.com/stefanpejcic/OpenPanel/issues/new/choose), hogy prioritást állíthassunk fel a következő frissítésnél.
* Ha önállóan szeretne további vizsgálatot végezni, kövesse az alábbi lépéseket.

---

## 1. lépés: Engedélyezze a Fejlesztői módot

Futtassa a következő parancsot a részletes hibakeresés bekapcsolásához mind az **OpenPanel**, mind az **OpenAdmin** felületeken:

```bash
opencli config update dev_mode on
```

---

## 2. lépés: Ellenőrizze a naplókat

Reprodukálja a problémát a felhasználói felületen, majd egyidejűleg tekintse át a naplókat a hibák észlelése érdekében.

* **OpenPanel naplók:**

``` bash
docker logs -f openpanel
  ```
* **OpenAdmin naplók:**

``` bash
tail -f /var/log/openpanel/admin/error.log
  ```

---

## 3. lépés: A Fejlesztői mód letiltása

A befejezés után kapcsolja ki a fejlesztői módot a szükségtelen naplónövekedés elkerülése érdekében:

```bash
opencli config update dev_mode off
```

