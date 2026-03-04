---
sidebar_position: 5
---

# nyelvek (nyelvek)

Kezelje az OpenPanel felhasználók számára elérhető nyelveket.

![openadmin admin panel locales](/img/admin/locales.png)

## Telepítse a nyelvi beállítást

Alapértelmezés szerint csak az **EN** területi beállítás van telepítve. Más területi beállítások engedélyezéséhez először azokat kell telepíteni.

<Tabs>
<TabItem value="openadmin-install-locale" label="OpenAdminnal" alapértelmezett>

A nyelvi beállítás telepítéséhez lépjen az **OpenAdmin > Beállítások > Területi beállítások** elemre, és kattintson a **Telepítés** gombra a kívánt területi beállítás mellett.

</TabItem>
<TabItem value="CLI-install-locale" label="OpenCLI-vel">

Ha a terminálról szeretne nyelvi beállítást telepíteni, használja a [Github] területi előtagját (https://github.com/stefanpejcic/openpanel-translations/tree/main), és futtassa:

```bash
opencli locale <LOCALE_HERE>
```

Példa: Telepítse a török ​​nyelvet:

```bash
opencli locale tr-tr
```

Példa: Telepítsen egyszerre több területi beállítást:

```bash
opencli locale sr-sr tr-tr zh-cn
```

</TabItem>
</Tabs>


## Alapértelmezett nyelv


<Tabs>
<TabItem value="openadmin-default-locale" label="OpenAdminnal" alapértelmezett>

Ha egy adott területi beállítást szeretne alapértelmezettként megadni, lépjen az **OpenAdmin > Beállítások > Területi beállítások** elemre, és kattintson a kívánt területi beállítás melletti **Alapértelmezett beállítás** gombra.

</TabItem>
<TabItem value="CLI-default-locale" label="Terminálról">

Alapértelmezett területi beállítás megadásához a terminálról:

```bash
echo <LOCALE_HERE> > /etc/openpanel/openpanel/default_locale
```

Példa: Állítsa be a török ​​nyelvet alapértelmezettként:

```bash
echo tr > /etc/openpanel/openpanel/default_locale
```

</TabItem>
</Tabs>

:::info
Az alapértelmezett módosítás **nem** frissíti automatikusan a meglévő felhasználók beállításait; böngészőbeállításai és fiókkonfigurációik élveznek elsőbbséget.
További részletekért lásd: [Útmutatók > Az alapértelmezett terület beállítása](/docs/articles/accounts/default-user-locales/#setting-the-default-locale).
:::

## Területi beállítás szerkesztése

A nyelvi beállítás szerkesztéséhez kattintson a táblázatban a mellette lévő GitHub ikonra. Ez megnyitja a forrást a GitHubon, ahol elágazhatja a tárat és szerkesztheti a fordítási fájlt.

## Hozzon létre egy nyelvi beállítást

Új nyelvi beállítás létrehozása:

1. Fork [a fordítások tárháza](https://github.com/stefanpejcic/openpanel-translations/).
2. Másolja az "en_us" mappát egy új területi mappába, például "es_es" mappába.
3. Fordítsa le az "messages.pot" fájlt.
4. Nyújtson le lehívási kérelmet.
