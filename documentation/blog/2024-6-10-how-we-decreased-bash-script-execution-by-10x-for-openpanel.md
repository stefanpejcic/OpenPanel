---
title: How we decreased account creation by 16x
description: Here are the tips how to drastically decrease bash script execution time
slug: how-we-decreased-bash-script-execution-by-10x-for-openpanel
authors: stefanpejcic
tags: [OpenPanel, news, dev]
image: https://openpanel.co/img/blog/how-we-decreased-bash-script-execution-by-10x-for-openpanel.png
hide_table_of_contents: true
---

How we decreased Docker account creation time for OpenPanel from an average of 40s to less than 3 seconds

<!--truncate-->


[OpenPanel](https://openpanel.co/) started as a LAMP stack that we built to fulfill our personal needs with a custom and reliable panel offering the freedom of VPS but with all the benefits of shared hosting: a low price tag and easy maintenance.

The very first script we developed for OpenPanel is the `opencli user-add` script, responsible for account creation, service setup, and file management.

## Initial Challenges

The script's original execution involved numerous steps, including:

- Username validation
- Password security checks
- Plan verification
- Docker image and configuration checks
- Container creation
- Firewall and permission setup
- Configuration file creation
- Database user addition

These processes required meticulous checks and conditional statements to ensure seamless execution.

## Execution time and debugging for bash

To troubleshoot bash scripts we use `bash -x` when running the script, example:

```bash
bash -x /usr/local/admin/scripts/user/add stefan stefan stefan default_plan_nginx --debug
```

With `-x` we get detailed output and can see where we have recursion steps that can be simplified and where our script takes the most time to execute.

To measure the time needed for the total execution we can prepend the `time` command:

```bash
time bash -x /usr/local/admin/scripts/user/add stefan stefan stefan default_plan_nginx --debug
```

This shows that the starting execution speed is 48s which is far from ideal, so lets run the script with `-x` and see what we can improve.

## Optimization Strategies

Here are the steps that we took to decrease the speed:

### 1. Simplified Argument Parsing

We streamlined the argument parsing process by directly checking for the `--debug` flag, eliminating unnecessary iterations through arguments.

Existing code checked each argument:

```bash
for arg in "$@"; do
    case $arg in
        --debug)
            DEBUG=true
            ;;
        *)
            ;;
    esac
done
```

but since we already expect each value on specific posisiton:
```
username="${1,,}"
password="$2"
email="$3"
plan_name="$4"
```

`--debug` can only really be on the 5th position, and by not checking all other, we run only 1 cehck instead of 5 (one per each arg).

```bash
# Parse optional flags to enable debug mode when needed
if [ "$5" = "--debug" ]; then
    DEBUG=true
fi
```




### 2. Enhanced Username Validation

Utilizing arrays and optimizing validation processes, we improved efficiency in checking username validity.


We can user [readarray](https://helpmanual.io/builtin/readarray/) for the forbidden usernames list and then merge all the checks:

```bash
is_username_forbidden() {
    local check_username="$1"
    readarray -t forbidden_usernames < "$FORBIDDEN_USERNAMES_FILE"

    # Check if the username meets all criteria
    if [[ "$check_username" =~ [[:space:]] ]] || [[ "$check_username" =~ [-_] ]] || \
       [[ ! "$check_username" =~ ^[a-zA-Z0-9]+$ ]] || \
       (( ${#check_username} < 3 || ${#check_username} > 20 )); then
        return 0
    fi

    # Check against forbidden usernames
    for forbidden_username in "${forbidden_usernames[@]}"; do
        if [[ "${check_username,,}" == "${forbidden_username,,}" ]]; then
            return 0
        fi
    done

    return 1
}

# Validate username
if is_username_forbidden "$username"; then
    echo "Error: The username '$username' is not valid. Ensure it is a single word with no hyphens or underscores, contains only letters and numbers, and has a length between 3 and 20 characters."
    exit 1
fi
```

The array does help with speed, and the rest is just pure cosmetic changes - since we are rewritting the entire script anyways. :)


### 3. Reused Common Operations

By sourcing a separate bash script for common operations, we minimized redundancy across scripts, enhancing maintainability.

In every script that deals with user accounts we need to connect to database and read the data for that user. We can create a separate bash script to hold just that logic, and then include it in our other scripts when needed:

```bash
DB_CONFIG_FILE="/usr/local/admin/scripts/db.sh"
source "$DB_CONFIG_FILE"
```

This saves time in the future, because we will only need to edit one file instead of all files that deal with database.


### 4. Efficient Database Data Retrieval

Optimizing database data retrieval using `read` and reducing reliance on multiple echo commands streamlined the process.

We read from database the selected plan information:

```bash
# Fetch DOCKER_IMAGE, DISK, CPU, RAM, INODES, BANDWIDTH and NAME for the given plan_name from the MySQL table
query="SELECT cpu, ram, docker_image, disk_limit, inodes_limit, bandwidth, name, storage_file, id FROM plans WHERE name = '$plan_name'"

# Execute the MySQL query and store the results in variables
cpu_ram_info=$(mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "$query" -sN)

# Check if the query was successful
if [ $? -ne 0 ]; then
    echo "Error: Unable to fetch plan information from the database."
    exit 1
fi

# Check if any results were returned
if [ -z "$cpu_ram_info" ]; then
    echo "Error: Plan with name $plan_name not found. Unable to fetch Docker image and CPU/RAM limits information from the database."
    exit 1
fi

# Extract DOCKER_IMAGE, DISK, CPU, RAM, INODES, BANDWIDTH and NAME,values from the query result
disk_limit=$(echo "$cpu_ram_info" | awk '{print $4}' | sed 's/ //;s/B//')
cpu=$(echo "$cpu_ram_info" | awk '{print $1}')
ram=$(echo "$cpu_ram_info" | awk '{print $2}')
inodes=$(echo "$cpu_ram_info" | awk '{print $6}')
bandwidth=$(echo "$cpu_ram_info" | awk '{print $7}')
name=$(echo "$cpu_ram_info" | awk '{print $8}')
storage_file=$(echo "$cpu_ram_info" | awk '{print $9}' | sed 's/ //;s/B//')
plan_id=$(echo "$cpu_ram_info" | awk '{print $11}')
```

This ends being 6 echo statements of the same thing! To avoid that, we can use [read](https://linuxcommand.org/lc3_man_pages/readh.html) to assign the values from the database results:



```bash
# Execute the MySQL query and store the results in variables
read cpu ram docker_image disk_limit inodes bandwidth name storage_file plan_id <<< $(mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "$query" -sN | tr -d '[:space:]')

# Check if the query was successful
if [ $? -ne 0 ] || [ -z "$name" ]; then
    echo "Error: Unable to fetch plan information from the database or plan with name $plan_name not found."
    exit 1
fi

# Extract and process values from the query result
disk_limit="${disk_limit//[!0-9]/}"  # Remove non-numeric characters
ram="${ram//[!0-9]/}"  # Remove non-numeric characters
storage_file="${storage_file//[!0-9]/}"  # Remove non-numeric characters
plan_id="${plan_id//[!0-9]/}"  # Remove non-numeric characters

# Convert RAM to GB if it's in the format "Xg"
if [[ "$ram" == *g ]]; then
    ram="${ram%g}"
fi

```

This saves IO.

### 5. Consolidated Conditions

Merging conditions optimized processing time and reduced redundancy, enhancing script efficiency.

We have a lot of `if else` checks for varius cases, like this one that checks if plan has disk limits, and if it does, then checks the set storage driver for Docker:

```bash

# create storage file
if [ "$storage_file" -eq 0 ]; then
    if [ "$DEBUG" = true ]; then
    echo "Storage file size is 0. Skipping storage file creation."
    fi
else
    if [ "$storage_driver" == "overlay" ] || [ "$storage_driver" == "overlay2" ]; then
        echo "Run without creating /home/storage_file_$username"
    elif [ "$storage_driver" == "devicemapper" ]; then
        if [ "$DEBUG" = true ]; then
            fallocate -l ${storage_file}g /home/storage_file_$username
            mkfs.ext4 -N $inodes /home/storage_file_$username
        else
            fallocate -l ${storage_file}g /home/storage_file_$username > /dev/null 2>&1
            mkfs.ext4 -N $inodes /home/storage_file_$username > /dev/null 2>&1
        fi
    fi

fi
```

These types of checks that can not be avoided, should be run later in the file only in case that all previous checks succeded.


### 6. Skipped Unnecessary Checks

We avoided redundant checks such as username length and Docker image existence, focusing on essential operations.


Since Docker already downloads image if it does not exist locally, and displays error when name is 2 characters or less, we don't really need to have the same checks in our code.

### 7. Optimized Docker Container Setup

Selective service startup in Docker containers based on user requirements reduced initialization time and resource usage.


Docker does recommend runign one process per contianer, but that is not what we are doing here. We need several services running in the user contianer when account is created. But since we only need some at the beginning, and not others, like nginx since user has no domains yet, mysql since user has no databases yet, etc. We can edit the Dockerfile to start with all services stopped, and use the entrypoint script to start what is needed.


### 8. Utilized Skeleton Files

Introduction of skeleton files for common user data minimized read/write operations during account creation, improving efficiency.

We store cached configuration files for each user on our host server, in order to avoid reading the data from container often. That data is created when we start the container, but instead of creating the files and folders for each user, we can just create a skeleton directory and place all needed files&fodlers in that directory. Then when creating a new account we just copy the directory to our new user.


### 9. Minimized Operations

Reducing firewall port operations by checking for IPv6 address existence and avoiding unnecessary port openings optimized resource utilization.

For each account we open a total of 8 ports on UFW:
- 4 ipv4 ports
- 4 ipv6 ports

And since UFW only allows 1 port to be specified at a time, we need to run a total of 4 commands. but since majority of servers don't yet have ipv6, we can add one additional check if ipv6 is enabled, and if it is we will have a total of 9 commands (8 existing to open ports and 1 more for checking ipv6), but if not, we lower it to just 5 (4 ipv4 ports and 1 checking ipv6).

This saves time for majority of users that do not have ipv6 yet.


### 10. Implemented Background Processing

Running post-account creation processes such as PHP version retrieval in the background optimized overall execution time.

After each account is created we run `opencli php-get_available_php_versions $username` to generate a list of php versions that user can install and have it ready when they enter the interface.

This is a time expensive opperation because we need to update package manager in users container and then check latest php versions:

```bash
root@stefan:~/my_project# time opencli php-get_available_php_versions stefan
PHP versions for user stefan have been updated and stored in /home/stefan/etc/.panel/php/php_available_versions.json.

real	0m8.660s
user	0m0.042s
sys	0m0.037s
```

We can append `&` to the end of the command and make it run in background, so that out script can continue execution:

```
opencli php-get_available_php_versions $username &
```

This saves on average up to 10s per account!

## Achievements

Through these optimizations, we achieved a significant reduction in Docker account creation time for OpenPanel, from an average of 48 seconds to less than 3 seconds. These enhancements not only improved user experience but also optimized resource utilization, marking a milestone in our journey of script optimization and efficiency improvement.

[OpenPanel 0.2.1 Changelog](https://openpanel.co/docs/changelog/0.2.1/#openpanel-blacklist)



