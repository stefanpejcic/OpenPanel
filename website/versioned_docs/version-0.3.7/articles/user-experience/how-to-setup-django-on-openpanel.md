# How to Setup Django on OpenPanel

OpenPanel features a custom AutoInstaller that provides one-click installations for a variety of popular CMS and frameworks. Unlike other autoinstallers that primarily support PHP-based applications, OpenPanel's AutoInstaller also supports Python, Ruby on Rails, and Node.js applications.

Here’s a step-by-step guide to installing a Python-based application like Django on OpenPanel:

### Step 1: Create a Domain

Navigate to **OpenPanel > Domains** to add the domain name you will use for your app.
Ensure the domain name is pointed to the server's IP address, which you can find on the Dashboard page. Verify that you can access the domain.

### Step 2: Open AutoInstaller

In the OpenPanel menu, click on **Auto Installer** and select **Install** under the **NodeJS & Python** section.

[![2024-08-02-12-05.png](https://i.postimg.cc/qq6JhywP/2024-08-02-12-05.png)](https://postimg.cc/N2YqZyWD)


### Step 3. Fill the form

In the form, choose your domain name under 'Application URL'.

![2024-08-02-12-07-1.png](https://i.postimg.cc/fLLTs69T/2024-08-02-12-07-1.png)

In the 'Application Startup File' field, enter domainname/manage.py (replace domainname with your actual domain name).

![2024-08-02-12-07-2.png](https://i.postimg.cc/SKVNwyjv/2024-08-02-12-07-2.png)

In the 'Optional flags' section, set the Django runserver command as follows: `runserver 0.0.0.0:3000` (replace 3000 with your desired port number).

![2024-08-02-12-07-3.png](https://i.postimg.cc/FKhHWfFS/2024-08-02-12-07-3.png)

For the 'Type' field, select **Python**:

![2024-08-02-12-08.png](https://i.postimg.cc/5NFy6dkX/2024-08-02-12-08.png)

and in the 'Port' field, enter the same port number you used in the runserver flag (e.g., 3000).

![2024-08-02-12-08-1.png](https://i.postimg.cc/52BtS5J1/2024-08-02-12-08-1.png)


Click **Create** and wait a few minutes for the process to complete.

![2024-08-02-12-07.png](https://i.postimg.cc/850kztRd/2024-08-02-12-07.png)


Note: The initial setup might be slow as it installs PM2, Python, Django admin, and other dependencies. Subsequent application setups will be faster.

### Step 4. Test in browser

Once the installation is complete, open your domain in a web browser to test the setup.

![2024-08-02-12-08-3.png](https://i.postimg.cc/CM2dhLJy/2024-08-02-12-08-3.png)
