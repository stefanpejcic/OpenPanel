# Apache, Nginx, OpenResty, OpenLiteSpeed ​​és Varnish

Az Nginx, Apache, OpenResty, OpenLiteSpeed ​​és Varnish konfigurálása felhasználónként az OpenPanelben.

Felhasználó létrehozásakor beállíthatja a használandó webszerver típusát:

- Apache
- Apache + lakk
- Nginx
- Nginx + lakk
- OpenResty
- OpenResty + Lakk
- OpenLitespeed
- OpenLitespeed + Lakk


**Webszerverek összehasonlítása**:

| Funkció / Webszerver | Apache | Nginx | OpenResty | OpenLiteSpeed ​​|
| --------------------------------- | ---------------------------------- | ----------------------------------------- | ----------------------------------------- | ------------------------------------------------- |
| **Lakk támogatás** | Igen | Igen | Igen | Igen |
| **.htaccess támogatás** | Teljes | Nincs (VHost fájlok szerkesztése szükséges) | Nincs (VHost fájlok szerkesztése szükséges) | Teljes ([bizonyos korlátozásokkal](https://docs.openlitespeed.org/config/rewriterules/)) |
| **Több PHP-verzió** | Igen | Igen | Igen | Nem, csak 1 verzió |
| **Teljesítmény (statikus tartalom)** | Mérsékelt | Magas | Magas | Magas |
| **Teljesítmény (dinamikus PHP)** | Mérsékelt | Magas | Magas | Nagyon magas |
| **A felhasználó felülbírálásának egyszerűsége** | .htaccess vagy VHost Editor | Csak a VHost szerkesztőn keresztül | Csak a VHost szerkesztőn keresztül | .htaccess vagy VHost Editor |
| **Memória lábnyom** | Közepes (~50–100 MB folyamatonként) | Alacsony (~10–30 MB dolgozónként) | Alacsony (~15–40 MB dolgozónként) | Közepes (~30–60 MB folyamatonként) |
| **Dokkó mérete** | ~200 MB | ~120 MB | ~150 MB | ~1000 MB || **Ajánlott felhasználási esetek** | Régi alkalmazások, .htaccess-heavy webhelyek | Nagy teljesítményű statikus vagy dinamikus oldalak | Fejlett Lua szkriptelés, nagy egyidejűség | Nagy teljesítményű PHP hosting, WordPress, Laravel |

> **Megjegyzések:**
>
> * A memória/lemez alapterület értéke hozzávetőleges, és a konfigurációtól és a forgalomtól függ.
> * A lakk kiürítheti a statikus tartalmat, jelentősen csökkentve a webszerver terhelését.
> * A „docker” modullal a felhasználók bármikor válthatnak a webszerver típusa között a **Webszerver típusának módosítása** oldalon.
> * Az OpenLiteSpeed ​​általában a legjobb a .htaccess támogatást igénylő nagy teljesítményű PHP-alkalmazásokhoz.

---

A rendszergazdák az [*OpenAdmin > Settings > User Defaults*](/docs/admin/settings/defaults/) oldalon választhatják ki az alapértelmezett webszervert.

---

## Apache

Az [Apache](https://httpd.apache.org/) alapértelmezés szerint hozzárendelhető az összes új felhasználóhoz, vagy az egyes felhasználókhoz a fiók létrehozása során.

### Beállítás egy felhasználó számára

Az Apache beállítása egyetlen felhasználó számára a létrehozás során:

1. Nyissa meg az **Új felhasználó** űrlapot.
2. A **Speciális** részben válassza ki az **Apache** webszervert.

![apache for user](/img/guides/apache.gif)

> **Tipp:** Ha engedélyezni szeretné a Lakk funkciót ennél a felhasználónál, egyszerűen kapcsolja be a **Lakk gyorsítótárat** ugyanazon az oldalon.

### Beállítás alapértelmezettként

Az Apache alapértelmezetté tétele az összes újonnan létrehozott felhasználó számára:

* Nyissa meg az [**OpenAdmin > Settings > User Defaults**](/docs/admin/settings/defaults/) menüpontot, és válassza ki az **Apache**-t alapértelmezett webszerverként.


---

## Nginx

Az [Nginx](https://httpd.apache.org/) alapértelmezés szerint hozzárendelhető az összes új felhasználóhoz, vagy az egyes felhasználókhoz a fiók létrehozása során.

### Beállítás egy felhasználó számára

1. Nyissa meg az **Új felhasználó** űrlapot.
2. A **Speciális** alatt válassza ki az **Nginx** webszervert.

![nginx for user](/img/guides/nginx.gif)

> **Tipp:** Ha az **Nginx + Varnish** funkciót szeretné használni, engedélyezze a Varnish Cache ON funkciót.

### Beállítás alapértelmezettként

* Nyissa meg az [**OpenAdmin > Beállítások > Felhasználói alapértelmezések**] (/docs/admin/settings/defaults/) menüpontot, és válassza ki az **Nginx**-t alapértelmezett webszerverként.

---

## OpenResty

Az [OpenResty](https://openresty.org/) alapértelmezés szerint hozzárendelhető az összes új felhasználóhoz, vagy az egyes felhasználókhoz a fiók létrehozása során.


### Beállítás egy felhasználó számára

1. Nyissa meg az **Új felhasználó** űrlapot.
2. A **Speciális** részben válassza az **OpenResty** lehetőséget.

![openresty for user](/img/guides/openresty.gif)

> **Tipp:** Kapcsolja be a **Lakk gyorsítótárat** az **OpenResty + Lakk** funkcióhoz.

### Beállítás alapértelmezettként

* Nyissa meg az [**OpenAdmin > Settings > User Defaults**](/docs/admin/settings/defaults/) menüpontot, és válassza az **OpenResty** lehetőséget alapértelmezettként.

---

## OpenLiteSpeed

Az [OpenLiteSpeed](https://openlitespeed.org/) alapértelmezés szerint hozzárendelhető az összes új felhasználóhoz, vagy az egyes felhasználókhoz a fiók létrehozása során.

### Beállítás egy felhasználó számára

1. Nyissa meg az **Új felhasználó** űrlapot.
2. A **Speciális** részben válassza az **OpenLiteSpeed** lehetőséget.

![openlitespeed for user](/img/guides/openlitespeed.gif)

> **Tipp:** Az OpenLiteSpeed ​​és a Varnish kombinálásához engedélyezze a **Varnish Cache ON** beállítást.

### Beállítás alapértelmezettként

* Nyissa meg az [**OpenAdmin > Beállítások > Felhasználói alapértelmezések**] (/docs/admin/settings/defaults/) menüpontot, és válassza ki az **OpenLiteSpeed** lehetőséget alapértelmezettként.

