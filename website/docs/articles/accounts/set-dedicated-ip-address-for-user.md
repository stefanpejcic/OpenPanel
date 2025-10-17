# Set Dedicated IP for User

Dedicated IP address for the user cna be set from OpenAdmin or OpenCLI:

## OpenAdmin

Navigate to **OpenAdmin > Users > *username* > Edit** and change the IP address:

![image](https://i.postimg.cc/65f9Jcsf/slika.png)


Click on save.

## OpenCLI

IP address can be changed form the terminal with command: [opencli user-ip](https://dev.openpanel.com/cli/users.html#Assign-Remove-IP-to-User).

```bash
opencli user-ip <USERNAME> <NEW_IP_ADDRESS>
```

This will edit all domain configuration files for the user and bind them to the new ip address. User services in their OpenPanel will also see
