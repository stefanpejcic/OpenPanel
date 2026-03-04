---
sidebar_position: 2
---


# FTP-fiók

Az FTP-fiókok részben megtekintheti és kezelheti az OpenPanel-felhasználókhoz társított összes FTP-alfiókot.

Ezzel az eszközzel létrehozhat, áttekinthet vagy távolíthat el FTP-hozzáférést bizonyos könyvtárakhoz, így biztosítva a biztonságos és ellenőrzött fájlátvitelt.

A fiókok megfelelő működéséhez az FTP-szolgáltatásnak futnia kell.

A táblázat a következő részleteket tartalmazza:
- **Fiók** – Az FTP-fiók felhasználóneve.
- **Tulajdonos** – Az FTP-fiókot birtokló OpenPanel felhasználói fiók.
- **Path** – A fájlrendszer elérési útja, amelyhez az FTP-fiók hozzáfér.
- **Művelet** – Az egyes fiókok kezelésének lehetőségei (pl. hozzáférés törlése vagy szerkesztése).

Az FTP-alfiókok listája időszakonként frissül az [`opencli ftp-users` cronjob ütemtervének] (https://dev.openpanel.com/crons.html#ftp-users) megfelelően.
