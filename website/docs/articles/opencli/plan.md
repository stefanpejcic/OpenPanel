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

<details>
  <summary>Example output</summary>

```json
[
  {
    "id": "1",
    "name": "Standard plan",
    "description": "Small plan for testing",
    "email_limit": "0",
    "ftp_limit": "10",
    "domains_limit": "0",
    "websites_limit": "0",
    "disk_limit": "5 GB",
    "inodes_limit": "1000000",
    "db_limit": "0",
    "cpu": "2",
    "ram": "2g",
    "bandwidth": "10",
    "feature_set": "basic",
    "max_email_quota": "10G"
  }
]
[
  {
    "id": "2",
    "name": "Developer Plus",
    "description": "4 cores, 6G ram",
    "email_limit": "0",
    "ftp_limit": "10",
    "domains_limit": "0",
    "websites_limit": "0",
    "disk_limit": "10 GB",
    "inodes_limit": "1000000",
    "db_limit": "0",
    "cpu": "4",
    "ram": "6g",
    "bandwidth": "100",
    "feature_set": "default",
    "max_email_quota": "0"
  }
]
[
  {
    "id": "3",
    "name": "example",
    "description": "ddfsds",
    "email_limit": "1",
    "ftp_limit": "1",
    "domains_limit": "1",
    "websites_limit": "1",
    "disk_limit": "10 GB",
    "inodes_limit": "1000000",
    "db_limit": "1",
    "cpu": "1",
    "ram": "1g",
    "bandwidth": "100",
    "feature_set": "default",
    "max_email_quota": "10G"
  }
]

```

</details>


## Create Plan

To create a new plan run the following command:

```bash
opencli plan-create name"<TEXT>" description="<TEXT>" emails=<COUNT> ftp=<COUNT> domains=<COUNT> websites=<COUNT> disk=<COUNT> inodes=<COUNT> databases=<COUNT> cpu=<COUNT> ram=<COUNT> bandwidth=<COUNT> feature_set=<NAME> max_email_quota=<COUNT>
```

| Parameter           | Description                                      | Type    | Notes                                                         |
|---------------------|--------------------------------------------------|---------|---------------------------------------------------------------|
| `name`              | Name of the plan                                 | String  | No spaces                                                     |
| `description`       | Plan description                                 | String  | Use quotes for multiple words                                |
| `feature_set`       | Feature set assigned to the plan                 | String  | Must match an existing feature set name                      |
| `email_limit`       | Max number of email accounts                     | Integer | `0` for unlimited                                            |
| `max_email_quota`   | Max size per email account                       | String  | Integer followed by `B`, `k`, `M`, `G`, or `T`; `0` unlimited |
| `ftp_limit`         | Max number of FTP accounts                       | Integer | `0` for unlimited                                            |
| `domains_limit`     | Max number of domains                            | Integer | `0` for unlimited                                            |
| `websites_limit`    | Max number of websites                           | Integer | `0` for unlimited                                            |
| `disk_limit`        | Disk space limit in GB                           | Integer |                                                               |
| `inodes_limit`      | Max number of inodes                             | Integer | `0` for unlimited (minimum recommended: 250000)             |
| `db_limit`          | Max number of databases                          | Integer | `0` for unlimited                                            |
| `cpu`               | CPU core limit                                   | Integer |                                                               |
| `ram`               | RAM limit in GB                                  | Integer |                                                               |
| `bandwidth`         | Port speed in Mbit/s                             | Integer |                                                               |

Example:
```bash
opencli plan-create name="New Plan" description="This is a new plan" emails=100 ftp=50 domains=20 websites=30 disk=100 inodes=100000 databases=10 cpu=4 ram=8 bandwidth=100 feature_set=default max_email_quota=2G"
```

## List Users on Plan

List all users that are currently using a plan:

```bash
opencli plan-usage
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
opencli plan-usage --json
```

<details>
  <summary>Example output</summary>
  
```json
[
  {
    "id": "3",
    "username": "demo",
    "email": "stefan@netops.com",
    "plan_name": "Standard plan",
    "registered_date": "2025-04-28 14:47:52"
  }
]
[
  {
    "id": "4",
    "username": "dummy",
    "email": "dummy",
    "plan_name": "Standard plan",
    "registered_date": "2025-04-28 15:20:19"
  }
]
```
</details>


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
opencli plan-edit --debug id=<ID> name"<TEXT>" description="<TEXT>" emails=<COUNT> ftp=<COUNT> domains=<COUNT> websites=<COUNT> disk=<COUNT> inodes=<COUNT> databases=<COUNT> cpu=<COUNT> ram=<COUNT> bandwidth=<COUNT> feature_set=<DEFAULT> max_email_quota=<COUNT>
```

| Parameter           | Description                                      | Type    | Notes                                                         |
|---------------------|--------------------------------------------------|---------|---------------------------------------------------------------|
| `name`              | Name of the plan                                 | String  | No spaces                                                     |
| `description`       | Plan description                                 | String  | Use quotes for multiple words                                |
| `feature_set`       | Feature set assigned to the plan                 | String  | Must match an existing feature set name                      |
| `email_limit`       | Max number of email accounts                     | Integer | `0` for unlimited                                            |
| `max_email_quota`   | Max size per email account                       | String  | Integer followed by `B`, `k`, `M`, `G`, or `T`; `0` unlimited |
| `ftp_limit`         | Max number of FTP accounts                       | Integer | `0` for unlimited                                            |
| `domains_limit`     | Max number of domains                            | Integer | `0` for unlimited                                            |
| `websites_limit`    | Max number of websites                           | Integer | `0` for unlimited                                            |
| `disk_limit`        | Disk space limit in GB                           | Integer |                                                               |
| `inodes_limit`      | Max number of inodes                             | Integer | `0` for unlimited (minimum recommended: 250000)             |
| `db_limit`          | Max number of databases                          | Integer | `0` for unlimited                                            |
| `cpu`               | CPU core limit                                   | Integer |                                                               |
| `ram`               | RAM limit in GB                                  | Integer |                                                               |
| `bandwidth`         | Port speed in Mbit/s                             | Integer |                                                               |



<details>
  <summary>Example output</summary>

```bash
# opencli plan-edit --debug id=1 name="New Plan" description="This is a new plan" emails=100 ftp=50 domains=20 websites=30 disk=100 inodes=100000 databases=10 cpu=4 ram=8 bandwidth=100 feature_set="default" max_email_quota="2G"
```
</details>
