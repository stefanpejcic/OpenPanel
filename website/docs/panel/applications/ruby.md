---
sidebar_position: 23
---

# Ruby Applications

Containerized [Ruby](https://www.ruby-lang.org/) applications can be created and managed in **OpenPanel Enterprise Edition**.

---

## Create an Application

To create a new Ruby application, navigate to **OpenPanel > AutoInstaller** and click **Setup Ruby Application**.

On the next page, you can configure the following settings:

* **Name** – The name of the application and container as displayed in OpenPanel.
* **Port** – Set a custom port if your app uses one. Otherwise, port 80 is used by default.
* **Domain Name / Subfolder** – The domain (and optional subfolder) where the application will be publicly accessible.
* **Startup File** – The file executed at startup with the `ruby` command. Defaults to `app.rb`.
* **Custom Startup Command** – Use a custom startup command instead of the default `ruby`.
* **Type** – Fixed to Ruby.
* **Version** – Select any available Ruby version from Docker Hub's official `ruby` image tags.
* **Run Install** – Run `bundle install` using the `Gemfile` before starting the application. Skip this if your application is already built or has no gem dependencies.
* **CPU Cores** – Number of CPU cores allocated to the application.
* **Memory** – Amount of memory (in GB) allocated to the application.

After completing the form, click **Start Installation**.
The installation process will be displayed below the form. Once complete, you’ll be redirected to the management page where you can view all your applications.

### Example App

A minimal [Sinatra](https://sinatrarb.com/) application.

Example `app.rb` file:

```ruby
require 'sinatra'

set :port, 3000
set :bind, '0.0.0.0'

get '/' do
  'Hello World from Ruby on port 3000!'
end
```

Example `Gemfile`:

```ruby
source 'https://rubygems.org'

gem 'sinatra'
```

Enable **Run Install** so `bundle install` runs before the app starts.

---

## Manage Applications

Once your application is created, you can manage it from **OpenPanel > Site Manager**.

Click **Manage** next to the application name to open its management page.

On this page, you can view important details such as:

* **Screenshot** – Preview of the application’s domain.
* **Status** – Current container status.
* **Version** – Ruby version in use.
* **CPU Limit** – Configured CPU allocation.
* **Memory Limit** – Configured memory allocation.
* **Speed** – Google PageSpeed Insights data for the website.
* **Files** – Current folder path and size.
* **Firewall** – WAF (Web Application Firewall) status for the domain (if enabled).

You also have several management options:

* **Actions** – Start, stop, or restart the container.
* **Overview** – Modify startup file or command, working directory, package installation settings (Bundler), version, and resource limits (CPU, Memory, PIDs).
* **Install Packages** – View and manage the `Gemfile`, and run `bundle install`.
* **Logs** – View container logs for troubleshooting.
* **Remove** – Delete the application.
