# Licenc átvitele új kiszolgálóra

Kövesse az alábbi lépéseket az **OpenPanel Enterprise Edition** licencének másik szerverre való átviteléhez:

1. Tiltsa le a licencet a jelenlegi kiszolgálón.
2. Adja ki újra a licencet [my.openpanel.com](https://my.openpanel.com) fiókjából.
3. Adja hozzá a licencet az új szerverhez.

---

## 1. Távolítsa el a licencet a régi kiszolgálóról

Azon a szerveren, ahol a licence jelenleg aktiválva van, távolítsa el a licencet az alábbi módszerek egyikével:

**Az OpenAdmintól:**
Lépjen az **OpenAdmin > Licenc** elemre, és kattintson a **Leépítés** gombra.
![Rendkívüli licenc](https://i.postimg.cc/RSsLWmLQ/downgrade.png)

**A terminálról:**

``` bash
opencli licenc törlése
  ```

Miután eltávolította a licencet a régi kiszolgálóról, folytassa a következő lépéssel.

---

## 2. Adja ki újra a licencet

Minden licenc automatikusan annak a szervernek az IP-címéhez van kötve, ahol először aktiválták.
Az új szerverre való átvitelhez **újra ki kell adnia** a licencet:

1. Jelentkezzen be [my.openpanel.com-fiókjába](https://my.openpanel.com/index.php?rp=/login).
2. Nyissa meg a [Licencek](https://my.openpanel.com/clientarea.php?action=services) részt.
![Licenclista](https://i.postimg.cc/KjG6p24j/lienses.png)
3. Keresse meg az áthelyezni kívánt licencet, és kattintson rá.
![Licenclista](https://i.postimg.cc/xC0pnPQZ/rows.png)
4. Kattintson az **Újrakiadás** gombra.
![Licenc újra kiadása](https://i.postimg.cc/xC0pnPQZ/rows.png)

Az újbóli kiadás után a licenc már nincs IP-címhez kötve.
Ez automatikusan csatlakozik a következő szerverhez, ahol aktiválja.

---

## 3. Adja hozzá a licencet az új kiszolgálóhoz

Most aktiválja a licencet az új szerveren. Ezt kétféleképpen teheti meg:

**Az OpenAdmintól:**
Lépjen az **OpenAdmin > Licenc** elemre, írja be a licenckulcsot, majd kattintson a **Kulcs mentése** gombra.
![Licenckulcs hozzáadása](/img/guides/add_key.png)

**A terminálról:**

``` bash
opencli licenc add hozzá a KULCSOT IDE
  ```

---

**Ennyi!** Az OpenPanel Enterprise licence sikeresen átkerült az új szerverre.
