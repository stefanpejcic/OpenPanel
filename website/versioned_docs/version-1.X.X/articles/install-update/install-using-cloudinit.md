# Installing OpenPanel via Cloud-Init

Use the following **cloud-init YAML** configuration to automatically install OpenPanel on your server during first boot:

```yaml
#cloud-config
packages:
  - curl
  - lsb-release
  - gnupg

runcmd:
  - curl -fsSL https://openpanel.org/ | bash
```
