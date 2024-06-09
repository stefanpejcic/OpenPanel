---
title: How we decreased account creation by 16x
description: Here are the tips how to drastically decrease bash script execution time
slug: how-we-decreased-bash-script-execution-by-10x-for-openpanel
authors: stefanpejcic
tags: [OpenPanel, news, dev]
image: https://openpanel.co/img/blog/openapnel_enterprise.png
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

## Optimization Strategies

### 1. Simplified Argument Parsing

We streamlined the argument parsing process by directly checking for the `--debug` flag, eliminating unnecessary iterations through arguments.

### 2. Enhanced Username Validation

Utilizing arrays and optimizing validation processes, we improved efficiency in checking username validity.

### 3. Reused Common Operations

By sourcing a separate bash script for common operations, we minimized redundancy across scripts, enhancing maintainability.

### 4. Efficient Database Data Retrieval

Optimizing database data retrieval using `read` and reducing reliance on multiple echo commands streamlined the process.

### 5. Consolidated Conditions

Merging conditions optimized processing time and reduced redundancy, enhancing script efficiency.

### 6. Skipped Unnecessary Checks

We avoided redundant checks such as username length and Docker image existence, focusing on essential operations.

### 7. Optimized Docker Container Setup

Selective service startup in Docker containers based on user requirements reduced initialization time and resource usage.

### 8. Utilized Skeleton Files

Introduction of skeleton files for common user data minimized read/write operations during account creation, improving efficiency.

### 9. Minimized Operations

Reducing firewall port operations by checking for IPv6 address existence and avoiding unnecessary port openings optimized resource utilization.

### 10. Implemented Background Processing

Running post-account creation processes such as PHP version retrieval in the background optimized overall execution time.

## Achievements

Through these optimizations, we achieved a significant reduction in Docker account creation time for OpenPanel, from an average of 40 seconds to less than 3 seconds. These enhancements not only improved user experience but also optimized resource utilization, marking a milestone in our journey of script optimization and efficiency improvement.



