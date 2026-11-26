---
sidebar_position: 5
---

# Email Filters

OpenPanel uses [Sieve](http://sieve.info/) (via Dovecot) to filter incoming email messages for your accounts.

![Email Filters page with no filters yet and the Create first filter button](/img/openpanel-screenshots/emails/filters-page.png#gh-light-mode-only)
![Email Filters page with no filters yet and the Create first filter button](/img/openpanel-screenshots/emails/filters-page_dark.png#gh-dark-mode-only)

Each email address can have its own set of filters. Filters are applied in order, top to bottom, and can match on headers, body content, spam status, and more.

## Creating a Filter

1. Go to **OpenPanel > Email > Filters**.
2. Select the email address you want to manage filters for.
3. Click **Add filter** and give the filter a name.
4. Configure one or more **conditions**, the fields and values that must match.
5. Add one or more **actions** to perform when conditions are met.
6. Click **Save filters**.

![An email filter that moves messages whose subject contains newsletter into the Newsletters folder](/img/openpanel-screenshots/emails/filters-rule.png#gh-light-mode-only)
![An email filter that moves messages whose subject contains newsletter into the Newsletters folder](/img/openpanel-screenshots/emails/filters-rule_dark.png#gh-dark-mode-only)

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
| **Forward to email (keep copy)** | Forward the message to another address while keeping a copy in the mailbox |
| **Discard message** | Silently delete the message |
| **Reject with message** | Bounce the message with a custom reason |
| **Mark as read** | Mark the message as seen |
| **Flag / star** | Add a flag to the message |
| **Auto-responder (vacation)** | Send an automatic reply with a custom subject and message, limited to once every N days per sender |
| **Stop processing** | Skip any remaining filters |

## Edit the Sieve file

To write the Sieve rules by hand, switch to **File editor** at the top of the page. It shows the account's Sieve file, edit it and click **Save file**.

:::tip Unsaved changes
Unsaved edits in this editor are kept as a copy in your browser. If you leave the page, close the tab or the save fails, the next time you open the File editor for that email address a notice shows when that copy was made, with three options:

- **View changes** shows what's different from the saved file, line by line.
- **Restore** puts the copy back in the editor, then you can save it.
- **Discard** deletes the copy.

The copy stays only in that browser and is removed once the file is saved. Leaving the page with unsaved edits also asks you to confirm first.
:::
