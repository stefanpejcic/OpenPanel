---
sidebar_position: 24
---

# Java Applications

Containerized [Java](https://www.java.com/) applications can be created and managed in **OpenPanel Enterprise Edition**.

---

## Create an Application

To create a new Java application, navigate to **OpenPanel > AutoInstaller** and click **Setup Java Application**.

On the next page, you can configure the following settings:

* **Name** – The name of the application and container as displayed in OpenPanel.
* **Port** – Set a custom port if your app uses one. Otherwise, port 80 is used by default.
* **Domain Name / Subfolder** – The domain (and optional subfolder) where the application will be publicly accessible.
* **Startup File** – The file executed at startup with the `java` command. Defaults to `Main.java`.
* **Custom Startup Command** – Use a custom startup command instead of the default `java` (for example `java -jar target/app.jar` for a Maven build).
* **Type** – Fixed to Java.
* **Version** – Select a JDK version from Docker Hub's official `eclipse-temurin` LTS tags.
* **Run Install** – Run `mvn install` using the project's `pom.xml` before starting the application. Skip this for a single-file app or one that's already built.
* **CPU Cores** – Number of CPU cores allocated to the application.
* **Memory** – Amount of memory (in GB) allocated to the application.

After completing the form, click **Start Installation**.
The installation process will be displayed below the form. Once complete, you’ll be redirected to the management page where you can view all your applications.

:::info
Since Java 11 ([JEP 330](https://openjdk.org/jeps/330)), `java SomeFile.java` runs a single source file directly — it's compiled in memory with no separate `javac` step and no build tool required. That's the default here, matching the "just run one file" model of the Node.js, Python, and Ruby application types for a quick app. Projects that do have a `pom.xml` can enable **Run Install** to run `mvn install` first, then use a custom startup command to run the resulting build.
:::

### Example App

A minimal HTTP server using only the JDK standard library — no build step or dependencies required.

Example `Main.java` file:

```java
import com.sun.net.httpserver.HttpServer;
import java.net.InetSocketAddress;

public class Main {
    public static void main(String[] args) throws Exception {
        HttpServer server = HttpServer.create(new InetSocketAddress(3000), 0);
        server.createContext("/", exchange -> {
            byte[] response = "Hello World from Java on port 3000!".getBytes();
            exchange.sendResponseHeaders(200, response.length);
            exchange.getResponseBody().write(response);
            exchange.getResponseBody().close();
        });
        server.start();
    }
}
```

No **Run Install** step is needed for this example — leave it disabled and set the startup file to `Main.java`.

---

## Manage Applications

Once your application is created, you can manage it from **OpenPanel > Site Manager**.

Click **Manage** next to the application name to open its management page.

On this page, you can view important details such as:

* **Screenshot** – Preview of the application’s domain.
* **Status** – Current container status.
* **Version** – JDK version in use.
* **CPU Limit** – Configured CPU allocation.
* **Memory Limit** – Configured memory allocation.
* **Speed** – Google PageSpeed Insights data for the website.
* **Files** – Current folder path and size.
* **Firewall** – WAF (Web Application Firewall) status for the domain (if enabled).

You also have several management options:

* **Actions** – Start, stop, or restart the container.
* **Overview** – Modify startup file or command, working directory, package installation settings (Maven), version, and resource limits (CPU, Memory, PIDs).
* **Install Packages** – View and manage the `pom.xml`, and run `mvn install`.
* **Logs** – View container logs for troubleshooting.
* **Remove** – Delete the application.
