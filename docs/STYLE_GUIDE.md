# OpenPanel Documentation Style Guide

## Terminology Standards

Always use these standard forms:
- **OpenPanel** - The hosting control panel product (not "openpanel" or "Openpanel")
- **OpenAdmin** - The administrator interface (not "openadmin" or "Openadmin")
- **OpenCLI** - The command-line interface (not "opencli" or "Opencli")
- **CSF** - ConfigServer Firewall (use full name on first mention)
- **CorazaWAF** - Web Application Firewall

## Grammar and Formatting

- Use the Oxford comma in lists
- Use three dots for ellipses (...)
- Capitalize first word of all sentences
- Place a single space between sentences
- Use active voice when possible
- End all sentences with appropriate punctuation

## Code Examples

Format shell commands as follows:
```bash
# Command description
command --flag value
```

## User Interface Terms

Refer to UI elements consistently:
- Button: "Click the **Save** button"
- Menu: "Navigate to the **Settings** menu"
- Tab: "Select the **Advanced** tab"
- Field: "Enter your domain in the **Domain Name** field"

## Technical Terms

Use consistent hyphenation:
- auto-discovery (not autodiscovery)
- web server (not webserver)
- command line (adjective) / command-line (noun)

## Script Conventions

### Bash/Shell Scripts
- Use meaningful variable names in `UPPER_CASE` for environment/global variables
- Use `lower_case` for local variables and function names
- Add comments for non-obvious code sections
- Group related functions together
- Document function parameters with comments:
  ```bash
  # Description of what the function does
  # $1: First parameter description
  # $2: Second parameter description
  function_name() {
      local param1="$1"
      local param2="$2"
      # Function body
  }
  ```
- Use proper error handling with meaningful error messages
- Initialize variables at the beginning of scripts or functions
- Use proper indentation (2 or 4 spaces) consistently

### Error Formatting
- Error messages should start with "Error:" in red text where possible
- Warning messages should use yellow/orange text where possible
- Success messages should use green text where possible
