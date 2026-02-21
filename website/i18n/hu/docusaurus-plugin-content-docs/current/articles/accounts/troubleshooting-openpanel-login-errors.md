# Bejelentkezési hibák

A következő hibaüzenetek jelenhetnek meg az OpenPanel bejelentkezési oldalán. Íme, mit jelentenek mindegyik, és hogyan kell megoldani:


### `Felhasználónév és jelszó szükséges.`

Ez a hiba azt jelzi, hogy az egyik vagy mindkét bejelentkezési mező üresen maradt.

**Megoldás:** Kérjük, adja meg felhasználónevét és jelszavát is.

---

### `Nem lehet csatlakozni az adatbázishoz.`

Ez arra utal, hogy a MySQL szolgáltatás nem fut, vagy a "felhasználók" tábla sérült.

**Megoldás:** Ellenőrizze az adatbázis-szolgáltatás állapotát az **OpenAdmin > Szolgáltatások** részben, és győződjön meg arról, hogy a MySQL megfelelően fut.

---

### `Fiókját felfüggesztettük. Kérjük, forduljon az ügyfélszolgálathoz.`

OpenPanel felhasználói fiókját felfüggesztettük, és a bejelentkezés jelenleg le van tiltva.

**Megoldás:** Forduljon tárhelyszolgáltatójához fiókja újraaktiválásához.

---

### `Ismeretlen fiók. Kérjük, ellenőrizze a felhasználónevet.`

A megadott felhasználónév nem létezik a szerveren.

**Megoldás:** Ellenőrizze, hogy a megfelelő bejelentkezési linket és felhasználónevet használja-e. Ha a probléma továbbra is fennáll, forduljon a tárhelyszolgáltató támogatási csapatához.

---

### `Érvénytelen jelszó. Kérjük, próbálja újra.`

A megadott jelszó helytelen.

**Megoldás:** Ha elfelejtette jelszavát, vagy továbbra is problémái vannak, lépjen kapcsolatba tárhelyszolgáltatójával, és kérje a jelszó visszaállítását.

---

### `Invalid 2FA code. Please try again.`

The Two-Factor Authentication (2FA) code provided is not valid.

**Solution:** Generate a new 2FA code using your authentication app and try again.

---

### `Too many failed login attempts. Please try again later.`

Too many incorrect login attempts have been made from your IP address.

**Solution:** Contact your server administrator to review the log file at `/var/log/openpanel/user/failed_login.log`. They can unblock your IP or adjust the failed login attempt limits.
