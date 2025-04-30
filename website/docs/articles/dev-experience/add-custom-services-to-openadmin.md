# Add custom services to OpenAdmin

Custom services can be added to OpenAdmin to be run and monitored from the admin panel.

In this example, We will add [netdata](https://learn.netdata.cloud/docs/netdata-agent/installation/docker) docker image.

## Add service

```
cd /root
nano docker-compose.yml
```

and add the netdata section:

```
  netdata:
    image: netdata/netdata
    container_name: netdata
    pid: host
    network_mode: host
    restart: unless-stopped
    cap_add:
      - SYS_PTRACE
      - SYS_ADMIN
    security_opt:
      - apparmor:unconfined
    environment:
      - DO_NOT_TRACK=1
      - NETDATA_CLOUD_ENABLE=no
    volumes:
      - ./netdata/config:/etc/netdata
      - ./netdata/lib:/var/lib/netdata
      - ./netdata/cache:/var/cache/netdata
      - /:/host/root:ro,rslave
      - /etc/passwd:/host/etc/passwd:ro
      - /etc/group:/host/etc/group:ro
      - /etc/localtime:/etc/localtime:ro
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /etc/os-release:/host/etc/os-release:ro
      - /var/log:/host/var/log:ro
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - /run/dbus:/run/dbus:ro
```

save and exit the file. then start the service:

```
docker --context=default compose up -d netdata
```

now, I suggest just whitelisting your IP on the firewall instead of opening the port to the internet.

Then open the IP:19999 in your browser:

![screenshot](https://i.postimg.cc/CFRC5zG9/netdata.png)


## Monitor service

Next step is to make the netdata service available from **OpenAdmin > Services** page so we can view its status, start/stop or restart it.

For this, we need to add the netdata service in `/etc/openpanel/openadmin/config/services.json` file:

```
    {
        "name": "Netdata",
        "type": "docker",
        "on_dashboard": false,
        "real_name": "netdata"
    },
```
`real_name` is the service/container name, and `name` is the human-readable name displayed in OpenAdmin services.

Save the file and refresh the page in OpenAdmin - service is immeditelly visible and we can start/stop it from the panel.

![services](https://i.postimg.cc/qpTMgY8r/2025-04-30-16-33.png)

We can also receive notifications when the service is not running. For this edit the file: `/etc/openpanel/openadmin/config/notifications.ini` and add service name (`netdata`) under services:

```
services=panel,admin,caddy,mysql,docker,csf,netdata
```

save the file and exit. Sentinel service will now also check if the netdata container is running and if not, send alerts to **OpenAdmin > Notifications** and if emails are enabled then to email as well.

