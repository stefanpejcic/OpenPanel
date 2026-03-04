# A Caddy webszerver hibaelhárítása

Az OpenPanel Caddy webszervert használ az SSL kezeléséhez, és fordított proxyt a felhasználónkénti webszerverekhez.

Ha a Caddy le van tiltva, a szerveren tárolt összes webhely nem működik.

A Caddyvel kapcsolatos problémák hibaelhárítása:

## 1. Ellenőrizze, hogy a Caddy fut-e

Fut:

```bash
curl -I localhost
```

Látnia kell a "Server: Caddy" szöveget a válaszban, például:

```
root@demo:~# curl -I localhost
HTTP/1.1 308 Permanent Redirect
Connection: close
Location: https://localhost/
Server: Caddy
Date: Thu, 23 Oct 2025 13:11:03 GMT
```

Ha megjelenik, a **Caddy megfelelően működik**. Ha egy adott webhely vagy domain nem működik, külön-külön végezzen hibaelhárítást a domainnel.

---

## 2. Ellenőrizze a Caddy állapotát

Ellenőrizze, hogy Caddy „aktív” vagy „fut”:

```bash
docker ps -a
```

Ha az állapot "kilépett", "újraindítás" vagy "indítás", az a tárolóval kapcsolatos problémát jelez.

---

## 3. Érvényesítse a konfigurációt

Ha a Caddy szintaktikai hibát észlel a fő konfigurációs fájlban vagy bármely tartománykonfigurációban, előfordulhat, hogy indításkor kilép, és ismételten újraindul.

A konfiguráció érvényesítéséhez:

```bash
docker exec caddy caddy validate --config /etc/caddy/Caddyfile
```

* A "warn" vagy "info" üzenetek rendben vannak.
* Minden "hiba" üzenetet meg kell oldani, mielőtt a Caddy megfelelően elindulna.

A Caddy fő konfigurációs fájlja: `/etc/openpanel/caddy/Caddyfile`, a tartományonkénti fájlok pedig a következő helyen találhatók: `/etc/openpanel/caddy/domains/*.conf`

---

## 4. Növelje az erőforrások korlátait

Alapértelmezés szerint a Caddy 1 CPU magra és 1 GB RAM-ra korlátozódik. A limitek növeléséhez:

1. Szerkessze az ".env" fájlt:

```bash
nano /root/.env
```

2. Állítsa be az értékeket:

```
CADDY_CPU="1.0"
CADDY_RAM="1.0G"
```

3. Mentse el a fájlt, és építse újra a tárolót:

```bash
cd /root
docker compose down caddy
docker compose up -d caddy
```

---

## 5. Naplók megtekintése

Ha a fenti lépések nem oldják meg a problémát, ellenőrizze a naplókat:

```bash
docker logs -f caddy
```

Másold ki a hibákat, és keress online, vagy tedd közzé a [Discord-csatornánkon] (https://discord.openpanel.com/) vagy az [OpenPanel fórumain] (https://community.openpanel.org/t/openadmin).
