# How To Free Up Disk Space

Low disk space can affect server stability and performance. This guide helps you identify and free up disk space on your OpenPanel server.

---

### 1. Check Disk Usage

Start by checking how much space is currently used:

```bash
df -h
```

To find large directories:

```bash
du -h / | sort -hr | head -n 20
```

---

### 2. Clean Up Docker

Docker can quickly consume space with unused containers, images, volumes, and networks.

* To remove **unused** Docker data:

```bash
docker system prune
```

* To also remove unused **volumes** (use with caution):

```bash
docker system prune --volumes
```

> 🧼 **Note:** This only removes resources not actively in use. Review the list before confirming.

* View what's taking space:

```bash
docker system df
```

---

### 3. Clear Logs

Log files often grow large over time:

* Delete rotated/compressed logs:

```bash
rm -f /var/log/*.gz /var/log/*.1
```

* Truncate current logs:

```bash
truncate -s 0 /var/log/syslog
truncate -s 0 /var/log/auth.log
```

---

### 4. Remove Unused Packages

Free space by removing orphaned packages:

* For Debian/Ubuntu:

```bash
apt autoremove
apt clean
```

* For CentOS/RHEL:

```bash
yum autoremove
yum clean all
```

---

### 5. Clear Cache

Remove stored .deb or .rpm files:

* APT cache:

```bash
rm -rf /var/cache/apt/archives/*
```

* YUM cache:

```bash
rm -rf /var/cache/yum
```

---

### 6. Delete Temporary Files

Clean up temp directories:

```bash
rm -rf /tmp/*
rm -rf /var/tmp/*
```

---

### 7. Clean User Trash

For all users:

```bash
rm -rf /home/*/.cache/*
rm -rf /home/*/.local/share/Trash/*
```

---

### 8. Remove Old Backups

Search for old or backup files:

```bash
find / -type f \( -name "*.bak" -o -name "*.old" \)
```

Inspect before deleting.

---

### 9. Analyze with ncdu

Use `ncdu` for a navigable summary:

```bash
apt install ncdu      # Debian/Ubuntu
yum install ncdu      # CentOS/RHEL

ncdu /
```

---

### 10. Update OpenPanel (Optional)

Updating OpenPanel removes previous docker images

```bash
opencli update --force
```

---

> **Tip:** Enable disk usage alerts via **OpenAdmin > Settings > Notifications**.
