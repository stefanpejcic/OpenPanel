---
sidebar_position: 5
---

# Create User

Easily create a new MongoDB user in just a few clicks. Users are used for interacting with databases or the MongoDB server.

Navigate to **MongoDB > Create User**:

![Create MongoDB User form with the username and password fields](/img/openpanel-screenshots/mongodb/new_user-form.png#gh-light-mode-only)
![Create MongoDB User form with the username and password fields](/img/openpanel-screenshots/mongodb/new_user-form_dark.png#gh-dark-mode-only)

1. **Enter a Username**
   Type your desired username into the input field. Only letters, numbers, and underscores are allowed, and the username must be between 1 and 31 characters long.

2. **Enter a Password**
   Type your desired password in the second input field. To generate a strong random password, click the button next to the input field.

3. **Click 'Create User'**
   Once you enter or generate the username and password, click the **Create User** button to create your new MongoDB user.

A new user is created with no roles - use [Assign User to DB](/docs/panel/mongodb/assign) to grant them access to a database.

---

## Best Practices

- Use short and descriptive names (e.g., `app_user`, `analytics_user`).
- Avoid using uppercase letters, spaces, or special characters.
- Make sure the username is meaningful for easy identification later.
