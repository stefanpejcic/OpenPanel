# Egyéni domain a phpMyAdmin számára

Az **OpenPanel**-ban minden felhasználó rendelkezik elszigetelt szolgáltatásokkal, beleértve egy dedikált **phpMyAdmin** példányt, amely egyedi porton keresztül érhető el.

Alapértelmezés szerint a felhasználó phpMyAdmin példánya a következő címen érhető el:

```
http://IP:PORT
```

> **Tipp:** A hozzárendelt cím az OpenPanel **MySQL > phpMyAdmin** alatt látható.

---

## Egyéni tartomány beállítása minden felhasználó számára

Konfigurálhat egyetlen egyéni tartományt az összes felhasználó phpMyAdmin példányához. A felhasználók ezután hozzáférhetnek a phpMyAdminhoz:

```
https://your-domain:PORT
```

**Példa:**

| Felhasználó | Cím |
| ------ | ------------------------------- |
| A felhasználó | https://phpmyadmin.my-domain.com:37746 |
| B felhasználó | https://phpmyadmin.my-domain.com:37752 |
| Felhasználó C | https://phpmyadmin.my-domain.com:37778 |

### Lépések

#### 1. Adja hozzá domainjét a Caddyben

Szerkessze a Caddy konfigurációs fájlját:

```bash
nano /etc/openpanel/caddy/Caddyfile
```

Adjon hozzá egy részt a phpMyAdmin domainhez (például: `phpmyadmin.my-domain.com`):

```caddy
# START PHPMYADMIN DOMAIN #
phpmyadmin.my-domain.com: {
    @withPort {
        host phpmyadmin.my-domain.com
    }
    reverse_proxy localhost:{http.request.port}
}
# END PHPMYADMIN DOMAIN #
```

![caddyfile](https://i.postimg.cc/G2K9YrBC/caddy-pma.png)

Ezután hozzon létre egy üres fájlt a Caddy számára az SSL automatikus kezeléséhez:

```bash
touch /etc/openpanel/caddy/domains/phpmyadmin.my-domain.com.conf
```

Indítsa újra a Caddy szolgáltatást:

```bash
docker restart caddy
```

---

#### 2. Állítsa be a `PMA_ABSOLUTE_URI`-t az új felhasználók számára

Szerkessze az ".env" fájlt új felhasználók számára:

```bash
nano /etc/openpanel/docker/compose/1.0/.env
```

A "# PHPMYADMIN" alatt adja hozzá:

```bash
PMA_ABSOLUTE_URI="http://phpmyadmin.my-domain.com:{PMA_PORT}/"
```

Mentse el a fájlt.

![env fájl](https://i.postimg.cc/RCjqHPF8/pma-env.png)

Mostantól minden új felhasználó számára elérhető lesz a phpMyAdmin az egyéni domainen keresztül. Ez az OpenPanel felhasználói felületén is megjelenik.

---

## Lehetővé teszi a felhasználók számára, hogy beállítsák saját egyéni domainjüket

A felhasználók saját tartományt választhatnak a phpMyAdmin számára az IP:PORT használata helyett.

**Példa:**

| Felhasználó | Cím |
| ------ | ---------------------------------- |
| A felhasználó | http://IP:PORT |
| B felhasználó | https://phpmyadmin.their-domain.com:37752 |
| Felhasználó C | http://phpmyadmin-awesome.net:37778 |

### Lépések

1. Engedélyezze az **edit_vhost** modult az **OpenAdmin > Beállítások > Modulok** részben:

![module](https://i.postimg.cc/c1DBGGFw/vhost-module.png)

és győződjön meg arról, hogy szerepel a tárhelycsomagok által használt szolgáltatáskészletekben:
![feature](https://i.postimg.cc/PJFWW1Lx/feature.png)

3. A felhasználók egy **Szerkesztés (ceruza)** gombot fognak látni a **MySQL > phpMyAdmin** oldalon.
![szerkesztés](https://i.postimg.cc/L4y3FFxz/user-pma.png)


5. Egyéni tartomány konfigurálásához a felhasználóknak:

1. Adja hozzá a domaint a **Domainek > Új** alatt.
2. Szerkessze a tartomány VirtualHostját, hogy a forgalmat a `phpmyadmin:80' címre proxyzza.
3. Állítsa be a tartományt a **MySQL > phpMyAdmin** pontban.
![szerkesztés](https://i.postimg.cc/gdpSRYRT/user-pma2.png)
4. Indítsa újra a phpMyAdmin szolgáltatást.

Ha elkészült, a phpMyAdmin elérhető a felhasználó által választott domainen keresztül.
