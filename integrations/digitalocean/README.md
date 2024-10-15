# DigitalOcean Marketplace Partner Tools

[![Apache license](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Pull Requests Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat)](http://makeapullrequest.com)

This repository contains resources for [DigitalOcean Marketplace](https://marketplace.digitalocean.com/) partners, like documentation on image requirements and creation, tools for image cleanup and validation, and templates for build automation.

## Getting Started creating your Kubernetes based 1-Click App

Please visit [our Kubernetes Marketplace repo](https://github.com/digitalocean/marketplace-kubernetes/blob/master/CONTRIBUTING.md) for instructions on how to create and submit your Kubernetes based 1-Click App.

Kubernetes 1-Click Apps can be submitted via our [Marketplace Kubernetes Github repo](https://github.com/digitalocean/marketplace-kubernetes).

## Getting Started creating your Droplet based 1-Click App

The overall process for creating an image that you can submit as a Droplet based 1-Click App is as follows:

1. Create and configure a build Droplet manually first to make sure your configuration works. You can create a build Droplet with any method, like the [control panel](https://cloud.digitalocean.com/), the [API](https://developers.digitalocean.com/), or command-line tools like [`doctl`](https://github.com/digitalocean/doctl). **We strongly encourage you to use a $6 Droplet as your build Droplet.** Using a $6 Droplet as your build Droplet will ensure that all Droplet types will be available for usage with your 1-Click App.

2. Clean up and validate the build Droplet with the provided scripts, `cleanup.sh` and `img_check.sh`. The scripts will check for and fix potential security concerns and verify that the image will be compatible with Marketplace.

3. Use Packer to create a fresh [snapshot](https://www.digitalocean.com/docs/images/snapshots/) of the image that you want to create. While there are several ways to create an image, we recommend using Packer as the most simple and consistent option.

4. Submit your final image to the Marketplace team for review. This can be made through [our Vendor Portal](https://cloud.digitalocean.com/vendorportal). If you've signed expressed interest in joining the Marketplace through [the form on this page](https://marketplace.digitalocean.com/vendors) but you've not received a login for the Vendor Portal, please reach out to one-clicks-team@digitalocean.com and we'll help you out.

<a name="dbaas-integration"></a>

5. **(Optional)** Integrate a DigitalOcean Managed Database into your 1-Click App</h4>

    As a Vendor, you can offer a DigitalOcean Managed Database (DBaaS) to any DigitalOcean customer at the time they spin up your Droplet based 1-Click App. You can customize your app image to integrate with the managed database directly, or let your customers complete the configuration themselves after the 1-Click App boots up. This benefits your customers in terms of database scalability and ease of management, and reduces the burden of database support for you as a vendor.

    To enable this option, use the checkboxes that are shown in the Vendor Portal related to enablement of managed databases. If you’ve checked at least 1 box, when a user attempts to create your 1-Click App, they’ll receive a prompt like this one:

    ![dbaas_1_click_offer](/images/dbaas_1_click_offer.png)

    When a user selects the Managed Database option, DigitalOcean handles the creation of the database cluster as well as the user's Droplet. The Droplet will have a `DATABASE_URL` environment variable configured including a database connection string, such as:

    `postgresql://doadmin:<password>@dbaas-db-11111-do-user-1111111-1.b.db.ondigitalocean.com:25060/defaultdb?sslmode=require`

    The user’s managed database configuration and credentials will be stored in `/root/.digitalocean_dbaas_credentials` in the following format.

    ```
    db_protocol=
    db_username=
    db_password=
    db_host=
    db_port=
    db_database=
    ```

    To disable this feature for any new users of your 1-Click App, simply visit the Vendor Portal to edit your 1-Click App, removing the checkboxes next to all database engines. Once you save the edits, customers will no longer be given the option to add a managed database. Note that existing customers who have already deployed a managed database in conjunction with your 1-Click App will not be affected by that removal, and their managed databases will continue to operate.

## Build Automation with Packer

[Packer](https://www.packer.io/intro) is a tool for creating images from a single source configuration. Using this Packer template reduces the entire process of creating, configuring, validating, and snapshotting a build Droplet to a single command:

```
packer build marketplace-image.json
```

By doing this, there is a reduced likelihood of having to submit an image multiple times as a result of falling in any of the next steps:

- Installing OS updates
- Deleting bash history.
- Removing log files and SSH keys from the root user
- Enabling the firewall (i.e. ufw if you use Ubuntu)

This repository is itself a Packer template for a LAMP stack. You can modify this template to use as a starting point for your image. Note that not all of the scripts/files in this repository are strictly necessary, as this aim at covering a broad case of application.

## Usage

To run the LAMP stack in this template, you'll need to [install Packer](https://www.packer.io/intro/getting-started/install.html) and [create a DigitalOcean personal access token](https://docs.digitalocean.com/reference/api/create-personal-access-token/) and set it to the `DIGITALOCEAN_TOKEN` environment variable. Running `packer build marketplace-image.json` without any other modifications will create a build Droplet configured with LAMP, clean and verify it, then power it down and snapshot it.

To start adapting this template for your own image, you can customize some variables in `marketplace-image.json`:

* `apt_packages` lists the APT packages to install on the build Droplet.
* `image_name` defines the name of the resulting snapshot, which by default is `marketplace-snapshot-` with a UNIX timestamp appended.

You can also modify these variables at runtime by using [the `-var` flag](https://www.packer.io/docs/templates/legacy_json_templates/user-variables).

Please see the [RStudio Server 1-Click Scripts](https://github.com/pachadotdev/rstudio-server-droplet) to see an example of Packer usage.

A successful run would look like this output:

```
pacha@pop-os:~/github/marketplace-partners$ packer build marketplace-image.json
digitalocean output will be in this color.

==> digitalocean: Creating temporary ssh key for droplet...
==> digitalocean: Creating droplet...
==> digitalocean: Waiting for droplet to become active...
==> digitalocean: Using ssh communicator to connect: 165.227.211.66
==> digitalocean: Waiting for SSH to become available...
==> digitalocean: Connected to SSH!
==> digitalocean: Provisioning with shell script: /tmp/packer-shell581341144
    digitalocean: .............................................................
    digitalocean: status: done
==> digitalocean: Uploading files/etc/ => /etc/
==> digitalocean: Uploading files/var/ => /var/
==> digitalocean: Provisioning with shell script: /tmp/packer-shell079619818
    digitalocean:
    digitalocean: WARNING: apt does not have a stable CLI interface. Use with caution in scripts.

    ...

    digitalocean: The following NEW packages will be installed:
    digitalocean:   linux-headers-5.4.0-66 linux-headers-5.4.0-66-generic

    digitalocean: The following packages will be upgraded:
    digitalocean:   accountsservice alsa-ucm-conf apport apt apt-utils base-files bind9-dnsutils

    ... MANY MORE LINES OMITTED HERE ...
    
    digitalocean: Distribution: Ubuntu
    digitalocean: Version: 20.04
    digitalocean:
    digitalocean: [PASS] Supported Operating System Detected: Ubuntu
    digitalocean: [PASS] Supported Release Detected: 20.04
    digitalocean: [PASS] Cloud-init is installed.
    digitalocean: [PASS] Firewall service (ufw) is active
    digitalocean:
    digitalocean: Updating apt package database to check for security updates, this may take a minute...
    digitalocean:
    digitalocean: [PASS] There are no pending security updates for this image.
    digitalocean:
    digitalocean:
    digitalocean: Checking for log files in /var/log
    digitalocean:
    digitalocean: [WARN] un-cleared log file, /var/log/auth.log found
    digitalocean: [WARN] un-cleared log file, /var/log/ufw.log found
    digitalocean:
    digitalocean:
    digitalocean: Checking all user-created accounts...
    digitalocean:
    digitalocean:
    digitalocean: Checking the root account...
    digitalocean: [PASS] User root has no password set.
    digitalocean: [ OK ] User root has no SSH keys present
    digitalocean: [PASS] root's Bash History appears to have been cleared
    digitalocean: [PASS] DigitalOcean Monitoring agent was not found
    digitalocean: [PASS] MongoDB is not installed
    digitalocean:
    digitalocean: ---------------------------------------------------------------------------------------------------
    digitalocean: Scan Complete.
    digitalocean: Some non-critical tests failed.  Please review these items.
    digitalocean: ---------------------------------------------------------------------------------------------------
    digitalocean: 8 Tests PASSED
    digitalocean: 2 WARNINGS
    digitalocean: 0 Tests FAILED
    digitalocean: ---------------------------------------------------------------------------------------------------
    digitalocean: Please review all [WARN] items above and ensure they are intended or resolved.  If you do not have a specific requirement, we recommend resolving these items before image submission
    digitalocean:
==> digitalocean: Gracefully shutting down droplet...
==> digitalocean: Creating snapshot: lemp-20-04-snapshot-1615212919
==> digitalocean: Waiting for snapshot to complete...
==> digitalocean: Destroying droplet...
==> digitalocean: Deleting temporary ssh key...
Build 'digitalocean' finished.
```

See that this output has two acceptable warnings, but something like having ufw disabled or present SSH keys means we couldn't accept the image.

## Configuration Details

By using [Packer's DigitalOcean Builder](https://www.packer.io/docs/builders/digitalocean.html) to integrate with the [DigitalOcean API](https://developers.digitalocean.com/), this template fully automates Marketplace image creation.

This template uses Packer's [file provisioner](https://www.packer.io/docs/provisioners/file.html) to upload complete directories to the Droplet. The contents of `files/var/` will be uploaded to `/var/`. Likewise, the contents of `files/etc/` will be uploaded to `/etc/`. One important thing to note about the file provisioner, from Packer's docs:

> The destination directory must already exist. If you need to create it, use a shell provisioner just prior to the file provisioner in order to create the directory. If the destination directory does not exist, the file provisioner may succeed, but it will have undefined results.

This template also uses Packer's [shell provisioner](https://www.packer.io/docs/provisioners/shell.html) to run scripts from the `/scripts` directory and install APT packages using an inline task.

Learn more about using Packer in [the official Packer documentation](https://www.packer.io/docs/).

## Other Examples

We also use Packer to build some of the Marketplace 1-Click Apps that DigitalOcean maintains. You can see the source code for these scripts [in this repo.](https://github.com/digitalocean/droplet-1-clicks)

## Update your App Image via API

The Vendor API makes it possible to update existing droplet 1-click apps programmatically. You can use this to ensure your listing features the most recently released version by tying it into your existing CI/CD pipeline to push an update to DigitalOcean.

Not all listing information can be updated via the API. For any changes outside the parameters of this request, you will still need to visit the Vendor Portal.

To update your app via the API, send a PATCH request to https://api.digitalocean.com/api/v1/vendor-portal/apps/<app_id>/versions/<version>

App ID can be obtained from your app’s listing in Vendor Portal, via the URL. An invalid app ID will return a 404 Not Found error. 

Apps in ‘pending’ or ‘in review’ state cannot be updated. Attempting to update an app in one of these states will return a 400 Bad Request error.

An authorization header with a bearer token is required to make the request. See [API documentation](https://docs.digitalocean.com/reference/api/api-reference/#section/Authentication) for more information on how to obtain this token.

### Request

**Authorizations**: bearer_auth (write)    
**Request Body schema**: application/json     

#### Parameters:
**imageId**    
_required_    
_integer_    
ID of the image to use for your app. The image must be a snapshot already uploaded to your DigitalOcean account, under the team you use to access the vendor portal. In addition to Packer, snapshots can be created via the API through [droplet actions](https://docs.digitalocean.com/reference/api/api-reference/#tag/Droplet-Actions). For additional information about manipulating snapshots via the API, view the [API documentation](https://docs.digitalocean.com/reference/api/api-reference/#tag/Snapshots) for snapshots.

**reasonForUpdate**    
_string_    
A brief description of the changes made which necessitate this update.     

**version**    
_string_    
The version to mark this update as.

**osVersion**    
_string_    
The version of the operating system your app runs on. A null value will not overwrite an existing value, but a blank string will.

**softwareIncluded**    
_Array of Software (see below)_    
Software types and versions included with your app. A null value will not overwrite an existing value, but an empty array will.


**Software**:

**name**    
_string_    
Name of this software. A null value will not overwrite an existing value, but a blank string will.

**version**    
_string_    
Version of this software in use by your app. A null value will not overwrite an existing value, but a blank string will.

**releaseNotes**    
_string_    
Any release notes to include alongside this software in your app’s listing. A null value will not overwrite an existing value, but a blank string will.

**website**    
_string_    
The website for this software, for further information. A null value will not overwrite an existing value, but a blank string will.

**licenseType**    
_string_    
The type of license this software uses. A null value will not overwrite an existing value, but a blank string will.

**licenseLink**    
_string_    
A link to details of the license, if applicable. A null value will not overwrite an existing value, but a blank string will. 


#### Example request:
```
curl -X PATCH \
 -H "Content-Type: application/json" \
 -H "Authorization: Bearer $DIGITALOCEAN_API_TOKEN" \
 -d '{
    "reasonForUpdate": "example", 
    "version": "3", 
    "imageId": 413639,
    "softwareIncluded": [
        {
            "name": "Ubuntu Linux",
            "version": "22.04"
        }
        ]
    }' \
    https://api.digitalocean.com/api/v1/vendor-portal/apps/$APP_ID/versions/$VERSION_ID
```

### Responses

**200 - success**

Returns app data in response.

_Example response body:_
```
{
    "appId": "60089fc6d333037bffa70d9b",
    "name": "Example App ",
    "version": 4,
    "type": "droplet",
      ...
    "status":
        {
        "value":"pending",
        "lastUpdated":1692302119556,
        "modifiedBy":"administrator",
        "reason":"example"
        },
    "emergencyContacts": 
        [
          ...
        ],
    "customData":
        {
        "version": "3",
        "osVersion": "Ubuntu 20.04",
        "description": "app desc",
        "summary": "app summary",
        "imageLabel": "sample-20-04",
        "imageId": 417346,
        "imageName": "Sample on Ubuntu 20.04",
        "imageDescription": "Sample 3 on Ubuntu 20.04",
        "reasonForUpdate": "Sample",
           ...
        }
}
```


**400 - Bad Request**
- Image ID was missing from request, or
- App in request was not a droplet 1-click, or 
- Attempted an update on an app in pending or in-review status


**403 - Unauthorized**
 - Authentication token was invalid, or
 - App or image does not belong to requestor’s team


**404 - Not Found**
 - App ID was not found or is invalid


### Automating with Packer

If you are using Packer, you can add a set of post-processor actions to automatically submit your new image to update your app, via the [manifest](https://developer.hashicorp.com/packer/docs/post-processors/manifest) and [shell-local](https://developer.hashicorp.com/packer/docs/post-processors/shell-local) post-processors, such as:

```
post-processors {     
    post-processor "manifest" {       
        output = "manifest.json"      
        strip_path = true     
    }     

    post-processor "shell-local" {        
        inline = [ "sh mp-submit.sh" ]     
    }   
}
```

where mp-submit.sh is

```
#!/bin/bash

IMG_ID=$(jq '.builds[-1].artifact_id | split(":")[1] | tonumber' manifest.json)

curl -X PATCH -H "Content-Type: application/json" -H "Authorization: Bearer ${DIGITALOCEAN_API_TOKEN}" -d "{\"reasonForUpdate\": \"new version\", \"version\": \"${APP_VERSION}\", \"imageId\": ${IMG_ID}}" https://api.digitalocean.com/api/v1/vendor-portal/apps/${APP_ID}/versions/${APP_VERSION}
```

You will need to set env variables DIGITALOCEAN_API_TOKEN, APP_VERSION, and APP_ID in your terminal to use this script. It uses jq to parse the manifest Packer creates for the snapshot ID, then uses cURL to submit it to the API endpoint.


## Supported Operating Systems

To maintain compatibility with Marketplace tools and processes, we support a limited number of Linux distributions and releases for Marketplace images. These options provide either `deb`- or `rpm`-based packaging and will have security patches and updates for a reasonable time period.

We currently support the following OSes:

- Debian 9 (stretch)
- Debian 10 (buster)
- Ubuntu 24.04 (LTS)
- Ubuntu 22.04 (LTS)
- Ubuntu 20.04 (LTS)
- Ubuntu 18.04 (LTS)
- Ubuntu 16.04 (LTS)
- CentOS 7.x
- CentOS 8.x
- CentOS Stream 8
- CentOS Stream 9
- AlmaLinux 8.x
- AlmaLinux 9.x
- Rocky Linux 8.x
- Rocky Linux 9.x

All supported operating systems are available as base images to build on in the DigitalOcean cloud.

## Software Prerequisites

The following software packages are necessary for the initial configuration of new Droplets and to ensure connectivity:

- `cloud-init` 0.76 or higher (0.79 or higher recommended)
- `openssh-server` (SFTP-enabled configuration recommended)

 All of these packages are provided by default in the default DigitalOcean base images.

 ## Image Configuration

 ### Running Commands on First Boot

 You can often pre-load much of what your image will need in your build system, but some setup (like setting database passwords or configuration that needs the Droplet's assigned IP address) will need to be run for each new Droplet created from your image.

 You can create scripts that run on first boot using cloud-init. Droplets will attempt to run any scripts located in the `/var/lib/cloud/scripts/per-instance` directory when they're first created. Scripts in that directory are run in alphanumerical order, so we recommend using a number as the beginning of the file names (e.g. `01-example-script.sh`).

 Make sure you can run the script from the command line successfully and that it has execute permissions.

 ### Running Commands on First Login

 Some of your image setup may require information that you can't get automatically, like the domain name to use for a service. You may also need to run interactive third-party scripts, like LetsEncrypt's Certbot.

 To run a script on the user's first login, we recommend adding a line to the root `.bashrc` file that runs the script and adding a line to the script that removes the line from the root `.bashrc` file.

 More specifically, at the end of the script you want to run on first login, add the following line. For consistency, we recommend putting first login scripts in the `/opt/your_company_name` directory. Make sure the script has execute permissions.

 ```
 cp -f /etc/skel/.bashrc /root/.bashrc
 ```

 Then add a line to the end of `/root/.bashrc` that runs your script by specifying the full path to the script.

 When the user first logs in, the system runs `.bashrc`, which will automatically run your script. The last line of the script overwrites the root `.bashrc` with the default `.bashrc` from the `/etc/skel` directory so the call to run your script no longer exists. Using this method, your script only runs once the first time the user logs in, but the file remains in the filesystem if they need to re-run or reference it later.

### Recommendations

* **Use the smallest suitable disk size**.

We don't support decreasing the size of a Droplet's disk because it poses data integrity issues. Building your image using the smallest disk size appropriate for your use case lets your users choose from the widest variety of Droplet plans.

* **Do not enable unnecessary DigitalOcean features on your build Droplet**.

By not enabling features like monitoring, IPv6, or private networking when you create your build Droplet, you retain more of your distribution's standard configuration, meaning you'll need to do less cleanup before you create the final image.    

* **Install software updates from the distribution's repositories** before creating your final image.

This secures the system and can save your users time when they create new Droplets from your image.

* **Use official package repositories** or well-maintained third-party repositories whenever possible. Packages installed through other means may not provide a mechanism for applying timely security updates.

For official distribution packages, we recommend maintaining the `mirrors.digitalocean.com` mirrors, which are direct mirrors of the distribution's package archive. These mirrors are provided by default and provide faster downloads because the mirrors are stored within our infrastructure.

* **If you need to provide a password to your user, consider configuring it so that it is randomly generated at boot time** and explain to users via your Getting Started instructions how to access the password. Here's an example of how you can generate a high quality, 12 character password on a Linux Droplet, and store it in a file on the Droplet.

```sh
gpg --gen-random --armor 2 12 > /root/.secrets.txt
```

* **Add a message of the day (MOTD)**, which is text displayed when a user logs into their Droplet. We recommend writing an MOTD which introduces your image's features and points users to its documentation.

You can add an MOTD to your image by creating a text file in `/etc/update-motd.d`. Naming the file beginning with `99`, like `99-image-readme`, will display the MOTD as the last text the user sees before the login prompt.

 ## Contributing

 We'd love to have your contribution to this project! [You can find more details here](https://github.com/digitalocean/marketplace-partners/blob/master/.github/CONTRIBUTING.md).

## Caveats

Avoid building architecture specific components into your 1-Click App, as your App may be run by DigitalOcean customers across a variety of operating systems and underlying hypervisors. You should use generic machine architecture to ensure consistent use across all infrastructure.

For example, avoid building ruby gems with native extensions as the underlying machine architecture may use flags that don't exist across all hypervisors.
