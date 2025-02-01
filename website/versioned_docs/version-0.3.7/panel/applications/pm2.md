---
sidebar_position: 2
---

# NodeJS and Python

The PM2 page allows you to run Python or NodeJS applications from within the OpenPanel interface.

![pm2_noapps.png](/img/panel/v1/applications/pm2_noapps.png)

## Create an Application

To create a new application click on the 'New Application' button.

In the form set:

- Application URL - domain on which the application will be visible
- Application startup file - this is the file that PM2 will start
- Type - NodeJS or Python Application
- Port - port on which the application will run and Nginx will proxy the _Application URL_ to
- Watch - wheather to auto restart the application if fails
- Enable Logs - wheather to collect logs for the application or not

And click on the 'Create' button.

![pm2_app_new.png](/img/panel/v1/applications/pm2_app_new.png)

### Python Applications

For python applications you need to set the Type to *Python* so that PM2 uses the [--interpreter](https://pm2.keymetrics.io/docs/tutorials/using-transpilers-with-pm2) flag.

#### Example python application

```python
from flask import Flask

app = Flask(__name__)

@app.route('/')
def hello():
    return "Hello world!"

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)

```


### NodeJS Application

#### NPM Install 

1.  Navigate to your project directory:
Open [Web Terminal](/docs/panel/advanced/terminal) or connect via [SSH](/docs/panel/advanced/ssh) to the server and navigate to the directory where your project is located. You can use the cd command to change directories.

```bash
cd path/to/your/project
```

2. Check for package.json file:
Make sure that your project has a package.json file. This file contains metadata about your project and the list of dependencies. If you don't have one, you can create it by running:

```bash
npm init
```
Follow the prompts to set up your package.json file.

3. Install dependencies:
Run the following command to install the dependencies listed in your package.json file:

```bash
npm install
```

#### Example NodeJS application

```js
const http = require('http');

const server = http.createServer((req, res) => {
  res.writeHead(200, { 'Content-Type': 'text/plain' });
  res.end('Hello, World!\n');
});

const port = process.env.PORT || 3000;

server.listen(port, () => {
  console.log(`Server is running on port ${port}`);
});

```




## View Applications

The PM2 page displays a list of all current NodeJS and Python applications.

- Application name - domain on which the application is visible
- Uptime - time passed since the application was started or created
- Restarts - number of times the PM2 restarted the script (only if watch was enabled for the application)
- CPU - current CPU % usage for the application process
- Memory - current RAM usage in bytes by the applciaiton process
- Watching - wheather the application was created with the watch option or not
- Actions - Start/Stop, Restart and Delete the application
- Type - NodeJS or Python application
- Port - port on which the application is running and Nginx is proxying requests to
- Startup file - main entrypoint for the application

![pm2_app.png](/img/panel/v1/applications/pm2_app.png)

## Manage Application

Currently only the following actions can be performed for existing applications in the PM2 page:

- [Start / Stop](/docs/panel/applications/pm2#stop)
- [Restart](/docs/panel/applications/pm2#restart)
- [Delete](/docs/panel/applications/pm2#delete)
- [View Logs](/docs/panel/applications/pm2#logs)

### Restart

To restart an application *(force stop and start it immediately) click on the 'Restart' button:

![pm2_app_restart.png](/img/panel/v1/applications/pm2_app_restart.png)

This will immediately stop the pm2 process for the applicaiton and start again the startup script.

### Stop

To stop the application simply click on the 'Stop' button. This will immediately stop the process but leave the Nginx proxy to the port.

![pm2_app_stop.png](/img/panel/v1/applications/pm2_app_stop.png)

### Delete

To completely remove an appliaiton click on the 'Delete' button. This will immediately stop the process and delete the Nginx proxy to the port.

![pm2_app_delete.png](/img/panel/v1/applications/pm2_app_delete.png)

### Logs

To view error logs for an applicaiton, click on the 'Logs' button:
![pm2_app_logs.png](/img/panel/v1/applications/pm2logs.png)
