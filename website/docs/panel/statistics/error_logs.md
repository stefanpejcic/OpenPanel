---
sidebar_position: 6
---

# Error Logs

The **Error Logs** page is available in your OpenPanel dashboard only if the **Error Logs** module is enabled on the server and your hosting plan includes it.

It shows errors for all of your websites in one place, so you don't have to know which service or container to check.

Errors are collected from:

* **PHP**: warnings, notices, deprecations, fatal errors (with the stack trace) and messages written with `error_log()`, from every PHP version you use
* **Web server**: errors and warnings from Apache, Nginx, OpenResty, OpenLiteSpeed or LiteSpeed. With OpenLiteSpeed or LiteSpeed, PHP runs inside the web server, so PHP errors are listed under that service instead of a PHP version.
* **Database**: errors and warnings from MySQL, MariaDB or PostgreSQL
* **HTTP 5xx responses**: requests to your domains that ended with a 500, 502, 503 or 504 response

Each entry shows the time (in your browser's timezone), the severity, the service it came from, the domain it belongs to and the message. Click a row to see the full message.

## Filtering

* **Domain**: show only errors for one of your domains. A PHP error is matched to a domain by the path of the file that caused it.
* **Service**: show only one source, for example a single PHP version.
* **Severity**: show everything, only warnings and errors, or only errors.
* **Time range**: the last hour, 6 hours, 24 hours (default), 7 days, 30 days, or a custom range.
* **Search**: show only entries whose message contains the text, for example `Fatal error`, `wp-content` or `502`.

The newest 500 entries are shown. If there are more, narrow the filters to see the rest.

## Sharing errors with support

* **Copy** puts the entries you ticked on your clipboard, or all visible entries if none are ticked.
* **Download** saves all entries that match the current filters as a `.txt` file.

Both give one line per entry with the time in UTC, severity, service, domain and message, ready to paste into a support ticket.

:::info
Applications that write their own log files are not shown here. For example, WordPress with `WP_DEBUG_LOG` enabled writes to `wp-content/debug.log`, which you can open from the File Manager.
:::
