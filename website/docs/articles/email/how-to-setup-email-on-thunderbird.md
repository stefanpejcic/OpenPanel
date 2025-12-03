# How to Set Up Email on Thunderbird

This guide explains how to set up your **OpenPanel-created email account** in the Thunderbird email app.

> **Note:** If you are setting up email on a different device or service, see the main guide:
> [How to setup your email client](/docs/articles/email/how-to-setup-your-email-client)

---

## Add a New Account

1. Open **Thunderbird → File → New → Existing Mail Account**  

2. Enter the following details:  
   - **Your Name**: Your display name  
   - **Email Address**: Your full email address  
   - **Password**: Your email account password  
   
3. Click **Continue**  

> Thunderbird will attempt to automatically detect your settings. If an error occurs, ignore it and enter the details manually.

---

## Choose Account Type

Thunderbird allows **IMAP** or **POP**.  
- If unsure, select **IMAP** (recommended).  
- Learn more: [📬 IMAP vs POP3](/docs/articles/email/imap-vs-pop3/)

---

## Incoming Mail Settings

### IMAP (Recommended)

| Setting        | Description                                      | Example                          |
|----------------|--------------------------------------------------|----------------------------------|
| Username       | Your full email address                          | user@domain.tld                  |
| Password       | Your email account password                      | ********                         |
| Server         | Incoming mail server address                     | mail.domain.tld                  |
| Port           | Port number for incoming mail                    | 993                              |
| Security Type  | Encryption method for secure connection          | SSL/TLS                          |
| Authentication | Authentication method used to log in             | Normal Password                  |

### POP

| Setting        | Description                                      | Example                          |
|----------------|--------------------------------------------------|----------------------------------|
| Username       | Your full email address                          | user@domain.tld                  |
| Password       | Your email account password                      | ********                         |
| Server         | Incoming mail server address                     | mail.domain.tld                  |
| Port           | Port number for incoming mail                    | 995                              |
| Security Type  | Encryption method for secure connection          | SSL/TLS                          |
| Authentication | Authentication method used to log in             | Normal Password                  |

After entering incoming mail settings, click **Done**.

---

## Outgoing Mail Settings (SMTP)

| Setting        | Description                                      | Example                          |
|----------------|--------------------------------------------------|----------------------------------|
| Username       | Your full email address                          | user@domain.tld                  |
| Password       | Your email account password                      | ********                         |
| Server         | Outgoing mail server address                     | mail.domain.tld                  |
| Port           | Port number for outgoing mail                    | 465 / 587                        |
| Security Type  | Encryption method for secure connection          | SSL/TLS                          |
| Authentication | Authentication method used to log in             | Normal Password                  |

---

Your email account is now ready to use in Thunderbird.
