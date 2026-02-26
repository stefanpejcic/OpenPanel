# Hogyan folyik a webes forgalom a felhasználói tárolókkal

A Caddy az összes bejövő webforgalom **belépési pontja**.
Az első tartomány hozzáadásakor a Caddy automatikusan elindul, és elkezdi a figyelést a **80** (HTTP) és **443** (HTTPS) portokon.

Caddy a következőkért felelős:

* **SSL-tanúsítványok és megújítások**
* **Átirányítások (HTTP → HTTPS)**
* **Webes alkalmazások tűzfala (CorazaWAF)**
* **Fordított proxyzás a felhasználói webszerverekhez**

---

## Felhasználónkénti webszerverek

Minden felhasználó saját, elszigetelt **webszerver-példányát** futtatja egy egyedi helyi porton.
A támogatott webszerverek a következők:

**Nginx**
**Apache**
**OpenResty**
**OpenLiteSpeed**
**lakk**

A Caddy a domain konfigurációja alapján továbbítja a kéréseket a megfelelő felhasználó webszerverére.

---

## Alkalmazástárolók

A felhasználó webszervere ezután **alkalmazástárolókba** proxy kéréseket küld, mint például:

* **PHP-FPM (több verzió támogatott)**
**Node.js**
* **Python/Django/Flask**
* **Dokker konténerek**

Példafolyamat egy PHP-alkalmazáshoz:

```
Client → Caddy → Nginx → PHP-FPM container
```

Példafolyamat egy Node.js alkalmazáshoz:

```
Client → Caddy → Nginx → Node.js container
```

---

## Több PHP verzió kezelése

Minden felhasználó saját **PHP-FPM tárolóverziót** futtat, amely biztosítja a kompatibilitást a különböző keretrendszerekkel vagy régi alkalmazásokkal. Több PHP verzió (7.4, 8.0, 8.1, 8.2, 8.3 stb.) biztonságosan együtt létezhet egy felhasználó számáraL

Példa:

```
Site A
Client → Caddy → Nginx → PHP-FPM-8.2

Site B
Client → Caddy → Nginx → PHP-FPM-7.4

Site C
Client → Caddy → Nginx → PHP-FPM-8.3
```
---

## Lakk használata (opcionális)

Ha a felhasználó engedélyezi a **Lakk gyorsítótárazást**, az a Caddy és a webszerver között helyezkedik el:

```
Client → Caddy → Varnish → Webserver → php container
```

Ez lehetővé teszi a gyorsítótárazást és a teljesítmény gyorsítását a háttérprogram elérése előtt.

---

## Példafolyamatok

### PHP (Nginx + PHP-FPM-8.2)

```
Client → Caddy          (SSL, WAF, redirects)  
       → Nginx          (per-user webserver)  
       → PHP-FPM-8.2    (specific PHP version)
```

### PHP (lakkkal + PHP-FPM-8.1)

```
Client → Caddy          (SSL, WAF, redirects)
       → Varnish        (caching)  
       → Nginx          (per-user webserver)  
       → PHP-FPM-8.1    (specific PHP version)
```

### Node.js

```
Client → Caddy          (SSL, WAF, redirects)
       → Nginx          (per-user webserver)  
       → Node.js        (user’s container)
```

### OpenLiteSpeed ​​(beépített LSPHP-vel)

```
Client → Caddy          (SSL, WAF, redirects)  
       → OpenLiteSpeed  (VHost + lsphp runtime)
```

---

✅ **Röviden:**
**A Caddy mindig az első belépési pont** → a forgalmat a felhasználó **webszerverére** irányítja, → amely a kéréseket a felhasználó **alkalmazástárolójába (PHP verzió, Node.js stb.)** proxyzza.
