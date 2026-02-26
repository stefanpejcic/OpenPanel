# Cloudflare Tunnel + OpenPanel

A Cloudflare Tunnel lehetővé teszi a szolgáltatások (webhelyek, API-k vagy belső eszközök) biztonságos elérhetővé tételét az internet számára **a tűzfalportok megnyitása** vagy a szerver IP-címének felfedése nélkül.

Ez az útmutató a Cloudflare Tunnel OpenPanel szerverhez való konfigurálását ismerteti.

> **MEGJEGYZÉS**: Ezt a cikket a mesterséges intelligencia hozta létre, és ellenőrzést igényelhet. Kattintson az oldal alján található linkre a szerkesztéshez és a hozzájáruláshoz.

---

## 1. Adja hozzá a "cloudflared" szolgáltatást

Szerkessze a meglévő `/root/docker-compose.yml` fájlt, és adja hozzá:

```yaml
  cloudflared:
    image: cloudflare/cloudflared:latest
    restart: unless-stopped
    command: tunnel --config /etc/cloudflared/config.yml run
    volumes:
      - ./cloudflared:/etc/cloudflared
    network_mode: host
```

Hozza létre a Cloudflared mappát:

```bash
mkdir -p /root/cloudflared
```

Hozd létre a `/root/cloudflared/config.yml' fájlt:

```yaml
tunnel: my-openpanel-tunnel
credentials-file: /etc/cloudflared/<TUNNEL-ID>.json

ingress:
  - hostname: site1.example.com
    service: http://localhost
  - hostname: site2.example.com
    service: http://localhost
```

---

## 2. Telepítés és bejelentkezés (egyszeri)

Futtassa a következőket a Cloudflare hitelesítéséhez:

```bash
docker run -it --rm \
  -v /root/cloudflared:/etc/cloudflared \
  cloudflare/cloudflared:latest tunnel login
```

Nyissa meg a hivatkozást a böngészőjében, és jelentkezzen be Cloudflare-fiókjával.

---

## 3. Hozzon létre egy alagutat

```bash
docker run -it --rm \
  -v /root/cloudflared:/etc/cloudflared \
  cloudflare/cloudflared:latest tunnel create my-openpanel-tunnel
```

Ez létrehozza az `<TUNNEL-ID>.json` fájlt a `/root/cloudflared` fájlban.

---

## 4. Frissítse a `config.yml` fájlt

* Cserélje ki az `<TUNNEL-ID>` elemet a tényleges azonosítóval.
* Frissítse a "szolgáltatás" URL-címét, hogy megfeleljen a belső webhelycímeknek.

---

## 5. Konfigurálja a DNS-t a Cloudflare-ben

Minden webhelyhez (`site1.example.com`, `site2.example.com`):

* Írja be: "CNAME".
* Érték: `<TUNNEL-ID>.cfargotunnel.com`
* Proxy állapota: **Proxy** (narancssárga felhő)

---

## 6. Indítsa el az alagutat

```bash
cd /root && docker compose up -d cloudflared
```

---

## 7. Közvetlen hozzáférés blokkolása

A tűzfalban blokkoljon minden bejövő HTTP/HTTPS forgalmat (és a "2083" és "2087" portokat) **kivéve** Cloudflare IP-tartományok:

* [IPv4-tartományok](https://www.cloudflare.com/ips-v4)
* [IPv6-tartományok](https://www.cloudflare.com/ips-v6)

## Ha még nem hozott létre alagutat:
## 8. Hozzon létre új alagutat az irányítópulton
* Jelentkezzen be a Cloudflare Dashboardba
* Nyissa meg a [https://dash.cloudflare.com](https://dash.cloudflare.com) webhelyet
* Jelentkezzen be fiókjával
* Lépjen a Zero Trust → Networks → Tunnels menüpontra
* Kattintson a Create alagút gombra
* Csatlakozótípusként válassza a Cloudflaredet
* Adjon nevet (pl. "my-openpanel-tunnel")
* Kattintson az Alagút mentése gombra
* A következő képernyőn megjelenik az Ön alagútazonosítója, és egy letölthető hitelesítő adatfájl
* Töltse le a JSON hitelesítő adatfájlt – a fájl neve ```<TUNNEL_ID>.json```
