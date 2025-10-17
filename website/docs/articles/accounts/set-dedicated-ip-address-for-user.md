# Assigning a Dedicated IP Address to a User

A dedicated IP address can then be assigned to a user through **OpenAdmin** or **OpenCLI**.

Changing IP updates all related configuration files — including DNS zones and Caddy VirtualHosts — ensuring that the user’s all current and newly added domains are bound to the new IP address.
The updated IP address will also be reflected immediately in the user’s OpenPanel interface.

## Using OpenAdmin

The OpenAdmin interface lists all IP addresses currently assigned to the server — these can also be viewed using the `hostname -I` command.
If you need to add additional IP addresses to the server, **follow the official network configuration documentation for your operating system distribution**.
No special configuration is required from the OpenPanel side — newly added IPs will appear automatically in the interface.

To change IP address for a user from OpenAdmin:

1. Navigate to **OpenAdmin → Users → *username* → Edit**.

2. Update the **IP Address** field with the desired dedicated IP.

   ![image](https://i.postimg.cc/65f9Jcsf/slika.png)

3. Click **Save** to apply the changes.

## Using OpenCLI

To change a user’s IP address from the command line, use the [`opencli user-ip`](https://dev.openpanel.com/cli/users.html#Assign-Remove-IP-to-User) command:

```bash
opencli user-ip <USERNAME> <NEW_IP_ADDRESS>
```

If the specified IP address is already assigned to another user, the command will abort with a warning.
To override this behavior and force the reassignment, include the `-y` flag:

```bash
opencli user-ip <USERNAME> <NEW_IP_ADDRESS> -y
```

To remove a dedicated IP from a user and restore them to the [main (shared) IP](/docs/articles/install-update/openpanel-main-ip-address), use:

```bash
opencli user-ip <USERNAME> delete
```
