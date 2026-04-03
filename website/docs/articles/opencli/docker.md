# Docker

Manage Docker settings: update docker images, set global resource limits for docker, etc.

### Images

Uses [Cup 🥤](https://github.com/sergi0g/cup) to check for image updates and displays them on **OpenPanel > Containers > Image Updates**

Check image updates for specific user:
```bash
opencli docker-images <USERNAME>
```

Example:
```bash
root@test:/usr/local/opencli# opencli docker-images s70nesto

Running Cup in context: s70nesto
Output for user s70nesto saved to /home/s70nesto/docker-data/cup/cup.json

{"images":[{"in_use":false,"parts":{"registry":"registry-1.docker.io","repository":"shinsenter/php","tag":"8.2-fpm"},"reference":"shinsenter/php:8.2-fpm","result":{"error":null,"has_update":true,"info":{"current_version":"8.2","new_tag":"8.4-fpm","new_version":"8.4","type":"version","version_update_type":"minor"}},"server":null,"time":276,"url":"https://hub.docker.com/r/shinsenter/php/tags"},{"in_use":true,"parts":{"registry":"ghcr.io","repository":"sergi0g/cup","tag":"latest"},"reference":"ghcr.io/sergi0g/cup:latest","result":{"error":null,"has_update":false,"info":null},"server":null,"time":162,"url":"https://github.com/sergi0g/cup"},{"in_use":true,"parts":{"registry":"registry-1.docker.io","repository":"mcuadros/ofelia","tag":"latest"},"reference":"mcuadros/ofelia:latest","result":{"error":null,"has_update":false,"info":null},"server":null,"time":575,"url":"https://github.com/mcuadros/ofelia"},{"in_use":false,"parts":{"registry":"registry-1.docker.io","repository":"library/mysql","tag":"8.0"},"reference":"mysql:8.0","result":{"error":null,"has_update":true,"info":{"current_version":"8.0","new_tag":"9.3","new_version":"9.3","type":"version","version_update_type":"major"}},"server":null,"time":549,"url":null},{"in_use":false,"parts":{"registry":"registry-1.docker.io","repository":"library/phpmyadmin","tag":"latest"},"reference":"phpmyadmin:latest","result":{"error":null,"has_update":true,"info":{"local_digests":["sha256:5f37deab81ddca73cb44de568ecbe0109fd738a76d614f41833f6b0788ad4012"],"remote_digest":"sha256:db901a16f662cfa7857b55353883e91ce1dc45cfadf57e9631bb44a38274e69b","type":"digest"}},"server":null,"time":550,"url":"https://github.com/phpmyadmin/docker#readme"},{"in_use":false,"parts":{"registry":"registry-1.docker.io","repository":"library/httpd","tag":"latest"},"reference":"httpd:latest","result":{"error":null,"has_update":true,"info":{"local_digests":["sha256:f6557a77ee2f16c50a5ccbb2564a3fd56087da311bf69a160d43f73b23d3af2d"],"remote_digest":"sha256:1ae8051591a5ded56e4a3d7399c423e940e8475ad0e5adb82e6e10893fe9b365","type":"digest"}},"server":null,"time":552,"url":null}],"metrics":{"major_updates":1,"minor_updates":1,"monitored_images":6,"other_updates":2,"patch_updates":0,"unknown":0,"up_to_date":2,"updates_available":4}}
```

Check image updates for all users:
```bash
opencli docker-images --all
```

### logs

View logs sizes for user and system containers.


```bash
# opencli docker-logs

Usage: opencli docker-logs [options]

Options:
  <USERNAME>                                    Display log sizes for specified user.
  --system                                      Display log sizes just for system containers.
  --users                                       Display log sizes just for user containers.
  --all                                         Display log sizes for all user and system containers.

Examples:
  opencli docker-logs stefan
  opencli docker-logs --users
  opencli docker-logs --system
  opencli docker-logs --all

```

### lazydocker (DEPRECATED)

To manage system containers *(run as root user):
```bash
opencli docker
```

To manage user containers:
```bash
opencli docker <USERNAME>
```

### Collect Stats

To collect docker resource usage information (cpu, ram, i/o) for all users:
```bash
opencli docker-collect_stats
```

`collect_stats` script will also rotate data according to [`resource_usage_retention` setting](/cli/config.html#resource-usage-retention)


### Limits

Set global docker limits (storage, ram and cpu) for all system containers.
```bash
opencli docker-limits [--apply | --read]
```


To view current limits: 
```bash
opencli docker-limits --read
```

Example:
```bash
root@stefan:~# opencli docker-limits --read
[DOCKER]
max_ram=90
max_cpu=95
max_disk=114
```

#### CPU and RAM limit

To apply new limits for CPU % and RAM %:

```bash
opencli docker-limits --apply
```

#### Storage (disk) limit

To increase storage (disk) allocated to Docker, pass the size in **GB**:

```bash
opencli docker-limits --apply 100
```
