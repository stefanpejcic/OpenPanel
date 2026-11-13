# openpanel/caddy-coraza: caddy with the coraza WAF and per-ip/per-site rate limiting
# the builder runs on the build host's arch and cross-compiles caddy for each target, so arm64 builds need no qemu
ARG CADDY_VERSION=2.11.9
# builder uses the newest go patch release instead of the one baked into caddy:*-builder, so go stdlib CVEs are fixed on every rebuild
ARG GO_VERSION=1.27

FROM --platform=$BUILDPLATFORM docker.io/library/golang:${GO_VERSION}-alpine AS builder
ARG CADDY_VERSION
ARG TARGETOS
ARG TARGETARCH
RUN apk add --no-cache git ca-certificates \
 && go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    xcaddy build v${CADDY_VERSION} --output /usr/bin/caddy \
    --with github.com/corazawaf/coraza-caddy/v2@v2.6.2 \
    --with github.com/mholt/caddy-ratelimit@v0.1.1-0.20260612195517-5625512f24f6

FROM docker.io/library/caddy:${CADDY_VERSION}
# pull patched alpine packages (musl, zlib, brotli, curl, ...) that landed after the upstream caddy image was built
RUN apk upgrade --no-cache
COPY --from=builder /usr/bin/caddy /usr/bin/caddy
