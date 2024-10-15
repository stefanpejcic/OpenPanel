#!/bin/bash

###
# This script will help you install any desired locale
#
# Usage:
#
# Installing single locale: bash <(curl -sSL https://raw.githubusercontent.com/stefanpejcic/openpanel-translations/main/install.sh) sr-sr
#
# Installing multiple locales at once: bash <(curl -sSL https://raw.githubusercontent.com/stefanpejcic/openpanel-translations/main/install.sh) sr-sr tr-tr
#
###

# might change in future
github_repo="stefanpejcic/openpanel-translations"

# locales dir since OpenPanel v.0.2.1
babel_translations="/etc/openpanel/openpanel/translations"

# at least 1 locale is needed
if [ "$#" -lt 1 ]; then
  if ! command -v jq &> /dev/null; then
    echo "jq command is required to parse JSON responses. Please install jq to use this feature."
    exit 1
  fi

  echo "Please provide at least one locale to the command, or a list"
  echo ""
  # list available locales from github repo
  echo "Available locales:"
  locales=$(curl -s "https://api.github.com/repos/$github_repo/contents" | jq -r '.[] | select(.type == "dir") | .name')
  echo "$locales"
  echo ""
  echo "Example for a single locale (DE): opencli locale de-de"
  echo ""
  echo "Example for multiple locales (DE & ES): opencli locale de-de es-es"
  echo ""
  
  exit 0
fi

cd /usr/local/panel

validate_locale() {
  # validate format (LL-LL or ll-ll)
  if [[ "$1" =~ ^[a-z]{2}-[a-z]{2}$ ]] || [[ "$1" =~ ^[A-Z]{2}-[A-Z]{2}$ ]]; then
    return 0  # ok
  else
    return 1  # not ok
  fi
}

# Loop through each provided locale
for locale in "$@"
do
  # must be lowercase
  formatted_locale=$(echo "$locale" | tr '[:upper:]' '[:lower:]')

  if validate_locale "$formatted_locale"; then
    # babel supports just 2 letters
    two_letter=$(echo "$locale" | cut -d'-' -f1 | tr '[:upper:]' '[:lower:]')

    echo "Creating directory for $formatted_locale locale.."
    echo ""
    mkdir -p $babel_translations/"$two_letter"/LC_MESSAGES/  &>/dev/null
    echo "Downloading $formatted_locale locale from https://raw.githubusercontent.com/$github_repo/main/$formatted_locale/messages.pot"
    wget -O $babel_translations/"$two_letter"/LC_MESSAGES/messages.pot "https://raw.githubusercontent.com/$github_repo/main/$formatted_locale/messages.pot" &>/dev/null
    pybabel init -i $babel_translations/$two_letter/LC_MESSAGES/messages.pot -d $babel_translations -l "$two_letter" &>/dev/null
    echo ""
  else
    echo "Invalid locale format: $locale. Skipping."
  fi
done

# Do this only once

echo "Compiling .mo files for all available locales in $babel_translations directory.."
pybabel compile -f -d $babel_translations  &>/dev/null
echo "Restarting OpenPanel to apply translations.."
docker restart openpanel  &>/dev/null
echo "DONE"
