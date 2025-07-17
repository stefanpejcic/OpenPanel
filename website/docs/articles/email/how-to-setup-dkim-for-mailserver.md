# Setup DKIM for Mailserver

Ensure that you're running the [**Enterprise edition**](https://openpanel.com/enterprise/) of OpenPanel. Email support is only available in this version.

Follow [this guide](/docs/articles/user-experience/how-to-setup-email-in-openpanel/) to enable Emails in OpenPanel.

---

OpenDKIM is currently enabled by default.

## 1. Generate DKIM keys

To generate DKIM keys for a domain, run:

```bash
opencli email-setup config dkim domain DOMAIN_NAME_HERE
```

---

## 2. DNS record

When mail signed with your DKIM key is sent from your mail server, the receiver needs to check a DNS `TXT` record to verify the DKIM signature is trustworthy.

### OpenPanel DNS

If you are using nameservers on the same OpenPanel server that mailserver is running on, then run:
```bash
docker cp openadmin_mailserver:/tmp/docker-mailserver/opendkim/keys/DOMAIN_NAME_HERE/mail.txt /tmp/mail.txt && cat /tmp/mail.txt >> /etc/bind/zones/DOMAIN_NAME_HERE.zone
```

Make sure to replace `DOMAIN_NAME_HERE` with *yourdomain.com*

This will append the DKIM record to the DNS zone.

---

### External DNS

If you are using external nameservers like [Cloudflare](https://www.cloudflare.com/), run this command to view the record:

```bash
docker cp openadmin_mailserver:/tmp/docker-mailserver/opendkim/keys/DOMAIN_NAME_HERE/mail.txt /tmp/mail.txt && cat /tmp/mail.txt
```

Add the TXT record to your **remote server's DNS zone**:

| **Field** | **Value**                                                                      |
| --------- | ------------------------------------------------------------------------------ |
| **Type**  | `TXT`                                                                          |
| **Name**  | `mail._domainkey`                                                              |
| **TTL**   | Default value (or `3600` seconds)                                              |
| **Data**  | Contents of the file inside parentheses `(...)`,<br>formatted as advised below |

**Formatting the TXT record value correctly:**
`TXT` records with values that are longer than 255 characters need to be split into multiple parts. This is why the generated `mail.txt` file (containing your public key for use with DKIM) has multiple value parts wrapped within double-quotes between `(` and `)`.

A DNS web-interface may handle this separation internally instead, and [could expect the value provided all as a single line](https://serverfault.com/questions/763815/route-53-doesnt-allow-adding-dkim-keys-because-length-is-too-long) instead of split. When that is required, you'll need to manually format the value as described below.

Your generated DNS record file (mail.txt) should look similar to this:

```
mail._domainkey IN TXT ( "v=DKIM1; k=rsa; "
"p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqQMMqhb1S52Rg7VFS3EC6JQIMxNDdiBmOKZvY5fiVtD3Z+yd9ZV+V8e4IARVoMXWcJWSR6xkloitzfrRtJRwOYvmrcgugOalkmM0V4Gy/2aXeamuiBuUc4esDQEI3egmtAsHcVY1XCoYfs+9VqoHEq3vdr3UQ8zP/l+FP5UfcaJFCK/ZllqcO2P1GjIDVSHLdPpRHbMP/tU1a9mNZ"
"5QMZBJ/JuJK/s+2bp8gpxKn8rh1akSQjlynlV9NI+7J3CC7CUf3bGvoXIrb37C/lpJehS39KNtcGdaRufKauSfqx/7SxA0zyZC+r13f7ASbMaQFzm+/RRusTqozY/p/MsWx8QIDAQAB"
) ;
```

Take the content between `( ... )`, and combine all the quote wrapped content and remove the double-quotes including the white-space between them. That is your `TXT` record value, the above example would become this:

```
v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqQMMqhb1S52Rg7VFS3EC6JQIMxNDdiBmOKZvY5fiVtD3Z+yd9ZV+V8e4IARVoMXWcJWSR6xkloitzfrRtJRwOYvmrcgugOalkmM0V4Gy/2aXeamuiBuUc4esDQEI3egmtAsHcVY1XCoYfs+9VqoHEq3vdr3UQ8zP/l+FP5UfcaJFCK/ZllqcO2P1GjIDVSHLdPpRHbMP/tU1a9mNZ5QMZBJ/JuJK/s+2bp8gpxKn8rh1akSQjlynlV9NI+7J3CC7CUf3bGvoXIrb37C/lpJehS39KNtcGdaRufKauSfqx/7SxA0zyZC+r13f7ASbMaQFzm+/RRusTqozY/p/MsWx8QIDAQAB
```
---

### Test DNS

To test that your new DKIM record is correct, query it with the `dig` command. The `TXT` value response should be a single line split into multiple parts wrapped in double-quotes:

```bash
$ dig +short TXT mail._domainkey.example.com
"v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqQMMqhb1S52Rg7VFS3EC6JQIMxNDdiBmO.."
```

---

## 3. Restart Mailserver

Currently mailserver needs to be stopped and started again in order to start signing your outgoing emails with DKIM. Navigate to **OpenAdmin > Services > Status** and click 'Stop' for mailserver, then again 'Start'.

[![2025-07-17-15-37.png](https://i.postimg.cc/d3hgY01F/2025-07-17-15-37.png)](https://postimg.cc/ctNF70wk)

---

## 4. Troubleshooting

[MxToolbox has a DKIM Verifier](https://mxtoolbox.com/dkim.aspx) that you can use to check your DKIM DNS record(s).
