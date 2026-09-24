---
sidebar_label: "Deploy a Flask App"
description: "How to deploy a Python Flask application with Gunicorn on OpenPanel: upload the code or deploy from Git, install requirements, connect a MySQL database and serve it on your domain with SSL."
---

# How to Deploy a Flask App (Python) with Gunicorn

This guide shows how to put a **Flask** application online with OpenPanel. The app runs in its own Python container, and OpenPanel's web server forwards your domain to it, with a free SSL certificate.

:::info
Python applications are available in **OpenPanel Enterprise**. The feature must be enabled for your hosting plan.
:::

---

## How It Works

- Your code lives in the domain's folder, e.g. `/var/www/html/example.com/`.
- OpenPanel starts an official `python` container with that folder as the working directory.
- On every start, the container can run `pip install -r requirements.txt` and then your **startup command**.
- The web server (Nginx, Apache, OpenResty or OpenLiteSpeed) proxies `https://example.com` to the **port** your app listens on.

---

## Step 1: Prepare the App

A minimal project:

```
example.com/
├── app.py
└── requirements.txt
```

`app.py`:

```python
from flask import Flask

app = Flask(__name__)

@app.route("/")
def index():
    return "Hello from Flask on OpenPanel!"
```

`requirements.txt`:

```
Flask==3.0.3
gunicorn==23.0.0
```

Flask's built-in server (`app.run()`) is for development only. In production, run the app with **Gunicorn**.

---

## Step 2: Add the Domain and Upload the Code

1. Add the domain in **OpenPanel → Domains** (see [Add a new domain](/docs/panel/domains/new/)) and point its DNS to the server.
2. Upload your files to the domain's folder with the [File Manager](/docs/panel/files/files/), [FTP](/docs/panel/files/FTP/), or skip this step and deploy from Git in the next step.

---

## Step 3: Create the Python Application

Go to **OpenPanel → Websites → Install App → Setup Python Application** and fill in:

| Field | Value |
|---|---|
| **Name** | e.g. `flaskapp` |
| **Domain** | `example.com` (optionally with a subfolder) |
| **Port** | `8000` |
| **Startup File** | `app.py` |
| **Custom Startup Command** | `gunicorn --bind 0.0.0.0:8000 --workers 2 app:app` |
| **Version** | a current Python version, e.g. 3.12 |
| **Run Install** | ✅ Yes - runs `pip install -r requirements.txt` on start |
| **Git repository** | optional - an `https://` URL of your repository |
| **CPU / Memory** | e.g. `1` core and `0.5` GB |

![Install Python Application form with the application details, domain, startup file and advanced options](/img/openpanel-screenshots/applications/python_install-form.png#gh-light-mode-only)
![Install Python Application form with the application details, domain, startup file and advanced options](/img/openpanel-screenshots/applications/python_install-form_dark.png#gh-dark-mode-only)

Click **Start Installation**. When it's done, open `https://example.com`.

`app:app` means *module* `app` (the file `app.py`) and the Flask *object* named `app`. If your project uses an application factory, wrap it in single quotes: `gunicorn --bind 0.0.0.0:8000 'app:create_app()'`. Don't use double quotes in the startup command.

:::tip Deploy from Git
If you fill in a **Git repository** URL, the container fetches the latest commit of the default branch **every time it starts**. To deploy a new version, push to the repository and click **Restart** on the application's page. Only public `https://` repositories are supported (no SSH keys).
:::

---

## Step 4: Connect a Database

Create a database and user in **OpenPanel → MySQL → Database Wizard** (or PostgreSQL). Inside containers, databases are reached by **service name**, not `localhost`:

| Database | Host | Port |
|---|---|---|
| MySQL | `mysql` | `3306` |
| MariaDB | `mariadb` | `3306` |
| PostgreSQL | `postgres` | `5432` |

Example with SQLAlchemy:

```python
import os
from flask_sqlalchemy import SQLAlchemy

app.config["SQLALCHEMY_DATABASE_URI"] = os.environ.get(
    "DATABASE_URL", "mysql+pymysql://dbuser:dbpass@mysql:3306/dbname"
)
db = SQLAlchemy(app)
```

Add `Flask-SQLAlchemy` and `PyMySQL` (or `psycopg2-binary` for PostgreSQL) to `requirements.txt`. See [connecting to MySQL from applications](/docs/articles/databases/how-to-connect-to-mysql-from-php-applications-in-openpanel/).

---

## Step 5: Secrets and Environment Variables

Keep secrets out of your code. Put them in a `.env` file in the app folder and load it with `python-dotenv`:

```
# /var/www/html/example.com/.env
SECRET_KEY=change-me
DATABASE_URL=mysql+pymysql://dbuser:dbpass@mysql:3306/dbname
```

```python
from dotenv import load_dotenv
load_dotenv()
```

Don't commit the `.env` file to Git. Restart the app after changing it.

---

## Managing the App

Open **OpenPanel → Websites → Sites** and click **Manage** next to the app:

- **Restart** after changing code or settings (Gunicorn doesn't reload automatically).
- **Install Packages** to edit `requirements.txt` and run pip.
- **Logs** to see Gunicorn output and Python tracebacks.
- **Overview** to change the startup command, Python version and CPU/memory limits.

![Python application page with its status, runtime version, CPU and memory limits, and Stop and Restart actions](/img/openpanel-screenshots/applications/python-site.png#gh-light-mode-only)
![Python application page with its status, runtime version, CPU and memory limits, and Stop and Restart actions](/img/openpanel-screenshots/applications/python-site_dark.png#gh-dark-mode-only)

More details: [Python applications](/docs/panel/applications/python/).

---

## Troubleshooting

| Problem | Fix |
|---|---|
| **502 Bad Gateway** | The app isn't listening on the configured port, or crashed on start. Check **Logs**, and make sure Gunicorn binds to `0.0.0.0:8000` - not `127.0.0.1`. See [502 errors](/docs/articles/domains/bad-gateway-502-error-troubleshooting/). |
| `ModuleNotFoundError` | The package is missing from `requirements.txt`, or **Run Install** is off. |
| Database connection refused | Use `mysql` / `mariadb` / `postgres` as the host, never `localhost`. |
| Container restarts in a loop / killed | The memory limit is too low. Increase it in **Overview**. |

---

## Related

- [Deploy a Django app](/docs/articles/websites/deploy-django-app/)
- [Deploy a Laravel app](/docs/articles/websites/deploy-laravel-app/)
- [Host a PHP website](/docs/articles/websites/hosting-a-php-website-with-openpanel/)
