---
sidebar_position: 2
---

# Firewall

View and edit firewall (UFW) rules

![openadmin firewall settings](/img/admin/adminpanel_firewall_settings.png)

The firewall settings page provides three tabs:

- IPv4 - that lists all IPv4 firewall rules
- IPv6 - that lists all IPv6 firewall rules
- Logs - displays the UFW service log

## View existing rules

The table shows firewall rules, showcasing information such as rule ID, action, ports, source/destination IP, and the username of the user utilizing the port.
For IPv6 rules, navigate to the IPv6 tab.

![openadmin firewall ipv6 rules](/img/admin/adminpanel_firewall_ipv6.png)

## Add Rules

To create a new rule click on the 'New Rule' button and in the modal choose 'ALLOW' to allow the IP address or port, and 'DENY' to block access for IP address or port.

![openadmin firewall add rule](/img/admin/adminpanel_firewall_add_rule.png)

## Delete Rules

To delete a rule click on the 'Delete' link next to it, and in the confirmaiton modal click on 'Delete' button.

![openadmin firewall delete rule](/img/admin/adminpanel_firewall_delete_rule.png)


## View logs

For logs, navigate to the 'Logs' tab.

![openadmin firewall logs](/img/admin/adminpanel_firewall_logs.png)
