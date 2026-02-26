# PHP beállítás webhelyenként (mappánként)

A `.user.ini` fájlok használatával különböző PHP-korlátokat állíthat be az egyes webhelyekhez vagy mappákhoz.

## `.user.ini` fájl létrehozása

Például, ha egyéni PHP-korlátokat szeretne beállítani csak egy adott webhelyhez, például `example.net`:

1. Nyissa meg az **OpenPanel > Fájlkezelő** elemet, és navigáljon a tartomány dokumentumgyökéréhez (az adott webhely fő mappájához).
2. Kattintson az **Új fájl** elemre, írja be a `.user.ini` nevet, jelölje be a **Megnyitás a fájlszerkesztőben létrehozás után** lehetőséget, majd kattintson a **Létrehozás** gombra.
3. A szerkesztőben adja hozzá a módosítani kívánt PHP konfigurációs beállításokat. Például:

   ```
feltöltési_maximális_fájlméret = 64M
post_max_size = 64M
maximális_végrehajtási_idő = 120
   ```
4. Mentse el a fájlt, és lépjen ki a szerkesztőből.

Végül nyissa meg webhelyét, és erősítse meg a változtatásokat:

* Bármely PHP webhely esetében ellenőrizheti a `phpinfo()` segítségével.
* WordPress esetén lépjen az **Eszközök > Webhely állapota** oldalra, és ellenőrizze, hogy az új korlátozások érvényesek-e.

---

> **MEGJEGYZÉS**: Mivel a `.user.ini`-t nyilvános könyvtárakból olvassuk be, a tartalma bárki számára ki lesz szolgáltatva, aki kéri, és potenciálisan bizalmas konfigurációs beállításokat jelenít meg. Blokkolja a hozzáférést az Apache .htaccess fájljával vagy a Vhost Editor for Nginx/OpenResty programmal.
