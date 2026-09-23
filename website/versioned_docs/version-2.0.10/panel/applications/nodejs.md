---
sidebar_position: 21
---

# Node.js

Containerized [Node.js](https://nodejs.org) applications can be created and managed in **OpenPanel Enterprise Edition**.

---

## Create an Application

To create a new Node.js application, navigate to **OpenPanel > AutoInstaller** and click **Setup Node.js Application**.

![Auto Installer page with cards for WordPress, Joomla, Drupal, Website Builder, PrestaShop, OpenCart and other applications](/img/openpanel-screenshots/applications/autoinstaller-page.png#gh-light-mode-only)
![Auto Installer page with cards for WordPress, Joomla, Drupal, Website Builder, PrestaShop, OpenCart and other applications](/img/openpanel-screenshots/applications/autoinstaller-page_dark.png#gh-dark-mode-only)

On the next page, you can configure the following settings:

* **Name** – The name of the application and container as displayed in OpenPanel.
* **Port** – Set a custom port (e.g., 3000) if your app uses one. Otherwise, port 80 is used by default.
* **Domain Name / Subfolder** – The domain (and optional subfolder) where the application will be publicly accessible.
* **Startup File** – The file executed at startup with the `node` command. Defaults to `index.js`.
* **Custom Startup Command** – Use a custom startup command instead of the default `node`.
* **Type** – Fixed to Node.js.
* **Version** – Select any available Node.js version from Docker Hub.
* **Run Install** – Run `npm install` using `package.json` before starting the application.
* **CPU Cores** – Number of CPU cores allocated to the application.
* **Memory** – Amount of memory (in GB) allocated to the application.

![Install Node.js Application form with the application details, domain, startup file and advanced options](/img/openpanel-screenshots/applications/nodejs_install-form.png#gh-light-mode-only)
![Install Node.js Application form with the application details, domain, startup file and advanced options](/img/openpanel-screenshots/applications/nodejs_install-form_dark.png#gh-dark-mode-only)

After completing the form, click **Start Installation**.
The installation process will be displayed below the form. Once complete, you’ll be redirected to the management page where you can view all your applications.

### Example App

An example Node.js (Express) application that is running on http://nodejs.openpanel.org/

Example settings:

Example `app.js` file:

```js
const express = require('express');
const app = express();
const port = 3000;

app.get('/', (req, res) => {
  res.send('Hello World from Node.js on port 3000!');
});

app.listen(port, () => {
  console.log(`Server is running at http://localhost:${port}`);
});
```

Example `package.json` file:

```json
{
  "name": "helloworld-node",
  "version": "1.0.0",
  "description": "Simple Node.js Hello World app using Express",
  "main": "app.js",
  "scripts": {
    "start": "node app.js"
  },
  "author": "Stefan Pejcic",
  "license": "MIT",
  "dependencies": {
    "express": "^4.18.2"
  }
}
```

---

## Manage Applications

Once your application is created, you can manage it from **OpenPanel > Site Manager**.

![Site Manager listing websites grouped by type with their version, creation date and PageSpeed scores](/img/openpanel-screenshots/applications/sites-list.png#gh-light-mode-only)
![Site Manager listing websites grouped by type with their version, creation date and PageSpeed scores](/img/openpanel-screenshots/applications/sites-list_dark.png#gh-dark-mode-only)

Click **Manage** next to the application name to open its management page.

![Node.js application page with its status, runtime version, CPU and memory limits, and Stop and Restart actions](/img/openpanel-screenshots/applications/nodejs-site.png#gh-light-mode-only)
![Node.js application page with its status, runtime version, CPU and memory limits, and Stop and Restart actions](/img/openpanel-screenshots/applications/nodejs-site_dark.png#gh-dark-mode-only)

On this page, you can view important details such as:

* **Screenshot** – Preview of the application’s domain.
* **Status** – Current container status.
* **Version** – Node.js version in use.
* **CPU Limit** – Configured CPU allocation.
* **Memory Limit** – Configured memory allocation.
* **Speed** – Google PageSpeed Insights data for the website.
* **Files** – Current folder path and size.
* **Firewall** – WAF (Web Application Firewall) status for the domain (if enabled).

You also have several management options:

* **Actions** – Start, stop, or restart the container.
* **Overview** – Modify startup file or command, working directory, package installation settings (NPM), version, and resource limits (CPU, Memory, PIDs).
* **Install Packages** – View and manage `package.json`, and run NPM/PNPM installations.
* **Logs** – View container logs for troubleshooting.
* **Remove** – Delete the application.
