#!/bin/bash
# php-rebuild.sh IMAGE TAG
#
# Rebuilds the PHP inside IMAGE from the newest php.net release of its own branch and tags the result
# TAG, for a PHP security release that lands before the image's maintainers rebuild it (PHP2Shell,
# CVE-2026-109618 and CVE-2026-109619). IMAGE is any docker-library based image: php, phpmyadmin,
# roundcube/roundcubemail, shinsenter/php. Only PHP changes: the image's own configure flags, its
# extensions and its config are kept. Needs podman (or CONTAINER_ENGINE=docker), curl, jq and the
# Dockerfile next to this script. Debian-based images only (the Dockerfile uses apt-get).
#
# Guide: https://openpanel.com/docs/articles/security/php2shell-patch-php-before-images-are-rebuilt
set -euo pipefail
image=${1:?usage: php-rebuild.sh IMAGE TAG}
tag=${2:?usage: php-rebuild.sh IMAGE TAG}
dir=$(cd "$(dirname "$0")" && pwd)
engine=${CONTAINER_ENGINE:-podman}

cur=$("$engine" run --rm --entrypoint php "$image" -r 'echo PHP_VERSION;')
branch=${cur%.*}
release=$(curl -fsS "https://www.php.net/releases/index.php?json&version=$branch")
new=$(jq -r '.version' <<<"$release")
sha=$(jq -r '[.source[] | select(.filename | endswith(".tar.xz")) | .sha256][0]' <<<"$release")
[[ $new =~ ^[0-9]+\.[0-9]+\.[0-9]+$ && $sha =~ ^[0-9a-f]{64}$ ]] || { echo "php.net has no release for PHP $branch"; exit 1; }
if [[ $(printf '%s\n' "$cur" "$new" | sort -V | tail -n 1) == "$cur" ]]; then
  echo "$image already has PHP $cur (newest $branch release: $new)"
  exit 0
fi

echo "$image: PHP $cur -> $new"
"$engine" build --build-arg BASE="$image" --build-arg PHP_NEW="$new" --build-arg PHP_NEW_SHA256="$sha" \
  -t "$tag" -f "$dir/Dockerfile" "$dir"
echo "$tag: PHP $("$engine" run --rm --entrypoint php "$tag" -r 'echo PHP_VERSION;')"
