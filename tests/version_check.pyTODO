import requests

'''
Check if current panel version is +1 then the latest
'''

local_version_file = '/usr/local/panel/version'
latest_version_url = 'https://openpanel.org/version'

def get_local_version():
    try:
        with open(local_version_file, 'r') as file:
            return file.readline().strip()
    except FileNotFoundError:
        print(f"Error: {local_version_file} not found.")
        return None

def get_remote_version():
    try:
        response = requests.get(latest_version_url)
        response.raise_for_status()
        return response.text.strip()
    except requests.exceptions.RequestException as e:
        print(f"Error fetching remote version: {e}")
        return None

def is_version_incremented(local, remote):
    local_parts = list(map(int, local.split('.')))
    remote_parts = list(map(int, remote.split('.')))
    
    if local_parts[0] > remote_parts[0]:  # Major version increment
        return True
    elif local_parts[0] == remote_parts[0] and local_parts[1] > remote_parts[1]:  # Minor version increment
        return True
    elif (local_parts[0] == remote_parts[0] and local_parts[1] == remote_parts[1] and 
          local_parts[2] == remote_parts[2] + 1):  # Patch increment by exactly +1
        return True
    else:
        return False

def main():
    local_version = get_local_version()
    remote_version = get_remote_version()

    if local_version and remote_version:
        if is_version_incremented(local_version, remote_version):
            print(f"Local version {local_version} is properly incremented over remote version {remote_version}.")
        else:
            print(f"Error: Local version {local_version} is not incremented over remote version {remote_version}.")
    else:
        print("Unable to verify version increment due to missing version data.")

if __name__ == "__main__":
    main()
  
