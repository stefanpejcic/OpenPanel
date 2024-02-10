---
sidebar_position: 6
---

# Development

Welcome to the Development section of OpenPanel, where developers can dive into extending the capabilities of our platform. OpenPanel is designed with flexibility in mind, allowing you to enhance its functionality and tailor it to your specific needs.

## Building Custom Modules

Developers have the opportunity to create custom modules that seamlessly integrate with OpenPanel. These modules can introduce new features, extend existing functionalities, and adapt the platform to unique use cases. 

Stay tuned for the upcoming addition of our Module Development Guide, providing a comprehensive walkthrough on creating and seamlessly integrating custom modules into OpenPanel. Exciting developments are on the horizon as we enhance our resources to empower developers in shaping the future of OpenPanel.



## Custom Code


### Custom CSS

To add custom CSS code to OpenPanel interface, edit file `/usr/local/panel/templates/custom_code/in_header.html` and put your custom code inside. Please note 


```bash
nano /usr/local/panel/templates/custom_code/custom.css
```

### Custom JS


```bash
nano /usr/local/panel/templates/custom_code/custom.js
```

### Code in Header

To insert custom code within the `<head>` tag of the OpenPanel interface, modify the content of the file located at `/usr/local/panel/templates/custom_code/in_header.html` and include your custom code within it.

```bash
nano /usr/local/panel/templates/custom_code/in_header.html
```

### Code in Footer

To insert custom code within the `<footer>` tag of the OpenPanel interface, modify the content of the file located at `/usr/local/panel/templates/custom_code/in_footer.html` and include your custom code within it.

```bash
nano /usr/local/panel/templates/custom_code/in_footer.html
```


### After installation

To execute custom code following the installation of OpenPanel, place your custom script on the server. When initiating the OpenPanel installation process, use the `--post_install=` flag and specify the path to your script within it.
Example:

```bash
 --post_install=/root/my_custom_script.sh
```



## Developer Support

Connect with fellow developers in the [OpenPanel community](https://community.openpanel.co/) through our Forum. Share your experiences, seek advice, and collaborate on innovative solutions. Together, we can expand the capabilities of OpenPanel and create a vibrant ecosystem.
