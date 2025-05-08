https://hub.docker.com/r/openpanel/caddy-coraza


docker buildx build --platform linux/amd64,linux/arm64 -t openpanel/caddy-coraza:latest --push .



```
docker buildx build --platform linux/amd64 -t openpanel/caddy-coraza:amd64 --output type=docker .
docker push yourname/caddy-coraza:amd64


docker buildx build --platform linux/arm64 -t openpanel/caddy-coraza:arm64 --output type=docker .
docker push yourname/caddy-coraza:amd64

```
