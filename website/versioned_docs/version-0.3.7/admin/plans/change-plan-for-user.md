---
sidebar_position: 2
---

# Change plan for a user

You can control the memory, CPUs, and disk space the docker container uses, and you can specify whether or not Kubernetes is supported. You can change the CPUs and memory on an existing docker container on the fly without any downtime.

To change a hosting plan (package) for an account click on 'Edit information' for that user and in the new modal select the new plan name then click on 'Save changes'.

![openadmin change plan for a user](/img/admin/change_plan.png)





If you need more disk space, you have to create a new docker container. Unfortunately **you can’t resize the existing docker container’s disk**.

This is a limitation with the Docker service itself, and not with OpenPanel. Support for resizing existing docker images is on the way.

You can [increase the size of the  devicemapper partition for the docker container](https://pcx3.com/linux/how-to-increase-docker-container-disk-size-devicemapper/), but please note that the change is not permanent and will be reverted on server reboot. 


When changing a plan for the user, the disk size value from the new plan is simply ignored. The value is used only for new user accounts.
