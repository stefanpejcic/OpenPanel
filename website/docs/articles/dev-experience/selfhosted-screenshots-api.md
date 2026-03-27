# Screenshots API

[OpenPanel SiteManager](/docs/panel/applications/) displays a website screenshot for each website. These screenshots are by default locally generated and stored on your server.

## Remote API

When installing OpenPanel on your server, you can use the `--screenshots=<url>` flag to use a custom api instead.

Later, you can update the setting using: `opencli config update screenshots <url_of_your_remote_service>`


This is recommended for small servers, as the local screenshots api will use additional 1GB of storage.

To setup your own screenshots API:

1. Deploy: [![Deploy with Vercel](https://vercel.com/button)](https://vercel.com/new/clone?repository-url=https%3A%2F%2Fgithub.com%2Fstefanpejcic%2Fscreenshot-v2%2F&project-name=openpanel-screenshots-api&repository-name=openpanel-screenshots-api) [![Deploy to netlify Vercel](https://www.netlify.com/img/deploy/button.svg)](https://app.netlify.com/start/deploy?repository=https%3A%2F%2Fgithub.com%2Fstefanpejcic%2Fscreenshot-v2%2F&project-name=openpanel-screenshots-api&repository-name=openpanel-screenshots-api)
2. Configure your custom domain
3. Update OpenPanel to use it, example if domain is `screenshots-v2.openpanel.com` then run:
   ```
   opencli config update screenshots http://screenshots-v2.openpanel.com/api/screenshot
   ```
