# How to limit access to OpenAdmin

To restrict OpenAdmin access only to your team, whitelist your server's IP addresses on the firewall, and then disable port `2087`.

## Whitelist your team's IP addresses

Replace *YOUR_TEAM_IP* with your actual IP addresses. Repeat for each team IP.

```
csf -a YOUR_TEAM_IP
```

## Block all other access to port 2087

Edit `/etc/csf/csf.conf` and remove *2087* from the allowed *TCP_IN* list, or run:

```
csf -d 0.0.0.0/0 2087
```

Then restart CSF:

```
csf -r
```
