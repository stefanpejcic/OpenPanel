---
sidebar_position: 3
---

# Switch Web Server

The **Web Server Config > Web Server Type** page allows you to switch your current web server between available options: **Nginx**, **OpenResty**, and **Apache**.

The currently active web server is displayed in the top-right corner of the page.

## Requirements

To access this feature:

- The **Docker** module must be enabled **server-wide** by an Administrator.
- Your account must have the **Docker** feature enabled.

## Usage

Before switching the web server, please ensure the following:

- **All existing domains must be removed.**
- The current web server container must be **stopped** before the new one can be started.

![Switch web server page with the conditions for switching to another web server](/img/openpanel-screenshots/containers/webserver-page.png#gh-light-mode-only)
![Switch web server page with the conditions for switching to another web server](/img/openpanel-screenshots/containers/webserver-page_dark.png#gh-dark-mode-only)

> ⚠️ If you already have domains configured, **back up all configurations**, remove all domains one by one, then proceed with switching the web server.  
> To avoid downtime, it's best to make this change **before adding any domains**.

### Steps to Switch

1. In the OpenPanel menu, navigate to **Web Server Config > Web Server Type**.
2. From the dropdown menu, select the new web server you want to use.
3. Click the **Switch** button to initiate the process.

After confirmation:

- The existing web server container will be stopped and its data will be removed.
- The new web server will be started.
- You can then re-add your domains under the new web server configuration.
