---
sidebar_position: 5
---

# ImunifyAV

[ImunifyAV](https://cloudlinux.zendesk.com/hc/en-us/articles/4716287786396-Imunify360-Standalone-installation-guide-with-integration-conf-examples) enhances your server’s security by allowing you to scan user webiste files for malicious content.

> Note: Imunify, its trademarks, and all related assets are the property of [CloudLinux Zug GmbH](https://cloudlinux.com/).

## Install

ImunifyAV is **not** included with OpenPanel by default. To install it, run:

```bash
opencli imunify install
```

This command installs the latest PHP version, the *imunify360-agent*, and configures access through OpenAdmin.

## Start

To start the Imunify graphical interface, use:

```bash
opencli imunify start
```

## Login

Access the Imunify GUI from **OpenAdmin > Security > Imunify** using your SSH credentials (typically the `root` username and its password).

## Manage

Imunify allows you to scan user files and detect any malicious content.

