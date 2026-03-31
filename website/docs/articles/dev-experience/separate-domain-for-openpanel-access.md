# Set a separate domain for OpenPanel UI

To set separate domain just for the OpenPanel UI, for example `pejcic.rs`:

1. Add the domain to Caddyfile

  ```bash
  DOMAIN="pejcic.rs" && cat << EOF >> "/etc/openpanel/caddy/Caddyfile"
  
  # START USERPANEL DOMAIN #
  $DOMAIN {
      reverse_proxy localhost:2083
  }
  
  http://$DOMAIN {
      reverse_proxy localhost:2083
  }
  # END USERPANEL DOMAIN #
  EOF
  ```

2. Create an empty file with the domain name so SSL can be generated and renewed:
  ```bash
  touch /etc/openpanel/caddy/domains/pejcic.rs.conf
  ```

3. Restart services
  ```bash
  docker restart caddy openpanel
  ```
