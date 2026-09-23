# CorazaWAF

`opencli waf` command can be used to enable/disable CorazaWAF on the server, enable/disable WAF per domain, view rules, view statistics, etc.

## Status

```bash
opencli waf status
```


## Enable
To enable the WAF and ensure it is used for new domains, run:

```bash
opencli waf enable
```


## Disable
To completely disable the WAF for all existing domains and prevent it from being applied to new domains, run:

```bash
opencli waf disable
```



## Domain


### Status
Check if CorazaWAF is enabled for domain:

```bash
opencli waf domain <DOMAIN_NAME>
```

### Enable
Enable WAF for a domain:

```bash
opencli waf domain <DOMAIN_NAME> enable
```

### Disable
Disable WAF for a domain:

```bash
opencli waf domain <DOMAIN_NAME> disable
```

## Update

Update OWASP CRS:

```bash
opencli waf update
```

### Update Log

View update log for OWASP CRS:

```bash
opencli waf update log
```

## IDs

List all rule IDs from all enabled sets:

```bash
opencli waf ids
```

## Tags

List all rule tags from all enabled sets:

```bash
opencli waf tags
```

## Stats

Get stats from the log.

### Country
```bash
opencli waf stats country
```

Example:
```
root@apolo2:/home/pcx3# opencli waf stats country
    2000 US
    1501 DE
    1033 BE
     809 SG
     705 SC
     606 ID
     603 IN
     505 NL
     500 RS
     209 CH
     109 FR
     108 PY
     102 BG
      70 CA
```

### IP
```bash
opencli waf stats ip
```

Example:
```
root@apolo2:/home/pcx3# opencli waf stats ip
   11069 154.83.103.102
    9038 154.83.103.15
    9038 154.83.103.111
    9038 154.83.103.106
    7007 154.83.103.19
    7007 154.83.103.109
    4069 154.83.103.202
    4069 154.83.103.10
    4062 108.162.221.121
    4055 154.83.103.14
```

### Path
```bash
opencli waf stats path
```

Example:
```
root@apolo2:/home/pcx3# opencli waf stats path
    1308 /modules/younitedpay/logo.png
     707 /.git/config
     608 /.env
     304 /.git/HEAD
     301 /wp-content/plugins/WordPressCore/include.php
     301 /php/how-to-fix-file_get_contents-wrapper-is-disabled-error-in-php/
     204 /wp-includes/images/include.php
     203 /wp-includes/widgets/include.php
     201 /.aws/credentials
     109 /wp-content/themes/include.php
```

### Agent
```bash
opencli waf stats agent
```

Example:
```
root@apolo2:/home/pcx3# opencli waf stats agent
   10400  Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36
    1909  Mozlila/5.0 (Linux; Android 7.0; SM-G892A Bulid/NRD90M; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/60.0.3112.107 Moblie Safari/537.36
    1308  Go-http-client/2.0
     809  Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 Edg/134.0.0.0
     702  Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36
     610  Mozilla/5.0 (Windows NT 10.0; WOW64) Gecko/20060309 Firefox/22.0
     600  Mozilla/5.0 (Macintosh; Intel Mac OS X 11_8_4) Gecko/20052101 Firefox/22.0
     544  Mozilla/5.0
     371  Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36 OPR/117.0.0.0
     350  Mozilla/5.0 (Windows; U; MSIE 9.0; Windows NT 10.0; Trident/5.0; WOW64)
```

### Hourly
```bash
opencli waf stats hourly
```

Example:
```
root@apolo2:/home/pcx3# opencli waf stats hourly
     14 2025/04/14 00
     14 2025/04/14 10
      7 2025/04/14 19
      7 2025/04/15 18
      7 2025/04/15 22
      7 2025/04/16 00
     14 2025/04/16 10
     14 2025/04/16 11
      7 2025/04/17 04
      7 2025/04/17 07
      7 2025/04/17 12
      7 2025/04/18 01
      7 2025/04/18 06
      7 2025/04/18 07
      7 2025/04/18 10
      7 2025/04/18 19
    469 2025/04/20 07
     14 2025/04/20 08
      7 2025/04/20 14
     14 2025/04/20 18
      7 2025/04/21 15
     14 2025/04/22 05
```

### Request
```bash
opencli waf stats request
```

Example:
```
root@apolo2:/home/pcx3# opencli waf stats request
    1803 POST /xmlrpc.php HTTP/1.1
    1408 POST /xmlrpc.php HTTP/2.0
    1038 GET /modules/younitedpay/logo.png HTTP/2.0
     608 GET /.git/config HTTP/1.1
     641 GET /.env HTTP/1.1
     553 POST /api/discussions/137 HTTP/1.1
     334 POST /api/discussions/128 HTTP/1.1
     324 GET /.git/HEAD HTTP/1.1
     301 POST /wp-comments-post.php HTTP/1.1
     214 POST //xmlrpc.php HTTP/2.0
```
