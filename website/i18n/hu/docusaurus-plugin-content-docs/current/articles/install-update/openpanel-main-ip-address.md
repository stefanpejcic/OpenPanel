# Fő IP az OpenPanelben

A szerver hálózati konfigurációja határozza meg a szerver fő IP-címét. A fő IP-cím **nem** konfigurálható az OpenPanel segítségével.

A hálózati rendszergazda vagy a tárhelyszolgáltató csak módosíthatja. A fő IP-cím az az IP-cím, amelyet a kernel alapértelmezett kimenő IP-címként választ ki. Emiatt ez az az IP-cím, amelyet az OpenPanel licencrendszer a szervere IP-címeként használ.

Az aktuális IP-cím ellenőrzéséhez:

```bash
curl -4Lk https://ip.openpanel.com
```

---

Ha egy adott IP-címet szeretne hozzárendelni egy felhasználóhoz, lásd: [Hogyan állítsunk be dedikált IP-címet egy felhasználó számára] (/docs/articles/accounts/set-dedicated-ip-address-for-user)
