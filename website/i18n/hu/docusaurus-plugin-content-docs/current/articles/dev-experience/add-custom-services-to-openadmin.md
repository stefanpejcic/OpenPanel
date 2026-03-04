# Egyedi szolgáltatások

Egyéni szolgáltatások hozzáadhatók az OpenAdminhoz, amelyek az adminisztrációs panelről futtathatók és felügyelhetők.

Ebben a példában hozzáadjuk a [netdata](https://learn.netdata.cloud/docs/netdata-agent/installation/docker) dokkolóképet.

## Szolgáltatás hozzáadása

```
cd /root
nano docker-compose.yml
```

és add hozzá a netdata részt:

```
  netdata:
    image: netdata/netdata
    container_name: netdata
    pid: host
    network_mode: host
    restart: unless-stopped
    cap_add:
      - SYS_PTRACE
      - SYS_ADMIN
    security_opt:
      - apparmor:unconfined
    environment:
      - DO_NOT_TRACK=1
      - NETDATA_CLOUD_ENABLE=no
    volumes:
      - ./netdata/config:/etc/netdata
      - ./netdata/lib:/var/lib/netdata
      - ./netdata/cache:/var/cache/netdata
      - /:/host/root:ro,rslave
      - /etc/passwd:/host/etc/passwd:ro
      - /etc/group:/host/etc/group:ro
      - /etc/localtime:/etc/localtime:ro
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /etc/os-release:/host/etc/os-release:ro
      - /var/log:/host/var/log:ro
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - /run/dbus:/run/dbus:ro
```

mentse és lépjen ki a fájlból. majd indítsa el a szolgáltatást:

```
docker --context=default compose up -d netdata
```

most azt javaslom, hogy ahelyett, hogy megnyitná az internethez vezető portot, csak tegye az IP-címét a tűzfalon az engedélyezőlistára.

Ezután nyissa meg az IP:19999-et a böngészőjében:

![screenshot](https://i.postimg.cc/CFRC5zG9/netdata.png)


## Monitor szolgáltatás

Következő lépésként elérhetővé kell tenni a netdata szolgáltatást az **OpenAdmin > Services** oldalon, hogy megnézhessük az állapotát, elindíthassuk/leállíthassuk vagy újraindíthassuk.

Ehhez hozzá kell adnunk a netdata szolgáltatást az "/etc/openpanel/openadmin/config/services.json" fájlhoz:

```
    {
        "name": "Netdata",
        "type": "docker",
        "on_dashboard": false,
        "real_name": "netdata"
    },
```
A „valódi_név” a szolgáltatás/tároló neve, a „név” pedig az OpenAdmin szolgáltatásokban megjelenő, ember által olvasható név.

Mentse el a fájlt és frissítse az oldalt az OpenAdminban - a szolgáltatás azonnal látható, és a panelről tudjuk elindítani/leállítani.

![szolgáltatások](https://i.postimg.cc/qpTMgY8r/2025-04-30-16-33.png)

Akkor is kaphatunk értesítést, ha a szolgáltatás nem fut. Ehhez szerkessze a következő fájlt: "/etc/openpanel/openadmin/config/notifications.ini", és adja hozzá a szolgáltatás nevét ("netdata") a szolgáltatások alatt:

```
services=panel,admin,caddy,mysql,docker,csf,netdata
```

mentse a fájlt és lépjen ki. A Sentinel szolgáltatás mostantól azt is ellenőrzi, hogy a netdata tároló fut-e, és ha nem, akkor riasztást küld az **OpenAdmin > Értesítések** oldalra, és ha az e-mailek engedélyezve vannak, akkor az e-mailekre is.

