import sys
import re

def remove_localhost_lines_and_replace_grant_usage(input_path, output_path):
    try:
        with open(input_path, 'r') as infile:
            lines = infile.readlines()

        with open(output_path, 'w') as outfile:
            for line in lines:
                if 'localhost' in line:
                    continue

                # Replace escaped underscores
                line = line.replace(r'\_', '_')
                
                # Check if the line starts with "GRANT USAGE ON"
                if line.startswith('GRANT USAGE ON'):
                    # Extract the username and password from the line using regex
                    match = re.match(r"GRANT USAGE ON \*\.\* TO '([^']+)'@'([^']+)' IDENTIFIED BY PASSWORD '([^']+)'", line)
                    if match:
                        user = match.group(1)
                        host = match.group(2)
                        password = match.group(3)
                        new_line = f"CREATE USER '{user}'@'{host}' IDENTIFIED WITH 'mysql_native_password' AS '{password}';\n"

                        outfile.write(new_line)
                    else:
                        # If regex doesn't match, write the original line
                        outfile.write(line)
                else:
                    outfile.write(line)
                    
        print(f"Processed file saved to {output_path}")
    
    except FileNotFoundError:
        print(f"Error: The file {input_path} was not found.")
    except Exception as e:
        print(f"An error occurred: {e}")

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python script.py <input_sql_file> <output_sql_file>")
    else:
        input_path = sys.argv[1]
        output_path = sys.argv[2]
        remove_localhost_lines_and_replace_grant_usage(input_path, output_path)
