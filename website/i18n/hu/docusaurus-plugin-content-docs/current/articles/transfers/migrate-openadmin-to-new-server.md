# Szerver áttelepítése

Az OpenPanel egy valóban operációs rendszer-agnosztikus hosting panel, ami azt jelenti, hogy zökkenőmentesen fut bármely Linux disztribúción. Ez egyszerűvé és hatékonysá teszi a panel és az összes felhasználói adat migrálását – egyik szerverről a másikra.

Az összes adat áttelepítése egy szerverről egy másikra:

## Készítse elő az új szervert

Az új szervernek tiszta környezettel kell rendelkeznie, **csak OpenPanel telepítve** – meglévő felhasználók vagy további konfigurációk nélkül.

Az OpenPanel telepítése a következő futtatással:

```bash
bash <(curl -sSL https://openpanel.org)
```

Várja meg, amíg a telepítés befejeződik, majd folytassa a forráskiszolgálóval.

## Migráció a régi kiszolgálóról

Az áttelepítés végrehajtható az OpenAdmin felhasználói felületén vagy a terminálról:

### Az OpenAdmin használatával

A régi szerveren jelentkezzen be az **OpenAdmin** oldalra, navigáljon a *Speciális > Migráció* oldalra, adja meg az új szerver IP-címét és gyökér hitelesítő adatait:

Kattintson a „Migráció indítása” gombra, és várja meg, amíg a folyamat befejeződik:

### Terminál használata

```bash
opencli server-migrate -h NEW_SERVER_IP --user root --password NEW_SERVER_ROOT_PASSWORD
```

Hagyja, hogy az áttelepítési folyamat befejeződjön.

## Ellenőrizze a migrációt

A befejezés után jelentkezzen be az OpenAdmin panelre az új kiszolgálón a régi kiszolgálón használt hitelesítési adatokkal.

Ellenőrizze, hogy minden szolgáltatás megfelelően indult-e el az OpenAdminban a Szolgáltatás állapota alatt, és teszteljen néhány felhasználói webhelyet, hogy megbizonyosodjon arról, hogy megfelelően működnek.

Ha minden rendben van, folytassa a tartományok DNS-rekordjainak frissítésével, hogy az új kiszolgálóra mutasson.

Ha valami nem fut, ellenőrizze az áttelepítési naplót a forráskiszolgálón, hogy nem tartalmaz-e hibákat: `/tmp/server_migrate.log`
