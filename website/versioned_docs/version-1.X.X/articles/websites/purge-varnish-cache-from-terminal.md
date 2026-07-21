# Purging Varnish Cache

If Varnish is enabled for a website, cache can be purged from the terminal or directly through OpenPanel.

## Purge from Terminal

You can purge cache entries manually using the `curl` command.

### Purge a Specific URL

This removes a single cached object (for a specific page, image, or file):
```bash
curl -X PURGE -H "Host: example.com" http://example.com/wp-content/uploads/test.jpg
```

### Purge a Specific Website

This removes **all cached objects** for one website (domain):
```
curl -X PURGE -H "Host: example.com" -H "X-Purge-Method: regex" http://example.com/.*
```

### Purge All Websites

To remove all cached objects across all domains handled by Varnish:
```
curl -X BAN http://localhost/.* 
```

:::info 
Alternatively, you can simply restart the Varnish service (see below).
:::

## Purge from Crons

To purge cache from a cronjob, simply use the above `curl` commands and make sure to use varnish/webserver or php for container when creating Crons on OpenPanel.

## Purge from WordPress

To purge a cache for a WordPress website, use the [CLP Varnish Cache plugin](https://wordpress.org/plugins/clp-varnish-cache/)

To configure the plugin:
1. Login to wp-admin and go to **Settings → CLP Varnish Cache**.
2. Set **varnish** as 'Varnish Server'then **Enable** again.

To purge cache using the plugin:
1. Login to wp-admin and go to **Settings → CLP Varnish Cache**.
2. Click on **Purge Entire Cache** button.

## Purge from Terminal

You can purge cache entries manually using the `curl` command.

---

### Notes

* The `Host` header is required because Varnish stores cache separately for each domain name.
* The `X-Purge-Method: regex` header allows purging multiple URLs using pattern matching.
* Purge requests must come from localhost (the website itself or php/webserver containers).
