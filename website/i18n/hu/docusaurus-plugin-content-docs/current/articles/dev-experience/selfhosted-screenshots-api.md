# Screenshots API

Az [OpenPanel SiteManager](/docs/panel/applications/) megjelenít egy webhely képernyőképet minden webhelyről. Ezeket a képernyőképeket a mi hostolt API-nkkal hozzuk létre.

## Telepítse az API-t

Amikor telepíti az OpenPanel-t a szerverére, használhatja a `--screenshots=<url>` jelzőt, hogy helyette egyéni API-t használjon.

Ez néhány TB tárhellyel és több száz webhelytel rendelkező megosztott szerverekhez ajánlott.

Saját képernyőképek API beállításához:

1. Telepítés Vercelen: [![Deploy with Vercel](https://vercel.com/button)](https://vercel.com/new/clone?repository-url=https%3A%2F%2Fgithub.com%2Fstefanpejcic%2Fscreenshot-v2%2F&project-name=openpanel-screenshots-api&repository-name)
2. [Egyéni domain hozzáadása](https://vercel.com/docs/domains/working-with-domains/add-a-domain)
3. Frissítse az OpenPanel alkalmazást a használatához:
   ```
opencli konfiguráció frissítés képernyőképek http://screenshots-v2.openpanel.com/api/screenshot
   ```
