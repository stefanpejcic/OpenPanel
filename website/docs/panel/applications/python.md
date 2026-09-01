---
sidebar_position: 22
---

# Python Applications

Containerized [Python](https://python.org/) applications can be created and managed in **OpenPanel Enterprise Edition**.

---

## Create an Application

To create a new Python application, navigate to **OpenPanel > AutoInstaller** and click **Setup Python Application**.

![screenshot](/img/docs-content/HmZh5ZMJ-new-tab.png)

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

![screenshot](/img/docs-content/x0PBW9qB-new-app.png)

After completing the form, click **Start Installation**.
The installation process will be displayed below the form. Once complete, you’ll be redirected to the management page where you can view all your applications.

### Example App

An example Python (Flask) application that is running on http://python.openpanel.org/

Example settings:
![example](/img/docs-content/D2Z3DNdW-example-python-settings.png)

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

Once your application is created, you can manage it from **OpenPanel > Site Manager**.

![screenshot](/img/docs-content/vYbbVP6T-manage-apps.png)

Click **Manage** next to the application name to open its management page.

![screenshot](/img/docs-content/bzMFXdpg-single-app.png)

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
