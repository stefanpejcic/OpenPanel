# A HSTS engedélyezése egy tartományon az OpenPanelben

A HSTS ([HTTP Strict Transport Security](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Strict-Transport-Security)) engedélyezhető globálisan az összes tartományhoz, vagy egyedileg az OpenPanel egyes tartományaihoz.

---

## HSTS engedélyezése egy adott tartományhoz

### Végfelhasználóként

Ha Ön OpenPanel felhasználó:

1. Győződjön meg arról, hogy rendelkezik hozzáféréssel az **edit_vhost** funkcióhoz. Ha engedélyezve van, a **Domainek** részben a **Edit VirtualHosts** felirat látható.
2. Kattintson az **Edit VirtualHosts** elemre, és válassza ki a konfigurálni kívánt tartományt.
3. A szerkesztőben adja hozzá a HSTS fejlécet **csak a HTTPS szakaszban**.
4. Használja a megfelelő szintaxist a webszervertől függően (Nginx, Apache, OpenLiteSpeed ​​stb.).
5. Mentse el a változtatásokat.

Példa konfigurációkra:

#### OpenLiteSpeed

Adja hozzá ezt a HTTPS-figyelő Átírás/fejlécek szakaszához:
```bash
Header always set Strict-Transport-Security "max-age=2592000; includeSubDomains; preload"
```

#### Apache
Adja hozzá a következőt a blokkon belül:
```bash
<IfModule mod_headers.c>
    Header always set Strict-Transport-Security "max-age=2592000; includeSubDomains; preload"
</IfModule>
```

#### Nginx
Adja hozzá ezt a szerver blokkba:
```bash
add_header Strict-Transport-Security "max-age=2592000; includeSubDomains; preload" always;
```


---

### Rendszergazdaként

Ha OpenAdmin vagy root SSH hozzáféréssel rendelkezik:

1. Nyissa meg a tartomány Caddy vhost konfigurációs fájlját:

```bash
/etc/openpanel/caddy/domains/DOMAIN_NAME.TLD.conf
```

2. Adja hozzá a következő HSTS-fejlécet **közvetlenül a `tls {` sor** elé:

```bash
# HSTS
header {
  Strict-Transport-Security "max-age=2592000; includeSubDomains; preload"
}
```

3. Mentse el a fájlt, és töltse be újra a Caddy-t.

---

## A HSTS engedélyezése minden tartományhoz

A HSTS automatikus alkalmazása a szerver összes tartományára:

1. Szerkessze az alapértelmezett tartománysablonokat:

```bash
/etc/openpanel/caddy/templates/domain.conf_with_modsec
/etc/openpanel/caddy/templates/domain.conf
```

2. Adja hozzá a HSTS fejlécet **közvetlenül a `tls {` sor** elé minden sablonban:

```bash
# HSTS
header {
  Strict-Transport-Security "max-age=2592000; includeSubDomains; preload"
}
```

3. Mentse el a sablonokat. Minden később létrehozott új tartomány örökli ezt a HSTS-konfigurációt.
