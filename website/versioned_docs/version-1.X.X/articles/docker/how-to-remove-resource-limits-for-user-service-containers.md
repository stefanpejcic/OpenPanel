# Unlimited CPU&RAM

Setting unlimited CPU and RAM resources allows an OpenPanel user account to consume all available CPU cores and physical memory on the server.

**This is strongly discouraged**, even on a single-user VPS, as a rogue process or compromised website could destabilize the entire system. It's important to set resource limits to ensure the system has enough capacity for essential services such as monitoring, container orchestration, DNS, and more.

If you still want to enable unlimited resources, follow these steps carefully:

## Remove Limits from the User Plan

Set both the CPU and RAM limits to `0` (zero) in the plan settings to remove restrictions:

![userlimits.png](/img/panel/v2/userlimits-guide.png)

----

## Remove Docker Resource Limits


### For existing user

Run this command to remove CPU, memory, and PID limits from all services defined in the user's Docker Compose file. Replace `dorotea` with the actual username:

```bash
USERNAME=dorotea && \
sed -i '/deploy:/,/^[^[:space:]]/{
  /deploy:/d
  /resources:/d
  /limits:/d
  /cpus:/d
  /memory:/d
  /pids:/d
}' /home/$USERNAME/docker-compose.yml
```

If the user already has running services, stop and restart them to apply the changes (replace `dorotea` with the username):

```bash
USERNAME=dorotea && \
cd /home/$USERNAME && \
running_services=$(docker --context=$USERNAME compose ps --services --filter "status=running") && \
docker --context=$USERNAME compose down && \
docker --context=$USERNAME compose up -d $running_services
```


### For All new users

To ensure all new users have unlimited CPU and RAM by default, apply the same `sed` commands to the Docker Compose template used for new accounts:

```bash
sed -i '/deploy:/,/^[^[:space:]]/{
  /deploy:/d
  /resources:/d
  /limits:/d
  /cpus:/d
  /memory:/d
  /pids:/d
}' /etc/openpanel/docker/compose/1.0/docker-compose.yml
```

---


Once these changes are applied, any new or restarted services will be able to consume all available system resources. The control panel will display a warning to users indicating that CPU and RAM limits are disabled:

![userlimits3.png](/img/panel/v2/userlimits-guide3.png)

