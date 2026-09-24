---
sidebar_label: "Add Extra IP Addresses"
description: "How to add additional IPv4 addresses to an OpenPanel server (Ubuntu netplan, Debian, AlmaLinux/Rocky), make them persistent, and assign a dedicated IP to a hosting account."
---

# How to Add Extra IP Addresses to Your Server

OpenPanel uses the server's main IP for all accounts by default. With additional IPs you can give a customer a **dedicated IP address** - for their own SSL setup, a separate sending reputation, or simply because their plan includes one.

There are two parts: adding the IP to the **server's network configuration**, then assigning it to an **account** in OpenPanel.

:::info IPv4 only
This guide covers **IPv4**. Additional IPv6 addresses are not tested with OpenPanel yet.
:::

---

## Step 1: Order the IP from Your Provider

The provider must route the additional IP to your server. It's called *Additional IP*, *Floating IP*, *Failover IP* or *Secondary IP* depending on the provider:

- **Hetzner Cloud**: *Primary IPs* / *Floating IPs*; **Hetzner Robot**: order *additional IPs* for the server.
- **OVHcloud**: *Additional IP*, then move it to your server.
- **DigitalOcean**: *Reserved IPs* (one per Droplet).
- **Vultr**, **Linode**: *Add IPv4 address* on the instance's network page.
- **AWS**: attach a secondary private IP to the instance and associate an Elastic IP with it.

Note the IP, the **netmask/prefix** and the **gateway**, if the provider gives you one.

---

## Step 2: Add the IP to the Server

Find your network interface name first:

```bash
ip -4 addr
```

It's usually `eth0`, `ens3`, `enp1s0` or similar.

### Ubuntu (netplan)

Edit the netplan file in `/etc/netplan/` (e.g. `50-cloud-init.yaml` or `01-netcfg.yaml`) and add the new IP under `addresses`:

```yaml
network:
  version: 2
  ethernets:
    eth0:
      dhcp4: true
      addresses:
        - 203.0.113.50/32
```

Apply it:

```bash
netplan try      # applies and rolls back automatically if you lose connection
netplan apply
```

:::tip Cloud-init
On some cloud images, cloud-init rewrites `50-cloud-init.yaml` on reboot. Put the extra IP in its own file instead, e.g. `/etc/netplan/60-extra-ips.yaml`, with the same structure.
:::

### Debian (ifupdown)

Add an alias to `/etc/network/interfaces`:

```
auto eth0:1
iface eth0:1 inet static
    address 203.0.113.50
    netmask 255.255.255.255
```

Then run `ifup eth0:1`.

### AlmaLinux / Rocky Linux (NetworkManager)

```bash
nmcli connection show                         # find the connection name, e.g. "System eth0"
nmcli connection modify "System eth0" +ipv4.addresses 203.0.113.50/32
nmcli connection up "System eth0"
```

Use the exact netmask and settings your provider documents - some require a `/32` with a specific gateway, others a normal subnet.

### Check

```bash
hostname -I
ping -c 3 -I 203.0.113.50 1.1.1.1
```

Test from another machine that the IP answers, and **reboot once** to make sure the setting survives a restart.

---

## Step 3: Assign the IP to an Account

OpenPanel detects new IPs automatically - no extra configuration needed.

1. Go to **OpenAdmin → Accounts → Users → *username* → Edit**.
2. Select the new IP in the **IP Address** field.
3. Click **Save**.

Or from the terminal:

```bash
opencli user-ip USERNAME 203.0.113.50
```

OpenPanel updates the user's DNS zones and proxy configuration, so all their current and future domains use the dedicated IP. If a domain's DNS is hosted elsewhere, update its `A` record there.

See [Assign a dedicated IP to a user](/docs/articles/accounts/set-dedicated-ip-address-for-user/).

---

## Good to Know

- **License**: only the server's **main** IP needs a license - additional IPs are free to use.
- **Firewall**: Sentinel Firewall (CSF) applies to all IPs on the server; no changes are needed.
- **Email** is still sent from the server's main IP, so its [reverse DNS](/docs/articles/email/reverse-dns-ptr-record/) is the one that matters for mail.
- **Removing** a dedicated IP: `opencli user-ip USERNAME delete` moves the user back to the main IP. Do this **before** removing the IP from the server.

---

## Related

- [Main IP in OpenPanel](/docs/articles/install-update/openpanel-main-ip-address/)
- [Change the server hostname or IP](/docs/articles/server/change-server-hostname-or-ip/)
