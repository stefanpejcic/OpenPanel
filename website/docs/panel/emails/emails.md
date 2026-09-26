---
sidebar_position: 1
---

# Email Accounts

If the **Emails** module is enabled on the server and your user account has access, you'll see a table listing all current email accounts, the total number of accounts, a search bar, and an option to add a new account.

![Email Accounts page listing mailboxes with their storage usage and the Webmail, Manage and Connect Devices buttons](/img/openpanel-screenshots/emails/emails-list.png#gh-light-mode-only)
![Email Accounts page listing mailboxes with their storage usage and the Webmail, Manage and Connect Devices buttons](/img/openpanel-screenshots/emails/emails-list_dark.png#gh-dark-mode-only)

- The total number of email accounts is displayed at the top left.  
- A search bar is available on the top right for quick filtering.

From this interface, you can view all your email accounts and their details:

* **Account @ Domain**: the full email address
* **Storage Used**: quota usage with a percentage indicator (unit depends on the mailbox's allocated quota: KiB, MiB, GiB, or TiB)
* **Options**: buttons for Webmail, Manage, and Connect Devices

---

## Create Email

To create a new email account, click the **New Email** button.

![Create an Email form with the domain, username, password and storage quota fields](/img/openpanel-screenshots/emails/new-form.png#gh-light-mode-only)
![Create an Email form with the domain, username, password and storage quota fields](/img/openpanel-screenshots/emails/new-form_dark.png#gh-dark-mode-only)

---

## Webmail

Clicking the **Webmail** link autologins you into an email address and opens the RoundCube Webmail client in a new tab.

![Webmail, Manage and Connect Devices buttons in the row of an email account](/img/openpanel-screenshots/emails/emails-buttons.png#gh-light-mode-only)
![Webmail, Manage and Connect Devices buttons in the row of an email account](/img/openpanel-screenshots/emails/emails-buttons_dark.png#gh-dark-mode-only)

---

## Manage

Clicking **Manage** for an account opens the management screen where you can:

- Change the account password  
- Update storage quota  
- Suspend incoming and/or outgoing emails

![Edit Email account page with storage quota, incoming and outgoing mail toggles and a password field](/img/openpanel-screenshots/emails/edit-form.png#gh-light-mode-only)
![Edit Email account page with storage quota, incoming and outgoing mail toggles and a password field](/img/openpanel-screenshots/emails/edit-form_dark.png#gh-dark-mode-only)

---

## Bulk Actions

Tick the checkbox of one or more accounts, or the checkbox in the table header to select every account shown by the current search. A bar appears at the bottom of the page with the number selected, a **Clear** link and these actions:

![Two email accounts selected with the bulk actions bar offering suspend and unsuspend of incoming and outgoing email, Change quota, Change password and Delete](/img/openpanel-screenshots/emails/emails-bulk.png#gh-light-mode-only)
![Two email accounts selected with the bulk actions bar offering suspend and unsuspend of incoming and outgoing email, Change quota, Change password and Delete](/img/openpanel-screenshots/emails/emails-bulk_dark.png#gh-dark-mode-only)

| Action | What it does |
|---|---|
| **Suspend incoming** | Stops the selected accounts from receiving emails. |
| **Unsuspend incoming** | Lets them receive emails again. |
| **Suspend outgoing** | Stops the selected accounts from sending emails. |
| **Unsuspend outgoing** | Lets them send emails again. |
| **Change quota** | Sets the mailbox size in GB, `0` is unlimited. It can't be more than your plan allows. |
| **Change password** | Sets the same new password on the selected accounts. |
| **Delete** | Permanently deletes the selected accounts and all their emails. |

Click an action, confirm it, and it runs on the selected accounts one after another. When it's done the page reloads with a notice listing the accounts it worked for, or which ones failed and why.

## Connect Devices

Click **Connect Devices** to view the email server settings for that account.

![Connect Devices page with the username, incoming and outgoing servers and the IMAP and SMTP ports](/img/openpanel-screenshots/emails/connect-page.png#gh-light-mode-only)
![Connect Devices page with the username, incoming and outgoing servers and the IMAP and SMTP ports](/img/openpanel-screenshots/emails/connect-page_dark.png#gh-dark-mode-only)

Specific guides on how to set an email client using a specific device or service:

- [How to setup email on iPhone](/docs/articles/email/how-to-setup-email-on-iphone)
- [How to setup email on Android](/docs/articles/email/how-to-setup-email-on-android)
- [How to setup email on Apple Mail app](/docs/articles/email/how-to-setup-email-on-apple-mail-app)
- [How to setup email on Thunderbird](/docs/articles/email/how-to-setup-email-on-thunderbird)
- [How to setup email on Outlook 365 desktop app](/docs/articles/email/how-to-setup-email-on-outlook-365-desktop-app)
- [How to setup email on Gmail using desktop / browser](/docs/articles/email/how-to-setup-email-in-gmail)
