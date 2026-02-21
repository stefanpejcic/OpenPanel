# RSPAMD GUI

Az [Rspamd](https://rspamd.com/) egy fejlett spamszűrő rendszer, amely számos különféle módot kínál az üzenetek szűrésére, beleértve a reguláris kifejezéseket és a statisztikai elemzést. Az rspamd minden üzenetet elemz, és spam-pontszámot kap.

![rspamd gui](https://i.postimg.cc/wHkc2j15/2025-08-11-17-56-1.png)

Az [OpenPanel Enterprise Edition](https://openpanel.com/enterprise/#compare) alapértelmezés szerint tartalmazza az RSPAMD-t.

## Az Rspamd engedélyezése

Az RSPAMD engedélyezése:

1. Nyissa meg az **OpenAdmin > E-mailek > Beállítások** menüpontot.
2. Engedélyezze az 'RSPAMD' opciót.
![rspamd enable openadmin](https://i.postimg.cc/M6gTbRyZ/2025-08-11-18-04.png)
3. Mentse el a változtatásokat, és indítsa újra az e-mail szervert.
![rspamd restart openadmin](https://i.postimg.cc/6KPyXsDs/2025-08-11-18-05.png)

> **MEGJEGYZÉS**: Ha az Rspamd engedélyezve van, akkor kapcsolja ki a következő szolgáltatásokat (mivel átfedő funkciókat biztosítanak): *Amavis, SpamAssassin, OpenDKIM, OpenDMARC*.

### Jelszó beállítása az Rspamd GUI-hoz

A bejelentkezés előtt be kell állítania a vezérlő jelszavát.

1. Nyisson meg egy terminált, és futtassa:

``` bash
docker exec -it openadmin_mailserver rspamadm pw
   ```

* Adja meg új jelszavát.
* Kivonatos értéket fog kapni, például:

     ```
$2$mu7yqw9bn9heied5aeh8utec173umcub$oesyhcdpayqob6emzctn76c3dfrr1ipsi1hmht4a9sm7ytui8wjy
     ```

2. Adja meg a levelezőszerver-tárolót:

``` bash
docker exec -it openadmin_mailserver sh
   ```

3. Adja hozzá a kivonatolt jelszót a konfigurációhoz:

``` bash
echo 'password = "HASH_HERE";' >> /etc/rspamd/local.d/worker-controller.inc
   ```

4. Lépjen ki a tárolóból, és indítsa újra a levelezőszervert:

``` bash
kijárat
docker indítsa újra az openadmin_mailserver-t
   ```

## Az RSPAMD GUI elérése

Az Rspamd tartalmaz egy webalapú grafikus felhasználói felületet, amely statisztikákat és szűrési adatokat jelenít meg.

Az interfész az "11334" porton keresztül érhető el.

![rspamd bejelentkezés](https://i.postimg.cc/HWSMgjMf/2025-08-11-17-56.png)

1. Nyissa meg böngészőjét, és keresse fel a következőt:

   ```
http://<SZERVER_IP_CÍM>:11334
   ```
2. Írja be a beállított jelszót.
3. Mostantól hozzáférhet az Rspamd irányítópultjához.
