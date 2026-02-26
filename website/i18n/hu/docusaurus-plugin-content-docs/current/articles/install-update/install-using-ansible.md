# Az OpenPanel telepítése az Ansible-n keresztül

Az OpenPanel egy vagy több célgépre történő telepítéséhez használja a következő **Ansible playbook**-t:

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

Mentse ezt "install-openpanel.yml" néven, és futtassa:

```bash
ansible-playbook -i inventory install-openpanel.yml
```

Ez automatikusan telepíti az OpenPanel-t a készletben lévő összes gépen.
