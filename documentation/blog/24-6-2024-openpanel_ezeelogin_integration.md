---
title: OpenPanel Ezeelogin Integration
description: Ezeelogin now supports auto-login to OpenPanel servers
slug: openpanel-ezeelogin-integration
authors: stefanpejcic
tags: [OpenPanel, news, ezeelogin]
image: https://openpanel.co/img/blog/openpanel_ezeelogin_integration.png
hide_table_of_contents: true
---

Ezeelogin 7.37.11 supports auto-login to OpenPanel servers!

<!--truncate-->

We've been relying on [Ezeelogin](https://www.ezeelogin.com/) since 2014 as our SSH Jump Server solution for managing access to Linux hosts, including control panels like cPanel and CyberPanel. However, we encountered issues with auto-login for OpenPanel due to overlapping port usage (2083 and 2087) with cPanel, causing session cookie conflicts.

After contacting Ezeelogin support, they promptly addressed the issue and introduced support for OpenPanel in Ezeelogin version 7.37.11 🎉

A big thank you to the Ezeelogin team for enabling support for OpenPanel!


----

How to upgrade Ezeelogin to the `7.37.11` version that supports OpenPanel:

1. Run the below command to fully backup the current installation.

```bash
root@gateway ~]# /usr/local/sbin/backup_ezlogin.php 
```
 

2. Download the latest package according to the PHP version you are running on the gateway server.

```bash
root@gateway ~]# wget https://downloads.ezeelogin.com/ezlogin_7.37.11_php81.bin

root@gateway ~]# wget https://downloads.ezeelogin.com/ezlogin_7.37.11_php71.bin

```

3. Run the below command to upgrade Ezeelogin using a single command. Replace the Ezeelogin package name, database root username, and password.

```bash
root@gateway ~]# sh ezlogin_7.37.11_php81.bin -- -dbsuser -dbspass <enter_(root)db_password> -skipgeolite -auto -force -ACCEPT_SETTINGS -I_ACCEPT_EULA -skipbackup -update
```

Refer detailed article to upgrade Ezeelogin to the latest version: https://www.ezeelogin.com/kb/article/upgrade-ezeelogin-jump-server-to-the-latest-version-136.html

