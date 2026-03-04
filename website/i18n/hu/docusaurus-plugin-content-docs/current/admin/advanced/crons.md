---
sidebar_position: 4
---

# Cron állások

**OpenAdmin > Speciális > System Cron Jobs** lehetővé teszi a rendszergazdák számára, hogy megtekintsék az OpenPanel által használt ütemezett cron-feladatokat, módosítsák azok ütemezését, vagy engedélyezzék/letiltsák a naplózást az `/etc/openpanel/openadmin/cron.log` fájlba.

![screenshot](/img/admin/openadmin_cronjobs.png)


Az ütemezések megváltoztatása azt eredményezheti, hogy egyes OpenPanel funkciók nem működnek megfelelően. Csak akkor módosítsa ezeket a beállításokat, ha:

- Korlátozott erőforrásokkal rendelkező szervereken kell finomhangolni a végrehajtást, ill
- Az OpenPanel támogatása ezt tanácsolta Önnek.

Ezek a cron-feladatok elengedhetetlenek az OpenPanel belső műveleteihez.

Ha saját egyéni cron-feladatait szeretné hozzáadni, biztonságosabb, ha közvetlenül a root felhasználó crontabjában adja hozzá őket.
