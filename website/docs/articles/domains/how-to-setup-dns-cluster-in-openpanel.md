# DNS Clustering

DNS clustering enables you to synchronize DNS records across multiple OpenPanel servers, providing redundancy and scalability for your DNS infrastructure.

All servers involved in the cluster must be running [BIND9](https://www.isc.org/bind/) - either installed by OpenPanel or as a standalone service or container.

## Share DNS Across Multiple OpenPanel Servers

The simplest way to build a redundant DNS cluster is to run OpenPanel on two or more servers, each managing DNS zones and nameservers in sync. No extra configuration is needed beyond what's described below.

### Step 1: Establish SSH Access Between Servers

Suppose you have two servers:

- Server #1 IP: `185.241.214.214`
- Server #2 IP: `95.217.216.36`

You need to first configure SSH key-based authentication both ways (from Server #1 to Server #2 and vice versa) so that root SSH access is possible without password prompts.

Generate SSH keys on each server (if not already created):

```bash
ssh-keygen -t rsa -b 4096
```

Then copy each server’s public key to the other:

```bash
ssh-copy-id root@185.241.214.214
ssh-copy-id root@95.217.216.36
```

Verify passwordless SSH connections work:

```bash
ssh root@185.241.214.214
ssh root@95.217.216.36
```


### Step 2: Create Nameservers in Your DNS Zone

Using the domain `yourdomain.com` as an example, define two nameservers:

- `dns1.yourdomain.com` → points to Server #1 IP (A record)
- `dns2.yourdomain.com` → points to Server #2 IP (A record)

Add these A records in your domain's DNS provider for `yourdomain.com`.


### Step 3: Register Nameservers in OpenPanel

On **master** server, open the OpenAdmin panel:
* Navigate to **Settings > OpenPanel > Nameservers**
* Add both `dns1.yourdomain.com` and `dns2.yourdomain.com`

[![2025-07-09-17-30.png](https://i.postimg.cc/kXnvzCwW/2025-07-09-17-30.png)](https://postimg.cc/jCFfnGpj)

### Step 4: Enable DNS Clustering

On **master** server:

* Go to **OpenAdmin > Domains > DNS Cluster**
* Click **Enable DNS Clustering**

[![2025-07-09-17-32.png](https://i.postimg.cc/FzG3NfG3/2025-07-09-17-32.png)](https://postimg.cc/2LbVxSsS)

* Click **Add Server** and enter the IP of the slave server, then **Add**

[![2025-07-09-17-33.png](https://i.postimg.cc/7PX2C2MT/2025-07-09-17-33.png)](https://postimg.cc/3W4RVWqK)

### Test Your Cluster

Add a new domain on either server via a OpenPanel user account.

Then verify the DNS zone is synchronized on both servers using the `dig` command:

```bash
dig A +short yourdomain.com @185.241.214.214
dig A +short yourdomain.com @95.217.216.36
```

Replace `yourdomain.com` with the domain you added.

If both servers return the correct IP, your DNS clustering setup is working!
