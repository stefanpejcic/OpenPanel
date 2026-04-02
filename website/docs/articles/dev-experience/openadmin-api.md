# OpenAdmin API Documentation

A comprehensive REST API for managing **OpenPanel** hosting accounts, plans, domains, services, and server resources. All endpoints require **Bearer JWT authentication**.

**Base URL:** `http://PANEL:2087`

**Authentication:**
All requests require a bearer token:

```
Authorization: Bearer YOUR_JWT_TOKEN
```

---

## Table of Contents

1. [Users](#users)
2. [Plans](#plans)
3. [Domains](#domains)
4. [Usage](#usage)
5. [System](#system)
6. [Services](#services)
7. [Notifications](#notifications)

---

## Users

### List All Accounts

`GET /api/users`
**Response:** `200 OK`

```json
[
  {
    "username": "stefan",
    "email": "stefan@pejcic.rs",
    "plan_name": "default_plan_nginx"
  }
]
```

### Create Account

`POST /api/users`
**Request Body:**

```json
{
  "email": "stefan@pejcic.rs",
  "username": "stefan",
  "password": "s32dsasaq",
  "plan_name": "default_plan_nginx"
}
```

**Response:**

```json
{
  "response": {
    "message": "Successfully added user stefan password: s32dsasaq"
  },
  "success": true
}
```

### Get Single Account

`GET /api/users/{username}`

### Suspend / Unsuspend / Change Password

`PATCH /api/users/{username}`
**Request Body Example:**

```json
{
  "action": "suspend",
  "password": "NEW_PASSWORD_HERE"
}
```

### Change Plan

`PUT /api/users/{username}`
**Request Body:**

```json
{
  "plan_name": "default_plan_nginx"
}
```

### Delete Account

`DELETE /api/users/{username}`

### Autologin to User Account

`CONNECT /api/users/{username}`

### List Users with Dedicated IPs

`GET /api/ips`

---

## Plans

### List All Plans

`GET /api/plans`
**Response:**

```json
{
  "plans": [
    {
      "id": 1,
      "name": "ubuntu_nginx_mysql",
      "description": "Unlimited disk space and Nginx",
      "bandwidth": 100,
      "cpu": "1",
      "ram": "1g",
      "disk_limit": "10 GB",
      "storage_file": "0 GB",
      "inodes_limit": 1000000,
      "db_limit": 0,
      "domains_limit": 0,
      "websites_limit": 10,
      "docker_image": "openpanel/nginx"
    }
  ]
}
```

### Get Single Plan

`GET /api/plans/{plan_id}`

---

## Domains

### List All Domains

`GET /api/domains`

### Add Domain

`POST /api/domains/new`
**Request Body:**

```json
{
  "username": "current_user",
  "domain": "example.com",
  "docroot": "/var/www/html/example.com"
}
```

### Suspend Domain

`POST /api/domains/suspend/{domain}`

### Unsuspend Domain

`POST /api/domains/unsuspend/{domain}`

### Delete Domain

`POST /api/domains/delete/{domain}`

---

## Usage

### API Usage Info

`GET /api/usage`

### Current Usage Stats

`GET /api/usage/stats`
**Response:**

```json
{
  "usage_stats": "{\"timestamp\": \"2024-09-03\", \"users\": 1, \"domains\": 2, \"websites\": 0}"
}
```

### CPU Usage

`GET /api/usage/cpu`
**Response Example:**

```json
{
  "core_0": 0,
  "core_1": 0
}
```

### Memory Usage

`GET /api/usage/memory`

### Disk Usage Per User

`GET /api/usage/disk`

### Server Disk Usage

`GET /api/usage/server`
**Response Example:**

```json
[
  {
    "device": "/dev/vda1",
    "mountpoint": "/",
    "fstype": "ext4",
    "total": 123690532864,
    "used": 63366230016,
    "free": 60307525632,
    "percent": 51.2
  }
]
```

---

## System

### Get System Information

`GET /api/system`
**Response Example:**

```json
{
  "hostname": "stefi",
  "os": "Ubuntu 24.04 LTS",
  "time": "2024-09-04 15:09:16",
  "kernel": "6.8.0-36-generic",
  "cpu": "DO-Premium-Intel(x86_64)",
  "openpanel_version": "0.2.7",
  "running_processes": 178,
  "package_updates": 98,
  "uptime": "18905"
}
```

---

## Services

### List Monitored Services

`GET /api/services`

### Edit Services List

`PUT /api/services`
**Request Body Example:**

```json
{
  "service1": { "name": "Service One", "status": "active" },
  "service2": { "name": "Service Two", "status": "inactive" }
}
```

### Check Status of Services

`GET /api/services/status`

### Start Service

`POST /api/service/start/{service_name}`

### Restart Service

`POST /api/service/restart/{service_name}`

### Stop Service

`POST /api/service/stop/{service_name}`

---

## Notifications

### List Notifications

`GET /api/notifications`

**Response Example:**

```json
[
  { "title": "Update available", "message": "New package updates available." }
]
```

---
