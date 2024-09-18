# How to fix: failed to create runc console socket: mkdir /tmp/pty160439680

When running command: 

```bash
opencli user-login stefan
```

If you receive an error:
> failed to create runc console socket: mkdir /tmp/pty1604396800: no space left on device: unknown

It indicates that the disk space is at 100% and you need to first free some space on the server in order to login.

First, find the partition that is full:

```bash
df -h
```

Example output:
```bash
df -h
Filesystem      Size  Used Avail Use% Mounted on
tmpfs           3.2G  307M  2.9G  10% /run
/dev/vda1       296G  296G     0 100% /
tmpfs            16G  252K   16G   1% /dev/shm
tmpfs           5.0M     0  5.0M   0% /run/lock
/dev/dm-1        10G  5.7G  4.4G  57% /var/lib/docker/devicemapper/mnt/ef7bedfb84a4b7d216944174518a29d79edbc11a17546f0a6f9f691128a0577d
/dev/dm-3        10G  1.1G  8.9G  11% /var/lib/docker/devicemapper/mnt/ff5c36635ee829daf9f9aef6b23b02afb7ee34431a2c37c4321b97a1568ba435
/dev/dm-5       160G  6.5G  154G   5% /var/lib/docker/devicemapper/mnt/77de40032c66f2f6926209de4e58a0eced5113cdc32bb363127857b1442a2989

```

In this example we see that the entire / is full, so we can use a tools such as `ncdu` do check what is using the disk space:

```bash
ncdu /
```

Navigate through directories and delete any old backups or log files by pressing `D` on the keyboard.

After freeing disk space, repeat the command.
