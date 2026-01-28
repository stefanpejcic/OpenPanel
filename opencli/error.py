import subprocess
import argparse

# TODO: https://github.com/stefanpejcic/opencli/issues/63

'''
Display logs for error code
'''

def extract_error_log_from_docker(error_code):
    # Use subprocess to get the container logs
    try:
        result = subprocess.run(
            ['docker', '--context', 'default',  'logs', '--since=60m', 'openpanel'],
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            #stderr=subprocess.PIPE,
            text=True,
            check=True
        )
    except subprocess.CalledProcessError as e:
        print(f"Error running docker logs: {e.stderr}")
        return f"Error while fetching logs: {e.stderr}"

    logs = result.stdout.splitlines()

    result_log = []
    found_error_code = False

    for line in reversed(logs):
        if found_error_code:
            result_log.append(line.strip())
            if 'ERROR' in line:
                break
        elif error_code.lower() in line.lower():
            found_error_code = True
            result_log.append(line.strip())

    result_log.reverse()

    if not found_error_code:
        print("\n=== NO LOGS FOR ERROR ID ===")
        return f"Error Code '{error_code}' not found in the OpenPanel UI logs."

    return result_log

def main():
    parser = argparse.ArgumentParser(description="Extract error logs from the OpenPanel container by error code.")
    parser.add_argument("error_code", help="The error code to search for in the logs")

    args = parser.parse_args()
    error_log = extract_error_log_from_docker(args.error_code)

    # Print the result
    if isinstance(error_log, str):
        print(error_log)
    else:
        print(f"\n=== LOGS FOR ERROR ID: '{args.error_code}' ===\n")
        for line in error_log[:-1]:
            print(line)

if __name__ == "__main__":
    main()
