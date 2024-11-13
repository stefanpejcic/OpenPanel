# Setting Up a Spring Boot Java Application on OpenPanel

Running Java applications on OpenPanel is simple and efficient. Follow the steps below to get your [Spring Boot](https://spring.io/projects/spring-boot) application up and running.

## Step 1: Upload the .jar File

Once your Spring Boot application is packaged into a `.jar` file, upload it to your OpenPanel account. You can do this using one of the following methods:
- [**FileManager**](/docs/panel/files/#upload-files) within OpenPanel
- [**FTP**](/docs/panel/files/FTP/) for remote file transfers
- [**WebTerminal**](/docs/panel/advanced/terminal/) for terminal-based file uploads
- [**Remote SSH**](/docs/panel/advanced/ssh/) for secure shell file uploads

## Step 2: Run the Application

1. Log in to **OpenPanel** and navigate to **AutoInstaller**.
2. In the **Type** dropdown, select **Java**.
3. In the **Startup Script** section, add the path to your `.jar` file.
4. Specify the **domain** and **port** on which you want the app to run locally.
5. Click **Create** and wait for the process to complete.

OpenPanel will automatically install the required tools, including **NPM**, **PM2**, **Java**, and **JDK**. It will then start your application, configure a reverse proxy for the domain to the specified port, and ensure that your app starts automatically when the system boots up.

By following these steps, your Spring Boot application will be fully set up and running on OpenPanel.
