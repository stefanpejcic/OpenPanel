---
sidebar_position: 2
---

# NodeJS and Python

Contianerized [Node.js](https://nodejs.org) and [Python](https://python.org/) applications can be created and managed on OpenPanel with the help of our integrated [pm2](https://pm2.io/) manager.

---

## Create a new Application

To create a new Node.js or Python application using pm2, first access the **Auto Installer** page and under Node.js or Python click on the "Setup Node.js/Python Application" button.

The next step is choosing an internal name for your new application, this name is only visible on the backend and is used to identify the container.

After naming your application you'll need to choose a domain and optionally a subdirectory where your new application will be installed.

On this page you can also choose which version of Node.js or Python to install.

If your application hasn't been built yet, you can check a box that tells the installer to run NPM or PIP install before starting your application.

With this option enabled, the installer will first run npm install using the package.json file or pip install using the requirements.txt file. 

If your application is already built, you can skip this step and leave the box unchecked.

When you're done configuring click on the "Start Installation" button to run the AutoInstaller, the installation log will be displayed below.

---

## Manage your Application

On the **Site Manager** page of your OpenPanel account you can find basic details about your applications.

Here you can access a management panel for each app by clicking on the "Manage" button besides your application.

Some of the basic operations available are Stopping, Restarting and Removing the application, there is also a Live preview feature along with a recent screenshot of your app and a PageSpeed statistics graph.

Within the overview tab of the manager you can change your application's startup settings such as it's Startup file, Work directory and whether NPM/PIP install will run on startup. 

Here you can also change the version of Node.js or Python that is used and the application's assigned resources such as CPU core and Memory limits.

NPM/PIP install can also be ran manually through the manager and there's a shortcut to open the appropriate package.json/requirements.txt file inside file editor.

Additionally your application container's logs are available to you on the "Logs" tab of the manager.




