---
sidebar_position: 9
---

# Docker

Manage users docker containers and their information.

## Collect Stats

The `collect_stats` script periodically checks resource usage for all users using the [docker stats](https://docs.docker.com/engine/reference/commandline/stats/) command.

For each user, data is logged within their respective folder in JSON files, located in the `/usr/local/panel/core/stats/` directory.

By default, the script is configured to operate in the background at 60-minute intervals using a cronjob:


```bash
0 * * * * opencli docker-collect_stats
```

To modify the script's execution time, [edit the crontab](https://www.airplane.dev/blog/how-to-edit-crontab) and adjust the cron schedule as needed to specify the desired execution frequency for the script.

To initiate the manual data collection process, execute the following command:

```bash
opencli docker-collect_stats
```

Example:
```bash
# opencli docker-collect_stats

Data for stefan written to /usr/local/panel/core/stats/stefan/2023-10-08-09-33-56.json
{"cpu_percent": 0.81, "mem_percent": 13.38, "net_io": "240.5k", "block_io": "409.2k"}
Data for radovan written to /usr/local/panel/core/stats/radovan/2023-10-08-09-33-56.json
{"cpu_percent": 0.01, "mem_percent": 0.02, "net_io": "50.59k", "block_io": "0"}
```

:::info
The `docker stats` command is a resource-intensive operation that consumes significant host server resources and should be executed sparingly, adhering to the established schedule.
:::



## Is port in use

The `is_port_in_use` script is used by the PM2 (Python / NodeJS Application Manager) when installing a new applicaiton to chekc if specified port is already in use in the docker contianer by another application or process.


```bash
opencli docker-is_port_in_use [username]
```


## usage_stats_cleanup

The `usage_stats_cleanup` script is used by the AdminPanel to rotate the 'Past Resource Usage' data for each user, with [the setting](https://docs.openpanel.co/docs/admin/scripts/openpanel_config#resource_usage_retention) specified by the Administrator.

If you need to manually rotate the data run:

```bash
opencli docker-usage_stats_cleanup
```
