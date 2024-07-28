#!/bin/bash

# Parse command line options
while getopts "yf" opt; do
  case $opt in
    y)
      confirm="yes"
      ;;
    f)
      confirm="force"
      ;;
    *)
      usage
      ;;
  esac
done

# If no flags are set, ask for user confirmation
if [ -z "$confirm" ]; then
  read -p "This will convert your website to host public OpenPanel demo. Do you want to proceed? (y/n) " user_input
  case $user_input in
    [Yy]*)
      #echo "User confirmed to proceed."
      
      ;;
    *)
      exit 1
      ;;
  esac
fi

echo "Setting demo..."

