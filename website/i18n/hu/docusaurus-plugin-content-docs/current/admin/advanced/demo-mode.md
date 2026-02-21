---
sidebar_position: 10
---

# Demo mód

Engedélyezze a Demo módot az OpenPanel és az OpenAdmin felületek csak olvasható módban történő zárolásához.

Ez a mód ideális azoknak a tárhelyszolgáltatóknak, akik biztonságos, nyilvános bemutatókörnyezetben szeretnék bemutatni az OpenPanel-t. A felhasználók felfedezhetik a felhasználói felületet, de módosításokat nem lehet végrehajtani – minden művelet le van tiltva mind az adminisztrációs, mind a felhasználói panelen.

Ha engedélyezve van, a Demo módot nem lehet kikapcsolni az adminisztrációs panelen keresztül.
A letiltásához futtassa a következő parancsot a terminálon:
```
opencli config update demo_mode off
```

Mielőtt bekapcsolná ezt a módot, győződjön meg arról, hogy konfigurálta a demótartalmat, és védi a szervert. 📘 [További információ](https://dev.openpanel.com/cli/config.html#Demo-mode)

