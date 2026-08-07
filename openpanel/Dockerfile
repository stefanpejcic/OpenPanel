FROM --platform=linux/$TARGETARCH alpine:3.22

ARG TARGETARCH

LABEL maintainer="stefan@pejcic.rs"
LABEL author="Stefan Pejcic"

WORKDIR /

# Runtime dependencies
RUN apk add --no-cache \
    bash \
    grep \
    procps \
    coreutils \
    findutils \
    wget \
    curl \
    mariadb-client \
    zip \
    unzip \
    jq \
    podman \
    podman-compose \
    tzdata
    gawk \
    tar \
    gzip \
    acl \
    attr \
    shadow \
    quota-tools

# OpenCLI
COPY opencli /usr/local/bin/opencli
RUN chmod +x /usr/local/bin/opencli

# Platform-specific binaries
COPY encoded-${TARGETARCH}/ /

# Podman socket
ENV CONTAINER_HOST=unix:///run/podman/podman.sock
ENV DOCKER_HOST=unix:///run/podman/podman.sock

# Port
EXPOSE 2083

# Entrypoint
CMD ["/openpanel"]
