# Plans

Scripts for creating and editing hosting plans (packages).

## List Plans

To list all current hosting packages (plans) run:

```bash
opencli plan-list
```

<details>
  <summary>Example output</summary>

```bash
# opencli plan-list
+----+----------------+------------------------+---------------+----------------+-------------+-----------+------------+--------------+----------+------+------+-----------+-------------+-----------------+
| id | name           | description            | domains_limit | websites_limit | email_limit | ftp_limit | disk_limit | inodes_limit | db_limit | cpu  | ram  | bandwidth | feature_set | max_email_quota |
+----+----------------+------------------------+---------------+----------------+-------------+-----------+------------+--------------+----------+------+------+-----------+-------------+-----------------+
|  1 | Standard plan  | Small plan for testing |             0 |             10 |           0 |         0 | 5 GB       |      1000000 |        0 | 2    | 2g   |        10 | basic       | 10G             |
|  2 | Developer Plus | 4 cores, 6G ram        |             0 |             10 |           0 |         0 | 10 GB      |      1000000 |        0 | 4    | 6g   |       100 | default     | 0               |
|  3 | example        | ddfsds                 |             1 |              1 |           1 |         1 | 10 GB      |      1000000 |        1 | 1    | 1g   |       100 | default     | 10G             |
+----+----------------+------------------------+---------------+----------------+-------------+-----------+------------+--------------+----------+------+------+-----------+-------------+-----------------+
```
</details>

You can also format the data as JSON:

```bash
opencli plan-list --json
```

## Create Plan

To create a new plan run the following command:

```bash
opencli plan-create name="<TEXT>" description="<TEXT>" emails=<COUNT> ftp=<COUNT> domains=<COUNT> websites=<COUNT> disk=<COUNT> inodes=<COUNT> databases=<COUNT> cpu=<COUNT> ram=<COUNT> bandwidth=<COUNT> feature_set=<NAME> max_email_quota=<COUNT> max_hourly_email=<COUNT> [reseller=<USERNAME>]
```

| Parameter           | Description                                      | Type    | Notes                                                         |
|---------------------|--------------------------------------------------|---------|---------------------------------------------------------------|
| `name`              | Name of the plan                                 | String  | No spaces                                                     |
| `description`       | Plan description                                 | String  | Use quotes for multiple words                                 |
| `feature_set`       | Feature set assigned to the plan                 | String  | Must match an existing feature set name                       |
| `emails`            | Max number of email accounts                     | Integer | `0` for unlimited                                             |
| `max_email_quota`   | Max size per email account                       | String  | Integer followed by `B`, `k`, `M`, `G`, or `T`; `0` unlimited |
| `max_hourly_email`  | Max outgoing emails per hour, across all domains | Integer |                                                               |
| `ftp`               | Max number of FTP accounts                       | Integer | `0` for unlimited                                             |
| `domains`           | Max number of domains                            | Integer | `0` for unlimited                                             |
| `websites`          | Max number of websites                           | Integer | `0` for unlimited                                             |
| `disk`              | Disk space limit in GB                           | Integer |                                                               |
| `inodes`            | Max number of inodes                             | Integer | `0` for unlimited (minimum recommended: 250000)               |
| `databases`         | Max number of databases                          | Integer | `0` for unlimited                                             |
| `cpu`               | CPU core limit                                   | Integer |                                                               |
| `ram`               | RAM limit in GB                                  | Integer |                                                               |
| `bandwidth`         | Port speed in Mbit/s                             | Integer |                                                               |
| `reseller`          | Reseller that can use this plan                  | String  | Optional, username of an existing reseller                    |

Example:
```bash
opencli plan-create name="New Plan" description="This is a new plan" emails=100 ftp=50 domains=20 websites=30 disk=100 inodes=100000 databases=10 cpu=4 ram=8 bandwidth=100 feature_set=default max_email_quota=2G max_hourly_email=1000
```

## List Users on Plan

List all users that are currently using a plan:

```bash
opencli plan-usage <PLAN_NAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli plan-usage 'Standard plan'
+----+----------+-------------------+---------------+---------------------+
| id | username | email             | plan_name     | registered_date     |
+----+----------+-------------------+---------------+---------------------+
|  3 | demo     | stefan@netops.com | Standard plan | 2025-04-28 14:47:52 |
|  4 | dummy    | dummy             | Standard plan | 2025-04-28 15:20:19 |
+----+----------+-------------------+---------------+---------------------+

```
</details>

You can also format the data as JSON:

```bash
opencli plan-usage <PLAN_NAME> --json
```


## Delete Plan

Delete a plan if no users are currently using it.

```bash
opencli plan-delete <PLAN_NAME> 
```

<details>
  <summary>Example output</summary>
  
```bash
# opencli plan-delete 'ubuntu_nginx_mysql'
Plan 'ubuntu_nginx_mysql' deleted successfully.
```
</details>

TIP: use `'` or `"` around the plan name if it contains spaces: `"plan name here"`.

`--json` flag can be passed to return the response as JSON.

<details>
  <summary>Example output</summary>

```bash
# opencli plan-delete 'ubuntu_nginx_mysql'  --json
{"message": "Plan 'ubuntu_nginx_mysql' deleted successfully."}
```
</details>


## Edit Plan

Change plan limits.

```bash
opencli plan-edit id=<ID> name="<TEXT>" description="<TEXT>" emails=<COUNT> ftp=<COUNT> domains=<COUNT> websites=<COUNT> disk=<COUNT> inodes=<COUNT> databases=<COUNT> cpu=<COUNT> ram=<COUNT> bandwidth=<COUNT> feature_set=<NAME> max_email_quota=<COUNT> max_hourly_email=<COUNT> [--debug]
```

| Parameter           | Description                                      | Type    | Notes                                                         |
|---------------------|--------------------------------------------------|---------|---------------------------------------------------------------|
| `id`                | ID of the plan to edit                           | Integer | Required                                                      |
| `name`              | Name of the plan                                 | String  | No spaces                                                     |
| `description`       | Plan description                                 | String  | Use quotes for multiple words                                 |
| `feature_set`       | Feature set assigned to the plan                 | String  | Must match an existing feature set name                       |
| `emails`            | Max number of email accounts                     | Integer | `0` for unlimited                                             |
| `max_email_quota`   | Max size per email account                       | String  | Integer followed by `B`, `k`, `M`, `G`, or `T`; `0` unlimited |
| `max_hourly_email`  | Max outgoing emails per hour, across all domains | Integer |                                                               |
| `ftp`               | Max number of FTP accounts                       | Integer | `0` for unlimited                                             |
| `domains`           | Max number of domains                            | Integer | `0` for unlimited                                             |
| `websites`          | Max number of websites                           | Integer | `0` for unlimited                                             |
| `disk`              | Disk space limit in GB                           | Integer |                                                               |
| `inodes`            | Max number of inodes                             | Integer | `0` for unlimited (minimum recommended: 250000)               |
| `databases`         | Max number of databases                          | Integer | `0` for unlimited                                             |
| `cpu`               | CPU core limit                                   | Integer |                                                               |
| `ram`               | RAM limit in GB                                  | Integer |                                                               |
| `bandwidth`         | Port speed in Mbit/s                             | Integer |                                                               |



<details>
  <summary>Example output</summary>

```bash
# opencli plan-edit --debug id=1 name="New Plan" description="This is a new plan" emails=100 ftp=50 domains=20 websites=30 disk=100 inodes=100000 databases=10 cpu=4 ram=8 bandwidth=100 feature_set="default" max_email_quota="2G" max_hourly_email=1000
```
</details>

## Apply Plan

Editing a plan (above) does not by itself change the limits already applied to users on that plan. To move users to a new plan and (re)apply that plan's limits to them:

```bash
opencli plan-apply <plan_id> <username1> <username2>... [--all] [--debug]
```

Use `--all` instead of listing usernames to apply it to every user currently on that plan.

By default all limits (CPU, RAM, disk, bandwidth, email) are applied. Restrict it to specific limits with:

- `--cpu` - apply the CPU limit only.
- `--ram` - apply the RAM limit only.
- `--dsk` - apply the disk limit only.
- `--net` - apply the bandwidth limit only.
- `--email` - apply the email rate limit only.

These flags can be combined, e.g. `--cpu --ram` applies only CPU and RAM limits.
