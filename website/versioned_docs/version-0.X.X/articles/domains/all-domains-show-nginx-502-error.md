# All Domains show Nginx 502 error

This happens when a third party user/service restarted docker and the floating ip service failed to reload the `/etc/hosts` file with new docker private IP addresses.

To resolve this issue, run:

```bash
opencli server-recreate_hosts
```

Then restart Nginx service:

- soft restart (reload without downtime):
  ```bash
  docker exec nginx bash -c "nginx -t && nginx -s reload"
  ```
- hard restart (stop and start again):
  ```bash
  docker restart nginx
  ```
