# Screenshots API

[OpenPanel SiteManager](/docs/panel/applications/) displays a website screenshot for each website. These screenshots are generated using our hosted API.

## Deploy your API

When installing OpenPanel on your server, you can use the `--screenshots=<url>` flag to use a custom api instead.

This is recommended for shared servers with a few TBs of storage and hundreds of websites.

To setup your own screenshots API:

1. Deploy on Vercel: [![Deploy with Vercel](https://vercel.com/button)](https://vercel.com/new/clone?repository-url=https%3A%2F%2Fgithub.com%2Fstefanpejcic%2Fscreenshot-v2%2F&project-name=openpanel-screenshots-api&repository-name=openpanel-screenshots-api)
2. [Add a custom domain](https://vercel.com/docs/domains/working-with-domains/add-a-domain)
3. Update OpenPanel to use it: 
   ```
   opencli config update screenshots http://screenshots-v2.openpanel.com/api/screenshot
   ```
