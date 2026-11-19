#!/bin/bash

# Sentinel firewall (csf) 15.17: CVE-2026-65638 and CVE-2026-65639 (command injection), upgrades before 15.17 never replaced the modules in /usr/local/csf/lib/Sentinel/ so 15.15 and 15.16 fixes didn't land,
# and installs from before 15.16 check an address that no longer gets releases, so csf -u doesn't see it: https://github.com/sentinelfirewall/sentinel/releases/tag/15.17
CSF_WANT="15.17"
CSF_URL="https://raw.githubusercontent.com/sentinelfirewall/sentinel/main/csf.tgz"
csf_version() { /usr/sbin/csf -v 2>/dev/null | grep -oE 'v[0-9]+\.[0-9]+' | head -1 | tr -d v; }
if [ -x /usr/sbin/csf ] && [ -d /etc/csf ]; then
    have=$(csf_version)
    if [ -n "$have" ] && [ "$(printf '%s\n%s\n' "$have" "$CSF_WANT" | sort -V | head -n1)" != "$CSF_WANT" ]; then
        echo "Updating Sentinel firewall from $have to $CSF_WANT..."
        tmp=$(mktemp -d)
        backup="/root/csf-backup-$(date +%Y%m%d%H%M%S)"
        mkdir -p "$backup"
        cp -a /etc/csf "$backup/etc-csf"
        cp -a /usr/local/csf/bin/csfpre.sh /usr/local/csf/bin/csfpost.sh "$backup/" 2>/dev/null
        if wget --timeout=30 --tries=3 -q -O "$tmp/csf.tgz" "$CSF_URL" && tar -xzf "$tmp/csf.tgz" -C "$tmp" && [ -f "$tmp/csf/install.sh" ]; then
            ( cd "$tmp/csf" && sh install.sh ) > "$backup/install.log" 2>&1
            # the container rules OpenPanel adds after every csf reload live in these two, an install must not lose them
            for f in csfpre.sh csfpost.sh; do
                if [ -f "$backup/$f" ] && ! cmp -s "$backup/$f" "/usr/local/csf/bin/$f"; then
                    cp -a "$backup/$f" "/usr/local/csf/bin/$f"
                    echo "Put back /usr/local/csf/bin/$f"
                fi
            done
            now=$(csf_version)
            if [ "$now" = "$CSF_WANT" ]; then
                /usr/sbin/csf -ra >/dev/null 2>&1 || /usr/sbin/csf -r >/dev/null 2>&1
                echo "Sentinel firewall updated to $now, the previous config is in $backup"
            else
                echo "Sentinel firewall still reports ${now:-no version} after the install, see $backup/install.log - update by hand: cd /root && wget $CSF_URL && tar -xzf csf.tgz && cd csf && sh install.sh"
            fi
        else
            echo "Could not download $CSF_URL, Sentinel firewall not updated - update by hand: cd /root && wget $CSF_URL && tar -xzf csf.tgz && cd csf && sh install.sh"
        fi
        rm -rf "$tmp"
    else
        echo "Sentinel firewall is already on ${have:-an unknown version}, nothing to update."
    fi
fi

# PHP security releases of 19 November 2026 (critical, all supported versions): https://www.php.net/archive/2026.php#2026-11-18-1
# pulls the new build of every php image an account uses and recreates only the running php-fpm containers on it, nobody moves to another image (--no-php-move)
# docker-update recreates a container only once its image really has the fixed PHP (PHP_FIXED), so running this before upstream rebuilt an image changes nothing
# it runs in its own unit: on a server with many accounts it takes longer than opencli update gives this script, newest versions go first
DOCKER_UPDATE=/usr/local/opencli/docker/update.sh
if [ -x "$DOCKER_UPDATE" ] && grep -q -- "--no-php-move" "$DOCKER_UPDATE"; then
    php_images=$(grep -hoE "^[[:space:]]+image:[[:space:]]*['\"]?(docker\.io/)?(shinsenter/php:[0-9.]+-fpm|openpanel/php-fpm:[0-9.]+)" /home/*/docker-compose.yml 2>/dev/null \
        | sed -E "s/^[[:space:]]+image:[[:space:]]*['\"]?(docker\.io\/)?//" | sort -u | sort -t: -k2 -rV | tr '\n' ' ')
    if [ -n "$php_images" ]; then
        PHP_LOG=/var/log/openpanel/updates/2.0.27-php.log
        mkdir -p "$(dirname "$PHP_LOG")"
        PHP_RUN='for image in '"$php_images"'; do echo "Updating $image..."; opencli docker-update "$image" -y --no-php-move || echo "Failed to update $image, run: opencli docker-update $image -y --no-php-move"; done; echo "PHP update finished."'
        echo "Updating PHP images in the background ($php_images), follow it with: tail -f $PHP_LOG"
        if ! systemd-run --unit="openpanel-php-update-$(date +%s)" --collect --quiet --setenv=HOME=/root \
                -p StandardOutput="append:$PHP_LOG" -p StandardError="append:$PHP_LOG" \
                /bin/bash -c "$PHP_RUN" 2>/dev/null; then
            setsid nohup /bin/bash -c "$PHP_RUN" >> "$PHP_LOG" 2>&1 </dev/null &
        fi
    fi
else
    echo "opencli docker-update has no --no-php-move, PHP images not updated - run opencli update --cli first, then this update again."
fi
