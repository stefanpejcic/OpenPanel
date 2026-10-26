---
sidebar_label: "Setup DKIM"
---

# How to Set Up DKIM for Your Mail Server

Ensure that you're running the [**Enterprise edition**](https://openpanel.com/enterprise/) of OpenPanel. Email support is only available in this version.

Follow [this guide](/docs/articles/user-experience/how-to-setup-email-in-openpanel/) to enable Emails in OpenPanel.

---

OpenDKIM is currently enabled by default.

## 1. Generate DKIM keys

To generate DKIM keys for a domain, run:

```bash
opencli email-setup config dkim domain DOMAIN_NAME_HERE
```

When the mailserver is installed, OpenPanel already generates a 2048-bit key for every newly added domain, so you only need to run this for domains that don't have a key yet.

### Key length

Keys are 2048-bit by default. To generate a different size, add the `keysize` option. Supported values are `1024`, `2048` and `4096`:

```bash
opencli email-setup config dkim keysize 1024 domain DOMAIN_NAME_HERE
```

A 2048-bit public key is longer than the 255 characters a single `TXT` string can hold, so the record has to be split into multiple parts (see [Formatting the TXT record value](#external-dns) below). If your DNS provider can't store a split or long `TXT` record, use `1024`: the record then fits in a single string. Otherwise keep `2048`, which is the recommended size; `1024` is the minimum that receivers accept.

:::warning
Existing keys are never overwritten. If the domain already has a DKIM key, the command above does nothing. To change the key length of an existing key, follow [Regenerate an existing DKIM key](#4-regenerate-an-existing-dkim-key).
:::

---

## 2. DNS record

When mail signed with your DKIM key is sent from your mail server, the receiver needs to check a DNS `TXT` record to verify the DKIM signature is trustworthy.

### OpenPanel DNS

If you are using nameservers on the same OpenPanel server that mailserver is running on, then run:
```bash
podman cp openadmin_mailserver:/tmp/docker-mailserver/opendkim/keys/DOMAIN_NAME_HERE/mail.txt /tmp/mail.txt && cat /tmp/mail.txt >> /etc/bind/zones/DOMAIN_NAME_HERE.zone
```

Make sure to replace `DOMAIN_NAME_HERE` with *yourdomain.com*

This will append the DKIM record to the DNS zone.

---

### External DNS

If you are using external nameservers like [Cloudflare](https://www.cloudflare.com/), run this command to view the record:

```bash
podman cp openadmin_mailserver:/tmp/docker-mailserver/opendkim/keys/DOMAIN_NAME_HERE/mail.txt /tmp/mail.txt && cat /tmp/mail.txt
```

Add the TXT record to your **remote server's DNS zone**:

| **Field** | **Value**                                                                      |
| --------- | ------------------------------------------------------------------------------ |
| **Type**  | `TXT`                                                                          |
| **Name**  | `mail._domainkey`                                                              |
| **TTL**   | Default value (or `3600` seconds)                                              |
| **Data**  | Contents of the file inside parentheses - formatted as advised below           |

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

Currently mailserver needs to be stopped and started again in order to start signing your outgoing emails with DKIM. Navigate to **OpenAdmin > Services > Service Status** and click 'Stop' for mailserver, then again 'Start'.

[![2025-07-17-15-37.png](/img/docs-content/d3hgY01F-2025-07-17-15-37.png)](/img/docs-content/d3hgY01F-2025-07-17-15-37.png)

---

## 4. Regenerate an existing DKIM key

To change the key length of a domain that already has a DKIM key (for example, to switch from 2048-bit to 1024-bit), delete the old key and generate a new one.

1. Delete the existing key:
   ```bash
   rm -rf /usr/local/mail/openmail/docker-data/dms/config/opendkim/keys/DOMAIN_NAME_HERE
   ```
   The mailserver keeps signing with the old key until it is restarted in step 4.

2. Generate the new key:
   ```bash
   opencli email-setup config dkim keysize 1024 domain DOMAIN_NAME_HERE
   ```

3. Replace the DNS record. The old public key no longer matches, so the existing `mail._domainkey` record has to be replaced, not added to.

   **OpenPanel DNS:** remove the old record from the zone, append the new one and reload the zone:
   ```bash
   sed -i '/^mail\._domainkey/{:a;/)/!{N;ba};d}' /etc/bind/zones/DOMAIN_NAME_HERE.zone
   cat /usr/local/mail/openmail/docker-data/dms/config/opendkim/keys/DOMAIN_NAME_HERE/mail.txt >> /etc/bind/zones/DOMAIN_NAME_HERE.zone
   opencli domains-dns check DOMAIN_NAME_HERE && opencli domains-dns reload DOMAIN_NAME_HERE
   ```
   If you use a secondary (slave) DNS server, also increase the serial number in the zone's `SOA` record before reloading so the change is transferred.

   **External DNS:** edit the existing `mail._domainkey` `TXT` record and replace its value with the new key, formatted as described in [External DNS](#external-dns).

4. Once `dig` returns the new key (see [Test DNS](#test-dns)), [restart the mailserver](#3-restart-mailserver) so it starts signing with the new key.

---

## 5. Troubleshooting

[MxToolbox has a DKIM Verifier](https://mxtoolbox.com/dkim.aspx) that you can use to check your DKIM DNS record(s).
