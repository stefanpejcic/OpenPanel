---
sidebar_position: 2
---

# Firewall

View and edit firewall rules.

OpenPanel supports both ConfigServer Firewall (CSF) and Uncomplicated Firewall (UFW). By default, CSF is installed, but you can choose to install UFW instead by [using the `--ufw` option during installation](/install).

Based on the installed firewall, the **OpenAdmin > Firewall** page will display either the ConfigServer Firewall UI or the custom UFW interface.


## CSF

If ConfigServer Security & Firewall (CSF) is installed, it's integrated UI will be displayed on **OpenAdmin > Firewall**.

For instructions on how to use the CSF UI, please refer to [ConfigServer Security & Firewall official documentation](https://download.configserver.com/csf/readme.txt).

![csf firewall](/img/admin/firewall_csf.png)

## UFW

If Uncomplicated Firewall (UFW) is installed, our custom interface will be displayed on **OpenAdmin > Firewall**.

![openadmin firewall settings](/img/admin/adminpanel_firewall_settings.png)

The firewall settings page provides multiple tabs:

- IPv4 - View and manage IPv4 firewall rules
- IPv6 - View and manage IPv6 firewall rules
- Settings - Manage UFW settings
- Blacklists - Enable/disable blaklists
- Logs - view the UFW service log

### View existing rules

The table shows firewall rules, showcasing information such as rule ID, action, ports, source/destination IP, and the username of the user utilizing the port.
For IPv6 rules, navigate to the IPv6 tab.

![openadmin firewall ipv6 rules](/img/admin/adminpanel_firewall_ipv6.png)

### Add Rules

To create a new rule click on the 'Add Rule' button and in the modal choose 'ALLOW' to allow the IP address or port, and 'DENY' to block access for IP address or port.

![openadmin firewall add rule](/img/admin/openadmin_ufw_ip.png)

### Delete Rules

To delete a rule click on the 'Delete' link next to it, and in the confirmaiton modal click on 'Delete' button.

![openadmin firewall delete rule](/img/admin/adminpanel_firewall_delete_rule.png)


### Settings




### Blacklists

Unless the `--skip-blacklists` flag is provided during the installation of OpenPanel, ipset-blacklists are automatically installed when the `--ufw` flag is used.

From the **OpenAdmin > Security > Firewall > Blacklists** page, administrators can easily add blacklists to block IP addresses from known malicious sources.

This feature utilizes the [ipset-blacklist](https://github.com/stefanpejcic/ipset-blacklist) service to automate the process of fetching and blocking IPs, providing a straightforward and effective method to enhance system security without manual intervention.

![openadmin ufw ipsetblacklists](/img/admin/openadmin_ufw_blacklists.png)

Default blacklists:

| Blacklist            | URL                                                                  |
|-----------------|----------------------------------------------------------------------|
| AbuseIPDB (DISABLED)       | [https://api.abuseipdb.com/api/v2/blacklist](https://api.abuseipdb.com/api/v2/blacklist) |
| OpenPanel       | [https://api.openpanel.co/blocklist.txt](https://api.openpanel.co/blocklist.txt) |
| Spamhaus DROP   | [https://www.spamhaus.org/drop/drop.lasso](https://www.spamhaus.org/drop/drop.lasso) |
| Spamhaus EDROP  | [https://www.spamhaus.org/drop/edrop.lasso](https://www.spamhaus.org/drop/edrop.lasso) |
| DShield         | [https://www.dshield.org/feeds/suspiciousdomains_Low.txt](https://www.dshield.org/feeds/suspiciousdomains_Low.txt) |
| FireHOL level1  | [https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level1.netset](https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level1.netset) |
| FireHOL level2  | [https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level2.netset](https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level2.netset) |
| FireHOL level3  | [https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level3.netset](https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level3.netset) |
| FireHOL level4  | [https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level4.netset](https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level4.netset) |
| Binary Defense   | [https://www.binarydefense.com/banlist.txt](https://www.binarydefense.com/banlist.txt) |
| blocklist.de    | [https://lists.blocklist.de/lists/all.txt](https://lists.blocklist.de/lists/all.txt) |


<Tabs>
  <TabItem value="openadmin-ufw-rbl" label="With OpenAdmin" default>

To enable or disable a blacklist in the OpenAdmin interface, click the 'Actions' button for the desired list, then select 'Enable' or 'Disable'.

To delete a blacklist from the OpenAdmin interface, click the 'Actions' button for the desired list, then select 'Delete'.

  </TabItem>
  <TabItem value="CLI-yfw-rbl" label="With OpenCLI">

To manage blacklists from the terminal:

- Download new IP addresses for all enabled blocklists:
  ```bash
  opencli blacklist --fetch
  ```
- Update all ipsets rules and reload UFW service:
  ```bash
  opencli blacklist --update_ufw
  ```
- Add a new blacklist:
  ```bash
  opencli blacklist --add-blacklist name=<name> url=<url>
  ```
- Enable a blacklist:
  ```bash
  opencli blacklist --enable-blacklist=<name>
  ```
- Disable a blacklist:
  ```bash
  opencli blacklist --disable-blacklist=<name>
  ```
- Delete a blacklist:
  ```bash
  opencli blacklist --delete-blacklist=<name>
  ```

  </TabItem>
</Tabs>



### View logs

For logs, navigate to the 'Logs' tab.

![openadmin firewall logs](/img/admin/adminpanel_firewall_logs.png)


## External Firewall

Some cloud providers, like [Hetzner](https://docs.hetzner.com/robot/dedicated-server/firewall/), offer their own external firewalls. If you are using an external firewall, ensure the following ports are open for OpenPanel services to be accessible: `53` `80` `443` `2083` `2087` `32768:60999`

If you are [using a custom port for OpenPanel instead of the default 2083](/docs/admin/settings/general/#change-openpanel-port), ensure that port is open as well.

## Restart rules

To re-open all necessary ports for OpenPanel services and users, run the command: `opencli firewall-reset`
