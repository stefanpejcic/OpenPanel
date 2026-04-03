# Dev

Overwrite HTML template or Python module within the OpenPanel UI container

- To list all files, run:
  ```bash
  opencli dev
  ```
- To overwrite a Python module:
  ```bash
  opencli dev modules/PATH.py
  ```
- To overwrite an HTML template:
  ```bash
  opencli dev templates/dashboard/dashboard.html
  ```

**Note:** This command is intended for development purposes, specifically when creating custom modules or modifying templates. It provides a quick way to overwrite files for testing. For persistent changes, refer to the [Customizing](/customize.html) section.
