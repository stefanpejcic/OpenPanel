---
sidebar_position: 7
---

# Értesítések

Konfigurálja az OpenAdmin értesítéseit és e-mailes riasztási beállításait.

<Tabs>
<TabItem value="openadmin-notifications-view" label="OpenAdminnal" alapértelmezett>

Az aktuális értesítési beállítások megtekintéséhez vagy szerkesztéséhez lépjen az **OpenAdmin > Beállítások > Értesítések** oldalra, vagy kattintson a „Beállítások szerkesztése” gombra az értesítési oldalon.
  
![openadmin értesítési beállításai](/img/admin/openadmin_notifications_settings.png)

</TabItem>
<TabItem value="CLI-notifications-view" label="OpenCLI-vel">

Az értesítési beállítások megtekintéséhez futtassa:

```bash
opencli admin notifications get <OPTION>
```

Példa:

```bash
# opencli admin notifications get reboot
yes
```

Az értesítési beállítások frissítéséhez futtassa:

```bash
opencli admin notifications update <OPTION> <NEW-VALUE>
```

Példa:

```bash
opencli admin notifications update load 10
Updated load to 10
```

</TabItem>
</Tabs>

---

## E-mail

Állítsa be a rendszerértesítések és riasztások fogadására használt e-mail címet.

Írja be e-mail címét az **E-mail értesítésekért** mezőbe. Hagyja üresen az e-mailes figyelmeztetések letiltásához.

Ha az e-mail cím be van állítva, a napi használati jelentés elküldésre kerül a címre, az ütemezés a 'Speciális > System Cron Jobs' menüpontban konfigurálható.

![report example](/img/admin/daily_report.png)


---

## Szolgáltatások

Értesítéseket kaphat, ha a szolgáltatások leállnak vagy nem reagálnak. A szolgáltatásokat 5 percenként ellenőrzik.

- **OpenPanel:** Értesítés, ha az OpenPanel UI meghibásodik.
- **OpenAdmin:** Értesítés, ha az OpenAdmin UI meghibásodik.
- **Caddy:** Értesítés, ha a webszerver nem válaszol.
- **MySQL:** Értesítés, ha az adatbázis nem érhető el.
- **Docker:** Értesítés, ha a Docker szolgáltatás nem működik.
- **BIND9:** Értesítés, ha a DNS szolgáltatás nem működik vagy nem válaszol.
- **Sentinel Firewall:** Értesítés, ha a Sentinel (CSF) le van tiltva.

---

## Erőforráshasználat

Értesítést kaphat, ha az erőforrás-használat meghaladja a küszöbértékeket (5 percenként ellenőrzik):

* Átlagos terhelés
* CPU %
*RAM %
* Lemezhasználat %
* SWAP %

---

## Műveletek

Értesítések fogadása konkrét műveletek esetén:

* A szerver újraindult
* A webhely támadás alatt áll
* A felhasználó eléri a terv korlátját
* Az OpenAdmin új IP-címről érhető el
* Új OpenPanel frissítés elérhető

---

## SMTP beállítások

Alapértelmezés szerint az e-mailes riasztások a `noreply@openpanel.com' címről kerülnek elküldésre.

Ha saját SMTP-kiszolgálóját szeretné használni az e-mailek kézbesítéséhez, konfigurálja a következőket:

<Tabs>
<TabItem value="openadmin-notifications-smtp" label="OpenAdminnal" alapértelmezett>
Állítsa be a hitelesítéshez használt szerverportot, TLS-t vagy SSL-t, felhasználónevet és jelszót.
</TabItem>
<TabItem value="CLI-notifications-smtp" label="OpenCLI-vel">

Konfigurálja az egyes értékeket az "opencli config update" parancsokkal, például:

```bash
opencli config update mail_server example.net
```

```bash
opencli config update mail_port 465
```

```bash
opencli config update mail_use_tls False
```

```bash
opencli config update mail_use_ssl True
```

```bash
opencli config update mail_username user@example.net
```

```bash
opencli config update mail_password strongpassword123
```

```bash
opencli config update mail_default_sender user@example.net
```

</TabItem>
</Tabs>
