---
sidebar_position: 21
---

# Node.js Applications

Containerized [Node.js](https://nodejs.org) applications can be created and managed in **OpenPanel Enterprise Edition**.

---

## Create an Application

To create a new Node.js application, navigate to **OpenPanel > AutoInstaller** and click **Setup Node.js Application**.

![screenshot](/img/docs-content/HmZh5ZMJ-new-tab.png)

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

![screenshot](/img/docs-content/x0PBW9qB-new-app.png)

After completing the form, click **Start Installation**.
The installation process will be displayed below the form. Once complete, you’ll be redirected to the management page where you can view all your applications.

### Example App

An example Node.js (Express) application that is running on http://nodejs.openpanel.org/

Example settings:
![example](/img/docs-content/cdC3Jxdp-example-nodejs-settings.png)

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

![screenshot](/img/docs-content/vYbbVP6T-manage-apps.png)

Click **Manage** next to the application name to open its management page.

![screenshot](/img/docs-content/bzMFXdpg-single-app.png)

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
