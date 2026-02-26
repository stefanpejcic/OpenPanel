---
sidebar_position: 6
---

# Biztonsági mentések

Az OpenPanelben a biztonsági mentéseket közvetlenül **a panelfelhasználók** konfigurálhatják, ellentétben más panelekkel, ahol a biztonsági mentéseket az adminisztrátor kezeli.

Ez lehetővé teszi a felhasználók számára, hogy meghatározzák saját biztonsági mentési ütemezéseiket, pontosan megválasszák, hogy miről készítsenek biztonsági másolatot, és a támogatott célhelyek széles skálájából válasszanak.
Ennek eredményeként az adminisztrátoroknak kevesebb feladatot kell kezelniük, a felhasználók pedig nagyobb irányítást és rugalmasságot kapnak.

![backups.png](/img/panel/v2/backups.png)

## Úticélok

Az OpenPanel a következő célokat támogatja:

- [**S3 tárhely**](#s3) - AWS S3, Filebase, MinIO stb.
- [**WebDAV**](#webdav) - Távoli biztonsági mentések WebDAV használatával.
- [**SSH**](#azure) - Távoli biztonsági mentések SSH-n keresztül egy másik szerverre.
- [**Azure**](#azure) - Távoli biztonsági mentések az Azure Blob Storage-ba.
- [**Dropbox**](/#dropbox) - Távoli biztonsági mentések a Dropbox felhőtárhelyére.

A felhasználó kiválaszthatja a célállomást a **Biztonsági mentések > Úticélok** oldalon.

![destinations.png](/img/panel/v2/destinations.png)

Miután kiválasztotta a biztonsági mentést, annak lehetőségei a **Biztonsági mentések > Beállítások oldalon** érhetők el.

### SSH/SFTP

| Beállítás | Leírás |
|----------|-----------|
| **SSH_HOST_NAME** | A távoli SSH-kiszolgáló FQDN-je, például: `server.local` |
| **SSH_PORT** | A távoli SSH-kiszolgáló portja |
| **SSH_REMOTE_PATH** | A könyvtár, ahová a biztonsági másolatokat el kell helyezni az SSH-kiszolgálón. Példa: `/home/user/backups` |
| **SSH_USER** | Az SSH-kiszolgáló felhasználóneve. Példa: "felhasználó" |
| **SSH_PASSWORD** | Az SSH szerver jelszava. Példa: "jelszó" |
| **SSH_IDENTITY_FILE** | A privát kulcs elérési útja a `/var/www/html/` könyvtárban az SSH-kiszolgálóhoz. A nem RSA kulcsok (pl. ed25519) is működni fognak.  |
| **SSH_IDENTITY_PASSPHRASE** | Az azonosító fájl jelmondata, ha van. Példa: "pass" |


### S3

| Beállítás | Leírás |
|----------|-----------|
| **AWS_S3_BUCKET_NAME** | A biztonsági mentések tárolására használandó távoli tároló neve. Ha ez nincs beállítva, a rendszer nem tárol távoli biztonsági másolatokat. Példa: `backup-bucket` |
| **AWS_S3_PATH** | Ha a biztonsági másolatot nem gyökér helyen szeretné tárolni a tárolóban, megadhat egy elérési utat. Az elérési út nem tartalmazhat bevezető perjelet. Példa: `saját/tartalék/helyszín` |
| **AWS_ACCESS_KEY_ID** **AWS_SECRET_ACCESS_KEY** | Határozza meg a hitelesítési adatokat a biztonsági mentési tárhely alapján történő hitelesítéshez és egy gyűjtőhelynevet. Bár ezek a kulcsok mindegyike "AWS" előtaggal rendelkezik, a beállítás bármely S3-kompatibilis tárolóval használható. |
| **AWS_IAM_ROLE_ENDPOINT** | A statikus hitelesítő adatok megadása helyett IAM-példányprofilokat vagy hasonlókat is használhat a hitelesítéshez. Néhány lehetséges konfigurációs lehetőség az AWS-ben: `- EC2: http://169.254.169.254` `- ECS: http://169.254.170.2` |
| **AWS_ENDPOINT** | Ez a tárolószerver FQDN-je, pl. `tárhely.example.com`. Ha egy adott (nem https) protokollt kell beállítania, akkor az `AWS_ENDPOINT_PROTO' opciót kell használnia. Az alapértelmezett érték a szabványos AWS S3 végpontra mutat. Példa: `s3.amazonaws.com` |
| **AWS_ENDPOINT_PROTO** | Az S3 tárolókiszolgálóval való kommunikáció során használandó protokoll. Alapértelmezés szerint „https”. Ezt a „http” értékre állíthatja be, ha egy másik, önállóan hosztolt s3 tárolóval kommunikál. |
| **AWS_ENDPOINT_INSECURE** | Ha ezt a változót "true"-ra állítja, akkor letiltja az SSL-tanúsítványok ellenőrzését a következőhöz: "AWS_ENDPOINT". Csak akkor használja ezt, ha önaláírt tanúsítványokat használ a távoli tároló háttérrendszeréhez. Ez csak akkor használható, ha az „AWS_ENDPOINT_PROTO” „https” értékre van állítva.  |
| **AWS_ENDPOINT_CA_CERT** | Ha önaláírt tanúsítványokat szeretne használni az S3 szerverén, átadhatja a PEM kódolású CA-tanúsítvány helyét, és azt a tanúsítványok érvényesítésére fogja használni. Alternatív megoldásként adja át a tanúsítványt tartalmazó PEM-kódolású karakterláncot. Példa: `/var/www/html/cert.pem` |
| **AWS_STORAGE_CLASS** | A kulcs értékének beállítása megváltoztatja az S3 tárolási osztály fejlécét. Az alapértelmezett viselkedés a szabványos osztály használata, ha nincs megadva érték. Példa: "GLACIER" |
| **AWS_PART_SIZE** | Ennek a változónak a beállítása megváltoztatja az S3 alapértelmezett alkatrészméretét a másolási lépéshez. Ez az érték akkor hasznos, ha nagy fájlokat szeretne feltölteni. Megjegyzés: Ha a Scaleway-t S3-szolgáltatóként használja, ügyeljen arra, hogy az alkatrészszámláló 1000-re van állítva. Míg a Minio keményen kódolt értéket használ 10 000-re. Megkerülő megoldásként próbáljon meg magasabb értéket beállítani. Alapértelmezésben `16` (MB), ha nincs beállítva (minio-ból), ezt az értéket az igényei szerint állíthatja be. A mértékegység MB és egy egész szám. |


### WebDAV

| Beállítás | Leírás |
|----------|-----------|
| **WEBDAV_URL** | A távoli WebDAV-kiszolgáló URL-címe. Példa: `https://webdav.example.com` |
| **WEBDAV_PATH** | A WebDAV-kiszolgálón a biztonsági másolatok elhelyezésére szolgáló könyvtár. Ha az elérési út nem található a szerveren, akkor létrejön. Példa: `/saját/könyvtár/` |
| **WEBDAV_USERNAME** | A WebDAV szerver felhasználóneve Példa: `user` |
| **WEBDAV_PASSWORD** | A WebDAV szerver jelszava. Példa: "jelszó" |
| **WEBDAV_URL_INSECURE** |  Ha ezt a változót "true"-ra állítja, akkor letiltja a WEBDAV_URL SSL-tanúsítványainak ellenőrzését. Csak akkor használja ezt, ha önaláírt tanúsítványokat használ a távoli tároló háttérrendszeréhez. |

### Azure

| Beállítás | Leírás |
|----------|-----------|
| **AZURE_STORAGE_ACCOUNT_NAME** | A hitelesítő adatok fiókneve az Azure Blob Storage használatakor. Ezt az Azure Blob Storage használatakor be kell állítani.  Példa: "fióknév" |
| **AZURE_STORAGE_PRIMARY_ACCOUNT_KEY** | A hitelesítő adatok elsődleges fiókkulcsa az Azure Blob Storage használatakor. Ha ez nincs megadva, a parancs megpróbál visszatérni egy kapcsolati karakterlánc (ha adott) vagy egy felügyelt identitás (ha egyik sincs beállítva) használatára. |
| **AZURE_STORAGE_CONNECTION_STRING** | Kapcsolati karakterlánc az Azure Blob Storage eléréséhez. Ha ez nincs megadva, a parancs megpróbál visszatérni egy elsődleges fiókkulcs (ha adott) vagy egy felügyelt identitás (ha egyik sincs beállítva) használatára. |
| **AZURE_STORAGE_CONTAINER_NAME** | A tároló neve az Azure Blob Storage használatakor.  Példa: "tárolónév" |
| **AZURE_STORAGE_ENDPOINT** | A szolgáltatás végpontja az Azure Blob Storage használatakor. Ez egy sablon, amely átadhatja a fiók nevét.  Példa: `https://{{ .AccountName }}.blob.core.windows.net/` |
| **AZURE_STORAGE_ACCESS_TIER** | A hozzáférési szint az Azure Blob Storage használatakor. A lehetséges értékek [itt vannak felsorolva](https://github.com/Azure/azure-sdk-for-go/blob/sdk/storage/azblob/v1.3.2/sdk/storage/azblob/internal/generated/zz_constants.go#L14-L30) Példa: "Cold" |


### Dropbox

| Beállítás | Leírás |
|----------|-----------|
| **DROPBOX_REMOTE_PATH** | Abszolút távoli elérési út a Dropboxban, ahol a biztonsági másolatokat tárolni kell. Megjegyzés: Ha nem rendelkezik globális hozzáféréssel, használja az alkalmazás alútvonalát a Dropboxban.  Példa: "/saját/könyvtár" |
| **DROPBOX_APP_KEY** **DROPBOX_APP_SECRET** | Alkalmazáskulcs és alkalmazástitok az alkalmazásból, amelyet a következő címen hoztak létre: [https://www.dropbox.com/developers/apps](https://www.dropbox.com/developers/apps) |
| **DROPBOX_CONCURRENCY_LEVEL** | A Dropbox egyidejű, csonkolt feltöltéseinek száma. A 6 feletti értékek általában nem eredményeznek javításokat. |
| **DROPBOX_REFRESH_TOKEN** |  Frissítse a tokent új, rövid élettartamú tokenek kéréséhez (OAuth2) |


Dropbox háttértár beállítása:

1. Hozzon létre egy új Dropbox alkalmazást az [App Console]-ban (https://www.dropbox.com/developers/apps)
2. Nyissa meg az új Dropbox alkalmazást, és másolja ki a DROPBOX_APP_KEY és a DROPBOX_APP_SECRET kulcsokat.
3. Kattintson az „Engedélyek” elemre az alkalmazásban, és győződjön meg arról, hogy a következő engedélyek (vagy több) vannak megadva:
- `files.metadata.write`
- `files.metadata.read`
- `files.content.write`
- `files.content.read`
4. Cserélje ki az APPKEY-t a `https://www.dropbox.com/oauth2/authorize?client_id=APPKEY&token_access_type=offline&response_type=code-ban a 2. lépésben szereplő alkalmazáskulcsra.
5. Keresse fel az URL-t, és erősítse meg az alkalmazás hozzáférését. Ez ad egy "auth kódot" -> mentsd el valahova!
6. Cserélje ki az AUTHCODE, APPKEY, APPSECRET kódokat ennek megfelelően, és hajtsa végre a kérést a termináljáról:
   ```
curl https://api.dropbox.com/oauth2/token \
-d kód=AUTHCODE \
-d grant_type=engedélyezési_kód \
-d client_id=APPKEY \
-d client_secret=APPSECRET
   ```
7. Hajtsa végre a kérést. JSON formátumú választ fog kapni. Használja a "refresh_token" értékét az utolsó "DROPBOX_REFRESH_TOKEN" beállításhoz
8. Most be kell állítania a DROPBOX_APP_KEY, DROPBOX_APP_SECRET és DROPBOX_REFRESH_TOKEN kulcsot. Ezek nem járnak le.

:::info
*Megjegyzés*: A „Létrehozott hozzáférési token” használata az alkalmazáskonzolban nem támogatott, mivel csak nagyon rövid élettartamú, ezért nem alkalmas automatikus biztonsági mentési megoldáshoz. A frissítési token ezt automatikusan kezeli – a fenti beállítási eljárásra csak egyszer van szükség.
:::

:::veszély
Fontos: Ha a fenti 1. lépésben a Dropbox alkalmazás létrehozásakor az "Alkalmazásmappa" hozzáférést választotta, a "DROPBOX_REMOTE_PATH" relatív elérési út lesz az App mappa alatt! (Például a DROPBOX_REMOTE_PATH=/somedir azt jelenti, hogy a biztonsági mentési fájl a /Apps/myapp/somedir mappába kerül feltöltésre.) Másrészt, ha a teljes Dropbox hozzáférést választotta, a `DROPBOX_REMOTE_PATH` értéke a Dropbox tárolóterületén belüli abszolút elérési utat jelenti. (A fenti példát figyelembe véve a biztonsági mentési fájl a Dropbox gyökérkönyvtárában lévő /somedir mappába kerül feltöltésre.)
:::


## Titkosítás

A biztonsági mentés titkosítása alapértelmezés szerint le van tiltva, és először a rendszergazdának kell engedélyeznie.

Az összes titkosítási lehetőség kölcsönösen kizárja egymást. Adjon meg egyetlen lehetőséget az Ön által választott titkosítási sémához.


| Beállítás | Leírás |
|----------|-----------|
| **GPG_PASSPHRASE** | A biztonsági másolatok szimmetrikusan titkosíthatók gpg használatával, ha jelszót adunk meg. |
| **GPG_PUBLIC_KEY_RING** |  A biztonsági másolatok aszimmetrikusan titkosíthatók gpg használatával, ha nyilvános kulcsokat adnak meg. Használhatja a cső szintaxisát többsoros érték átadására. |
| **AGE_PASSPHRASE** | A biztonsági másolatok szimmetrikusan titkosíthatók az életkor használatával, ha jelszót adnak meg. |
| **AGE_PUBLIC_KEYS** | A biztonsági másolatok aszimmetrikusan titkosíthatók az életkor használatával, ha nyilvános kulcsokat adnak meg. Több kulcsot vesszővel elválasztott listaként kell megadni. Jelenleg ez támogatja az "age" és "ssh" kulcsokat. |

## Forgatás

:::veszély
**FONTOS, KÉRJÜK, OLVASSA EL EZT A FUNKCIÓ HASZNÁLATA ELŐTT**:
A régi biztonsági másolatok levágására használt mechanizmus nem túl kifinomult, és alapértelmezés szerint a **a célkönyvtárban lévő összes fájlra** alkalmazza a szabályait, ami azt jelenti, hogy ha a biztonsági másolatokat más fájlok mellett tárolja, akkor ezeket is törölni lehet. Ha ezt az opciót használja, győződjön meg arról, hogy a biztonsági mentési fájlok egy kizárólag az ilyen fájlok számára használt könyvtárban vannak tárolva, vagy állítsa be a "BACKUP_PRUNING_PREFIX" beállítást, hogy az eltávolítást bizonyos fájlokra korlátozza.
:::

| Beállítás | Leírás |
|----------|-----------|
| **BACKUP_RETENTION_DAYS** | Adjon meg nullát vagy pozitív egész értéket a régi biztonsági mentések automatikus elforgatásának engedélyezéséhez. Az érték megadja, hogy hány napig tartanak biztonsági másolatot. Alapértelmezett: "-1" |
| **BACKUP_PRUNING_LEEWAY** | Abban az esetben, ha a biztonsági mentés időtartama észrevehetően ingadozik a beállításban, módosíthatja ezt a beállítást, hogy megbizonyosodjon arról, hogy nincsenek versenyfeltételek az időablak szélén lévő tartalék ülés között. Állítsa be ezt az értéket egy olyan időtartamra, amely a biztonsági mentések befejezését és a biztonsági mentéseket nem törlő forgatást jelenti, és amely várhatóan nagyobb, mint a biztonsági mentések maximális különbsége. Az érvényes értékek utótagja (s)seconds, (m)nutes vagy (h)ours. Alapértelmezés szerint egy percet használnak. |
| **BACKUP_PRUNING_PREFIX** | Abban az esetben, ha a célcsoport vagy -könyvtár más fájlokat is tartalmaz, mint amilyeneket ez a tároló kezel, az előtag értékének beállításával korlátozhatja a forgatás hatókörét. Ez általában a BACKUP_FILENAME nem paraméterezett része. Például. ha a BACKUP_FILENAME `db-backup-%Y-%m-%dT%H-%M-%S.tar.gz`, akkor a BACKUP_PRUNING_PREFIX értéket `db-backup-` értékre állíthatja, és győződjön meg arról, hogy a nem kapcsolódó fájlokat nem érinti az elforgatási mechanizmus. |

## Forrás

Alapértelmezés szerint mindenről biztonsági másolat készül. Abban az esetben, ha csak egy adott adatról kell biztonsági másolatot készítenie, állítsa be a „BACKUP_SOURCES” értéket.

A választható lehetőségek a következők:

| Beállítás | Leírás |
|----------|-----------|
| **Mind** |  Biztonsági másolatot készít az összes docker kötet tartalmáról: webhelyfájlok, mysql/mariadb adatbázisok, minecraft, postgre adatbázisok, vhosts fájlok. |
| **Fájlok** |  Csak a `/var/www/html/` könyvtárban lévő webhelyfájlokról készít biztonsági másolatot. |
| **E-mailek** |  Csak a `/var/mail/` könyvtárban lévő e-mail fájlokról készít biztonsági másolatot. |
| **MySQL adatbázisok** |  Csak a MySQL/MariaDB adatbázisokról készít biztonsági másolatot. |
| **Postgres adatbázisok** |  Csak a PostgreSQL adatbázisokról készít biztonsági másolatot. |
| **MsSQL adatbázisok** |  Csak az MsSQL adatbázisokról készít biztonsági másolatot. |
| **VirtualHosts** |  Csak az Apache/Nginx VirtualHostokról készít biztonsági másolatot a tartományokhoz. |
| **Minecraft adatok** |  Csak a `/data` mappát menti a Minecraft tárolóból. |


## Ütemezés

A biztonsági mentések rögzített ütemezetten futtathatók, amelyek cron kifejezésként vannak definiálva.
A "BACKUP_CRON_EXPRESSION" a biztonsági mentési feladat ütemezésére szolgál, a cron kifejezés egy időkészletet jelent, 5 vagy 6 szóközzel elválasztott mező használatával.

| Mezőnév | Kötelező? | Engedélyezett értékek | Engedélyezett speciális karakterek |
| ---------- | ---------- | -------------- | --------------------------- |
| Másodpercek | Nem | 0-59 | `* / , -` |
| Jegyzőkönyv | Igen | 0-59 | `* / , -` |
| Óra | Igen | 0-23 | `* / , -` |
| A hónap napja | Igen | 1-31 | `* / , - ?` |
| Hónap | Igen | 1-12. vagy JAN-DEC. | `* / , -` |
| A hét napja | Igen | 0-6 vagy V-SZO | `* / , - ?` |


A hónap és a hét napja mező értékei nem különböznek egymástól. A „SUN”, „Sun” és „sun” egyformán elfogadott. Segítségért látogasson el az olyan webhelyekre, mint a [crontab.guru](https://crontab.guru).

Ha nincs beállítva érték, akkor a „@daily” lesz használva, amely minden nap éjfélkor fut.

## Tömörítés


| Beállítás | Leírás |
|----------|-----------|
| `BACKUP_COMPRESSION` |  A tar-val együtt használt tömörítési algoritmus. Az érvényes lehetőségek a következők: "gz" (Gzip), "zst" (Zstd) vagy "none" (csak tar).  Az alapértelmezett a "gz". Vegye figyelembe, hogy a kiválasztás a fájl kiterjesztését érinti. |
| "GZIP_PARALLELISM" | A "gz" (Gzip) tömörítés párhuzamossági szintje. Meghatározza, hogy hány adatblokkot dolgozzon fel egyidejűleg. A magasabb értékek gyorsabb tömörítést eredményeznek. Nincs hatása a dekompresszióra. Alapértelmezett = `1`. Ha ezt 0-ra állítja, az összes elérhető szálat felhasználja. |

## Nevek

Az itt található lehetőségek alapértelmezés szerint le vannak tiltva, és először a rendszergazdának kell manuálisan engedélyeznie őket.

| Beállítás | Leírás |
|----------|-----------|
| `BACKUP_FILENAME` |  A biztonsági másolat fájl kívánt neve, beleértve a kiterjesztést. A formátumú igék az „strftime”-hez hasonlóan le lesznek cserélve. Az összes ige elhagyása ugyanazt a fájlnevet eredményezi minden biztonsági mentés futtatásakor, ami azt jelenti, hogy a korábbi verziók felülíródnak a következő futtatások során. A kiterjesztés definiálható szó szerint vagy a `{{ .Extension }}` sablonon keresztül, ebben az esetben "tar.gz", "tar.zst" vagy ".tar" lesz (a `BACKUP_COMPRESSION` beállításától függően). Az alapértelmezett `backup-%Y-%m-%dT%H-%M-%S.{{ .Extension }}` a következő fájlneveket eredményezi: `backup-2021-08-29T04-00-00.tar.gz`. |


## Kizárás

`BACKUP_EXCLUDE_REGEXP` – Érték megadásakor a `BACKUP_SOURCES` minden olyan fájlja, amelynek teljes elérési útja megegyezik a reguláris kifejezéssel, ki lesz zárva az archívumból. A reguláris kifejezések [a Go standard könyvtárából] (https://pkg.go.dev/regexp) használhatók. Példa: `\.log$`.

## Értesítések

Értesítések (e-mail, Slack stb.) küldhetők ki, amikor a biztonsági mentés befejeződik.

### NOTIFICATION_URLS

A konfiguráció a [shourrrr] által felhasznált URL-ek vesszővel elválasztott listájaként jelenik meg (https://containrrr.dev/shourrrr/v0.8/services/overview/)

| Szolgáltatás | URL formátum |
|--------------|-------------------------------------------- ------------------------------------------------------------|
| [Ugatás](#ugatás) | `bark://devicekey@host` |
| [Discord](#discord) | `discord://token@id` |
| [E-mail](#email) | `smtp://username:password@host:port/?from=fromAddress&to=recipient1[,recipient2,...]` |
| Gotify | `gotify://gotify-host/token` |
| Google Chat | `googlechat://chat.googleapis.com/v1/spaces/FOO/messages?key=bar&token=baz` |
| IFTTT | `ifttt://key/?events=event1[,event2,...]&value1=value1&value2=value2&value3=value3` |
| Csatlakozz | `join://shoutrrr:api-key@join/?devices=device1[,device2,...][&icon=icon][&title=title]` |
| Legfontosabb | `mattermost://[felhasználónév@]mattermost-host/token[/csatorna]` |
| Mátrix | `mátrix://username:password@host:port/[?rooms=!roomID1[,roomAlias2]]` |
| Ntfy | `ntfy://username:password@ntfy.sh/topic` |
| OpsGenie | `opsgenie://host/token?responders=responder1[,responder2]` |
| Pushbullet | `pushbullet://api-token[/device/#channel/email]` |
| Pushover | `pushover://shourrrr:apiToken@userKey/?devices=device1[,device2,...]` |
| Rocketchat | `rocketchat://[felhasználónév@]rocketchat-host/token[/channel|@címzett]` |
| Slack | `slack://[botnév@]token-a/token-b/token-c` |
| Csapatok | `teams://group@tenant/altId/groupOwner?host=organization.webhook.office.com` |
| Távirat | `telegram://token@telegram?chats=@channel-1[,chat-id-1,...]` |
| Zulip Chat | `zulip://bot-mail:bot-key@zulip-domain/?stream=name-or-id&topic=name` |

#### E-mail

URL formátum:

```
smtp://username:password@host:port/?from=fromAddress&to=recipient1[,recipient2,...]
```

URL mezők:

- **Felhasználónév** – SMTP szerver felhasználónév
**Alapértelmezett:** _üres_
**URL rész:** `smtp://felhasználónév:jelszó@host:port/`

- **Jelszó** – SMTP-szerver jelszó vagy hash (OAuth2-hez)
**Alapértelmezett:** _üres_
**URL rész:** `smtp://felhasználónév:jelszó@host:port/`

- **Host** – SMTP-szerver gazdagépneve vagy IP-címe (**Kötelező**)
**URL rész:** `smtp://felhasználónév:jelszó@host:port/`

- **Port** – SMTP szerver port
**Általános értékek:** 25, 465, 587, 2525
**Alapértelmezett:** `25`
**URL rész:** `smtp://felhasználónév:jelszó@host:port/`

Lekérdezés/paraméter kellékek:

A kellékek megadhatók a "params" argumentummal vagy közvetlenül az URL-ben a "?key=value&key=value" használatával.

- **FromAddress** - E-mail cím, ahonnan a levelet küldték (**Kötelező**)
**Álnevek:** `from`

- **ToAddresses** – A címzett e-mailek vesszővel elválasztott listája (**Kötelező**)
**Álnevek:** `to`

- **Auth** – SMTP hitelesítési módszer
**Alapértelmezett:** `Ismeretlen`
**Lehetséges értékek:** "Nincs", "Sima", "CRAMMD5", "Ismeretlen", "OAuth2"

- **ClientHost** – Az SMTP-szervernek a HELO fázis során elküldött gazdagépnév
**Alapértelmezett:** `localhost`
Az operációs rendszer gazdagépnevének használatához állítsa "auto"-ra

- **Titkosítás** - Titkosítási módszer
**Alapértelmezett:** "Automatikus".
**Lehetséges értékek:** "Nincs", "ExplicitTLS", "ImplicitTLS", "Auto"

- **FromName** – A feladó neve
**Alapértelmezett:** _üres_

- **Tárgy** – Az e-mail tárgysora
**Alapértelmezett:** "Shoutrrr Notification".
**Álnevek:** "cím".

- **UseHTML** – Az üzenet HTML formátumú-e
**Alapértelmezett:** ❌ Nem

- **UseStartTLS** – Használja-e a StartTLS titkosítást
**Alapértelmezett:** ✔ Igen
**Álnevek:** `starttls`
  

#### Bark

URL mezők:

- **DeviceKey** – Az egyes eszközök kulcsa (**Kötelező**)
**URL rész:** `bark://:devicekey@host/path`

- **Host** – Szerver gazdagépneve és portja (**Kötelező**)
**URL rész:** `bark://:devicekey@host/path`

- **Path** – Szerver elérési útja
**Alapértelmezett:** `/`
**URL rész:** `bark://:devicekey@host/path`

Lekérdezés/paraméter kellékek:

A kellékek megadhatók a "params" argumentum használatával, vagy az URL-en keresztül a "?key=value&key=value" stb. használatával.

- **Jelvény** – Az alkalmazás ikonja mellett megjelenő szám
**Alapértelmezett:** `0`

- **Kategória** – Fenntartott mező, még nincs használatban
**Alapértelmezett:** _üres_

- **Másolás** – A másolandó érték
**Alapértelmezett:** _üres_

- **Csoport** – Az értesítés csoportja
**Alapértelmezett:** _üres_

- **Ikon** – Az ikon URL-je, amely csak iOS 15 vagy újabb rendszeren érhető el
**Alapértelmezett:** _üres_

- **Séma** – Szerverprotokoll, "http" vagy "https".
**Alapértelmezett:** `https`

- **Hang** – A [Bark Sounds] értéke (https://github.com/Finb/Bark/tree/master/Sounds)
**Alapértelmezett:** _üres_

- **Cím** – Az értesítés címe, opcionálisan a feladó által beállítva
**Alapértelmezett:** _üres_

- **URL** – az értesítésre kattintva megnyíló URL
**Alapértelmezett:** _üres_

#### Egyenetlenség

URL formátum:

```
discord://token@webhookid
```

URL mezők:

- **Token** – (**Kötelező**)
**URL rész:** `discord://token@webhookid/`

- **WebhookID** – (**Kötelező**)
**URL rész:** `discord://token@webhookid/`

Lekérdezés/paraméter kellékek:

A kellékek megadhatók a "params" argumentum használatával, vagy az URL-en keresztül a "?key=value&key=value" stb. használatával.

- **Avatar** – A webhook alapértelmezett avatárjának felülírása a megadott URL-lel
**Alapértelmezett:** _üres_
**Álnevek:** `avatarurl`

- **Szín** – Az egyszerű üzenetek bal oldali szegélyének színe
**Alapértelmezett:** `0x50D9ff`

- **ColorDebug** – A bal oldali szegély színe hibakeresési üzeneteknél
**Alapértelmezett:** `0x7b00ab`

- **ColorError** – A hibaüzenetek bal oldali keretének színe
**Alapértelmezett:** `0xd60510`

- **ColorInfo** – Az információs üzenetek bal oldali keretének színe
**Alapértelmezett:** `0x2488ff`

- **ColorWarn** – A figyelmeztető üzenetek bal oldali keretének színe
**Alapértelmezett:** `0xffc441`

- **JSON** – A teljes üzenetet JSON-adatként kell-e elküldeni, ahelyett, hogy a "tartalom" mezőként használnák.
**Alapértelmezett:** ❌ Nem

- **SplitLines** – Az egyes sorokat külön beágyazott elemként kell-e elküldeni
**Alapértelmezett:** ✔ Igen

- **Cím**
**Alapértelmezett:** _üres_

- **Felhasználónév** – A webhook alapértelmezett felhasználónevének felülírása
**Alapértelmezett:** _üres_

Webhook létrehozása Discordban:

1. Nyissa meg csatornabeállításait a csatorna neve melletti fogaskerék ikonra kattintva.
![screenshot 1](https://i.postimg.cc/FmYrJfhZ/sc-1.png)
2. A bal oldali menüben kattintson az **Integrációk** lehetőségre.
![screenshot 2](https://i.postimg.cc/NtJGM8vF/sc-2.png)
3. A jobb oldalon kattintson a **Webhook létrehozása** lehetőségre.
![3. képernyőkép](https://i.postimg.cc/7Djx1x6y/sc-3.png)
4. Állítsa be a kívánt nevet, csatornát és ikont, majd kattintson a **Webhook URL másolása** lehetőségre.
![4. képernyőkép](https://i.postimg.cc/2rvC2PbL/sc-4.png)
5. Nyomja meg a **Változtatások mentése** gombot.
![screenshot 5](https://i.postimg.cc/wqG9GzQL/sc-5.png)
6 Formázza meg a szolgáltatás URL-jét:
   ```
https://discord.com/api/webhooks/693853386302554172/W3dE2OZz4C13_4z_uHfDOoC7BqTW288s-z1ykqI0iJnY_HjRqMGO8Sc7YDqvf_KVKjhJ
└────────────────┘ └───────────────────────────── ───────────────────────────────
webhook azonosító token
   
discord://W3dE2OZz4C13_4z_uHfDOoC7BqTW288s-z1ykqI0iJnY_HjRqMGO8Sc7YDqvf_KVKjhJ@693853386302554172
└───────────────────────────── ─────────────────────────────── └────────────────┘
token webhook azonosító
   ```



### NOTIFICATION_LEVEL

Alapértelmezés szerint a rendszer csak akkor küld értesítést, ha a biztonsági mentés sikertelen. Ha minden futtatásról értesítést szeretne kapni, állítsa a "NOTIFICATION_LEVEL" értékét "info" értékre az alapértelmezett "error" helyett.


## Biztonsági mentés elindítása

Az ütemezett beállításoktól függetlenül bármikor manuálisan indíthat biztonsági mentést.
Ehhez győződjön meg arról, hogy a Docker funkció engedélyezve van. Majd:

1. Lépjen a **Containers > Terminal** elemre.
2. Válassza ki a **Backup** tárolót.
3. A terminálba írja be a következő parancsot, és nyomja meg az Enter billentyűt:

```bash
backup
```

## Hibaelhárítás

Hiba *hiba az archívum másolásakor: ssh.(*sshStorage).Másolás: hiba a fájl létrehozásakor: a fájl nem létezik* azt jelzi, hogy a beállított könyvtár nem létezik a célhelyen:

```
time=2025-07-22T11:21:05.117Z level=ERROR msg="Fatal error running command: file does not exist" error="main.(*command).runAsCommand: error running script: main.runScript.func4: error running script: main.(*script).copyArchive: error copying archive: ssh.(*sshStorage).Copy: error creating file: file does not exist"
```
