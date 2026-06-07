---
sidebar_position: 4
---

# Email Filters

OpenPanel uses [Sieve](http://sieve.info/) (via Dovecot) to filter incoming email messages for your accounts.

Each email address can have its own set of filters. Filters are applied in order, top to bottom, and can match on headers, body content, spam status, and more.

## Creating a Filter

1. Go to **OpenPanel > Emails > Email Filters**.
2. Select the email address you want to manage filters for.
3. Click **Add filter** and give the filter a name.
4. Configure one or more **conditions** — the fields and values that must match.
5. Add one or more **actions** to perform when conditions are met.
6. Click **Save filters**.

### Conditions

Each filter can match on **any** or **all** of its conditions:

| Field | Description |
|---|---|
| **From** | Sender address |
| **Subject** | Email subject line |
| **To** | Primary recipient |
| **Any recipient** | To, Cc, or Bcc |
| **Body** | Message body text |
| **Any header** | Any email header field |
| **Spam status** | SpamAssassin spam flag |
| **Spam score** | SpamAssassin numeric score |
| **List ID** | Mailing list identifier header |

Each condition supports: `contains`, `does not contain`, `equals`, `begins with`, `ends with`, `matches regex`, `does not match regex`.

### Actions

When conditions are met, one or more actions are executed:

| Action | Description |
|---|---|
| **Deliver to folder** | Move the message into a specific mailbox folder |
| **Redirect to email** | Forward the message to another address |
| **Discard message** | Silently delete the message |
| **Reject with message** | Bounce the message with a custom reason |
| **Mark as read** | Mark the message as seen |
| **Flag / star** | Add a flag to the message |
| **Stop processing** | Skip any remaining filters |

