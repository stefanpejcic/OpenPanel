# Troubleshooting DNS

Make sure the DNS-server on the host server responds to requests for the domain zone:

`dig <domain> @<IP address> ANY +short`


## servers could not be reached
Response message `servers could not be reached` indicates that the bind9 service is not running.

Try to restart the service from **OpenAdmin > Services**

And if service restart failed check the service logs.

The usual culprit is an invalid DNS zone file for a domain, so edit the file mentioned in the error, save changes then restart service again.

----

## empty dig response

An empty response for the dig command  indicates that the DNS-server does not have information about the domain. Perhaps, it is not added to any account or its DNS zone file is missing.

To check if domain is added to any account use the command:
`opencli domains-whoowns <domain_name>`

---

## Can't connect to the server

Try to connect to port 53 of the server through telnet:

`telnet <IP address of the server> 53`

If you can not connect, check the Firewall settings on the admin panel.


---

## zone not loaded due to errors

When manually editing the zone files or importing then from another provider, it's important to check the files for syntax errors. If there are any errors the zone file will be skipped on service reload but will cause fail in service restart.

To check a domain zone file syntax:
```bash
named-checkzone <DOMAIN> /etc/bind/zones/<DOMAIN>.zone
```

Example:
```bash
# named-checkzone example.com /etc/bind/zones/example.com.zone
zone example.com/IN: NS 'ns17.openpanel.example.com' has no address records (A or AAAA)
zone example.com/IN: not loaded due to errors.
```
