# Installing OpenPanel via Ansible

Use the following **Ansible playbook** to install OpenPanel on one or more target machines:

```yaml
---
- name: Install OpenPanel on target machine
  hosts: all
  become: true
  vars:
    # Customize installation flags as needed. Full list: https://openpanel.com/install
    openpanel_install_flags: "--debug --username=admin --password=super123"

  tasks:
    - name: Download and run OpenPanel installer
      shell: |
        curl -sSL https://openpanel.org | bash -s -- {{ openpanel_install_flags }}
      args:
        executable: /bin/bash
```

Save this as `install-openpanel.yml` and run:

```bash
ansible-playbook -i inventory install-openpanel.yml
```

This will automatically install OpenPanel on all machines in your inventory.
