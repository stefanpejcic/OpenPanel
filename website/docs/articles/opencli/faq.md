# FAQ 

`opencli faq` displays answers to the most frequently asked questions:

```bash
opencli faq
```

<details>
  <summary>Example output</summary>

```bash
# opencli faq

Frequently Asked Questions

1. What is the login link for admin panel?

LINK: https://demo.openpanel.org:2087/
------------------------------------------------------------

2. What is the login link for user panel?

LINK: https://demo.openpanel.org:2083/
------------------------------------------------------------

3. How to restart OpenAdmin or OpenPanel services?

- OpenPanel: docker restart openpanel
- OpenAdmin: service admin restart
------------------------------------------------------------

4. How to reset admin password?

execute command opencli admin password USERNAME NEW_PASSWORD
------------------------------------------------------------

5. How to create new admin account?

execute command opencli admin new USERNAME PASSWORD
------------------------------------------------------------

6. How to list admin accounts?

execute command opencli admin list
------------------------------------------------------------

7. How to check OpenPanel version?

execute command opencli --version
------------------------------------------------------------

8. How to update OpenPanel?

execute command opencli update --force
------------------------------------------------------------

9. How to disable automatic updates?

execute command opencli config update autoupdate off
------------------------------------------------------------

10. Where are the logs?

- User panel errors:      /var/log/openpanel/user/error.log
- User panel access log:  /var/log/openpanel/user/access.log
- Admin panel errors:     /var/log/openpanel/admin/error.log
- Admin panel access log: /var/log/openpanel/admin/access.log
- Admin panel API access: /var/log/openpanel/admin/api.log
- Admin panel logins:     /var/log/openpanel/admin/login.log
- Admin panel alerts:     /var/log/openpanel/admin/notifications.log
- Admin panel crons:      /var/log/openpanel/admin/cron.log
------------------------------------------------------------

11. How to enable detailed logs?

- opencli config update dev_mode yes
------------------------------------------------------------
```
</details>
