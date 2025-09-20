---
title: Database Only Hosting with OpenPanel
description: How to setup Database-only Hosting plans with OpenPanel Enterprise edition
slug: database-hosting-with-openpanel
authors: radovan
tags: [OpenPanel, MySQL, MariaDB, Databases]
image: https://openpanel.com/img/blog/database-hosting-with-openpanel.png
hide_table_of_contents: true
---

Databases are at the heart of every modern application—whether you’re running WordPress, a custom PHP app, or Node.js microservices. With **OpenPanel**, you can offer secure, isolated, and scalable **Database-only hosting plans** that let users manage MySQL/MariaDB databases with ease.  

From one-click database creation to user management, phpMyAdmin access, and secure remote connections, OpenPanel makes database hosting simple and powerful.

<!--truncate-->

![database-manager]()

---

## Setting Up Database Hosting Plans

OpenPanel allows you to create **plans dedicated only to database usage**. These are perfect for developers, staging environments, or applications that only need database services without a full web stack.

### Step 1. Enable Database Modules

Go to **OpenAdmin > Settings > Modules** and enable:

- **MySQL** – Core database management  
- **phpMyAdmin** – Web-based database management (optional, advanced users)  
- **Remote MySQL** – Secure external access to databases (optional)  
- **Docker / Terminal** – Optional for advanced imports (>1GB SQL dumps)  

---

### Step 2: Create a Feature Set

Create a new feature set for database-only hosting:

1. Go to **OpenAdmin > Hosting Plans > Feature Manager**  
2. Add a new feature list: `db_only`  
3. Under *Manage feature set*, enable:  
   - MySQL  
   - phpMyAdmin (optional)  
   - Remote MySQL (optional)  

This ensures database users only get access to the tools they need.  

---

### Step 3: Create Hosting Plans

With the feature set ready, create new plans:

1. Go to **OpenAdmin > Hosting Plans > User Packages** and click **Create New**  
2. Define plan limits (number of DBs, users, disk quota)  
3. Select the `db_only` feature set  
4. Click **Create Plan**  

Example plans:  

- **Basic DB Hosting** – 1 database, 1 user, phpMyAdmin only  
- **Pro DB Hosting** – 5 databases, 5 users, remote access enabled  
- **Enterprise DB Hosting** – Unlimited databases, remote access + terminal imports  

---

### Step 4: Create User Accounts

Accounts can be created via **OpenAdmin**, the **terminal**, or third-party billing systems like WHMCS/FOSSBilling.  
Each user gets an isolated container running their MySQL/MariaDB service with secure credentials.

You can set different servers: MySQL, Percona or MariaDB and also different versions per user.
---

## Key Features of Database Hosting Plans

### Database Management Made Simple

On the **Databases page**, users can:

* Create and delete databases  
* View database sizes and assigned users  
* Assign or remove users from databases  
* Reset user passwords  
* Access the **Database Wizard** (auto-creates DB, user, and secure password)  

⚠️ Deleting a database or user is permanent, and OpenPanel provides warnings to prevent accidental data loss.

---

### Secure Application Connections

Because OpenPanel isolates services in containers, apps must connect using container hostnames:

- Hostname: `mysql` (for MySQL) or `mariadb` (for MariaDB)  
- Port: `3306`  
- Credentials: Created in the OpenPanel UI  

❌ Never use `localhost` or `127.0.0.1`  

✅ Works with PHP, Node.js, Python, WordPress, and more.  

**Example (PHP PDO):**

```php
$dsn = "mysql:host=mysql;dbname=my_database;charset=utf8mb4";
$pdo = new PDO($dsn, "my_user", "my_password");
```

**Example (Node.js mysql2):**

```js
const mysql = require('mysql2');
const conn = mysql.createConnection({
  host: 'mysql',
  user: 'my_user',
  password: 'my_password',
  database: 'my_database'
});
```

---

### Remote MySQL Access

By default, databases are accessible only inside the server.  
With **Remote MySQL Access**, users can securely connect from external servers.

- Enabled via **Databases > Remote MySQL**  
- Unique random port per instance (not 3306)  
- Can be tested via CLI, PHP, or WordPress wp-config.php  

⚠️ For security, strong passwords and encryption are required.  

---

### phpMyAdmin for Advanced Users

OpenPanel integrates **phpMyAdmin** for advanced database management:

* Run SQL queries  
* Import/export databases  
* Manage tables and indexes  
* View active queries  

⚠️ Direct table edits can break applications—intended for experienced users only.

---

### Importing Databases

Users can import `.sql` backups using:

1. **phpMyAdmin** (upload/import file)  
2. **Import Database** option

---

### Monitoring and Performance

Database users get access to:

* Active query monitoring (`SHOW PROCESSLIST`)  
* Database size overview  
* Ability to restart the MySQL service in their container  

---

## Why Choose OpenPanel for Database Hosting?

* **Isolation:** Each user gets their own MySQL/MariaDB container  
* **Security:** No shared localhost access, optional remote connections with custom ports  
* **Flexibility:** Works with PHP, Python, Node.js, WordPress, and more  
* **Developer-Friendly:** phpMyAdmin, CLI, Docker, and API integration  
* **Scalable Plans:** From 1 DB to unlimited, tailored for agencies and enterprises  

Ready to offer **Database-only hosting plans**? [Start Here](https://openpanel.com/enterprise/)
