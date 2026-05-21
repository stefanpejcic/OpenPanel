#!/bin/bash

# ENABLE GZIP
DIR="/etc/openpanel/caddy/domains"
mkdir -p /tmp/domain_backups_1.7.59

for file in "$DIR"/*.conf; do
  [ -f "$file" ] || continue

  cp "$file" /tmp/domain_backups_1.7.59/

  awk '
  function reset_block() {
    mode=""
    after_header=0
    encode_present=0
    inserted=0
  }

  BEGIN {
    reset_block()
  }

  # Detect HTTP block header
  /^http:\/\/.*\{$/ {
    print
    mode="http"
    after_header=1
    encode_present=0
    inserted=0
    next
  }

  # Detect HTTPS block header
  /^https:\/\/.*\{$/ {
    print
    mode="https"
    after_header=1
    encode_present=0
    inserted=0
    next
  }

  {
    print

    # detect existing encode line
    if ($0 ~ /encode zstd gzip/) {
      encode_present=1
    }

    # insert right after header if missing
    if (after_header && !inserted) {
      if (!encode_present) {
        print "  encode zstd gzip"
      }
      inserted=1
      after_header=0
    }

    # reset when block ends
    if ($0 ~ /^}/) {
      reset_block()
    }
  }
  ' "$file" > "$file.tmp" && mv "$file.tmp" "$file"

done
