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
