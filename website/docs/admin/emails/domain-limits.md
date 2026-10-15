---
sidebar_position: 5
---

# Domain Limits

Set how many emails each domain can send per hour.

Use **OpenAdmin > Emails > Domain Limits** to access it.

Every domain gets the hourly email limit of its owner's plan. The limits are enforced by Postfwd, so it has to be enabled in [Emails > Settings](/docs/admin/emails/settings). A domain without a rule can send any number of emails.

## Domains Table

Each row shows the domain, its owner, and the emails sent in the current hour against its limit. The bar turns yellow at 80% of the limit and red once it's reached.

When a domain hits its limit, a **Limit hit** label appears. Click it to show the last lines from the mail log where Postfwd rejected that domain's emails.

## Change a Limit

- **Edit**: set a custom hourly limit for one domain, then click **Save**.
- **Restore**: set the domain's limit from its owner's plan again.
- **Remove**: delete the domain's rule. Click it a second time to confirm.

## Restore from Plan

Click **Restore from plan**, pick **All users** or one user, and click **Restore**. Every domain of the selection gets the limit from its owner's plan again, and custom limits are replaced.

Use it after changing the hourly email limit of a plan, or to create the rules when the table is empty.

## Remove Rules

Click **Remove rules**, pick **All users** or one user, and click **Remove** twice to confirm. Those domains can send any number of emails until their rules are restored from their plan.
