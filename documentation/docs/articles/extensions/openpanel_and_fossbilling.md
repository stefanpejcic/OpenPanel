# OpenPanel and FOSSBilling

OpenPanel Enterprise edition has billing integrations with WHMCS and FOSSBilling.

## Whitelist FOSSBillling on OpenPanel

Before you can setup the OpenPanel server manager in FOSSBilling, you need to first whitelist your FOSSBilling server's IP address inside of OpenAdmin interface, and enable API access. 

To enable access to the FOSSBilling server's IP, first check the ip address on that server, from terminal you can run:

```bash
curl ip.openpanel.co
```

Login to OpenAdmin and under **Settings > Firewall** add the FOSSBilling server's IP under **Allow IP address**:

![whitelist ip](https://i.postimg.cc/433M6LBr/2024-08-04-16-10.png)

## Enable API access on OpenAdmin

To enable API access on OpenPanel, navigate to **Settings > API Access** from the OpenAdmin interface and click on 'Enable API access' button:

![enable api](https://i.postimg.cc/VsthWbWL/2024-08-04-16-14.png)


## Configure FOSSBilling with OpenPanel



## Test connection
