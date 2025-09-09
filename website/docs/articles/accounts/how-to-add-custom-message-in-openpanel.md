# How to display a custom message

OpenPanel supports both **per-user** and **global** custom messages that can be displayed in your users’ OpenPanel interfaces.

![example message](https://i.postimg.cc/QN0XSW0t/2025-08-14-14-49.png)

## Custom Message for a Specific User

To display a custom message in **OpenPanel > Dashboard** for a specific user:

1. Navigate to **OpenAdmin > Users > *username* > Overview**.

   ![add message](https://i.postimg.cc/KZBxwWJC/2025-08-14-14-50.png)

2. Enter the message — either plain text or HTML — in the **'Custom message for user'** area.

3. Click **Save**.

The message will appear immediately in the user’s Dashboard.

### Adding a Message via Terminal

You can also add a custom message by creating a file:

```bash
/etc/openpanel/openpanel/core/users/USERNAME/custom_message.html
```

Replace `USERNAME` with the actual user’s username.

---

## Global Message for All Users

To display a message for **all users**, copy or create a symbolic link of the custom message file in each user’s directory: `/etc/openpanel/openpanel/core/users/USERNAME/custom_message.html`.
