# Dedikált IP-cím hozzárendelése a felhasználóhoz

Ezt követően az **OpenAdmin** vagy **OpenCLI** segítségével egy dedikált IP-cím hozzárendelhető a felhasználóhoz.

Az IP-cím megváltoztatása frissíti az összes kapcsolódó konfigurációs fájlt – beleértve a DNS-zónákat és a Caddy VirtualHostokat –, így biztosítva, hogy a felhasználó minden jelenlegi és újonnan hozzáadott tartománya az új IP-címhez legyen kötve.
A frissített IP-cím azonnal megjelenik a felhasználó OpenPanel felületén is.

## Az OpenAdmin használatával

Az OpenAdmin felület felsorolja a szerverhez jelenleg hozzárendelt összes IP-címet – ezek a `hostname -I` paranccsal is megtekinthetők.
Ha további IP-címeket kell hozzáadnia a szerverhez, **kövesse az operációs rendszer disztribúciójához tartozó hivatalos hálózati konfigurációs dokumentációt**.
Nincs szükség speciális konfigurációra az OpenPanel oldalról – az újonnan hozzáadott IP-címek automatikusan megjelennek a felületen.

Egy felhasználó IP-címének megváltoztatása az OpenAdminból:

1. Lépjen az **OpenAdmin → Felhasználók → *felhasználónév* → Szerkesztés** elemre.

2. Frissítse az **IP-cím** mezőt a kívánt dedikált IP-vel.

![image](https://i.postimg.cc/65f9Jcsf/slika.png)

3. Kattintson a **Mentés** gombra a módosítások alkalmazásához.

## OpenCLI használata

A felhasználó IP-címének parancssorból történő módosításához használja az [`opencli user-ip`](https://dev.openpanel.com/cli/users.html#Assign-Remove-IP-to-User) parancsot:

```bash
opencli user-ip <USERNAME> <NEW_IP_ADDRESS>
```

Ha a megadott IP-cím már hozzá van rendelve egy másik felhasználóhoz, a parancs figyelmeztetéssel megszakad.
Ennek a viselkedésnek a felülbírálásához és az újbóli hozzárendelés kikényszerítéséhez adja meg az "-y" jelzőt:

```bash
opencli user-ip <USERNAME> <NEW_IP_ADDRESS> -y
```

Ha el szeretne távolítani egy dedikált IP-címet a felhasználótól, és visszaállítani a [fő (megosztott) IP-címre] (/docs/articles/install-update/openpanel-main-ip-address), használja a következőket:

```bash
opencli user-ip <USERNAME> delete
```
