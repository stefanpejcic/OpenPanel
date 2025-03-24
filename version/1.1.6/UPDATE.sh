#!/bin/bash

wget -O /etc/openpanel/apache/httpd.conf https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/apache/httpd.conf




DOLE:


```
LoadModule proxy_module modules/mod_proxy.so
LoadModule proxy_http_module modules/mod_proxy_http.so
LoadModule proxy_fcgi_module modules/mod_proxy_fcgi.so
```

u /home/*/httpd.conf


nadje 

```
Listen 80
```

cd /gome/USER 

docker --context USER compose down apache
docker --context USER compose up -d apache
