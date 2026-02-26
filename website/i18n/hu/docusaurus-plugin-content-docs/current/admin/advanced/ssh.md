---
sidebar_position: 3
---

# SSH hozzáférés

Az *OpenAdmin > Kiszolgáló > SSH-hozzáférés* lehetővé teszi a rendszergazdák számára, hogy megtekintsék és módosítsák a szerver aktuális SSH-konfigurációját.

### Alapvető SSH beállítások

![screenshot](/img/admin/ssh_access.png)

Ezen az oldalon a következők jelennek meg:

- **Port** - aktuális SSH port
- **PermitRootLogin** - bejelentkezés engedélyezése *root* felhasználó számára
- **PasswordAuthentication** - jelszavak használatának engedélyezése az ssh számára
- **PubkeyAuthentication** - ssh-kulcsok használatának engedélyezése

Bármely értéket megváltoztathatja, és az alkalmazáshoz kattintson a mentés gombra.

### Speciális SSH-beállítások

Itt szerkesztheti az SSH konfigurációs fájlt: `/etc/ssh/sshd_config`

### Engedélyezett SSH-kulcsok

Itt megtekintheti az aktuális engedélyezett ssh-kulcsokat, eltávolíthatja őket, vagy új kulcsot adhat hozzá.
