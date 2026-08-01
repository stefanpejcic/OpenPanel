---
sidebar_position: 1  
---

# Select PHP Version  

Assign a PHP version to each domain. Changing the version updates the domain's configuration file and reloads the web server, which may temporarily interrupt website processes for that domain.  

For each version, the current development status is displayed, highlighting domains using outdated or unsupported PHP versions that may require an update.

Ensure you check your site's requirements before selecting the appropriate PHP version.

![change version for domains](/img/panel/v2/openpanel_change_default_php.gif)

NOTE: If you are using OpenLitespeed or Litespeed as webserver, PHP version can not be set per-domain — only a single PHP version is used for all domains. Use the [**Default PHP Version**](/docs/panel/php/default/) page instead to change it for all domains at once.
