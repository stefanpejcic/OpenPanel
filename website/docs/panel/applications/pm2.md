---
sidebar_position: 2
---

# Node.js and Python

Containerized [Node.js](https://nodejs.org) and [Python](https://python.org/) applications can be created and managed in **OpenPanel Enterprise Edition**.

---

## Create an Application

To create a new Python or Node.js application, navigate to **OpenPanel > AutoInstaller** and select **Python** or **NodeJS**.

![screenshot](https://i.postimg.cc/HmZh5ZMJ/new-tab.png)

On the next page, you can configure the following settings:

* **Name** – The name of the application and container as displayed in OpenPanel.
* **Port** – Set a custom port (e.g., 3000 or 5000) if your app uses one. Otherwise, port 80 is used by default.
* **Domain Name / Subfolder** – The domain (and optional subfolder) where the application will be publicly accessible.
* **Startup File** – The file executed at startup with the `node` or `py` command.
* **Custom Startup Command** – Use a custom startup command instead of the default `node` or `py`.
* **Type** – Choose between Node.js or Python.
* **Version** – Select any available version from Docker Hub.
* **Run Install** – Run `npm install` or `pip install` before starting the application.
* **CPU Cores** – Number of CPU cores allocated to the application.
* **Memory** – Amount of memory (in GB) allocated to the application.

![screenshot](https://i.postimg.cc/x0PBW9qB/new-app.png)

After completing the form, click **Start Installation**.
The installation process will be displayed below the form. Once complete, you’ll be redirected to the management page where you can view all your applications.

---

## Manage Applications

Once your application is created, you can manage it from **OpenPanel > Site Manager**.

![screenshot](https://i.postimg.cc/vYbbVP6T/manage-apps.png)

Click **Manage** next to the application name to open its management page.

![screenshot](https://i.postimg.cc/bzMFXdpg/single-app.png)

On this page, you can view important details such as:

* **Screenshot** – Preview of the application’s domain.
* **Status** – Current container status.
* **Version** – Node.js or Python version in use.
* **CPU Limit** – Configured CPU allocation.
* **Memory Limit** – Configured memory allocation.
* **Speed** – Google PageSpeed Insights data for the website.
* **Files** – Current folder path and size.
* **Firewall** – WAF (Web Application Firewall) status for the domain (if enabled).

You also have several management options:

* **Actions** – Start, stop, or restart the container.
* **Overview** – Modify startup file or command, working directory, package installation settings (NPM/PIP), version, and resource limits (CPU, Memory, PIDs).
* **Install Packages** – View and manage `package.json` or `requirements.txt`, and run NPM/PNPM or PIP installations.
* **Logs** – View container logs for troubleshooting.
* **Remove** – Delete the application.
