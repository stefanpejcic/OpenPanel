# Az OpenPanel telepítése a Cloud-Init segítségével

Használja a következő **cloud-init YAML** konfigurációt az OpenPanel automatikus telepítéséhez a szerveren az első rendszerindításkor:

```yaml
#cloud-config
packages:
  - curl
  - lsb-release
  - gnupg

runcmd:
  - curl -fsSL https://openpanel.org/ | bash
```
