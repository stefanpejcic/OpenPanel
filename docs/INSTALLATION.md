# OpenPanel Installation Guide

This document provides detailed information about installing OpenPanel on your server.

## System Requirements

- **Supported Operating Systems:**
  - Ubuntu 24.04 (officially supported)
  - Fedora 38 (officially supported)
  - Ubuntu 22.04, AlmaLinux 9, RockyLinux 9, CentOS 9.4, Debian 11/12 (community supported)
- **Architecture:** x86_64 or aarch64 (ARM)
- **Minimum Requirements:**
  - 2 GB RAM (4+ GB recommended)
  - 20 GB storage (SSD recommended)
  - 1 CPU core (2+ cores recommended)
- **Public IPv4 Address:** Required for binding Nginx configuration files

## Quick Installation

To install OpenPanel with default options, run:

```bash
bash <(curl -sSL https://openpanel.org)
```

## Installation Options

The installation script supports various options to customize your installation:

```bash
bash <(curl -sSL https://openpanel.org) [OPTIONS]
```

### Core Options

| Option | Description |
|--------|-------------|
| `--key=<license>` | Set license key for OpenPanel Enterprise |
| `--domain=<domain>` | Set server hostname for panel access |
| `--username=<username>` | Set admin username (random if not provided) |
| `--password=<password>` | Set admin password (random if not provided) |
| `--email=<email>` | Set email for notifications and credentials |
| `--version=<version>` | Install specific version instead of latest |

### Configuration Options

| Option | Description |
|--------|-------------|
| `--csf` | Install ConfigServer Firewall (default) |
| `--ufw` | Install Uncomplicated Firewall instead of CSF |
| `--skip-firewall` | Skip firewall installation entirely |
| `--no-waf` | Skip CorazaWAF installation |
| `--no-ssh` | Disable port 22 (SSH) after installation |
| `--skip-dns-server` | Skip DNS server (Bind9) setup |
| `--swap=<size>` | Set swap size in GB |
| `--enable-dev-mode` | Enable developer mode |

### Advanced Options

| Option | Description |
|--------|-------------|
| `--skip-requirements` | Skip system requirements check |
| `--skip-panel-check` | Skip checking for existing panels |
| `--skip-apt-update` | Skip package manager update |
| `--post_install=<path>` | Run custom script after installation |
| `--debug` | Show detailed debug information |
| `--repair` or `--retry` | Retry installation with cleanup |
| `--uninstall` | Uninstall OpenPanel |

## Installation Process

The installation process follows these steps:

1. **Pre-installation checks:**
   - System requirements validation
   - Checking for existing control panels
   - OS detection and package manager selection

2. **Component installation:**
   - Python 3.12 and required packages
   - Docker and Docker Compose
   - Setting up configuration files
   - Installing OpenAdmin panel

3. **Service configuration:**
   - Setting up DNS (Bind9) if not skipped
   - Configuring firewall (CSF or UFW)
   - Setting up Docker containers
   - Configuring web server with CorazaWAF

4. **Finalization:**
   - Creating admin user
   - Setting up SSL for panels if domain provided
   - Configuring system services and cron jobs

## Troubleshooting

If you encounter issues during installation:

- Run with `--debug` flag to see detailed output
- Check the log file at `/root/openpanel_install.log`
- Try running with `--repair` flag to attempt fixing issues
- Post your log file to [the community forums](https://community.openpanel.org) for help

## Uninstallation

To remove OpenPanel from your server:

```bash
bash <(curl -sSL https://openpanel.org) --uninstall
```

This will remove all OpenPanel components, services, and configurations.

## Additional Resources

- [OpenPanel Documentation](https://openpanel.com/docs)
- [Community Forums](https://community.openpanel.org)
- [OpenAdmin Guide](https://openpanel.com/docs/admin/intro/)
- [OpenPanel User Guide](https://openpanel.com/docs/panel/intro/)
