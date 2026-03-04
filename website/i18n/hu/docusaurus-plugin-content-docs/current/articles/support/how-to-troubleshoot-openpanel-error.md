# Az OpenPanel felhasználói felület hibáinak elhárítása

## 500 Hiba

Ha **500-as hiba** fordul elő az OpenPanel felhasználói felületén, egy egyedi hibakód jelenik meg az oldalon.

> ⚠️ Ez a hibakód az Ön gépére vonatkozik. Csak a szerver rendszergazdája férhet hozzá a részletes információkhoz.

A hiba teljes részleteinek megtekintéséhez futtassa a következő parancsot a terminálon:

```bash
opencli error ERROR_ID_HERE
```

Ellenőrizze a kimeneten a hibaüzenetet. Ha segítségre van szüksége, átmásolhatja az üzenetet [támogatási fórumunkra](https://community.openpanel.org/) vagy a [Discord-csatornára](https://discord.openpanel.com/), ahol segítségre van szüksége a hibaelhárításhoz.

---

## A felhasználói felület nem válaszol

Ha egy funkció nem a várt módon működik (például egy gombra kattintva nem reagál), az valószínűleg **front-end probléma**.

A hibaelhárításhoz:

1. Nyissa meg böngészője **Fejlesztői eszközöket** (általában `F12` vagy `Ctrl+Shift+I` / `Cmd+Option+I`).
2. Lépjen a **Hálózat** lapra.
3. Ismételje meg a nem működő műveletet.
4. Ellenőrizze a kéréseket a **Hálózat** lapon és a **Konzol** naplójában található hibákat.

* Ha kiválaszt egy kérést a Hálózat lapon, megtekintheti a háttérrendszer által visszaadott választ.
* Ha a válasz nem tartalmaz elegendő információt a probléma diagnosztizálásához, engedélyezze a **dev_mode** beállítást a kiszolgálón, és ellenőrizze a Docker-naplókat.

---

## Fejlesztői mód

A **dev_mode** engedélyezésével az OpenPanel és az OpenAdmin felületek részletes hibakeresési információkat naplózhatnak minden kérés esetén. Ez segít a rendszergazdáknak látni a panel által futtatott parancsokat és a kapott válaszokat.

A dev_mode engedélyezése:

```bash
opencli config update dev_mode on
```

Ezután indítsa újra a panelt.

Ha a dev_mode engedélyezve van, a felhasználói panel részletes naplói a következő címen érhetők el:

```bash
docker logs -f openpanel
```

Ezek a naplók részletes hibakeresési információkat tartalmaznak a hibaelhárításhoz.

