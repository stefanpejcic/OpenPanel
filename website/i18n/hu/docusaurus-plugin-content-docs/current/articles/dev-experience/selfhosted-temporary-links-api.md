# Ideiglenes linkek API

Az [OpenPanel SiteManager](/docs/panel/applications/) rendelkezik egy ideiglenes hivatkozási lehetőséggel, amely lehetővé teszi a felhasználó számára, hogy tesztelje a webhelyet a szerver IP-címéről, mielőtt megváltoztatná a DNS-t. A proxydomainek az `.openpanel.org` által tárolt API aldomainjei.

## Használja a Self-hosted Remote API-t

Ha több OpenPanel-kiszolgálója van, és saját tartományát szeretné használni az ideiglenes hivatkozásokhoz, kijelölhet egy kiszolgálót az API hosztolására, és beállíthatja, hogy az OpenPanel-kiszolgálók ezt az API-t használják.

Ehhez állítsa be az [Ideiglenes hivatkozások szolgáltatást az OpenPanel szerverhez] (https://github.com/stefanpejcic/OpenPanel/blob/main/services/proxy/README.md) egy szerveren, adja hozzá a domaint, majd frissítse a [temporary_links](https://dev.open/panel.com/yourf-cli)#temporaryf.com/ OpenPanel szerverek az új példány használatához:

```
opencli config update temporary_links "https://preview.openpanel.org/index.php"
```
