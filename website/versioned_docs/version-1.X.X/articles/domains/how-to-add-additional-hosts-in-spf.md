
# Adding Additional Hosts to SPF Records

To include a specific host in SPF records across all domains, follow the instructions below.

---

## For Existing Domains

Use the following `sed` command to append a new host (e.g., `include:example.com`) to all SPF records in the zone files on the server:

```
sed -i 's/\(v=spf1\)/\1 include:example.com/' /etc/bind/zones/*.zone
```

---

## For New Domains

To ensure new domains include this host in their SPF records automatically:

1. Go to:
   `OpenAdmin > Domains > DNS Zone Templates`

2. Edit the zone template, include the desired hosts in the SPF entry:

   ```
   "v=spf1 include:example.com ~all"
   ```

This will ensure any newly created domain uses the updated SPF policy.
