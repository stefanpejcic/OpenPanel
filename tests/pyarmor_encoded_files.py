import os

app_path = '/usr/local/panel'

def is_pyarmor_encoded(file_path):
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            first_line = f.readline()
            return '# PyArmor' in first_line
    except UnicodeDecodeError:
        return True

def check_all_files_encoded():
    unencoded_files = []

    for root, dirs, files in os.walk(app_path):
        for file in files:
            if file.endswith('.py'):
                file_path = os.path.join(root, file)
                if not is_pyarmor_encoded(file_path):
                    unencoded_files.append(file_path)

    return unencoded_files

unencoded_files = check_all_files_encoded()

if unencoded_files:
    print("The following files are not encoded with PyArmor:")
    for file in unencoded_files:
        print(file)
else:
    print("All Python files are properly encoded with PyArmor!")
  
