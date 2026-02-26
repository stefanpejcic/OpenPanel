---
sidebar_position: 1
---

# Cron Jobs

A cron job egy Linux-parancs, amellyel a feladatokat ütemezheti a jövőbeni végrehajtáshoz. Lehetővé teszi az ismétlődő feladatok automatizálását, például értesítések küldését vagy szkriptek meghatározott időközönkénti futtatását.

![cronjobs.png](/img/panel/v2/cronjobsmain.png)

A CronJobs oldalon megtekintheti az aktuálisan ütemezett feladatokat, újakat hozhat létre, szerkesztheti vagy törölheti azokat.

:::info
Az Időzóna beállítás hasznos az ütemezett [cronjobs] (/docs/panel/advanced/cronjobs) helyi időzónában történő futtatásához.
:::


## Hozzáadás

Új cronjob létrehozásához kattintson a 'CronJob létrehozása' gombra, és az új űrlapon állítsa be a végrehajtandó szkriptet, válasszon egy tárolót a szkript végrehajtásához és a kívánt ütemezést.

![cronjobs_new.png](/img/panel/v2/cronjobs.png)

Az első mezőben kiválaszthatja azt a tárolót, amelyen a szkript futni fog.

![cronjobs_container.png](/img/panel/v2/cronjobs_container.png)

A második mező lehetővé teszi egy előre meghatározott (közös) ütemezés beállítását:

- 30-onként
- Minden percben
- 5 percenként
- 30 percenként
- Óránként
- Naponta
- Hetente
- Havonta
- Évente

![cronjobs_new_predefined.png](/img/panel/v2/cronjobs_common.png)

beállíthat egy szabványos cron kifejezést is, amely az időkészletet reprezentálja, **6** szóközzel elválasztott mezők használatával:

| Mezőnév | Kötelező? | Engedélyezett értékek | Engedélyezett speciális karakterek |
| ------------ | ---------- | ------------------ | -------------------------- |
| Másodpercek | Igen | 0-59 | `*` `/` `,` `-` |
| Jegyzőkönyv | Igen | 0-59 | `*` `/` `,` `-` |
| Óra | Igen | 0-23 | `*` `/` `,` `-` |
| A hónap napja | Igen | 1-31 | `*` `/` `,` `-` `?` |
| Hónap | Igen | 1-12. vagy JAN-DEC. | `*` `/` `,` `-` |
| A hét napja | Igen | 0-6 vagy V-SZO | `*` `/` `,` `-` `?` |

:::info
A szabványos Unix cronban található szokásos 5 mező helyett 6 mező van. Ennek az az oka, hogy az OpenPanel cron jobok is támogatják a másodpercenkénti ütemezést.
:::

További információért lásd: [CRON_Expression_Format](https://pkg.go.dev/github.com/robfig/cron#hdr-CRON_Expression_Format)

## Szerkesztés

Meglévő cronjob szerkesztéséhez kattintson a mellette lévő "Szerkesztés" gombra. Ez a művelet lehetővé teszi az adott cron feladat szerkesztését.

![cronjobs_edit.gif](/img/panel/v2/cron_edit_v2.gif)

A szkript végrehajtásának ütemezésének módosításához használhat egy eszközt, például a https://crontab.guru/.

Ha végzett, kattintson a "Mentés" gombra, hogy frissítse a crontab fájlt a módosításokkal.

## Törlés

Egy cronjob törléséhez kattintson a mellette lévő „Törlés” gombra. Ez a művelet megnyit egy modált, amely a törlés megerősítését kéri.

![cronjobs_delete.gif](/img/panel/v2/cron_delete.gif)

## Naplók

Minden cron-feladat-végrehajtás JSON-formátumban kerül rögzítésre.

A naplók megtekintéséhez kattintson a *„Naplók megtekintése”* gombra. Ez megnyit egy új lapot, amely megjeleníti az összes cron munkanaplót.

Ha egy adott feladatnév (megjegyzés) alapján szeretné szűrni a naplókat, fűzze hozzá a következő paramétert az URL-hez:
`?job=`, majd a munka neve. Példa: `/cronjobs/log?job=whmcs-cron`

## Fájlszerkesztő

A *Fájlszerkesztő* opció lehetővé teszi a cronok tárolására szolgáló fájl szerkesztését. A fájl formátuma:

```
[job-exec "JOB_NAME"]
schedule = 
container = 
command =
```

Példa:
```
[job-exec "example"]
schedule = @daily
container = nginx
command = curl https://ip.openpanel.com
```

Egy másik példa:
```
[job-exec "whmcs-cron"]
schedule = @every 5m
container = php-fpm-8.4
command = php -q /var/www/html/whmcs-paths/crons/cron.php
```

:::info
A cronjobhoz beállíthatja az átfedés nélküli beállítást is, hogy elkerülje a feladat többszöri párhuzamos futtatását, ha az előző végrehajtás nem fejeződött be.
:::


## Import / Export

A Cronjobok tömegesen szerkeszthetők egy fájlon keresztül, ami megkönnyíti több feladat egyidejű szerkesztését vagy átvitelét a szerverek között.
Egyszerűen kattintson a *„Váltás fájlszerkesztőre”* gombra a szerkesztő megnyitásához.

![fájlszerkesztő](https://i.postimg.cc/zXx0LDMm/slika.png)
