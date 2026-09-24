---
sidebar_label: "Setting the Max Processes (PIDs) Limit for a User"
---

# How to Set the Max Processes (PIDs) Limit for a User

Every user's `user-<uid>.slice` cgroup has a **TasksMax** ceiling — the maximum number of processes/threads (PIDs) the user's account can run at once. This isn't a plan tier customers shop for like CPU, RAM or disk — it's a safety net against fork bombs and runaway processes.

## How it's calculated

By default the PID limit is **autocalculated from the user's plan RAM**, at a rate of **500 tasks per 1GB of RAM**, with a floor of **100 tasks** for any plan with less than 1GB assigned. This keeps the limit proportional automatically whenever a plan's RAM value changes.

For example, a plan with 4GB RAM gets a TasksMax ceiling of 2000. A plan with 0GB (unlimited) RAM gets an unlimited TasksMax as well.

## Overriding the limit for a single user

If a specific user needs a custom PID limit — higher or lower than what their plan's RAM would derive — create a `TasksMax` file in their home directory containing a single integer:

```bash
echo "1000" > /home/<USERNAME>/TasksMax
```

This file, when present, always takes priority over the autocalculated value for that user.

## Applying the change

Creating or editing the `TasksMax` file does not take effect on its own — the plan must be (re)applied to the user for the new ceiling to be written to their `user-<uid>.slice`. Run:

```bash
opencli plan-apply <PLAN_ID> <USERNAME>
```

Using the user's current plan ID re-applies all of the plan's limits, including the new PID ceiling.

## Removing the override

Delete the `TasksMax` file and re-run `opencli plan-apply` for the user to fall back to the autocalculated value derived from their plan's RAM:

```bash
rm /home/<USERNAME>/TasksMax
opencli plan-apply <PLAN_ID> <USERNAME>
```
