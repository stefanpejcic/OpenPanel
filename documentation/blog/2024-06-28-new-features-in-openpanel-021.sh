---
title: What is new in OpenPanel 0.2.1
description: OpenPanel 0.2.1 brings a lot of new features and security enhancements.
slug: new-features-in-openpanel-021
authors: stefanpejcic
tags: [OpenPanel, news, install]
image: https://openpanel.co/img/blog/new-features-in-openpanel-02.png
hide_table_of_contents: true
---

OpenPanel 0.2.1 is officially released today and is available for installation: https://openpanel.co/install

<!--truncate-->

Version 0.2.1 is the most complex version that we have so far, and brings a lot of improvements to both the OpenAdmin interface and the backend logic.



Lets jump right in and see exactyl what improvements are made:


### Improved install

Installation script is now optimized and average installation speed is less that 5 minutes!

![faster install](https://i.postimg.cc/nLd55VjJ/2024-06-28-16-45.png)

We managed to lower the dependencies for OpenPanel to just:

"docker.io" "default-mysql-client" "nginx" "zip" "bind9" "unzip" "python3-pip" "pip" "gunicorn" "jc" "certbot" "python3-certbot-nginx" "sqlite3" "geoip-bin" "ufw"

### Customization that stick

All configuration files ofr OpenPanel are now stored under `/etc/openpanel/` and not modified on updates. This ensures that any customizations are permanent.

```bash
/etc/openpanel/
├── bind9/       - DNS configuration files and templates 
├── docker/      - Docker configuration
├── goaccess/    - Configuration for Domain Access Reports
├── mysql/       - MySQL login information and backups
├── nginx/       - Nginx configurtion and domain templates
├── openadmin/   - Configuration files for OpenAdmin service
├── openpanel/   - Configuration files for OpenPanel
├── skeleton/    - Files that are used for new users
├── ssh/         - SSH configuration
└── ufw/         - Firewall configuration
```

