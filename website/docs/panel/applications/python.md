---
sidebar_position: 22
---

# Python

Containerized [Python](https://python.org/) applications can be created and managed in **OpenPanel Enterprise Edition**.

Step-by-step guides: [Deploy a Flask app](/docs/articles/websites/deploy-flask-app/) · [Deploy a Django app](/docs/articles/websites/deploy-django-app/)

---

## Create an Application

To create a new Python application, navigate to **OpenPanel > Websites > Install App** and click **Setup Python Application**.

![Auto Installer page with cards for WordPress, Joomla, Drupal, Website Builder, PrestaShop, OpenCart and other applications](/img/openpanel-screenshots/applications/autoinstaller-page.png#gh-light-mode-only)
![Auto Installer page with cards for WordPress, Joomla, Drupal, Website Builder, PrestaShop, OpenCart and other applications](/img/openpanel-screenshots/applications/autoinstaller-page_dark.png#gh-dark-mode-only)

On the next page, you can configure the following settings:

* **Name** – The name of the application and container as displayed in OpenPanel.
* **Port** – Set a custom port (e.g., 5000) if your app uses one. Otherwise, port 80 is used by default.
* **Domain Name / Subfolder** – The domain (and optional subfolder) where the application will be publicly accessible.
* **Startup File** – The file executed at startup with the `py` command. Defaults to `app.py`.
* **Custom Startup Command** – Use a custom startup command instead of the default `py`.
* **Type** – Fixed to Python.
* **Version** – Select any available Python version from Docker Hub.
* **Run Install** – Run `pip install -r requirements.txt` before starting the application.
* **CPU Cores** – Number of CPU cores allocated to the application.
* **Memory** – Amount of memory (in GB) allocated to the application.

![Install Python Application form with the application details, domain, startup file and advanced options](/img/openpanel-screenshots/applications/python_install-form.png#gh-light-mode-only)
![Install Python Application form with the application details, domain, startup file and advanced options](/img/openpanel-screenshots/applications/python_install-form_dark.png#gh-dark-mode-only)

After completing the form, click **Start Installation**.
The installation process will be displayed below the form. Once complete, you’ll be redirected to the management page where you can view all your applications.

### Example App

An example Python (Flask) application that is running on http://python.openpanel.org/

Example settings:

Example `app.py` file:

```py
from flask import Flask

app = Flask(__name__)

@app.route('/')
def hello():
    return "Hello World from Flask on port 5000!"

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
```

Example `requirements.txt` file:

```
Flask==2.3.3
```

---

## Manage Applications

Once your application is created, you can manage it from **OpenPanel > Websites > Sites**.

![Site Manager listing websites grouped by type with their version, creation date and PageSpeed scores](/img/openpanel-screenshots/applications/sites-list.png#gh-light-mode-only)
![Site Manager listing websites grouped by type with their version, creation date and PageSpeed scores](/img/openpanel-screenshots/applications/sites-list_dark.png#gh-dark-mode-only)

Click **Manage** next to the application name to open its management page.

![Python application page with its status, runtime version, CPU and memory limits, and Stop and Restart actions](/img/openpanel-screenshots/applications/python-site.png#gh-light-mode-only)
![Python application page with its status, runtime version, CPU and memory limits, and Stop and Restart actions](/img/openpanel-screenshots/applications/python-site_dark.png#gh-dark-mode-only)

On this page, you can view important details such as:

* **Screenshot** – Preview of the application’s domain.
* **Status** – Current container status.
* **Version** – Python version in use.
* **CPU Limit** – Configured CPU allocation.
* **Memory Limit** – Configured memory allocation.
* **Speed** – Google PageSpeed Insights data for the website.
* **Files** – Current folder path and size.
* **Firewall** – WAF (Web Application Firewall) status for the domain (if enabled).

You also have several management options:

* **Actions** – Start, stop, or restart the container.
* **Overview** – Modify startup file or command, working directory, package installation settings (PIP), version, and resource limits (CPU, Memory, PIDs).
* **Install Packages** – View and manage `requirements.txt`, and run PIP installations.
* **Logs** – View container logs for troubleshooting.
* **Remove** – Delete the application.
