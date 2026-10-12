---
sidebar_position: 1
---

# Manage Websites

Websites page lists every website in Site Manager across all users on the server: WordPress and other CMS installs, Website Builder sites and Node.js, Python, Ruby or Java apps.

## List websites

<Tabs>
  <TabItem value="openadmin-websites-list" label="With OpenAdmin" default>

To list all websites navigate to **Websites** in OpenAdmin.

The table shows these columns:

| Column              | Description                                                                 |
|---------------------|-----------------------------------------------------------------------------|
| **Site**            | Favicon and address of the website, with its folder on the server below it. Click the address to open the site. |
| **Type**            | Icon of the application, for example WordPress, Joomla, Website Builder or Node.js. Hover over it to see the name. |
| **Owner**           | The OpenPanel user who owns the website, marked **SUSPENDED** when that user is suspended. |
| **Version**         | Installed version of the application. WordPress sites show when a newer WordPress version is available. |
| **Vulnerabilities** | Known vulnerable plugins, themes or core versions from the last [vulnerability scan](#scan-for-vulnerabilities). WordPress only. |
| **ImunifyAV**       | Result of the last [ImunifyAV scan](#scan-with-imunifyav) of the owner's files. Shown only when ImunifyAV is installed. |
| **PageSpeed**       | Google PageSpeed score from the last [PageSpeed check](#check-pagespeed). |
| **Created**         | When the website was added to Site Manager.                                 |
| **Actions**         | Menu with the [website actions](#website-actions).                          |

Click **Show Columns** to show or hide any column except Site and Actions. All columns are shown by default and your choice is remembered for your browser.

Use the search box (**Search sites, owners, domains...**) to filter the list by site address, owner, domain, admin email or type. When more than one type of website exists, buttons above the table (**All**, **WordPress**, **Website Builder**...) filter the list by type.

Click a column header to sort by it.

  </TabItem>
  <TabItem value="CLI-websites-list" label="With OpenCLI">

To list all websites run:

```bash
opencli websites-all
```

  </TabItem>
</Tabs>

## Scan for websites

Websites installed by users outside of Site Manager, for example uploaded over FTP or the File Manager, are not listed until they are detected.

<Tabs>
  <TabItem value="openadmin-websites-scan" label="With OpenAdmin" default>

Click **Scan for Websites**, enter a username or leave the field empty to scan all users, and click **Start Scan**. The scan runs in the background and new sites show up in the list once it finishes.

  </TabItem>
  <TabItem value="CLI-websites-scan" label="With OpenCLI">

To scan a single user run:

```bash
opencli websites-scan <USERNAME>
```

To scan all users run:

```bash
opencli websites-scan -all
```

  </TabItem>
</Tabs>

## Website actions

Click the **⋯** button at the end of a row to open the actions menu:

| Action                       | Description                                                                 |
|------------------------------|-----------------------------------------------------------------------------|
| **Visit site**               | Opens the website in a new tab.                                             |
| **Log in to wp-admin**       | Logs you in to the WordPress dashboard as an administrator. WordPress only. |
| **Manage as owner**          | Logs in to OpenPanel as the owner and opens this website in Site Manager.  |
| **Owner account**            | Opens the owner's account page in OpenAdmin.                                |
| **Edit Virtual Host**        | Opens the web server configuration of the website's domain.                 |
| **Check PageSpeed**          | Runs a [Google PageSpeed check](#check-pagespeed).                          |
| **Scan with ImunifyAV**      | Queues an [ImunifyAV scan](#scan-with-imunifyav) of the owner's files. Shown only when ImunifyAV is installed. |
| **Scan for vulnerabilities** | Runs a [vulnerability scan](#scan-for-vulnerabilities). WordPress only.     |
| **Detach**                   | Removes the website from Site Manager. Its files and databases are kept and it can be added back with [Scan for websites](#scan-for-websites). |

## Check PageSpeed

Hover over the PageSpeed score to see the Performance score, First Contentful Paint, Speed Index and Time to Interactive for desktop and mobile, and click **Update** to run the check again.

<Tabs>
  <TabItem value="openadmin-websites-pagespeed" label="With OpenAdmin" default>

Select **Check PageSpeed** from the actions menu, or select several websites and use the bulk action with the same name.

  </TabItem>
  <TabItem value="CLI-websites-pagespeed" label="With OpenCLI">

To check a single website run:

```bash
opencli websites-pagespeed <DOMAIN>
```

To check all websites run:

```bash
opencli websites-pagespeed -all
```

  </TabItem>
</Tabs>

## Scan for vulnerabilities

WordPress sites can be scanned for known vulnerabilities in their core version, plugins and themes. When something is found, the Vulnerabilities column shows how many vulnerabilities were found in how many components. Click it to see each vulnerable plugin or theme with details from Wordfence, WPScan, Patchstack and CVE.

<Tabs>
  <TabItem value="openadmin-websites-vulnerability" label="With OpenAdmin" default>

Select **Scan for vulnerabilities** from the actions menu, or select several websites and use the bulk action with the same name. Websites that are not WordPress are skipped.

  </TabItem>
  <TabItem value="CLI-websites-vulnerability" label="With OpenCLI">

To scan a single website run:

```bash
opencli websites-vulnerability <DOMAIN>
```

To scan all WordPress websites run:

```bash
opencli websites-vulnerability --all
```

  </TabItem>
</Tabs>

## Scan with ImunifyAV

When [ImunifyAV](/docs/admin/security/imunify/) is installed, the ImunifyAV column shows the result of the last malware scan of the owner's files: **Clean**, the number of infected files, or **Scanning...** while a scan is running. Click the number of infected files to open ImunifyAV and handle them.

ImunifyAV scans the whole home directory of a user, so the result is the same for every website of that user.

To start a scan, click the scan icon in the ImunifyAV column or select **Scan with ImunifyAV** from the actions menu. To scan several users at once, select their websites and use the **Scan with ImunifyAV** bulk action. Each owner is scanned once, even when several of their websites are selected.

Users can also be scanned from the **Users** page by selecting them and using the **Scan with ImunifyAV** bulk action.

## Bulk actions

Select websites with the checkboxes, or select all with the checkbox in the table header, and choose an action from the bar at the bottom of the page:

| Action                       | Description                                                       |
|------------------------------|-------------------------------------------------------------------|
| **Check PageSpeed**          | Runs a PageSpeed check for each selected website.                 |
| **Scan for vulnerabilities** | Scans each selected WordPress website, other types are skipped.   |
| **Scan with ImunifyAV**      | Queues an ImunifyAV scan for the owners of the selected websites. Shown only when ImunifyAV is installed. |
| **Detach**                   | Removes the selected websites from Site Manager, keeping their files and databases. |
