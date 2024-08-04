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


## Download OpenPanel Server Manager

On the FOSSBilling server navigate to the directory where FOSSBilling is installed and run this command to download the latest OpenPanel Server Manager:

```bash
wget -O library/Server/Manager/OpenPanel.php https://raw.githubusercontent.com/stefanpejcic/FOSSBilling-OpenPanel/main/OpenPanel.php
```

## Add OpenPanel Server

Login to your FOSSBilling admin panel and go to **System -> Hosting plans and servers** from within the navigation bar, then click on 'New server':

![new server](https://i.postimg.cc/bYV8DngC/2024-08-04-15-19.png)

in the new form we need to set:

![add openpanel server](https://i.postimg.cc/jKcjYwHJ/2024-08-04-15-21.png)


4. Name: anything that you want to identify the server
5. Hostname: if you are using domain to access OpenPanel, add it here, otherwise add the IP address.
6. IP: set the IP address of the OpenPanel server.
7. Assigned IP addresses: set IP addresses that are added on OpenPanel server.
8. Nameserver 1: Set nameserver to use for domains
9. Nameserver 2: Set nameserver to use for domains
10. Server manages: **Select OpenPanel**
11. Username: Set OpenAdmin panel username
12. Password: Set OpenAdmin panel password
13. Connection port: Set to `2087`
14. Use secure connection: *Yes* fi you are using domain name for panel access, otherwise *No*

and click on the 'Add server' button.

## Add Hosting Plan for OpenPanel

From your FOSSBilling admin panel go to **System -> Hosting plans and servers** from within the navigation bar, then click on 'New hosting plan'.

Set the name for the plan **same as on OpenPanel hosting plan** and click on 'Create hosting plan'.

![add plan](https://i.postimg.cc/02LsZqL7/2024-08-04-15-23.png)

## Assign OpenPanel Server to plan

From your FOSSBilling admin panel go to **Products -> Products & Services** from within the navigation bar, then click on the edit icon for the plan:

![edit plan](https://i.postimg.cc/N0twqkGM/2024-08-04-15-24.png)

Click on 'Configuration' and for Server select the OpenPanel server and for hosting plan set it to match the plan name on OpenPanel server:

![important](https://i.postimg.cc/GmG155CV/2024-08-04-15-26.png)

## Test connection

Create a new client and order the product that is configured to create OpenPanel account.

User should be able to login to their OpenPanel account:

![login](https://i.postimg.cc/x882pjf3/2024-08-04-15-17.png)

and to reset the password:

![reset password](https://i.postimg.cc/PJ7kgGNs/2024-08-04-15-17-1.png)
