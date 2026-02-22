# Configuring OpenPanel Backups

OpenPanel has a unique feature where end-users can configure their remote backups. This provides users with more freedom and control over the schedule, what to backup and finally more privacy as Admin does not have access to their destination.

Backups can be configured either by the system administrator (admin-configured) or by end users (user-configured). Each mode has distinct setup and restore procedures.

---

## Konfigurációs lehetőségek

| Funkció | Rendszergazda által konfigurált biztonsági mentések | Felhasználó által konfigurált biztonsági mentések |
| --------------------------- | ------------------------------- | -------------------------------------- |
| Biztonsági mentés konfiguráció | A rendszergazda szerkeszti a `backups.env` | A felhasználók a Biztonsági mentések oldalon konfigurálják |
| Biztonsági mentési modul állapota | Le kell tiltani a felhasználók számára | Engedélyezni kell a felhasználók számára |
| Ki állítja be a biztonsági mentés ütemezését | Admin | Felhasználó |
| Biztonsági mentési cél vezérlése | Admin | Felhasználó |
| A visszaállítást | Admin | Felhasználó |
| Rendszergazdai hozzáférés a biztonsági másolatokhoz | Teljes | Nincs |


### 1. Admin konfigurált

Ebben a módban az **adminisztrátor teljes mértékben felügyeli** a biztonsági mentés ütemezését, megőrzését és a cél beállításait. A végfelhasználók **nem** módosíthatják a biztonsági mentési konfigurációkat.

---

#### 1: A biztonsági mentések modul letiltása

Ha meg szeretné akadályozni, hogy a felhasználók módosítsák a biztonsági mentési beállításokat, kapcsolja ki a **Biztonsági mentések** modult az adminisztrációs felületen.

**Útvonal:**
"OpenAdmin > Beállítások > Modulok".
**Művelet:** Kapcsolja ki a **Biztonsági mentések** modult.

---

#### 2: Sablon szerkesztése

Módosítsa az új fiókokra vonatkozó biztonsági mentési konfigurációs sablont. Ez a fájl határozza meg az alapértelmezett biztonsági mentési beállításokat, például a távoli tárolási célokat.

**Fájl elérési útja:**
`/etc/openpanel/backups/backups.env`

Távoli SSH-célhely engedélyezéséhez és konfigurálásához törölje a megjegyzéseket, és frissítse a következő változókat:

```env
########### SSH/SFTP STORAGE
# SSH_HOST_NAME=""
# SSH_PORT="22"
# SSH_REMOTE_PATH=""
# SSH_USER=""
# SSH_PASSWORD=""
# SSH_IDENTITY_FILE="/var/www/html/id_rsa"
# SSH_IDENTITY_PASSPHRASE=""
```

**Példa:**

```env
########### SSH/SFTP STORAGE
SSH_HOST_NAME=""
SSH_PORT="22"
SSH_REMOTE_PATH=""
SSH_USER=""
SSH_PASSWORD=""
# SSH_IDENTITY_FILE="/var/www/html/id_rsa"
# SSH_IDENTITY_PASSPHRASE=""
```

> 🔗 További céltípusokért és példákért lásd: [Biztonsági másolatok dokumentációja](/docs/panel/files/backups/#destinations)

---

#### 3: Ütemezés szerkesztése

A biztonsági mentés gyakoriságának beállításához lépjen a következő helyre:

**Útvonal:**
"OpenAdmin > Speciális > System Cron Jobs".

Keresse meg a parancs cron feladatát:

```bash
opencli docker-backup
```

Szükség szerint állítsa be az ütemtervet. Ez a parancs biztonsági mentést indít el a megadott ütemezés szerint a kiszolgálón lévő **minden aktív felhasználó** számára.

---

### 2. Felhasználó által konfigurált

Ebben a módban a **Biztonsági mentések modul engedélyezve van**, amely lehetővé teszi a felhasználók számára, hogy igényeik szerint konfigurálják saját biztonsági mentéseiket.

**Beállítás:**

* Az adminisztrátornak **engedélyeznie kell a Biztonsági mentések modult** az OpenPanelben.
* A biztonsági mentés funkciót engedélyezni kell a tárhelycsomagokhoz kapcsolódó összes releváns szolgáltatáskészleten, hogy lehetővé tegye a felhasználói hozzáférést.
* A végfelhasználók beállíthatják:

* Biztonsági mentési cél (pl. távoli tárhely, egyéni elérési utak)
* Biztonsági mentés ütemezése (amikor a biztonsági mentés fut)
* Milyen adatokról kell biztonsági másolatot készíteni (fájlok, adatbázisok vagy mindkettő)
* Erőforrás korlátok (pl. a biztonsági mentés során használt sávszélesség vagy CPU)

> 🔗 A végfelhasználói konfigurációhoz lásd a [Biztonsági másolatok dokumentációja] (/docs/panel/files/backups/)

**Megjegyzések:**

* A felhasználók felelősek a biztonsági mentéseik kezeléséért.
* A felhasználók bármikor manuálisan elindíthatják a biztonsági mentési folyamatot, ha a *Docker* funkció engedélyezve van.
* Az adminisztrátorok **nem** férhetnek hozzá a felhasználók biztonsági mentési célhelyeihez vagy konfigurációihoz.

---

## Visszaállítási eljárások

### Visszaállítás rendszergazda által konfigurált biztonsági mentési módban

* Az adminisztrátor manuálisan hajtja végre a visszaállításokat, akár terminálparancsokkal, akár az OpenPanel UI terminálján keresztül.
* A gyakori visszaállítási lépések a következők:

* Adatbázisok esetén: a megfelelő táblák eldobása és az adatbázis kiíratásának importálása a biztonsági mentési fájlokból.
* Fájlok esetén: a FileManager vagy a parancssor használata a sérült fájlok törléséhez és a biztonsági másolatok újbóli feltöltéséhez.

---

### Visszaállítás felhasználó által konfigurált biztonsági mentési módban

* A végfelhasználók felelősek saját biztonsági másolataik visszaállításáért, mivel a biztonsági másolatok a felhasználó által meghatározott, a rendszergazdák számára elérhetetlen helyeken tárolódnak.
* A felhasználók hasonló visszaállítási lépéseket követnek, mint az adminisztrátori módban, de maguknak kell végrehajtaniuk a műveleteket a biztosított eszközök vagy utasítások segítségével.
* Az adminisztrátorok ebben a módban nem állíthatják vissza és nem férhetnek hozzá a biztonsági másolatokhoz a felhasználók nevében.

---

