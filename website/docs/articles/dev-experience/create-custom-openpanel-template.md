# Custom OpenPanel Templates

Templates offer a flexible way to customize the appearance of your control panel interface. Whether you're looking to align the panel's look with your brand identity or simply want to refresh its appearance, our template system provides the tools necessary to achieve your desired aesthetics.

## Using Templates

OpenPanel templates are designed to be easily interchangeable, allowing users to swiftly change the look and feel of their control panel without affecting the underlying functionality. This ensures that you can update the appearance of your control panel as often as you like, without any downtime or disruption to your operations.

To change theme for OpenPanel, run command:

```opencli config update template PATH_HERE```

instead of *PATH_HERE* set either just the folder name or a full path, example:

```opencli config update template "/home/custom_template/"```


and restart service to apply: `docker restart openpanel`.

---

To change theme for OpenAdmin, run command:

```opencli config update admin_template PATH_HERE```

instead of *PATH_HERE* set either just the folder name or a full path, example:

```opencli config update admin_template "/home/custom_admin_template/"```

and restart service to apply: `service admin restart`.

---

## Creating Custom Templates

For those who require a more personalized touch, OpenPanel allows the creation of custom templates. This option is perfect for users who want to integrate their brand colors, logos, and other design elements into their control panel.

To create new templates copy the default templates folders:

For OpenPanel:
`docker cp openpanel:/templates/ /home/custom_template/`

For OpenAdmin:
`cp /user/local/admin/templates/ /home/custom_admin_template/`

and then make the changes in html/css files.
