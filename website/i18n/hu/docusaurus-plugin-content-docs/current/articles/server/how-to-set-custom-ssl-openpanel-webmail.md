# Egyéni SSL-tanúsítványok beállítása az OpenPanel, az OpenAdmin és a Webmail számára

⚠️ Jelenleg egyéni SSL-tanúsítványok hozzáadása az OpenPanelhez és a Webmailhez csak a terminálról lehetséges.

- A Let's Encrypt tanúsítványok a következő helyen találhatók: `/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/`
- Az egyéni tanúsítványokat az `/etc/openpanel/caddy/ssl/custom/` mappában kell elhelyezni.

---

## Egyéni SSL az OpenAdmin és az OpenPanel számára

Kövesse az alábbi lépéseket egyéni SSL-tanúsítvány konfigurálásához az OpenAdmin és az OpenPanel felületekhez:

### 1. Adja hozzá a tanúsítványt

Hozzon létre egy "custom" könyvtárat az `/etc/openpanel/caddy/ssl/` mappában:

```bash
mkdir -p /etc/openpanel/caddy/ssl/custom/
```

Ezután hozzon létre egy könyvtárnevet, amely megegyezik a domainjével
```bash
mkdir -p /etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/YOUR_DOMAIN_HERE/
```

Töltse fel SSL fájljait ebbe a könyvtárba, és nevezze el őket:

* `$DOMAIN.crt`
* `$DOMAIN.kulcs`

**Példa:**

```
/etc/openpanel/caddy/ssl/custom/srv.openpanel.com/srv.openpanel.com.crt
/etc/openpanel/caddy/ssl/custom/srv.openpanel.com/srv.openpanel.com.key
```

### 2. Állítsa be a tartományt

```bash
opencli domain set <DOMAIN_NAME>
```

**Példa:**
```bash
opencli domain set srv.openpanel.com
```

### 3. A Caddy konfigurálása

Nyissa meg a **Caddyfile-t**:
```bash
nano /etc/openpanel/caddy/Caddyfile
```

Keresse meg a domainjének megfelelő szakaszt:

```
# START HOSTNAME DOMAIN #
example.net {
    reverse_proxy localhost:2087
}
```

Adja hozzá a "tls" sort a "reverse_proxy" után:

```
tls /etc/openpanel/caddy/ssl/custom/srv.openpanel.com/srv.openpanel.com.crt /etc/openpanel/caddy/ssl/custom/srv.openpanel.com/srv.openpanel.com.key
```

👉 Cserélje ki az `srv.openpanel.com` címet a tényleges domainjére.

Indítsa újra a Caddyt a tanúsítvány alkalmazásához:

```bash
docker restart caddy
```

---

## Egyéni SSL a webmailhez

### 1. Adja hozzá a tanúsítványt

Hozzon létre egy könyvtárat a domainhez:

```bash
mkdir -p /etc/openpanel/caddy/ssl/custom/YOUR_DOMAIN_HERE/
```

Töltse fel SSL fájljait, és nevezze el őket:

* `$DOMAIN.crt`
* `$DOMAIN.kulcs`

**Példa:**

```
/etc/openpanel/caddy/ssl/custom/webmail.openpanel.com/webmail.openpanel.com.crt
/etc/openpanel/caddy/ssl/custom/webmail.openpanel.com/webmail.openpanel.com.key
```

### 2. Rendelje hozzá a tartományt

```bash
opencli email-webmail domain <DOMAIN_NAME>
```

**Példa:**

```bash
opencli email-webmail domain webmail.openpanel.com
```

### 3. A Caddy konfigurálása

Nyissa meg a **Caddyfile-t**:

```bash
nano /etc/openpanel/caddy/Caddyfile
```

Keresse meg a Webmail részt:

```
# START WEBMAIL DOMAIN #
webmail.example.net {
    reverse_proxy localhost:8080
}
# END WEBMAIL DOMAIN #
```

Adja hozzá a "tls" sort a "reverse_proxy" után:

```
tls /etc/openpanel/caddy/ssl/custom/webmail.openpanel.com/webmail.openpanel.com.crt /etc/openpanel/caddy/ssl/custom/webmail.openpanel.com/webmail.openpanel.com.key
```

👉 Cserélje ki a `webmail.openpanel.com` címet a saját domainjére.

A változtatások alkalmazásához indítsa újra a Caddyt:

```bash
docker restart caddy
```

---

## Egyéni SSL domainekhez (végfelhasználók)

A végfelhasználók közvetlenül az **OpenPanel felületről** adhatják hozzá saját SSL-tanúsítványaikat, ha az ssl modul és a szolgáltatások engedélyezve vannak.

📖 Dokumentáció: [Custom SSL for Domains](https://openpanel.com/docs/panel/domains/ssl/#custom-ssl)
