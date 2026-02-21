# WordPress Themes and Plugins Sets

As an OpenPanel user, you can create **Theme Sets** and **Plugin Sets** for your account. Plugins and themes from these sets will automatically be installed on every new WordPress site you create using OpenPanel's WP Manager. This is especially handy if you frequently use the same plugins and themes across multiple sites, such as the Elementor theme with a child theme, Elementor plugin, and so on.

:::info
The 'filemanager' and 'wordpress' features need to be enabled for your account in order to use sets.
:::
---

## Themes Set

To create a Theme Set:

1. Go to **OpenPanel > WP Manager** and click on **Themes**.
2. A new page will open with a file editor, allowing you to configure the themes you want installed automatically.

**Instructions:**

* Add **one theme per line**.
* To install free themes from WordPress.org, simply use the theme slug:
  ![theme slug](https://i.postimg.cc/JnRzksmy/theme-slug.png)

  ```
  startup-business-elementor
  ```
* To install premium or third-party themes, provide the direct download URL:

  ```
  http://s3.amazonaws.com/bucketname/my-theme.zip?AWSAccessKeyId=123&Expires=456&Signature=abcdef
  ```
* Lines starting with `#` are treated as comments and ignored.

**Example Theme Set file:**

```bash
# Free themes from wordpress.org
hello-elementor
startup-business-elementor

# Premium theme from a remote URL
http://s3.amazonaws.com/bucketname/my-theme.zip?AWSAccessKeyId=123&Expires=456&Signature=abcdef
```

You can add as many themes as you like. Once done, **click the 'Save' button**.

---

## Plugins Set

To create a Plugin Set:

1. Go to **OpenPanel > WP Manager** and click on **Plugins**.
2. A new page will open with a file editor, allowing you to configure the plugins you want installed automatically.

**Instructions:**

* Add **one plugin per line**.
* To install free plugins from WordPress.org, use the plugin slug:
  ![plugin slug](https://i.postimg.cc/9fdf0d3G/plugin-slug.png)

  ```
  enable-wp-debug-toggle
  ```
* To install premium or third-party plugins, provide the direct download URL:

  ```
  https://github.com/envato/wp-envato-market/archive/master.zip
  ```
* Lines starting with `#` are treated as comments and ignored.

**Example Plugin Set file:**

```bash
# Free plugins from wordpress.org
akismet
bbpress
enable-wp-debug-toggle

# Premium plugin from GitHub
https://github.com/envato/wp-envato-market/archive/master.zip

# Plugin from S3 with access key
http://s3.amazonaws.com/bucketname/my-plugin.zip?AWSAccessKeyId=123&Expires=456&Signature=abcdef
```

You can add as many plugins as you like. Once done, **click the 'Save' button**.

