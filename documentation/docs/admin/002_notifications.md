---
sidebar_position: 6
---

# Notifications

Notifications are accesible from the notification icon in upper right corner. 

![notifications center](/img/admin/notifications_center.png)


OpenPanel records the following actions:

- server reboot
- service is inactive
- update is available
- high memory usage
- high averageload 
- high cpu usage
- high disk usage

Each notification  type can be disabled and treshold limits can be set by the Admin user.

<Tabs>
  <TabItem value="openadmin-notifications-view" label="With OpenAdmin" default>

To view current notification settings, click on the bell icon in the top menu.
You can view the current settings, and modify them.

  </TabItem>
  <TabItem value="CLI-notifications-view" label="With OpenCLI">

To view current notification settings run:

```bash
opencli admin notifications get <OPTION>
```

Example:

```bash
# opencli admin notifications get reboot
yes
```

To change the notification settings run:

```bash
opencli admin notifications update <OPTION> <NEW-VALUE>
```

Example:
```bash
opencli admin notifications update load 10
Updated load to 10
```

  </TabItem>
</Tabs>
    
To confirm receipt of a notification, select the checkmark icon located in front of it. Once a notification is confirmed, subsequent notifications of the same type will be logged if the issue persists. For instance, if a service is unavailable, the system will generate an initial notification. However, if you acknowledge the notification and the service remains unrecovered, the next time the check is executed, it will log another notification.

Example notifications:

On Server Reboot
![reboot](/img/admin/dashboard/reboot.png)

If service is inactive:
![service](/img/admin/dashboard/service.png)

If CPU usage is over a treshold:
![cpu](/img/admin/dashboard/cpu.png)

If new version of OpenPanel is available:
![update](/img/admin/dashboard/update.png)

If Memory usage is over a treshold:
![ram](/img/admin/dashboard/ram.png)

If system is running out of disk space:
![disk](/img/admin/dashboard/disk.png)
