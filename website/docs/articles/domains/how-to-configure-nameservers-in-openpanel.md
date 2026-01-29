# Configure Nameservers

Before adding any domains, it is important to first create nameservers to ensure valid DNS zone files and proper propagation.

Configuring Nameservers involves two key steps:

1. **Create private nameservers (glue DNS records)** for your domain through your domain registrar.
2. **Add the nameservers** to the OpenPanel configuration.

Tutorials for Popular Domain Providers

- [Cloudflare](https://developers.cloudflare.com/dns/additional-options/custom-nameservers/zone-custom-nameservers/)  
- [GoDaddy](https://uk.godaddy.com/help/add-custom-hostnames-12320)  
- [NameCheap](https://www.namecheap.com/support/knowledgebase/article.aspx/768/10/how-do-i-register-personal-nameservers-for-my-domain/#:~:text=Click%20on%20the%20Manage%20option,5.)

To add nameservers in OpenPanel, navigate to **OpenAdmin > Settings > OpenPanel** and enter your nameservers in the `ns1`, `ns2`, `ns3`, `ns4` fields, then click **Save changes**.

Alternatively, you can set nameservers via terminal commands:

```bash
opencli config update ns1 your_ns1.domain.com
opencli config update ns2 your_ns2.domain.com
opencli config update ns3 your_ns3.domain.com
opencli config update ns4 your_ns4.domain.com
```

Currently, you can add up to 4 nameservers.

:::info
After creating nameservers it can take up to 12h for the records to be globally accessible. Use a tool such as [whatsmydns.net](https://www.whatsmydns.net/) to monitor the status.
:::
