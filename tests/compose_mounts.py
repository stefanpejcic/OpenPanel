import os
import yaml

compose_file_path = '/etc/openpanel/docker/compose/new-docker-compose.yml'

def load_compose_file():
    try:
        with open(compose_file_path, 'r') as file:
            return yaml.safe_load(file)
    except FileNotFoundError:
        print(f"Error: {compose_file_path} not found.")
        return None
    except yaml.YAMLError as e:
        print(f"Error parsing YAML: {e}")
        return None

def check_mount_paths(compose_data):
    missing_paths = []
    
    for service_name, service_data in compose_data.get('services', {}).items():
        print(f"\nChecking service: {service_name}")
        
        volumes = service_data.get('volumes', [])
        for volume in volumes:
            host_path = volume.split(':')[0]
            
            if not os.path.exists(host_path):
                print(f"Missing path: {host_path}")
                missing_paths.append(host_path)
            else:
                print(f"Path exists: {host_path}")
    
    return missing_paths

def main():
    compose_data = load_compose_file()
    if not compose_data:
        print("Error: Unable to load compose data.")
        return
    
    missing_paths = check_mount_paths(compose_data)
    
    if missing_paths:
        print("\nThe following paths are missing:")
        for path in missing_paths:
            print(f" - {path}")
    else:
        print("\nAll mount paths exist.")

if __name__ == "__main__":
    main()
  
