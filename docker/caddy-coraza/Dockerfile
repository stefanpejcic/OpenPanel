FROM caddy:2.11.2-builder AS builder

LABEL maintainer="stefan@pejcic.rs"
LABEL author="Stefan Pejcic"
   
RUN xcaddy build \
    --with github.com/corazawaf/coraza-caddy/v2

FROM caddy:2.11.2

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
