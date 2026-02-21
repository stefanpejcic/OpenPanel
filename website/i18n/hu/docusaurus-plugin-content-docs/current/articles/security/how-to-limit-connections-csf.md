# A CSF-vel való kapcsolatok korlátozása

A CSF lehetővé teszi, hogy korlátozza az egyidejű kapcsolatok számát a szerver adott portjaira.

Például beállíthatja a CSF-et úgy, hogy a 80-as portot legfeljebb 5 egyidejű kapcsolatra, a 443-as portot pedig 20 egyidejű kapcsolatra korlátozza, ha hozzáadja a következőket a konfigurációs fájlhoz:

```
CONNLIMIT = "80;5,443;20"
```

**A CONNLIMIT működése:**

A „CONNLIMIT” beállítás a „port;limit” párok vesszővel elválasztott listáját használja.

Például:
```
CONNLIMIT = "22;5,80;20"
```

eszközök:

1. Minden IP-címnek legfeljebb 5 egyidejű új kapcsolata lehet a 22-es porthoz.
2. Minden IP-címnek legfeljebb 20 egyidejű új kapcsolata lehet a 80-as porthoz.


