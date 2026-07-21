# DNS Clustering

DNS clustering enables you to synchronize DNS records across multiple OpenPanel servers, providing redundancy and scalability for your DNS infrastructure.

All servers involved in the cluster must be running [BIND9](https://www.isc.org/bind/) - either installed by OpenPanel or as a standalone service or container.

## Share DNS Across Multiple OpenPanel Servers

The simplest way to build a redundant DNS cluster is to run OpenPanel on two or more servers, each managing DNS zones and nameservers in sync. No extra configuration is needed beyond what's described below.

### Step 1: Allow connections between the servers

Suppose you have two servers:

- Server #1 IP: `MASTER_IP`
- Server #2 IP: `SLAVE_IP`

You need to allow connection between the servers, on MASTER_IP whitelist the SLAVE_IP and on SLAVE_IP whitelist the MASTER_IP:

- on master: `csf -a SLAVE_IP`
- on slave: `csf -a MASTER_IP`


### Step 2: Prepare Slave Server

On MASTER server get the secret token from rndc.key file: `cat /etc/bind/rndc.key` 

On SLAVE:

- edit `/root/docker-compose.yml` and add `- "953:953/tcp"` under ports for bind9 service.
- edit `/etc/bind/named.conf.options` and add `allow-new-zones yes;` inside options block, and append at the end of the file: 
  ```
  controls {
      inet 0.0.0.0 port 953 allow { MASTER_IP_HERE; } keys { "rndc-key"; };
  };
  ```
-  restart service with: `docker exec openpanel_dns rndc reconfig | docker compose up -d bind9`

Test connection from MASTER_IP with:

```bash
docker exec openpanel_dns rndc -s SLAVE_IP -k rndc-key status
```

### Step 3: Create Nameservers in Your DNS Zone

Using the domain `yourdomain.com` as an example, define two nameservers:

- `dns1.yourdomain.com` → points to Server #1 IP (A record)
- `dns2.yourdomain.com` → points to Server #2 IP (A record)

Add these A records in your domain's DNS provider for `yourdomain.com`.


### Step 4: Register Nameservers in OpenPanel

On **master** server, open the OpenAdmin panel:
* Navigate to **Settings > OpenPanel > Nameservers**
* Add both `dns1.yourdomain.com` and `dns2.yourdomain.com`

[![2025-07-09-17-30.png](/img/docs-content/kXnvzCwW-2025-07-09-17-30.png)](/img/docs-content/kXnvzCwW-2025-07-09-17-30.png)

### Step 5: Enable DNS Clustering

On **master** server:

* Go to **OpenAdmin > Domains > DNS Cluster**
* Click **Enable DNS Clustering**

[![2025-07-09-17-32.png](/img/docs-content/FzG3NfG3-2025-07-09-17-32.png)](/img/docs-content/FzG3NfG3-2025-07-09-17-32.png)

* Click **Add Server** and enter the IP of the slave server, then **Add**

[![2025-07-09-17-33.png](/img/docs-content/7PX2C2MT-2025-07-09-17-33.png)](/img/docs-content/7PX2C2MT-2025-07-09-17-33.png)

### Test Your Cluster

Add a new domain on either server via a OpenPanel user account.

Then verify the DNS zone is synchronized on both servers using the `dig` command:

```bash
dig A +short yourdomain.com @185.241.214.214
dig A +short yourdomain.com @95.217.216.36
```

Replace `yourdomain.com` with the domain you added.

If both servers return the correct IP, your DNS clustering setup is working!
