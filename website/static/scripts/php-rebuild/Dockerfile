# Rebuilds an image's PHP from a newer upstream patch release, in place, for a PHP security release
# that lands before the image's maintainers rebuild (PHP2Shell, CVE-2026-109618 and CVE-2026-109619).
# Works on any docker-library based image (php, phpmyadmin, roundcube/roundcubemail, shinsenter/php):
# the image's own configure flags (`php-config --configure-options`), the source checked against
# php.net's SHA-256 and the release managers' GPG keys ($GPG_KEYS, set by the base image). A patch
# release keeps the extension API, so the image's own shared extensions and config load unchanged.
#
#   podman build --build-arg BASE=<image> --build-arg PHP_NEW=<version> \
#     --build-arg PHP_NEW_SHA256=<sha256 of php-<version>.tar.xz> -t <tag> .
#
# php-rebuild.sh fills in the version and checksum. Guide:
# https://openpanel.com/docs/articles/security/php2shell-patch-php-before-images-are-rebuilt
ARG BASE
FROM ${BASE}
ARG PHP_NEW
ARG PHP_NEW_SHA256
RUN set -eux; \
	apt-mark showmanual > /usr/src/apt-manual.saved; \
	apt-get update; \
	apt-get install -y --no-install-recommends $PHPIZE_DEPS ca-certificates curl gnupg xz-utils \
		libargon2-dev libcurl4-openssl-dev libonig-dev libreadline-dev libsodium-dev libsqlite3-dev \
		libssl-dev libxml2-dev zlib1g-dev \
		$(case "$(php-config --configure-options)" in *--with-apxs2*) echo apache2-dev;; esac); \
	cd /usr/src; \
	curl -fsSL -o php.tar.xz "https://www.php.net/distributions/php-$PHP_NEW.tar.xz"; \
	curl -fsSL -o php.tar.xz.asc "https://www.php.net/distributions/php-$PHP_NEW.tar.xz.asc"; \
	echo "$PHP_NEW_SHA256 *php.tar.xz" | sha256sum -c -; \
	export GNUPGHOME="$(mktemp -d)"; \
	for key in $GPG_KEYS; do gpg --batch --keyserver keyserver.ubuntu.com --recv-keys "$key"; done; \
	gpg --batch --verify php.tar.xz.asc php.tar.xz; \
	gpgconf --kill all; rm -rf "$GNUPGHOME"; \
	mkdir php-new; tar -Jxf php.tar.xz -C php-new --strip-components=1; cd php-new; \
# --with-pear only fetches PEAR from the network at install; the image's PEAR stays as it is.
	opts="$(php-config --configure-options | sed 's/ build_alias=.*//; s/ --with-pear//')"; \
	CFLAGS="$PHP_CFLAGS" CPPFLAGS="$PHP_CPPFLAGS" LDFLAGS="$PHP_LDFLAGS" \
		PHP_BUILD_PROVIDER='https://github.com/docker-library/php' PHP_UNAME='Linux - Docker' \
		./configure $opts; \
	make -j"$(nproc)"; \
	find -type f -name '*.a' -delete; \
	make install; \
	cd /; rm -rf /usr/src/php-new
# docker-library's own cleanup, in its own layer so a failure here does not cost the compile: keep
# every library something in /usr/local links against (diversion lines are skipped, as upstream does).
RUN set -eux; \
	apt-mark auto '.*' > /dev/null; \
	[ ! -s /usr/src/apt-manual.saved ] || apt-mark manual $(cat /usr/src/apt-manual.saved); \
	rm /usr/src/apt-manual.saved; \
	find /usr/local -type f -executable -exec ldd '{}' ';' \
		| awk '/=>/ { so = $(NF-1); if (index(so, "/usr/local/") == 1) { next }; gsub("^/(usr/)?", "", so); printf "*%s\n", so }' \
		| sort -u | xargs -r dpkg-query --search | awk 'sub(":$", "", $1) { print $1 }' | sort -u | xargs -r apt-mark manual; \
	apt-get purge -y --auto-remove -o APT::AutoRemove::RecommendsImportant=false; \
	rm -rf /var/lib/apt/lists/*; \
	php -r 'exit(PHP_VERSION === getenv("PHP_NEW") ? 0 : 1);'; \
# An extension that no longer loads is a startup warning, not an exit code.
	if php -m 2>&1 | grep -iE 'warning|unable to'; then exit 1; fi
LABEL php-rebuilt="${PHP_NEW}"
