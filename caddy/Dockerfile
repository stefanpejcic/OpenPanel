# openpanel/caddy-coraza: caddy with the coraza WAF and per-ip/per-site rate limiting
# the builder runs on the build host's arch and cross-compiles caddy for each target, so arm64 builds need no qemu
ARG CADDY_VERSION=2.11.7

FROM --platform=$BUILDPLATFORM docker.io/library/caddy:${CADDY_VERSION}-builder AS builder
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH xcaddy build \
    --with github.com/corazawaf/coraza-caddy/v2@v2.6.2 \
    --with github.com/mholt/caddy-ratelimit@v0.1.1-0.20260612195517-5625512f24f6

FROM docker.io/library/caddy:${CADDY_VERSION}
COPY --from=builder /usr/bin/caddy /usr/bin/caddy
