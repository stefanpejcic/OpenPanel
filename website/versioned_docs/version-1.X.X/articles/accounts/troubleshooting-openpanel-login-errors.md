# Login Errors

The following error messages may appear on the OpenPanel login page. Here's what each one means and how to resolve it:


### `Username and password are required.`

This error indicates that one or both of the login fields were left empty. 

**Solution:** Please enter both your username and password.

---

### `Unable to connect to database.`

This suggests that the MySQL service is not running, or the `users` table may be corrupted.

**Solution:** Check the database service status in **OpenAdmin > Services** and ensure MySQL is running correctly.

---

### `Your account is suspended. Please contact support.`

Your OpenPanel user account has been suspended, and login is currently disabled.

**Solution:** Contact your hosting provider to reactivate your account.

---

### `Unrecognized account. Please check username.`

The username you entered does not exist on the server.

**Solution:** Verify that you're using the correct login link and username. If the issue persists, contact your hosting provider’s support team.

---

### `Invalid password. Please try again.`

The password entered is incorrect.

**Solution:** If you’ve forgotten your password or continue to have issues, contact your hosting provider to request a password reset.

---

### `Invalid 2FA code. Please try again.`

The Two-Factor Authentication (2FA) code provided is not valid.

**Solution:** Generate a new 2FA code using your authentication app and try again.

---

### `Too many failed login attempts. Please try again later.`

Too many incorrect login attempts have been made from your IP address.

**Solution:** Contact your server administrator to review the log file at `/var/log/openpanel/user/failed_login.log`. They can unblock your IP or adjust the failed login attempt limits.
