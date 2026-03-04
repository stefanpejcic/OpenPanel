# Az OpenPanel telepítése

Mielőtt elkezdené, győződjön meg arról, hogy szervere megfelel [a minimális követelményeknek] (/docs/admin/intro/#requirements), és támogatott disztribúciót futtat.

Az OpenPanel telepítése a szerverre:

1. Jelentkezzen be az új szerverére;
- rootként SSH-n keresztül vagy
- mint felhasználó sudo jogosultságokkal és írja be a "sudo -i" parancsot
2. Másolja és illessze be az openpanel telepítési parancsot a terminálba
```shell
bash <(curl -sSL https://openpanel.org)
```

A telepítő szkript támogatja az [opcionális zászlókat] (/install), amelyek segítségével konfigurálható az openpanel, kihagyható bizonyos telepítési lépések, vagy egyszerűen megjeleníthetők a hibakeresési információk.

Ha bármilyen hibát észlelt a telepítőszkript futtatása közben, kérjük, másolja és illessze be a telepítési naplófájlt [a közösségi fórumokba] (https://community.openpanel.org).
