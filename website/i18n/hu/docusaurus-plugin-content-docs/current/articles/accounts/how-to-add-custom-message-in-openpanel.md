# Egyéni üzenet megjelenítése

Az OpenPanel támogatja mind a **felhasználónkénti**, mind a **globális** egyéni üzeneteket, amelyek megjeleníthetők a felhasználók OpenPanel felületén.

![példaüzenet](https://i.postimg.cc/QN0XSW0t/2025-08-14-14-49.png)

## Egyéni üzenet egy adott felhasználó számára

Egyéni üzenet megjelenítése az **OpenPanel > Irányítópult** oldalon egy adott felhasználó számára:

1. Lépjen az **OpenAdmin > Felhasználók > *felhasználónév* > Áttekintés** oldalra.

![üzenet hozzáadása](https://i.postimg.cc/KZBxwWJC/2025-08-14-14-50.png)

2. Írja be az üzenetet – egyszerű szöveges vagy HTML-formátumban – a **„Egyéni üzenet a felhasználó számára”** területen.

3. Kattintson a **Mentés** gombra.

Az üzenet azonnal megjelenik a felhasználó irányítópultján.

### Üzenet hozzáadása terminálon keresztül

Egyéni üzenetet is hozzáadhat egy fájl létrehozásával:

```bash
/etc/openpanel/openpanel/core/users/USERNAME/custom_message.html
```

Cserélje ki a „FELHASZNÁLÓNÉV” szót a tényleges felhasználó felhasználónevére.

---

## Globális üzenet minden felhasználó számára

Ha üzenetet szeretne megjeleníteni **minden felhasználó** számára, másolja ki vagy hozzon létre egy szimbolikus hivatkozást az egyéni üzenetfájlhoz az egyes felhasználók könyvtárában: `/etc/openpanel/openpanel/core/users/USERNAME/custom_message.html`.
