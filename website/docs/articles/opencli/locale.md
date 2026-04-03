# Locale

List or install new locale(s) for OpenPanel interface:

- List all available locales:
  ```bash
  opencli locale
  ```
- Installing single locale:
  ```bash
  opencli locale sr-sr
  ```
- Installing multiple locales at once:
  ```bash
  opencli locale sr-sr tr-tr
  ```
- Installing ALL available locales at once:
  ```bash
  opencli locale $(curl -s "https://api.github.com/repos/stefanpejcic/openpanel-translations/contents" | jq -r '.[] | select(.type=="dir") | .name' | tr '\n' ' ')
  ```
