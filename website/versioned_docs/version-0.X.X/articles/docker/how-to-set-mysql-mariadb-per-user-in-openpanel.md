# Configuring MySQL or MariaDB per user in OpenPanel

OpenPanel utilizes Docker images as the foundation for hosting packages. Think of Docker images as ISO templates for VPS: they come with pre-installed services and cannot be modified after creation.

## Available Docker Images in OpenPanel

### OpenPanel Community Edition:
- **Nginx with MySQL**
- **Apache with MySQL**

### OpenPanel Enterprise Edition (additional images):
- **Nginx with MariaDB**
- **Apache with MariaDB**

## How to Assign MariaDB to a User

To use MariaDB for a specific user instead of MySQL, follow these steps on a server running OpenPanel Enterprise Edition:

1. [Create a new hosting plan](/docs/admin/plans/hosting_plans/#create-a-plan).
2. Select one of the MariaDB-based Docker images (Nginx with MariaDB or Apache with MariaDB) for the plan.

Any user created under this plan will automatically use a Docker image with MariaDB instead of MySQL.
