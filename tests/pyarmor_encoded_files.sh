#!/bin/bash
########################################################
# Make sure all python files are encoded with PyArmor! #
########################################################

# Function to set app_path based on the input attribute
set_app_path() {
    case "$1" in
        admin)
            echo "/usr/local/admin"
            ;;
        panel)
            check_directory "/usr/local/panel/panel"
            echo "/usr/local/panel/panel"
            ;;
        *)
            echo "Invalid attribute. Please use 'admin' or 'panel'."
            exit 1
            ;;
    esac
}

# Function to check if a file is encoded with PyArmor
is_pyarmor_encoded() {
    local file_path="$1"
    if [[ ! -f "$file_path" ]]; then
        return 1
    fi

    # Read the first line of the file and check for PyArmor
    read -r first_line < "$file_path"
    [[ "$first_line" == *"# Pyarmor"* ]]
}

# Function to check if directory exists
check_directory() {
    local dir="$1"
    if [[ ! -d "$dir" ]]; then
        cd /root && docker compose up -d openpanel >/dev/null 2>&1
        docker cp openpanel:/usr/local/panel /usr/local/panel >/dev/null 2>&1
    fi
}

# Main script logic
if [[ $# -ne 1 ]]; then
    echo "Usage: $0 <admin|panel>"
    exit 1
fi

# Set app_path based on user input
app_path=$(set_app_path "$1")

unencoded_files=()

# Traverse the app_path and check each .py file
while IFS= read -r -d '' file; do
    if ! is_pyarmor_encoded "$file"; then
        unencoded_files+=("$file")
    fi
done < <(find "$app_path" -type f -name '*.py' -print0)

# Output the results
if [[ ${#unencoded_files[@]} -gt 0 ]]; then
    echo -e "The following files are not encoded with PyArmor:"
    for file in "${unencoded_files[@]}"; do
        echo -e "- \e[31m$file\e[0m"
    done
else
    echo "✓ All Python files are properly encoded with PyArmor!"
fi

rm -rf /usr/local/panel/panel >/dev/null 2>&1
