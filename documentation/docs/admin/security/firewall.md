---
sidebar_position: 2
---

# Firewall

View and edit firewall rules.

OpenPanel supports both ConfigServer Firewall (CSF) and Uncomplicated Firewall (UFW). By default, CSF is installed, but you can choose to install UFW instead by [using the `--ufw` option during installation](/install).

Based on the installed firewall, the **OpenAdmin > Firewall** page will display either the ConfigServer Firewall UI or the custom UFW interface.


## CSF

If ConfigServer Security & Firewall (CSF) is installed, it's integrated UI will be displayed on **OpenAdmin > Firewall**.

![csf firewall](/img/admin/firewall_csf.png)

For instructions on how to use the CSF UI, please refer to [ConfigServer Security & Firewall official documentation](https://download.configserver.com/csf/readme.txt).

## UFW

If Uncomplicated Firewall (UFW) is installed, our custom interface will be displayed on **OpenAdmin > Firewall**.

![openadmin firewall settings](/img/admin/adminpanel_firewall_settings.png)

The firewall settings page provides three tabs:

- IPv4 - that lists all IPv4 firewall rules
- IPv6 - that lists all IPv6 firewall rules
- Logs - displays the UFW service log

### View existing rules

The table shows firewall rules, showcasing information such as rule ID, action, ports, source/destination IP, and the username of the user utilizing the port.
For IPv6 rules, navigate to the IPv6 tab.

![openadmin firewall ipv6 rules](/img/admin/adminpanel_firewall_ipv6.png)

### Add Rules

To create a new rule click on the 'New Rule' button and in the modal choose 'ALLOW' to allow the IP address or port, and 'DENY' to block access for IP address or port.

![openadmin firewall add rule](/img/admin/adminpanel_firewall_add_rule.png)

## Delete Rules

To delete a rule click on the 'Delete' link next to it, and in the confirmaiton modal click on 'Delete' button.

![openadmin firewall delete rule](/img/admin/adminpanel_firewall_delete_rule.png)


### View logs

For logs, navigate to the 'Logs' tab.

![openadmin firewall logs](/img/admin/adminpanel_firewall_logs.png)




## Restart rules

To re-open all needed ports for OpenPanel services adn users, run command: `opencli firewall-reset`
