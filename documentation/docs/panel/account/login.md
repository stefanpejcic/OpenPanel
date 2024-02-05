---
sidebar_position: 9
---

# Login and Password Reset

By default, the port on which OpenPanel is accessible is 2083, so you should be able to access your account under `IP:2083` or `DOMAIN:2083`

However, the port can be changed by the administrator, if that is the case, you would log in via `IP:custom-port` or `DOMAIN:custom-port`

## Login

To log in to OpenPanel, you need a username and password.

![login.png](/img/panel/v1/account/login.png)
![login_failed.png](/img/panel/v1/account/login_failed.png)

If 2FA is enabled on your account, you will also be prompted to provide the 2FA code after providing the correct username and password.

![login_2fa.png](/img/panel/v1/account/login_2fa.png)


## Password Reset

In case you forgot your OpenPanel password, it can be reset in the following ways:

- From the "Forgot Password" link on the login form. *If enabled by the Administrator.
- From the "Account > Settings" when you are logged in.
- Manually by the Administrator. 

If you forgot your password and the password reset option is enabled on the server, you can use your email address to reset the password.

Simply click on the 'Forgot Password?' link and input your email address. Within seconds, you will receive a reset link to your email. Click on the reset link within the next 30 minutes to set a new password.

![reset page](/img/reset_pass.png)

In case you do not see the "Forgot Password" link on the login form, this means that the option is disabled by the Administrator, and you need to contact them to perform the password reset for you.


## Logout

To log out of the OpenPanel, simply click on the logout link under your profile.
