# How to Set Up Email on Android (Gmail App)

This guide explains how to set up your **OpenPanel-created email account** in the Gmail app on Android.

> **Note:** If you are setting up email on a different device or service, see the main guide: [How to setup your email client](/docs/articles/email/how-to-setup-your-email-client)

---

## Add a New Account

1. Open the **Gmail App**  
   Go to:  
   **Menu → Settings → Add Account → Personal (IMAP/POP) → Next**  

2. Enter your **email address** → **Next**  

3. Choose **IMAP** or **POP**  
   - If unsure, select **IMAP** (recommended).  
   - Learn more: [📬 POP3 vs IMAP: Email Access Protocols](/docs/articles/email/imap-vs-pop3/)  

4. Enter your **email password** → **Next**  

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

---

### POP

| Setting        | Description                                      | Example                          |
|----------------|--------------------------------------------------|----------------------------------|
| Username       | Your full email address                          | user@domain.tld                  |
| Password       | Your email account password                      | ********                         |
| Server         | Incoming mail server address                     | mail.domain.tld                  |
| Port           | Port number for incoming mail                    | 993                              |
| Security Type  | Encryption method for secure connection          | SSL/TLS                          |
| Authentication | Authentication method used to log in             | Normal Password                  |

After entering your incoming mail settings, tap **Next**.

---

## Outgoing Mail Settings (SMTP)

Make sure **"Require Sign-in"** is enabled.  

| Setting        | Description                                      | Example                          |
|----------------|--------------------------------------------------|----------------------------------|
| Username       | Your full email address                          | user@domain.tld                  |
| Password       | Your email account password                      | ********                         |
| Server         | Outgoing mail server address                     | mail.domain.tld                  |
| Port           | Port number for outgoing mail                    | 465 / 587                        |
| Security Type  | Encryption method for secure connection          | SSL/TLS                          |
| Authentication | Authentication method used to log in             | Password                         |

After entering your outgoing mail settings, tap **Next**.

---

Your OpenPanel email account is now set up and ready to use in the Gmail app.  
