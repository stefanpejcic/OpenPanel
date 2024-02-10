---
sidebar_position: 5
---

# Troubleshooting



## Logs

OpenPanel generates the following logs:


| Log file | Description |
|----------|-------------|
|`/var/log/openpanel/user/access.log`|OpenPanel access log|
|`/var/log/openpanel/user/error.log`|OpenPanel error log|
|`/var/log/openpanel/admin/access.log`|OpenAdmin access log|
|`/var/log/openpanel/admin/error.log`|OpenAdmin error log|

Logs for services:

| Log file | Description |
|----------|-------------|
|`/var/log/nging/access.log`|Nginx access log|
|`/var/log/nging/error.log`|Nginx error log|
|`/var/log/mysql/error.log`|MySQL error log|
|`journalctl -u named`|Named service logs|
|`/var/log/ufw.log`|UFW log|
|`docker logs $USERNAME`|Docker logs for user|

```
multitail /var/log/nginx/access.log -I /var/log/nginx/error.log
```
