
import os

destination_config_dir = '/usr/local/RapidBackup/config/destinations'
destinacija = "/etc/passwd /etc/passwd"

# Show file path
def delete_destination():
    file_path = os.path.join(destination_config_dir, f"{destinacija}.conf")
    print(file_path)

    try:
        with open(file_path, 'r') as file:
            print(file.read())
    except FileNotFoundError:
        print("nema")


delete_destination()
