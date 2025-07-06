#!/bin/bash
################################################################################
# Script Name: resources.sh
# Description: View services limits for user.
# Usage: opencli user-resources <CONTEXT> [--activate=<SERVICE_NAME>] [--deactivate=<SERVICE_NAME>] [--update_cpu=<FLOAT>] [--update_ram=<FLOAT>] [--service=<NAME>] [--json]
# Author: Stefan Pejcic
# Created: 26.02.2025
# Last Modified: 04.07.2025
# Company: openpanel.com
# Copyright (c) openpanel.com
# 
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
# 
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
# 
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.
################################################################################

# Process the OS service (first argument)
context=$1
env_file="/home/${context}/.env"
json_output=false
new_service=""
update_cpu=""
update_ram=""
service_to_update_cpu_ram=""
message=""

json_print=false
message_print=false

usage() {
    echo "Usage: opencli user-resources <context> [options]"
    echo ""
    echo "Arguments:"
    echo "  <context>          The docker context for user."
    echo ""
    echo "Options:"
    echo "  --json                       Output the result in JSON format."
    echo "  --update_cpu=<value>         Update CPU allocation for the service (e.g., --update_cpu=2 --service=apache)."
    echo "  --update_ram=<value>         Update RAM allocation for the service (e.g., --update_ram=4 --service=apache)."
    echo "  --activate=<service>         Activate the specified docker service."
    echo "  --service=<service_name>     Specify a service to update its CPU and RAM configuration."
    echo "  --deactivate=<service>       Deactivate the specified service."
    echo "  --activate=<service> --force Pull image again and activate."
    echo ""
    echo "Example usage:"
    echo "  opencli user-resources stefan --json"
    echo "  opencli user-resources stefan --activate=apache"
    echo "  opencli user-resources stefan --deactivate=apache"
    echo "  opencli user-resources stefan --activate=apache --force"
    echo "  opencli user-resources stefan --update_cpu=2"
    echo "  opencli user-resources stefan --update_ram=4"
    echo "  opencli user-resources stefan --service=apache --update_ram=1.2"
    echo "  opencli user-resources stefan --service=apache --update_cpu=0.5"
    echo "  opencli user-resources stefan --json --service=mysql --update_cpu=2 --update_ram=1.5"
    echo ""
    exit 1
}


shift

# Parse flags and arguments
for arg in "$@"; do
    case $arg in
        --json)
            json_output=true
            ;;
        --update_cpu=*)
            update_cpu="${arg#--update_cpu=}"
            ;;
        --update_ram=*)
            update_ram="${arg#--update_ram=}"
            ;;
        --activate=*)
            new_service="${arg#--activate=}"
            ;;
        --force*)
            FORCE_PULL=true
            ;;
        --service=*)
            service_to_update_cpu_ram="${arg#--service=}"
            ;;
        --deactivate=*)
            stop_service="${arg#--deactivate=}"
            ;;
        *)
            echo "Error: Invalid argument '$arg'."
            usage
            ;;
    esac
done


check_context_and_env_exist() {
    if [ -z "$context" ]; then
        echo "Error: docker context name must be provided as the first argument!"
        exit 1
    fi
    
    if [ ! -f "$env_file" ]; then
        echo "Error: $env_file file not found!"
        exit 1
    fi
}

# used for both cpu and ram
validate_number() {
    local num="$1"
    
    # matches 123, 45.67, .5, 0.3
    if [[ "$num" =~ ^[0-9]*\.?[0-9]+$ ]]; then
        if awk -v n="$num" 'BEGIN {exit !(n >= 0 && n <= 512)}'; then
            return 0
        fi
    fi

    return 1
}


update_cpu_for_service_or_total() {
    if [[ -n "$update_cpu" ]]; then
        if validate_number "$update_cpu"; then
            if [[ -n "$service_to_update_cpu_ram" ]]; then
                message+="<br> Updating CPU limit for service $service_to_update_cpu_ram to: $update_cpu Core"
                if [[ "$service_to_update_cpu_ram" == "mariadb" ]]; then
                    service_to_update_cpu_ram="mysql"
                elif [[ "$service_to_update_cpu_ram" == php-fpm-* ]]; then
                    service_to_update_cpu_ram="${service_to_update_cpu_ram//./_}"
                fi
                service_to_update_cpu_ram="${service_to_update_cpu_ram//-/_}"
                sed -i 's/^'"${service_to_update_cpu_ram^^}"'_CPU=".*"/'"${service_to_update_cpu_ram^^}"'_CPU="'"$update_cpu"'"/' "$env_file"
            else
                message+="<br> Updating total CPU limit to: $update_cpu Core"
                sed -i 's/^TOTAL_CPU=".*"/TOTAL_CPU="'"$update_cpu"'"/' "$env_file"
            fi
        else
            echo "Error: Invalid CPU value. Must be a number between 0 and 512."
            exit 1
        fi
    fi
}

update_ram_for_service_or_total() {
    if [[ -n "$update_ram" ]]; then
        update_ram="${update_ram//[gG]/}"  # Remove g or G
        if validate_number "$update_ram"; then
            update_ram="${update_ram}g"  # https://i.pinimg.com/736x/35/52/72/355272d3d4ddd508433781ee038d008c.jpg
            if [[ -n "$service_to_update_cpu_ram" ]]; then
                message+="<br> Updating RAM limit for service $service_to_update_cpu_ram to: $update_ram"

                if [[ "$service_to_update_cpu_ram" == "mariadb" ]]; then
                    service_to_update_cpu_ram="mysql"
                elif [[ "$service_to_update_cpu_ram" == php-fpm-* ]]; then
                    service_to_update_cpu_ram="${service_to_update_cpu_ram//./_}"
                fi
                service_to_update_cpu_ram="${service_to_update_cpu_ram//-/_}"

                sed -i 's/^'"${service_to_update_cpu_ram^^}"'_RAM=".*"/'"${service_to_update_cpu_ram^^}"'_RAM="'"$update_ram"'"/' "$env_file"
            else
                message+="<br> Updating total RAM limit to: $update_ram"
                sed -i 's/^TOTAL_RAM=".*"/TOTAL_RAM="'"$update_ram"'"/' "$env_file"
            fi
        else
            echo "Error: Invalid RAM value. Must be a number between 0 and 512."
            exit 1
        fi
    fi
}

load_env_file_now() {
    # Load .env variables now after the update!
    if [ -f $env_file ]; then
        export $(grep -v '^#' $env_file | xargs)
    fi
}



get_total_cpu_and_ram() {
    
    # Ensure TOTAL_CPU and TOTAL_RAM are set
    if [ -z "$TOTAL_CPU" ] || [ -z "$TOTAL_RAM" ]; then
        echo "Error: TOTAL_CPU or TOTAL_RAM not set in $env_file!"
        exit 1
    fi
    
    TOTAL_RAM=$(echo "$TOTAL_RAM" | sed 's/[gG]//g')
    TOTAL_RAM=$(echo "$TOTAL_RAM" | awk '{print int($1)}')
    TOTAL_CPU=$(echo "$TOTAL_CPU" | awk '{print int($1)}')
    
    
    if ! [[ "$TOTAL_CPU" =~ ^[0-9]+$ ]]; then
        echo "Error: TOTAL_CPU is not an integer."
        exit 1
    fi
    
    
    if ! [[ "$TOTAL_RAM" =~ ^[0-9]+$ ]]; then
        echo "Error: TOTAL_RAM is not an integer."
        exit 1
    fi
    
    TOTAL_USED_CPU=0
    TOTAL_USED_RAM=0

}

get_active_services_and_their_usage() { 
    
    RUNNING_SERVICES=$(docker --context $context ps --format "{{.Names}}")
    if [ $? -ne 0 ]; then
        message+="<br> Failed to retrieve the list of running services. Please ensure Docker is installed and the context $context is valid."
        if $json_output; then
            json_data="{\"context\": \"$context\", \"message\": \"$message\"}"
            echo "$json_data" | jq .
            json_print=true
        else
            echo "$message"
            message_print=true
        fi
        exit 1
    fi

    if [ -z "$RUNNING_SERVICES" ]; then
        message+="<br> No services are currently running in context $context"
        if $json_output; then
            :
        else
            echo "$message"
        fi
    fi

    json_data="{\"context\": \"$context\", \"services\": [], \"limits\": {\"cpu\": {\"used\": \"$TOTAL_USED_CPU\", \"total\": \"$TOTAL_CPU\"}, \"ram\": {\"used\": \"$TOTAL_USED_RAM\", \"total\": \"$TOTAL_RAM\"}}, \"message\": \"$message\"}"


    if [ -n "$RUNNING_SERVICES" ]; then
        services_data=""
    
        if $json_output; then
            :
        else
            echo "Services:"
        fi
    
        for service in $RUNNING_SERVICES; do
            # Replace any dots or hyphens with underscores in the service name to match .env variables
            service_name=$(echo "$service" | sed 's/[.-]/_/g')

            if [[ "$service_name" == "phpmyadmin" ]]; then
                service_name="PMA"
            fi
    
            if [[ "$service_name" == "$context" ]]; then
                service_name="OS"
            fi
            
            if [[ "$service_name" == "mariadb" ]]; then
                service_name="mysql"
            fi
            cpu_var="${service_name^^}_CPU"  # Convert service name to uppercase for matching .env variable
            ram_var="${service_name^^}_RAM"
    
            cpu_value=${!cpu_var:-0}
            ram_value=${!ram_var:-0}
    
            if [ -z "${!cpu_var}" ] || [ -z "${!ram_var}" ]; then
                # If either the CPU or RAM value is missing in the .env file, abort
                projected_cpu="0"
                projected_ram="0"
                message+="<br> Error: Service $service_name does not have CPU or RAM limits defined in .env file!"
                display_json_or_message
                exit 1 
            fi
    
    
            # Strip "G" from RAM values
            ram_value=${ram_value//G/}
            
            TOTAL_USED_CPU=$(awk "BEGIN {print $TOTAL_USED_CPU + $cpu_value}")
            TOTAL_USED_RAM=$(awk "BEGIN {print $TOTAL_USED_RAM + $ram_value}")

    
            if [[ "$service_name" == "OS" ]]; then
                service_name=$context
            fi
   
            # Convert service name to display format (underscores to hyphens and numbers with dots)
            display_service_name=$(echo "$service_name" | sed 's/_/-/g' | sed -E 's/([0-9]+)-([0-9]+)/\1.\2/g')
    
            service_data="{\"name\": \"$display_service_name\", \"cpu\": \"$cpu_value\", \"ram\": \"$ram_value\"}"
            
            if $json_output; then
                services_data="$services_data$service_data,"
            else
                echo "- $display_service_name - CPU: $cpu_value cores, RAM: $ram_value G"
            fi
        done
    
        # Remove last comma for valid JSON
        services_data=$(echo "$services_data" | sed 's/,$//')
    
        # Add services to the JSON structure
        if $json_output; then
            json_data="{\"context\": \"$context\", \"services\": [$services_data], \"limits\": {\"cpu\": {\"used\": \"$TOTAL_USED_CPU\", \"total\": \"$TOTAL_CPU\"}, \"ram\": {\"used\": \"$TOTAL_USED_RAM\", \"total\": \"$TOTAL_RAM\"}}, \"message\": \"$message\"}"
        else
            echo ""
            echo "Total usage:"
            echo "- CPU: $TOTAL_USED_CPU / $TOTAL_CPU"
            echo "- RAM: $TOTAL_USED_RAM / $TOTAL_RAM"
            echo ""
        fi
    
    else
        message+="<br> No currently running services."
        echo "$message"
        message_print=true
    fi
}



check_if_service_exists_or_running() {
    service_name="$1"
    action="$2"
    sanitized_service_name=${service_name//_/-}
    service_exists=$(docker --context "$context" compose -f "/home/$context/docker-compose.yml" config --services | grep -w "$sanitized_service_name")

    if [ -z "$service_exists" ]; then
        echo "Service $sanitized_service_name does not exist in the Docker Compose file."
        return 1  # Service doesn't exist
    fi

    # Action "check" – just verify if the service exists
    if [ "$action" == "check" ]; then
        return 0
    fi

    # Action "status" – check if the service is running
    if [ "$action" == "status" ]; then
        running_service=$(docker --context $context ps --filter "name=$sanitized_service_name" --format '{{.Names}}')

        if [ "$running_service" == "$sanitized_service_name" ]; then
            return 0  # Service is running
        else
            return 2  # Service exists but is not running
        fi
    
    fi
}


start_service_now() {
    service_name="$1"

    if $FORCE_PULL; then
        docker --context $context compose -f /home/$context/docker-compose.yml pull $service_name > /dev/null 2>&1  
    fi
    
    docker --context $context compose -f /home/$context/docker-compose.yml up -d $service_name > /dev/null 2>&1   
}

stop_service_now() {
    service_name="$1"
    docker --context $context compose -f /home/$context/docker-compose.yml down $service_name > /dev/null 2>&1   
}

display_json_or_message() {
        if $json_output; then
            # Remove last comma for valid JSON
            services_data=$(echo "$services_data" | sed 's/,$//')
        
            # Build CPU and RAM JSON parts conditionally
            cpu_json="\"used\": $TOTAL_USED_CPU, \"total\": $TOTAL_CPU"
            if [[ "$projected_cpu" =~ ^[0-9]+(\.[0-9]+)?$ ]]; then
                cpu_json+=", \"after\": $projected_cpu"
            fi
        
            ram_json="\"used\": $TOTAL_USED_RAM, \"total\": $TOTAL_RAM"
            if [[ "$projected_ram" =~ ^[0-9]+(\.[0-9]+)?$ ]]; then
                ram_json+=", \"after\": $projected_ram"
            fi
        
            # Final JSON construction
            json_data="{\"context\": \"$context\", \"services\": [$services_data], \"limits\": {\"cpu\": {${cpu_json}}, \"ram\": {${ram_json}}}, \"message\": \"$message\"}"
        
            echo "$json_data" | jq .
            json_print=true
        else
            echo "$message"
            message_print=true
        fi
}    

add_new_service() {
    # Handle new service addition if --activate=<service_name> is provided
    if [[ -n "$new_service" ]]; then
        # Replace dots and hyphens with underscores in the new service name
        new_service_name=$(echo "$new_service" | sed 's/[.-]/_/g')

        check_if_service_exists_or_running "$new_service" "check"

        if [ $? -eq 1 ]; then
            echo "Error starting: service $new_service_name does not exist in docker-compose.yml file. Contact the administrator."
        fi

        new_cpu_var="${new_service_name^^}_CPU"
        new_ram_var="${new_service_name^^}_RAM"

        new_cpu_value=${!new_cpu_var:-0}
        new_ram_value=${!new_ram_var:-0}
        new_ram_value=${new_ram_value//G/}

        # Check if the CPU value is a valid float or integer
        if ! [[ "$new_cpu_value" =~ ^[0-9]+(\.[0-9]+)?$ ]]; then
            message+="<br> Error: Service $new_service_name does not have a valid CPU limit defined!" 
        # Check if CPU is <= 0 or RAM is not defined or <= 0
        elif ! awk -v n="$new_cpu_value" 'BEGIN {exit (n > 0 ? 0 : 1)}' || ! awk -v r="$new_ram_value" 'BEGIN {exit (r > 0 ? 0 : 1)}'; then
            message+="<br>Error activating: service $new_service_name does not have CPU or RAM limits defined!"
        fi

        projected_cpu=$(awk "BEGIN {print $TOTAL_USED_CPU + $new_cpu_value}")
        if awk -v cpu="$projected_cpu" -v total="$TOTAL_CPU" 'BEGIN {exit !(cpu > total)}'; then
            if [ "$TOTAL_CPU" -eq 0 ]; then
                message+="<br>Warning: User has unlimited CPU limits: $projected_cpu / $TOTAL_CPU cpus"
            else
                message+="<br> Error: Adding $new_service would exceed CPU limits: $projected_cpu / $TOTAL_CPU cpus"
            fi
        fi

        projected_ram=$(awk "BEGIN {print $TOTAL_USED_RAM + $new_ram_value}")
        if awk -v ram="$projected_ram" -v total="$TOTAL_RAM" 'BEGIN {exit !(ram > total)}'; then
            if [ "$TOTAL_RAM" -eq 0 ]; then
                message+="<br> Warning: User has unlimited RAM limits: $projected_ram G / $TOTAL_RAM G"
            else
                message+="<br> Error: Adding $new_service would exceed RAM limits: $projected_ram G / $TOTAL_RAM G"
            fi
        fi

        if [ "$TOTAL_RAM" -eq 0 ] || [ "$TOTAL_CPU" -eq 0 ]; then
            # START SERVICE
            start_service_now "$new_service"
            # CHECK IF RUNNING
            check_if_service_exists_or_running "$service_name" "status"
            status=$?
            if [ $status -eq 0 ]; then
                message+="<br>Service $service_name started successfully."
                opencli docker-collect_stats "$USERNAME" >/dev/null 2>&1 &
            elif [ $status -eq 2 ]; then
                message+="<br>Service $service_name did not start. Contact Administrator."
                display_json_or_message
                exit 1
            fi
            display_json_or_message
            exit 1
        
        elif awk -v cpu="$projected_cpu" -v total_cpu="$TOTAL_CPU" -v ram="$projected_ram" -v total_ram="$TOTAL_RAM" \
            'BEGIN {exit !(cpu > total_cpu || ram > total_ram)}'; then
            display_json_or_message
            exit 1
        else
            # START SERVICE
            start_service_now "$new_service"
            # CHECK IF RUNNING
            check_if_service_exists_or_running "$service_name" "status"
            status=$?
            if [ $status -eq 0 ]; then
                message+="<br>Service $service_name started successfully."
                opencli docker-collect_stats "$USERNAME" >/dev/null 2>&1 &
            elif [ $status -eq 2 ]; then
                message+="<br>Service $service_name did not start. Contact Administrator."
                display_json_or_message
                exit 1
            fi
            display_json_or_message
            exit 1
        fi
    fi
}



stop_container() {
    # if --deactivate=<service_name> is provided
    if [[ -n "$stop_service" ]]; then        

        check_if_service_exists_or_running "$stop_service" "check"
        
        if [ $? -eq 1 ]; then
            echo "Error stopping: service $stop_service does not exist in docker-compose.yml file. Contact the administrator."
        fi
        
        # STOP SERVICE
        stop_service_now "$stop_service"
        
        # CHECK IF RUNNING
        check_if_service_exists_or_running "$stop_service" "status"
        status=$?
        if [ $status -eq 0 ]; then
            message+="<br>Service $stop_service did not stop. Contact Administrator."
            display_json_or_message
            exit 1
        elif [ $status -eq 2 ]; then
            message+="<br>Service $stop_service stopped successfully."
            opencli docker-collect_stats "$USERNAME" >/dev/null 2>&1 &
        fi

        display_json_or_message
    fi
}


final_output_for_json() {
    if [ "$json_output" = true ] && [ "$json_print" != true ]; then # skip if alreay printed json!
        echo "$json_data" | jq .
    fi
}

# MAIN
check_context_and_env_exist                   # first checks
update_cpu_for_service_or_total               # set TOTAL_CPU or $SERVICE_CPU
update_ram_for_service_or_total               # set TOTAL_RAM or $SERVICE_RAM
load_env_file_now                             # load the data from .env file after (if) we did updates
get_total_cpu_and_ram                         # get total cpu/ram usage allocated to the user
get_active_services_and_their_usage           # get combined cpu/ram usage for all active services
stop_container                                # stop container
add_new_service                               # check if starting new service is within user limits and start it

final_output_for_json                         # pretty print the data

exit 0

