# Offload Email Storage to a Remote NFS Server

## Architecture

- **OpenPanel server** — runs websites and mail services (Postfix/Dovecot)
- **NFS server** — stores `/var/mail/` data

---

## Before You Start

**Check your email storage path** in OpenAdmin > Emails > Settings you can configure where mail is stored. The default since v1.0 is `/var/mail/`. This setting is locked once email accounts exist, so check the current value and substitute it in the paths below.

**Check your package manager**: the commands below assume a Debian-based distro. Replace `apt` with your distro's package manager if needed.

---

## Step 1: Prepare the NFS Server

Install NFS:
```bash
apt update && apt install nfs-kernel-server
```

Create the mail storage directory:
```bash
mkdir -p /var/mail
```

Add the export — edit `/etc/exports`:
```bash
/var/mail  <OPENPANEL_SERVER_IP>(rw,sync,no_subtree_check,no_root_squash)
```
Apply:
```bash
exportfs -ra
systemctl restart nfs-kernel-server
```

---

## Step 2: Configure the OpenPanel Server

Install the NFS client:
```bash
apt update && apt install nfs-common
```

If mail is already running, stop it:
```bash
cd /usr/local/openmail && docker --context=default compose down mailserver
```

Back up existing mail data:
```bash
cp -a /var/mail /var/mail_backup
```

Create the mount point and mount the NFS share:
```bash
mkdir -p /var/mail
mount <NFS_SERVER_IP>:/var/mail /var/mail
```

Make the mount persistent — add to `/etc/fstab`:
```bash
<NFS_SERVER_IP>:/var/mail  /var/mail  nfs  defaults,_netdev  0  0
```

Apply:
```bash
mount -a
```

---

## Step 3: Restart Mail Services

```bash
cd /usr/local/openmail && docker --context=default compose up -d mailserver
```

Verify the mount:
```bash
df -h | grep mail
```

---

## Test

- Create a new email account and confirm it works
- Delete an existing account and confirm cleanup
- Reboot and confirm the NFS mount persists
