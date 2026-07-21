# Troubleshooting Guide: Caddy / Let’s Encrypt SSL Validation Failure

## Issue

SSL certificate generation in **Caddy** fails with both `http-01` and `tls-alpn-01` ACME challenges.  
The following errors appear in Caddy logs:

```json
{
  "level": "error",
  "msg": "challenge failed",
  "identifier": "test1.test.com",
  "challenge_type": "http-01",
  "problem": {
    "type": "urn:ietf:params:acme:error:connection",
    "detail": "103.67.236.210: Fetching http://test1.test.com/.well-known/acme-challenge/...: Connection reset by peer"
  }
}
```
```
{
  "level": "error",
  "msg": "challenge failed",
  "identifier": "test1.test.com",
  "challenge_type": "tls-alpn-01",
  "problem": {
    "type": "urn:ietf:params:acme:error:connection",
    "detail": "103.67.236.210: Connection reset by peer"
  }
}
```

Despite these failures, Caddy successfully creates an ACME order:
```
https://acme-staging-v02.api.letsencrypt.org/acme/order/262155913/31213094193
```

##Cause

The connection resets indicate a network-level issue:

- Firewalls or WAF blocking inbound connections on ports 80/443
- Manual ACME challenge configuration overriding Caddy’s internal ACME handling
- Repeated failed attempts may trigger Let’s Encrypt rate limits (HTTP 429)

##Workaround / Steps to Resolve

1. Remove manual ACME challenge configuration

If you have something like this in your Caddyfile:
```
@acme {
    path /.well-known/acme-challenge/*
}
handle @acme {
    reverse_proxy http://127.0.0.1:32773
}
tls {
    issuer acme {
        email dev@test.com
    }
}
```
Remove it entirely and let Caddy manage ACME challenges automatically.

2. Ensure firewall allows ACME validation

Let’s Encrypt must be able to reach your server on ports 80 and 443:
```
iptables -I INPUT -p tcp --dport 80 -j ACCEPT
iptables -I INPUT -p tcp --dport 443 -j ACCEPT
iptables -I OUTPUT -p tcp --dport 80 -j ACCEPT
iptables -I OUTPUT -p tcp --dport 443 -j ACCEPT
```

Also check:

- Cloud / hosting provider firewalls
- WAF or DDoS protection rules

3. Restart Caddy

```docker restart caddy```

4. Wait for automatic ACME retry

Caddy will retry issuing the certificate automatically after fixing network issues.

5. Optional workarounds if rate limits are hit:

- Wait for Let’s Encrypt limits to reset
- Switch to Cloudflare proxy
- Use ZeroSSL:
```
tls {
    issuer acme {
        ca https://acme.zerossl.com/v2/DV90
        email admin@example.com
    }
}
```
##Disable WAF

ModSecurity rules do not affect SSL generation: Caddy automatically tries to generate SSL when a website is visited via HTTPS, and this happens before the request is examined by any WAF rules.

However, you can temporarily disable all rules - or disable the WAF completely by simply:

- Disabling 'waf' module in OpenAdmin > Settings > Modules 
- Changing image tag from openpanel/caddy-coraza to caddy:latest in OpenAdmin > Services > Services Limits > CADDY image.
- Stop & Start the Web Server from OpenAdmin > Services > Services Status.

Or use opencli command: ```opencli waf disable```

Note that you need to remove existing domains as well, or edit their files to not have any # modsecurity parts: https://github.com/stefanpejcic/openpanel-configuration/blob/main/caddy/templates/domain.conf_with_modsec#L17-L33

We don’t recommend disabling WAF unless you have another firewall in place - for example, if you’re using a Cloudflare proxy & firewall for all domains.


##Verification

After completing the steps above:

- Check Caddy logs for successful ACME issuance
- Confirm HTTPS is working:

```https://test1.test.com```

Optional: verify certificate details with:

```curl -v https://test1.test.com ```


##Notes
Avoid manual ACME challenge configurations unless absolutely necessary
Ensure firewall and provider-level settings do not block Let’s Encrypt validation IPs




