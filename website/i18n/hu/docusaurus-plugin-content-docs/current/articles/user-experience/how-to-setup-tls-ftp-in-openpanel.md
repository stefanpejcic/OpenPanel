# Setup TLS for FTP

This guide explains how to enable Explicit FTP over TLS (FTPS) in OpenPanel using VSFTPD.

---

> [OpenPanel Enterprise](https://openpanel.com/enterprise/) supports FTP access. Once enabled, a shared FTP server becomes available for all users.
> Elsőként kövesse ezt az útmutatót az FTP-szerver konfigurálásához: [FTP szerver beállítása](./how-to-setup-ftp-in-openpanel.md)

---

## Step 1: Add a Domain

To use TLS, you must have an SSL certificate available on the server. Start by adding a domain in one of the following ways:

* **As an OpenAdmin user** (for OpenPanel/OpenAdmin access):
  [https://openpanel.com/docs/admin/settings/general/#domain](https://openpanel.com/docs/admin/settings/general/#domain)

* **As an OpenPanel user**:
  [https://openpanel.com/docs/panel/domains/new/](https://openpanel.com/docs/panel/domains/new/)

A hozzáadást követően az OpenPanel automatikusan létrehoz egy SSL-tanúsítványt a tartományhoz (ha támogatott).

---

## 2. lépés: Ellenőrizze az SSL-tanúsítványfájlokat

Győződjön meg arról, hogy az SSL-tanúsítvány és a privát kulcs fájlok léteznek a kiszolgálón.
Ha a tanúsítványt automatikusan generálták, akkor az útvonalaknak így kell kinézniük:

```bash
/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/DOMAIN_NAME/DOMAIN_NAME.crt
/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/DOMAIN_NAME/DOMAIN_NAME.key
```

Cserélje le a `DOMAIN_NAME' nevet a tényleges domain nevére.

---

## 3. lépés: Engedélyezze a TLS-t a VSFTPD-ben

1. Lépjen az **OpenAdmin → Szolgáltatások → FTP → Konfiguráció** elemre.
![image](https://i.postimg.cc/DZyrjCXF/1.png)
3. Illessze be a következő konfigurációs beállításokat a TLS engedélyezéséhez:
  
⚠️ **Fontos:** Cserélje le a `DOMAIN_NAME'-t a domainnevével mindkét fájlútvonalban.
  
``` bash
# SSL engedélyezése
ssl_enable=IGEN
rsa_cert_file=/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/DOMAIN_NAME/DOMAIN_NAME.crt
rsa_private_key_file=/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/DOMAIN_NAME/DOMAIN_NAME.key
   
allow_anon_ssl=NEM
force_local_data_ssl=IGEN
force_local_logins_ssl=IGEN
  
ssl_tlsv1=NEM
ssl_sslv2=NEM
ssl_sslv3=NEM
ssl_ciphers=MAGAS
   ```

3. Kattintson a **Mentés** gombra a módosítások alkalmazásához.
![image2](https://i.postimg.cc/52KQ9xYt/2.png)

---

## 4. lépés: Indítsa újra az FTP szolgáltatást

1. Keresse meg az **OpenAdmin → Szolgáltatások → Szolgáltatások állapota** lehetőséget.
2. Keresse meg az **FTP** szolgáltatást.
3. Kattintson a **Stop**, majd a **Start** gombra a szolgáltatás újraindításához.

![image2](https://i.postimg.cc/0vTJnjCB/3.png)
