# Customizing OpenPanel Interface (Branding and White-Label)

Everything in OpenPanel is modular and can easily be modified or disabled without breaking the rest of the functionalities.

To customize OpenPanel, you have the following options:

- [Display personalized message per user](#personalized-messages)
- [Enable/disable features and pages from the OpenPanel interface](#enabledisable-features)
- [Set pre-installed services for users](#set-pre-installed-services)
- [Localize the interface](#localize-the-interface)
- [Set custom branding](#set-custom-branding)
- [Set a custom color scheme](#set-a-custom-color-scheme)
- [Replace How-to articles with your knowledge base](#replace-how-to-articles-with-your-knowledge-base)
- [Customize any page](#edit-any-page-template)
- [Customize login page](#customize-login-page)
- [Add custom CSS or JS code to the interface](#create-custom-pages)
- [Self-hosted temporary links for SiteManager](/docs/articles/dev-experience/selfhosted-temporary-links-api/)
- [Self-hosted screenshots for SiteManager](/docs/articles/dev-experience/selfhosted-screenshots-api/)


## Personalized messages

Administrators can set a custom message to be displayed for any OpenPanel user by creating/editing file `/etc/openpanel/openpanel/core/users/USERNAME/custom_message.html` file. Both plain text and HTML code are accepted:

![custom_message_per_user](https://i.postimg.cc/5xTqkddZ/2024-09-05-15-52.png)

## Enable/disable features

Administrators have the ability to enable or disable each feature (page) in the OpenPanel interface. To activate a feature, navigate to [OpenAdmnin > Settings > OpenPanel](/docs/admin/settings/openpanel/#enable-features) and select service name in the "Enable Features" section and click save. 

Once enabled, the feature becomes instantly available to all users, appearing in the OpenPanel interface sidebar, search results, and dashboard icons.




## Set pre-installed services

OpenPanel uses [docker images](https://dev.openpanel.com/images/) as the base for each hosting plan. Based on the docker image, different services can be set per plan/user. For examples, we provide 2 docker iamges, one that has nginx pre-installed and another that uses apache. By creating a custom docker image, you can set in that image what to be pre-installed when you create a new user, for exmaple, set mariadb instead of mysql or install php ioncube loader extension.

To add a custom service pre-installed for users:

- [Create a custom docker image](/docs/articles/docker/building_a_docker_image_example_include_php_ioncubeloader/)
- [Create a new hosting plan with that docker image](/docs/admin/plans/hosting_plans/#create-a-plan)
- [Create a new user on that plan](/docs/admin/users/openpanel/#create-users)



## Localize the interface

OpenPanel is localization ready and can easily be translated into any language.

OpenPanel is shipped with the EN locale, [additional locales can be installed by the Administrator](https://dev.openpanel.com/localization.html#How-to-translate).

To install additional locales (replace `de-de` with your locale):

```bash
opencli locale de-de
```

To translate OpenPanel to another language follow these steps:

1. Fork [this repository](https://github.com/stefanpejcic/openpanel-translations).
2. Copy `en_us` to your locale e.g. `es_es`
3. Translate the `messages.pot` file
4. Send a [pull request](https://github.com/stefanpejcic/openpanel-translations/pulls)




## Set custom branding

Custom brand name and logo can be set from [OpenAdmnin > Settings > OpenPanel](/docs/admin/settings/openpanel/#branding) page.

To set a custom name visible in the OpenPanel sidebar and on login pages, enter the desired name in the "Brand name" option. Alternatively, to display a logo instead, provide the URL in the "Logo image" field and save the changes.




## Set a custom color scheme

To set a custom color-scheme for OpenPanel interface, edit the `/etc/openpanel/openpanel/custom_code/custom.css` file and in it set your preferred color scheme.

```bash
nano /etc/openpanel/openpanel/custom_code/custom.css
```

Set the custom css code, save and restart openpanel to apply changes:

```bash
cd /root && docker compose up -d openpanel
```

Example:

![custom_css_code](https://i.postimg.cc/YprhHZhg/2024-06-18-15-04.png)





## Replace How-to articles with your knowledge base

[OpenPanel Dashboard page](/docs/panel/dashboard) displays [How-to articles](/docs/panel/dashboard/#how-to-guides) from the OpenPanel Docs, however these can be changed to display your knowledgebase articles instead. 

Edit the file `/etc/openpanel/openpanel/conf/knowledge_base_articles.json` and in it set your links:

```json
{
    "how_to_topics": [
        {"title": "How to install WordPress", "link": "https://openpanel.com/docs/panel/applications/wordpress#install-wordpress"},
        {"title": "Publishing a Python Application", "link": "https://openpanel.com/docs/panel/applications/pm2#python-applications"},
        {"title": "How to edit Nginx / Apache configuration", "link": "https://openpanel.com/docs/panel/advanced/server_settings#nginx--apache-settings"},
        {"title": "How to create a new MySQL database", "link": "https://openpanel.com/docs/panel/databases/#create-a-mysql-database"},
        {"title": "How to add a Cronjob", "link": "https://openpanel.com/docs/panel/advanced/cronjobs#add-a-cronjob"},
        {"title": "How to change server TimeZone", "link": "https://openpanel.com/docs/panel/advanced/server_settings#server-time"}
    ],
    "knowledge_base_link": "https://openpanel.com/docs/panel/intro/?source=openpanel_server"
}
```


## Edit any page template

Each OpenPanel template code is open and can easily be edited by just editing the HTML code.

Templates are located inside a Docker container named `openpanel`, so you first need to find the template that has the code that you want to edit.

For example, to edit the sidebar and hide the OpenPanel logo, follow these steps:

1. Create a new folder/file locally for your modified code.
   ```bash
   mkdir /root/custom_template/
   ```
2. Copy the existing template code.
   ```bash
   docker cp openpanel:/usr/local/panel/templates/partials/sidebar.html /root/custom_template/sidebar.html
   ```
3. Edit the code.

4. Configure OpenPanel to use your template.
   Edit the `/root/docker-compose.yml` file and in it setyour file to overwrite the original template:
   ```bash
   nano /root/docker-compose.yml
   ```
   and in the file under [openpanel > volumes](https://github.com/stefanpejcic/openpanel-configuration/blob/180c781bfb7122c354fd339fbee43c1ce6ec017f/docker/compose/new-docker-compose.yml#L31) set local path and original:
   ```bash
   - /root/custom_theme/sidebar.html:/usr/local/panel/templates/partials/sidebar.html
   ```
6. Restart OpenPanel to apply the new template.
   ```bash
   cd /root && docker compose up -d openpanel
   ```


## Customize login page


OpenPanel login page template code is located at `/usr/local/panel/templates/user/login.html` inside the docker container.

To edit the login page:

1. Create a new folder/file locally for your modified code.
   ```bash
   mkdir /root/custom_template/
   ```
2. Copy the existing template code.
   ```bash
   docker cp openpanel:/usr/local/panel/templates/user/login.html /root/custom_template/login.html
   ```
3. Edit the code.

4. Configure OpenPanel to use your template.
   Edit the `/root/docker-compose.yml` file and in it setyour file to overwrite the original template:
   ```bash
   nano /root/docker-compose.yml
   ```
   and in the file under [openpanel > volumes](https://github.com/stefanpejcic/openpanel-configuration/blob/180c781bfb7122c354fd339fbee43c1ce6ec017f/docker/compose/new-docker-compose.yml#L31) set local path and original:
   ```bash
   - /root/custom_theme/login.html:/usr/local/panel/templates/user/login.html
   ```
6. Restart OpenPanel to apply the new login template.
   ```bash
   cd /root && docker compose up -d openpanel
   ```


## Add custom CSS or JS code

To add custom CSS code to the OpenPanel interface, edit the file `/etc/openpanel/openpanel/custom_code/custom.css`:

```bash
nano /etc/openpanel/openpanel/custom_code/custom.css
```

To add custom JavaScript code to the OpenPanel interface, edit the file `/etc/openpanel/openpanel/custom_code/custom.js`:

```bash
nano /etc/openpanel/openpanel/custom_code/custom.js
```

To insert custom code within the `<head>` tag of the OpenPanel interface, modify the content of the file located at `/etc/openpanel/openpanel/custom_code/in_header.html` and include your custom code within it:

```bash
nano /etc/openpanel/openpanel/custom_code/in_header.html
```

To insert custom code within the `<footer>` tag of the OpenPanel interface, modify the content of the file located at `/etc/openpanel/openpanel/custom_code/in_footer.html` and include your custom code within it:

```bash
nano /etc/openpanel/openpanel/custom_code/in_footer.html
```

