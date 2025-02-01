# Self-hosted Temporary Links API for OpenPanel SiteManager

[OpenPanel SiteManager](/docs/panel/applications/) has a temporary-links option that allows user to test website from the server IP, prior to changing DNS. Proxy domains are subdomains on our hosted api `.openpanel.org`.

## Use Self-hosted Remote API

If you have multiple OpenPanel servers and want to use your own domain for temporary links, you can designate one server to host the api, and set OpenPanel servers to use that API instead.

To do this, set up the [Temporary Links service for OpenPanel server](https://github.com/stefanpejcic/OpenPanel/blob/main/services/proxy/README.md) on one server, add the domain to it, and then update the [temporary_links](https://dev.openpanel.com/cli/config.html#temporary-links) service on your OpenPanel servers to use the new instance:

```
opencli config update temporary_links "https://preview.openpanel.org/index.php"
```
