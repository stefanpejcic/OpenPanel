---
sidebar_label: "Restrict OpenAdmin"
---

# How to Restrict Admin Panel Access by IP Address

To restrict OpenAdmin access only to your team, close port `2087` on the firewall for everyone and then allow it only from your team's IP addresses.

## Close port 2087

Edit `/etc/csf/csf.conf` and remove *2087* from the *TCP_IN* list (and from *TCP6_IN* if you use IPv6).

## Allow your team's IP addresses

Add one line per team IP to `/etc/csf/csf.allow`, replacing *YOUR_TEAM_IP* with the actual address:

```
tcp|in|d=2087|s=YOUR_TEAM_IP
```

This opens only port `2087` for that IP. `csf -a YOUR_TEAM_IP` also works, but it allows that IP on every port.

## Restart CSF

```
csf -r
```
