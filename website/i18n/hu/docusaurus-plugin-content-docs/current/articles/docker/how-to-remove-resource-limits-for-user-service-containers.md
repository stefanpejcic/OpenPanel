# Korlátlan CPU és RAM

A korlátlan CPU- és RAM-erőforrások beállítása lehetővé teszi az OpenPanel felhasználói fiók számára, hogy a kiszolgálón elérhető összes CPU-magot és fizikai memóriát felhasználja.

**Ez határozottan ellenjavallt**, még egyfelhasználós VPS-en sem, mivel egy csaló folyamat vagy feltört webhely destabilizálhatja az egész rendszert. Fontos az erőforrás-korlátok beállítása annak biztosítása érdekében, hogy a rendszer elegendő kapacitással rendelkezzen az olyan alapvető szolgáltatásokhoz, mint a figyelés, a konténerek összehangolása, a DNS és egyebek.

Ha továbbra is korlátlan erőforrást szeretne engedélyezni, kövesse figyelmesen az alábbi lépéseket:

## Távolítsa el a korlátokat a felhasználói tervből

A korlátozások eltávolításához állítsa a CPU és a RAM korlátját 0-ra (nulla) a tervbeállításokban:

![userlimits.png](/img/panel/v2/userlimits-guide.png)

----

## Távolítsa el a Docker erőforráskorlátait


### Meglévő felhasználó számára

Futtassa ezt a parancsot a CPU-, memória- és PID-korlátok eltávolításához a felhasználó Docker Compose fájljában meghatározott összes szolgáltatásból. Cserélje le a „dorotea” szót a tényleges felhasználónévvel:

```bash
USERNAME=dorotea && \
sed -i '/deploy:/,/^[^[:space:]]/{
  /deploy:/d
  /resources:/d
  /limits:/d
  /cpus:/d
  /memory:/d
  /pids:/d
}' /home/$USERNAME/docker-compose.yml
```

Ha a felhasználónak már vannak futó szolgáltatásai, állítsa le, majd indítsa újra azokat a módosítások alkalmazásához (a „dorotea” szót cserélje ki a felhasználónévre):

```bash
USERNAME=dorotea && \
cd /home/$USERNAME && \
running_services=$(docker --context=$USERNAME compose ps --services --filter "status=running") && \
docker --context=$USERNAME compose down && \
docker --context=$USERNAME compose up -d $running_services
```


### Minden új felhasználó számára

Annak érdekében, hogy minden új felhasználó alapértelmezés szerint korlátlan CPU-val és RAM-mal rendelkezzen, alkalmazza ugyanazokat a `sed' parancsokat az új fiókokhoz használt Docker Compose-sablonra:

```bash
sed -i '/deploy:/,/^[^[:space:]]/{
  /deploy:/d
  /resources:/d
  /limits:/d
  /cpus:/d
  /memory:/d
  /pids:/d
}' /etc/openpanel/docker/compose/1.0/docker-compose.yml
```

---


A változtatások alkalmazását követően minden új vagy újraindított szolgáltatás fel tudja használni az összes rendelkezésre álló rendszererőforrást. A vezérlőpult figyelmeztetést jelenít meg a felhasználók számára, jelezve, hogy a CPU és a RAM korlátai le vannak tiltva:

![userlimits3.png](/img/panel/v2/userlimits-guide3.png)

