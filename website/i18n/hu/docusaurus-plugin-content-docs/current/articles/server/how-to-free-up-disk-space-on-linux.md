# Hogyan szabadíthatunk fel lemezterületet

A kevés lemezterület hatással lehet a szerver stabilitására és teljesítményére. Ez az útmutató segít azonosítani és lemezterületet felszabadítani az OpenPanel szerveren.

---

### 1. Ellenőrizze a Lemezhasználatot

Kezdje azzal, hogy ellenőrizze, mennyi hely van jelenleg felhasználva:

```bash
df -h
```

Nagy könyvtárak keresése:

```bash
du -h / | sort -hr | head -n 20
```

---

### 2. Clean Up Docker

A Docker gyorsan elfoglalhatja a helyet a fel nem használt tárolókkal, képekkel, kötetekkel és hálózatokkal.

* A **fel nem használt** Docker-adatok eltávolítása:

```bash
docker system prune
```

* A fel nem használt **kötetek** eltávolításához (használja óvatosan):

```bash
docker system prune --volumes
```

> 🧼 **Megjegyzés:** Ezzel csak a nem használt erőforrásokat távolítja el. A megerősítés előtt tekintse át a listát.

* Nézze meg, mi foglal helyet:

```bash
docker system df
```

---

### 3. Naplók törlése

A naplófájlok idővel gyakran nagyra nőnek:

* Elforgatott/tömörített naplók törlése:

```bash
rm -f /var/log/*.gz /var/log/*.1
```

* Az aktuális naplók csonkolása:

```bash
truncate -s 0 /var/log/syslog
truncate -s 0 /var/log/auth.log
```

---

### 4. Távolítsa el a fel nem használt csomagokat

Szabad hely az árva csomagok eltávolításával:

* Debian/Ubuntu esetén:

```bash
apt autoremove
apt clean
```

* CentOS/RHEL esetén:

```bash
yum autoremove
yum clean all
```

---

### 5. Törölje a gyorsítótárat

Távolítsa el a tárolt .deb vagy .rpm fájlokat:

* APT gyorsítótár:

```bash
rm -rf /var/cache/apt/archives/*
```

* YUM gyorsítótár:

```bash
rm -rf /var/cache/yum
```

---

### 6. Törölje az ideiglenes fájlokat

Az ideiglenes könyvtárak tisztítása:

```bash
rm -rf /tmp/*
rm -rf /var/tmp/*
```

---

### 7. Tisztítsa meg a felhasználói szemetet

Minden felhasználó számára:

```bash
rm -rf /home/*/.cache/*
rm -rf /home/*/.local/share/Trash/*
```

---

### 8. Távolítsa el a régi biztonsági másolatokat

Régi vagy biztonsági másolat fájlok keresése:

```bash
find / -type f \( -name "*.bak" -o -name "*.old" \)
```

Törlés előtt ellenőrizze.

---

### 9. Elemezze az ncdu segítségével

Használja az ncdu-t navigálható összefoglalóhoz:

```bash
apt install ncdu      # Debian/Ubuntu
yum install ncdu      # CentOS/RHEL

ncdu /
```

---

### 10. Az OpenPanel frissítése (opcionális)

Az OpenPanel frissítése eltávolítja a korábbi docker-képeket

```bash
opencli update --force
```

---

> **Tip:** Enable disk usage alerts via **OpenAdmin > Settings > Notifications**.
