# Névszerverek konfigurálása

Bármely tartomány hozzáadása előtt fontos először névszervereket létrehozni az érvényes DNS-zónafájlok és a megfelelő terjedés biztosítása érdekében.

A névszerverek konfigurálása két fontos lépésből áll:

1. **Hozzon létre privát névszervereket (DNS rekordok összeragasztása)** a domainhez a domainregisztrátoron keresztül.
2. **Adja hozzá a névszervereket** az OpenPanel konfigurációjához.

Oktatóanyagok a népszerű domainszolgáltatóknak

- [Cloudflare](https://developers.cloudflare.com/dns/additional-options/custom-nameservers/zone-custom-nameservers/)
- [GoDaddy](https://uk.godaddy.com/help/add-custom-hostnames-12320)
- [NameCheap](https://www.namecheap.com/support/knowledgebase/article.aspx/768/10/how-do-i-register-personal-nameservers-for-my-domain/#:~:text=Click%20on%20the%20Manage%20option,5.

Ha névszervereket szeretne hozzáadni az OpenPanelhez, nyissa meg az **OpenAdmin > Beállítások > OpenPanel** elemet, írja be a névszervereket az „ns1”, „ns2”, „ns3”, „ns4” mezőkbe, majd kattintson a **Változtatások mentése** gombra.

Alternatív megoldásként beállíthatja a névszervereket terminálparancsokkal:

```bash
opencli config update ns1 your_ns1.domain.com
opencli config update ns2 your_ns2.domain.com
opencli config update ns3 your_ns3.domain.com
opencli config update ns4 your_ns4.domain.com
```

Jelenleg legfeljebb 4 névszervert adhat hozzá.

:::info
A névszerverek létrehozása után akár 12 óráig is eltarthat, amíg a rekordok globálisan elérhetővé válnak. Az állapot figyeléséhez használjon olyan eszközt, mint a [whatsmydns.net](https://www.whatsmydns.net/).
:::
