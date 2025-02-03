FROM caddy:2.9.1-builder AS builder

LABEL maintainer="stefan@pejcic.rs"
LABEL author="Stefan Pejcic"
   
RUN xcaddy build \
    --with github.com/corazawaf/coraza-caddy/v2 \
    --with github.com/caddyserver/caddy/v2/modules/caddy.logging.encoders.single_field

FROM caddy:2.9.1

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
