# Temporary Links service for OpenPanel server

proxy service used for website preview on specific IP

## Install with Docker

Download proxy directory content and inside it run:

```bash
docker build -t openpanel_proxy . && \
docker run -d \
  -e DOMAIN="your-domain.com" \
  -e PREVIEW_DOMAIN="preview.your-domain.com" \
  -p 80:80 \
  -p 443:443 \
  --name openpanel_proxy \
  openpanel_proxy
```

- **DOMAIN** - on which the api will be available - exmaple if set to `example.net` then you would point example.net to the server and on OpenPanel set: `opencli config update temporary_links "https://example.net/index.php"`
- **PREVIEW_DOMAIN** - domain on which sub-domains will be created. - if set to `example.com` then you would point *.example.com to the server and generate SSL for it.



## Install on RockyLinux

downlaod content, follow instructions from install.sh and then run it

```
bahs install.sh
```
