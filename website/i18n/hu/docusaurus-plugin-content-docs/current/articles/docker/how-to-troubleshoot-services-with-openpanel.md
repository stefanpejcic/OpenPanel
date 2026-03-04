# Az OpenPanel szolgáltatással kapcsolatos problémák elhárítása

Ha problémákba ütközik egy, az OpenPanel által kezelt szolgáltatással, és fiókjában engedélyezve van a Docker funkció, akkor ezzel újraindíthatja a szolgáltatástárolót, és további hibaelhárítás érdekében ellenőrizheti a naplókat.

Egyszerűen lépjen az Openpanel > Docker > Containers menüpontra, és próbálja meg leállítani/indítani a szolgáltatástárolót:

![containersGUI.png](https://i.postimg.cc/650sKYW3/containersgui.png)

A naplók ellenőrzéséhez lépjen az Openpanel > Docker > Logs menüpontra, és válasszon ki egy tárolót a legördülő menüből a naplók megtekintéséhez:

![containersLogsGUI.png](https://i.postimg.cc/Fzh4bqF1/containerlogsgui.png)

Ha a Docker funkció nincs engedélyezve a fiókjában, forduljon a rendszergazdához, hogy további ellenőrzéseket végezhessenek.

## A szolgáltatással kapcsolatos problémák hibaelhárítása rendszergazdaként

Jelentkezzen be az OpenAdminba a 2087-es porton, lépjen a Felhasználók oldalra, és kattintson egy felhasználónévre, és kezdje meg a fiók kezelését:

![containersAdmin1.png](https://i.postimg.cc/rpYjNt8p/containers-Admin1.png)

A felhasználókezelési oldalon lépjen a Szolgáltatások fülre, és engedélyezze/letiltja a szolgáltatást:

![containersAdmin2.png](https://i.postimg.cc/hv4G4hdV/containers-Admin2.png)

Ha a tároló újraindítása nem oldja meg a problémát, további hibaelhárítást kell végeznie, lépjen be a kiszolgálóhoz SSH-n keresztül, és vizsgálja meg a következő parancsokat:

`machinectl shell $felhasználónév@ /bin/bash -c 'systemctl --user status'` (a $username helyére az OpenPanel felhasználót írjon)

Megjelenik a felhasználó saját rendszeres szolgáltatáskezelőjének állapotáttekintése:

![containersAdminSSH1](https://i.postimg.cc/FKRzQY0S/containers-Admin-SSH1.png)

`machinectl shell $username@ /bin/bash -c 'journalctl --user -u $docker'` (a $username helyére az OpenPanel felhasználót írjuk)

Megjelennek a Docker szolgáltatásnaplók:

![containersAdminSSH2](https://i.postimg.cc/6qjwc0mJ/containers-Admin-SSH2.png)

Ha nem talál hibát vagy lehetséges okokat ezen parancsok kimenetén belül, akkor manuálisan indítsa el a szolgáltatástárolót, és nézze meg a naplókat a következő paranccsal:

`cd /home/$felhasználónév && docker --context=$felhasználónév összeállítása: SERVICE_NAME` (a $username helyére az OpenPanel felhasználót, a SERVICE_NAME helyére pedig annak a szolgáltatásnak a nevét, amelyet Ön a hibaelhárítás alatt tart)


