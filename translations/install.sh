#!/bin/bash

###
# This script will help you install any desired locale
#
# Usage:
#
# Installing single locale: opencli locale sr-sr
#
# Installing multiple locales at once: opencli locale sr-sr tr-tr
#
###

github_repo="stefanpejcic/openpanel-translations"
babel_translations="/etc/openpanel/openpanel/translations"

if [ "$#" -lt 1 ]; then
  if ! command -v jq &> /dev/null; then
    echo "jq command is required to parse JSON responses. Please install jq to use this feature."
    exit 1
  fi

  echo "Please provide at least one locale to the command, or a list"
  echo ""
  echo "Available locales:"
  locales=$(curl -s "https://api.github.com/repos/$github_repo/contents" | jq -r '.[] | select(.type == "dir" and (.name | test("^\\.") | not)) | .name')
  echo "$locales"
  echo ""
  echo "Example for a single locale (DE): opencli locale de-de"
  echo "Example for multiple locales (DE & ES): opencli locale de-de es-es"
  echo ""
  
  exit 0
fi

validate_locale() {
  # validate format (LL-LL or ll-ll)
  if [[ "$1" =~ ^[a-z]{2}-[a-z]{2}$ ]] || [[ "$1" =~ ^[A-Z]{2}-[A-Z]{2}$ ]]; then
    return 0  # ok
  else
    return 1  # not ok
  fi
}

for locale in "$@"
do
  formatted_locale=$(echo "$locale" | tr '[:upper:]' '[:lower:]')

  if validate_locale "$formatted_locale"; then
    two_letter=$(echo "$locale" | cut -d'-' -f1 | tr '[:upper:]' '[:lower:]')

    echo "Creating directory for $formatted_locale locale.."
    echo ""
    mkdir -p $babel_translations/"$two_letter"/LC_MESSAGES/  &>/dev/null
    echo "Downloading $formatted_locale locale from https://raw.githubusercontent.com/$github_repo/main/$formatted_locale/messages.pot"
    wget -O $babel_translations/"$two_letter"/LC_MESSAGES/messages.po "https://raw.githubusercontent.com/$github_repo/main/$formatted_locale/messages.po" &>/dev/null
    docker --context=default exec openpanel sh -c "pybabel update -i $babel_translations/$two_letter/LC_MESSAGES/messages.po -d $babel_translations -l $two_letter &>/dev/null"
    echo ""
  else
    echo "Invalid locale format: $locale. Skipping."
  fi
done

echo "Compiling .mo files for all available locales in $babel_translations directory.."
docker --context=default exec openpanel sh -c "pybabel compile -f -d $babel_translations  &>/dev/null"
echo "Restarting OpenPanel to apply translations.."
docker --context=default restart openpanel  &>/dev/null
echo "DONE"
