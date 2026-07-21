---
sidebar_position: 1
---

# WAF

Install ModSecurity and enable it for user domains.

The Settings > ModSecurity page allows you to install ModSecurity for Nginx and configures the [OWASP core ruleset](https://owasp.org/www-project-modsecurity-core-rule-set/)

The OWASP ModSecurity Core Rule Set (CRS) is a set of generic attack detection rules for use with ModSecurity that will increase the security of user domains and websites.

## Install ModSecurity

Upon the initial access to the ModSecurity page, you will be prompted to install the ModSecurity plugin.

:::danger
The installation process may require up to 10 minutes and involves rebuilding the Nginx configuration. It's important to note that any customizations to the service will be permanently removed during this process. It is advisable to perform the installation during off-peak hours to minimize the risk of causing downtime for websites.
:::

To install ModSecurity click on the 'Install' button.

![openadmin modsec install](/img/admin/adminpanel_modsec_install.png)

Or from terminal run: [opencli nginx-install_modsec](/docs/admin/scripts/webserver#install-modsecurity)

## Activate ModSecurity

Upon ModSecurity installation, all new domains will have ModSecurity enabled by default. However, individual users can choose to disable ModSecurity for their domains at any time through their OpenPanel interface. [More information](/docs/panel/advanced/server_settings#modsecurity-settings)


## Customize ModSecurity rules

Adjusting ModSecurity rules means fine-tuning security settings for your specific needs, giving administrators the power to better protect against specific threats and reduce false positives.

You can follow user-friendly guides to easily customize ModSecurity rules, adapting security settings to your specific needs.

- [Nginx Docs: Using the OWASP CRS with the NGINX ModSecurity WAF](https://docs.nginx.com/nginx-waf/admin-guide/nginx-plus-modsecurity-waf-owasp-crs/)
- [Nginx Docs: Using the ModSecurity Rules from Trustwave SpiderLabs with the NGINX ModSecurity WAF](https://docs.nginx.com/nginx-waf/admin-guide/nginx-plus-modsecurity-waf-trustwave-spiderlabs-rules/)
- [ModSecurity Documentation](https://github.com/SpiderLabs/ModSecurity/wiki)
- [ProSec Blog: Modsecurity Core Rule Sets and Custom Rules](https://www.prosec-networks.com/en/blog/modsecurity-core-rule-sets-und-eigene-regeln/)

## Enable ModSecurity for existing domains

After installing ModSecurity only new domains that users add will by default have ModSecurity activate, and for existing users this process can be performed by the administrator from this page or from each user panel individually. To enable ModSecurity on all domains owned by a user, select the user anc click on 'Enable' button.

![openadmin modsec settings](/img/admin/adminpanel_modsec_use.png)

Or from terminal run: [opencli domains-enable_modsec](/docs/admin/scripts/domains#enable-modsecurity)


