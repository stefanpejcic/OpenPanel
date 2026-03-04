# Lakkgyorsítótár tisztítása

Ha a Varnish engedélyezve van egy webhelyen, a gyorsítótár üríthető a terminálról vagy közvetlenül az OpenPanelen keresztül.

## Tisztítás a terminálról

A gyorsítótár bejegyzéseit manuálisan is törölheti a "curl" paranccsal.

### Adott URL törlése

Ez eltávolít egyetlen gyorsítótárban lévő objektumot (egy adott oldal, kép vagy fájl esetében):
```bash
curl -X PURGE -H "Host: example.com" http://example.com/wp-content/uploads/test.jpg
```

### Egy adott webhely megtisztítása

Ezzel eltávolítja az összes gyorsítótárazott objektumot** egy webhelyről (domain):
```
curl -X PURGE -H "Host: example.com" -H "X-Purge-Method: regex" http://example.com/.*
```

### Tisztítsa meg az összes webhelyet

Az összes gyorsítótárazott objektum eltávolítása a Varnish által kezelt összes tartományból:
```
curl -X BAN http://localhost/.* 
```

:::info
Alternatív megoldásként egyszerűen újraindíthatja a Lakk szolgáltatást (lásd alább).
:::

## Tisztítás a Cronstól

A cronjob gyorsítótárának törléséhez egyszerűen használja a fenti "curl" parancsokat, és ügyeljen arra, hogy a varnish/webservert vagy a php-t használja a tárolóhoz, amikor Cronokat hoz létre az OpenPanelen.

## Tisztítás a WordPressből

Egy WordPress webhely gyorsítótárának kiürítéséhez használja a [CLP Varnish Cache beépülő modult](https://wordpress.org/plugins/clp-varnish-cache/)

A bővítmény konfigurálásához:
1. Jelentkezzen be a wp-adminba, és lépjen a **Beállítások → CLP Lakk gyorsítótár** menüpontra.
2. Állítsa be a **lakk** elemet a 'Lakkszerver' értékre, majd ismét az **Engedélyezés** lehetőséget.

A gyorsítótár törlése a bővítmény használatával:
1. Jelentkezzen be a wp-adminba, és lépjen a **Beállítások → CLP Lakk gyorsítótár** menüpontra.
2. Kattintson a **Purge Entire Cache** gombra.

## Tisztítás a terminálról

A gyorsítótár bejegyzéseit manuálisan is törölheti a "curl" paranccsal.

---

### Megjegyzések

* A "Host" fejléc azért szükséges, mert a Varnish minden domainnévhez külön tárolja a gyorsítótárat.
* Az "X-Purge-Method: regex" fejléc lehetővé teszi több URL törlését mintaegyeztetés segítségével.
* A törlési kérelmeknek a localhosttól kell származniuk (magától a webhelytől vagy php/webszerver tárolóktól).
