# Hungarian Documentation Translation

A Hungarian folder structure (`i18n/hu/`) has been created for the OpenPanel documentation.

## How to Translate the Documentation Files

There are 324 `.md` files in `i18n/hu/docusaurus-plugin-content-docs/current`.

To automate the translation, you can use the `translate_docs.py` script located in the root of this repository.

1. **Install Argos Translate** (or modify the script to use a different API):
   ```bash
   pip install argostranslate
   ```
   *(Ensure you have downloaded the en -> hu argos model)*

2. **Run the script**:
   ```bash
   python3 translate_docs.py
   ```
   *Note: I have set the script to only translate the first 3 files for testing. To translate everything, remove `[:3]` from the `md_files` loop at the bottom of the script.*

3. **To test the local website**:
   ```bash
   npm install
   npm run start -- --locale hu
   ```
