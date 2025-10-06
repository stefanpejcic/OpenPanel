#!/bin/bash
################################################################################
# Script Name: email/server.sh
# Description: Manage mailserver
# Usage: opencli email-server <install|start|restart|stop|uninstall> [--debug]
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 18.08.2024
# Last Modified: 05.10.2025
# Company: openpanel.co
# Copyright (c) openpanel.co
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

# CONFIG
APP="opencli email-server"                             # this script
GITHUB_REPO="https://github.com/stefanpejcic/openmail" # download files
DIR="/usr/local/mail/openmail"                         # compose.yaml directory
CONTAINER=openadmin_mailserver                         # DMS container name
TIMEOUT=3600                                           # for graceful stop
DOCKER_COMPOSE="docker compose"                        # compose plugin

set -ueo pipefail

    ensure_jq_installed() {
        # Check if jq is installed
        if ! command -v jq &> /dev/null; then
            # Detect the package manager and install jq
            if command -v apt-get &> /dev/null; then
                sudo apt-get update > /dev/null 2>&1
                sudo apt-get install -y -qq jq > /dev/null 2>&1
            elif command -v yum &> /dev/null; then
                sudo yum install -y -q jq > /dev/null 2>&1
            elif command -v dnf &> /dev/null; then
                sudo dnf install -y -q jq > /dev/null 2>&1
            else
                echo "Error: No compatible package manager found. Please install jq manually and try again."
                exit 1
            fi
    
            # Check if installation was successful
            if ! command -v jq &> /dev/null; then
                echo "Error: jq installation failed. Please install jq manually and try again."
                exit 1
            fi
        fi
    }

ensure_jq_installed

_checkBin() {
	local cmd
	for cmd in "$@"; do
		hash "$cmd" 2>/dev/null || {
			echo "Error: '$cmd' not found."
			echo
			exit 1
		} >&2
	done

	# docker compose
	$DOCKER_COMPOSE version &>/dev/null || {
		echo "Error: '$DOCKER_COMPOSE' not available."
		echo
		exit 1
	} >&2
}

# Dependencies
_checkBin "cat" "cut" "docker" "fold" "jq" "printf" "sed" "tail" "tput" "tr"


DEBUG=false  # Default to false

# Check if --debug flag is provided
for arg in "$@"; do
    if [ "$arg" == "--debug" ]; then
        DEBUG=true
        echo "--debug flag provided: displaying verbose information."
    elif [ "$arg" == "-x" ]; then
        DEBUG=true
	set -x
	echo "-x flag provided: display each command as it execute."
    fi
done









# Check if container is running
if [ -n "${1:-}" ] && [ "${1:-}" != "install" ] && [ "${1:-}" != "status" ] && [ "${1:-}" != "start" ] && [ "${1:-}" != "stop" ] && [ "${1:-}" != "restart" ]; then
	if [ -z "$(docker ps -q --filter "name=^$CONTAINER$")" ]; then
		echo -e "Error: Container '$CONTAINER' is not up.\n" >&2
		exit 1
	fi
fi




# Print status
_status() {
	# $1	name
	# $2	status
	local indent spaces status
	indent=14

	# Wrap long lines and prepend spaces to multi line status
	spaces=$(printf "%${indent}s")
	status=$(echo -n "$2" | fold -s -w $(($(tput cols) - 16)) | sed "s/^/$spaces/g")
	status=${status:$indent}

	printf "%-${indent}s%s\n" "$1:" "$status"
}

_ports() {
	docker port "$CONTAINER"
}

_container() {
	if [ "$1" == "-it" ]; then
		shift
		docker exec -it "$CONTAINER" "$@"
	else
		docker exec "$CONTAINER" "$@"
	fi
}

_getDMSVersion() {
	# shellcheck disable=SC2016
	# todo 'cat /VERSION' is kept for compatibility with DMS versions < v13.0.1; remove in the future
	_container bash -c 'cat /VERSION 2>/dev/null || printf "%s" "$DMS_RELEASE"'
}



# ENTERPRISE
ENTERPRISE="/usr/local/opencli/enterprise.sh"
PANEL_CONFIG_FILE="/etc/openpanel/openpanel/conf/openpanel.config"
key_value=$(grep "^key=" $PANEL_CONFIG_FILE | cut -d'=' -f2-)


# Check if 'enterprise edition'
if [ -n "$key_value" ]; then
    :
else
    echo "Error: OpenPanel Community edition does not support emails. Please consider purchasing the Enterprise version that allows unlimited number of email addresses."
    source $ENTERPRISE
    echo "$ENTERPRISE_LINK"
    exit 1
fi



# ENABLE EMAILS MODULE AND EMAILS PAGES
	enable_emails_if_not_yet() {
	  if [ "$DEBUG" = true ]; then
	      echo ""
	      echo "----------------- CHECKING ENABLED MODULES ------------------"
	      echo ""
	  fi
	    config_file="/etc/openpanel/openpanel/conf/openpanel.config"
	    enabled_modules=$(grep '^enabled_modules=' "$config_file" | cut -d'=' -f2)
    	    if echo "$enabled_modules" | grep -q 'emails'; then
	        echo "'emails' module is already in enabled modules."
	        :
	    else
	        new_modules="${enabled_modules},emails"
	 	echo "'emails' module is not enabled. Enabling.."
	        sed -i "s/^enabled_modules=.*/enabled_modules=${new_modules}/" "$config_file"
	        echo "Restarting OpenPanel container to enable email pages.."
		if [ "$(docker ps -q -f name=openpanel)" ]; then
		    docker restart openpanel  >/dev/null 2>&1
		else
		    cd /root && docker --context default compose up -d openpanel  >/dev/null 2>&1
		fi  
	    fi
	}



# SUMMARY LOGS 
pflogsumm_get_data() {
	cd /tmp
	rm -rf PFLogSumm-HTML-GUI
	git clone https://github.com/stefanpejcic/PFLogSumm-HTML-GUI.git  > /dev/null 2>&1

 	docker cp PFLogSumm-HTML-GUI/pflogsummUIReport.sh openadmin_mailserver:/opt/pflogsummUIReport.sh   > /dev/null 2>&1
	
	echo "Generating email statistics reports.. This can take a while."
	
	docker exec openadmin_mailserver sh -c "bash /opt/pflogsummUIReport.sh"
	
	echo "Done, adding reports to OpenAdmin interface"
	mkdir -p /usr/local/admin/static/reports /usr/local/admin/templates/emails > /dev/null 2>&1
	docker cp openadmin_mailserver:/usr/local/admin/static/reports/reports.html /usr/local/admin/templates/emails/reports.html > /dev/null 2>&1
	docker cp openadmin_mailserver:/usr/local/admin/static/reports/data /usr/local/admin/templates/emails/ > /dev/null 2>&1
 
	#echo "Reloading admin panel.."
	#service admin restart   > /dev/null 2>&1
	echo "Completed"
 	rm -rf PFLogSumm-HTML-GUI
}



set_ssl_for_mailserver() {
	readonly MAILSERVER_ENV="/usr/local/mail/openmail/mailserver.env"	
	current_hostname=$(opencli domain)
	
	if [[ $current_hostname =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
	    # an IP
	    echo "Configuring mailserver to use IP address for IMAP/SMTP ..."
	    sed -i '/^SSL_TYPE=/c\SSL_TYPE=' "$MAILSERVER_ENV"
	    sed -i '/^SSL_CERT_PATH=/d' "$MAILSERVER_ENV"
	    sed -i '/^SSL_KEY_PATH=/d' "$MAILSERVER_ENV"
	else
	    # a letsencrypt
        local cert_path="/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/${current_hostname}/${current_hostname}.crt"
        local key_path="/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/${current_hostname}/${current_hostname}.key"

        # custom
        local fallback_cert_path="/etc/openpanel/caddy/ssl/custom/${current_hostname}/${current_hostname}.crt"
        local fallback_key_path="/etc/openpanel/caddy/ssl/custom/${current_hostname}/${current_hostname}.key"
        
        if [[ -f "$cert_path" && -f "$key_path" ]]; then
            echo "Using Let's Encrypt certs for $current_hostname"
        elif [[ -f "$fallback_cert_path" && -f "$fallback_key_path" ]]; then
            echo "Using custom certs for $current_hostname"
            cert_path="$fallback_cert_path"
            key_path="$fallback_key_path"
        else
		    echo "Warning: Domain $current_hostname is configured for panel access but has no SSL, it will not be used for mailserver IMAP/SMTP ..."
		    [[ ! -f "$cert_path" ]] && echo "- Missing: $cert_path"
		    [[ ! -f "$key_path" ]] && echo "- Missing: $key_path"
            return 0
        fi
		
		echo "Configuring mailserver to use domain $current_hostname for IMAP/SMTP ..."

		sed -i '/^SSL_TYPE=/c\SSL_TYPE=manual' "$MAILSERVER_ENV"

		grep -q '^SSL_CERT_PATH=' "$MAILSERVER_ENV" \
			&& sed -i "s|^SSL_CERT_PATH=.*|SSL_CERT_PATH=$cert_path|" "$MAILSERVER_ENV" \
			|| echo "SSL_CERT_PATH=$cert_path" >> "$MAILSERVER_ENV"

		grep -q '^SSL_KEY_PATH=' "$MAILSERVER_ENV" \
			&& sed -i "s|^SSL_KEY_PATH=.*|SSL_KEY_PATH=$key_path|" "$MAILSERVER_ENV" \
			|| echo "SSL_KEY_PATH=$key_path" >> "$MAILSERVER_ENV"

	fi
 }




# INSTALL
install_mailserver(){
  if [ "$DEBUG" = true ]; then
      echo ""
      echo "----------------- INSTALLING MAILSERVER ------------------"
      echo ""
      echo "Downloading from $GITHUB_REPO"
      echo ""
      mkdir -p /usr/local/mail/
      cd /usr/local/mail/ && git clone $GITHUB_REPO
      set_ssl_for_mailserver
      mkdir -p /etc/openpanel/email/snappymail
      cd /usr/local/mail/openmail && docker --context default compose up -d mailserver roundcube
  else
      mkdir -p /usr/local/mail/  >/dev/null 2>&1
      cd /usr/local/mail/ && git clone $GITHUB_REPO >/dev/null 2>&1
      set_ssl_for_mailserver
      mkdir -p /etc/openpanel/email/snappymail >/dev/null 2>&1
      cd /usr/local/mail/openmail && docker --context default compose up -d mailserver roundcube >/dev/null 2>&1
  fi

  enable_emails_if_not_yet
  configure_csf_ports

  if [ "$DEBUG" = true ]; then
      echo ""
      echo "----------------- CONFIGURING FIREWALL ------------------"
      echo ""
  fi
  
  function open_port_csf() {
      local port=$1
      local csf_conf="/etc/csf/csf.conf"
      
      # Check if port is already open
      port_opened=$(grep "TCP_IN = .*${port}" "$csf_conf")
      if [ -z "$port_opened" ]; then
          # Open port
          sed -i "s/TCP_IN = \"\(.*\)\"/TCP_IN = \"\1,${port}\"/" "$csf_conf"
          echo "Port ${port} opened in CSF."
          ports_opened=1
      else
          echo "Port ${port} is already open in CSF."
      fi
  }
  
  
  if [ "$DEBUG" = true ]; then
      echo ""
      echo "----------------- CONFIGURING FIREWALL ------------------"
      echo ""
   # CSF
    if command -v csf >/dev/null 2>&1; then
        open_port_csf 25
        open_port_csf 143
        open_port_csf 465
        open_port_csf 587
        open_port_csf 993
    else
        echo "Error: CSF is not installed. make sure ports 25 243 465 587 and 993 are opened on external firewall, or email will not work."
    fi
  else

    # CSF
    if command -v csf >/dev/null 2>&1; then
        open_port_csf 25 >/dev/null 2>&1
        open_port_csf 143 >/dev/null 2>&1
        open_port_csf 465 >/dev/null 2>&1
        open_port_csf 587 >/dev/null 2>&1
        open_port_csf 993 >/dev/null 2>&1
    else
        echo "Error: CSF is not installed. make sure ports 25 243 465 587 and 993 are opened on external firewall, or email will not work."
    fi

fi

#########

  if [ "$DEBUG" = true ]; then
      echo ""
      echo "----------------- ENABLE MAIL FOR EXISTING USERS ------------------"
      echo ""
  fi
  
  user_list=$(opencli user-list --json)
  

# at end lets add all domains
process_all_domains_and_start
  
}

# START
process_all_domains_and_start(){
  CONFIG_DIR="/etc/openpanel/caddy/domains"
  COMPOSE_FILE="/usr/local/mail/openmail/compose.yml"
  new_volumes="    volumes:\n      - ./docker-data/dms/mail-state/:/var/mail-state/\n      - ./var/log/mail/:/var/log/mail/\n      - ./docker-data/dms/config/:/tmp/docker-mailserver/\n      - /etc/localtime:/etc/localtime:ro\n"

  cp "$COMPOSE_FILE" "$COMPOSE_FILE.bak"
  
  if [ "$DEBUG" = true ]; then
      echo ""
      echo "----------------- MOUNT USERS HOME DIRECTORIES ------------------"
      echo ""
      echo "Re-mounting mail directories for all domains:"
      echo ""
      echo "- DOMAINS DIRECTORY: $CONFIG_DIR" 
      echo "- MAIL SETTINGS FILE: $COMPOSE_FILE"
      printf "%b" "- DEFAULT VOLUMES:\n$new_volumes"
  fi
    
 
echo "Processing domains in directory: $CONFIG_DIR"
for file in "$CONFIG_DIR"/*.conf; do
    echo "Processing file: $file"
    if [ ! -L "$file" ]; then
        # Extract the username and domain from the file name
        BASENAME=$(basename "$file" .conf)
	whoowns_output=$(opencli domains-whoowns "$BASENAME")
	owner=$(echo "$whoowns_output" | awk -F "Owner of '$BASENAME': " '{print $2}')
	if [ -n "$owner" ]; then
	        echo "Domain $BASENAME skipped. No user."
   	else
	        USERNAME=$owner
	        DOMAIN=$BASENAME
	 
	        DOMAIN_DIR="/home/$USERNAME/mail/$DOMAIN/"
	        new_volumes+="      - $DOMAIN_DIR:/var/mail/$DOMAIN/\n"	
	        echo "Mount point added: - $DOMAIN_DIR:/var/mail/$DOMAIN/"
    	fi

    fi
    
done

if [ $? -ne 0 ]; then
    echo "Error encountered while processing $file"
fi


  
  if [ "$DEBUG" = true ]; then
      echo ""
      echo "----------------- EMAIL DIRECTORIES ------------------"
      echo ""
      printf "%b" "- DEFAULT VOLUMES + VOLUMES PER DOMAIN:\n$new_volumes"
      echo ""
      echo "----------------- UPDATE COMPOSE ------------------"
      echo ""
  fi
  
  
  
  awk -v new_volumes="$new_volumes" '
  BEGIN { in_mailserver=0; }
  /^  mailserver:/ { in_mailserver=1; print; next; }
  /^  [a-z]/ { in_mailserver=0; }  # End of mailserver section if a new service starts
  {
      if (in_mailserver) {
          if ($1 == "volumes:") {
              print new_volumes
              while (getline > 0) {
                  if (/^[ ]{6}-[ ]+\/home/) {
                      continue
                  }
                  if (!/^[ ]{6}-/) {
                      print $0
                      break
                  }
              }
              in_mailserver=0
          } else {
              print
          }
      } else {
          print
      }
  }
  ' "$COMPOSE_FILE.bak" > "$COMPOSE_FILE"
  
  
  if [ "$DEBUG" = true ]; then
  	echo "compose.yml has been updated with the new volumes."
      echo ""
      echo "----------------- RESTART MAILSERVER ------------------"
      echo ""
  	cd $DIR && docker --context default compose  up -d mailserver
      echo ""
  else
  	cd $DIR && docker --context default compose up -d mailserver >/dev/null 2>&1
  	echo "MailServer started successfully."
  fi
  

}

# STOP
stop_mailserver_if_running(){

  if [ "$DEBUG" = true ]; then
      echo ""
      echo "----------------- STOP MAILSERVER ------------------"
      echo ""
  	  cd $DIR && docker --context default compose down mailserver
      echo ""
  else
  	cd $DIR && docker --context default compose down mailserver >/dev/null 2>&1
  	echo "MailServer stopped succesfully."
  fi
  
}


open_port_csf() {
    local port=$1
    local conf="/etc/csf/csf.conf"
    grep -q "TCP_IN = .*${port}" "$conf" || sed -i "s/TCP_IN = \"/TCP_IN = \",${port}/" "$conf"
}

configure_csf_ports() {
    if command -v csf &>/dev/null; then
        for p in 25 143 465 587 993; do open_port_csf "$p"; done
    else
        echo "Warning: CSF not installed. Ensure email ports are open externally."
    fi
}

# UNINSTALL
remove_mailserver_and_all_config(){
  if [ "$DEBUG" = true ]; then
      echo ""
      echo "----------------- UNINSTALL MAILSERVER ------------------"
      echo ""
  fi

  echo "Are you sure you want to uninstall the MailServer and remove all its configuration? (yes/no)"
  read -t 10 -n 1 user_input

  if [ $? -ne 0 ]; then
    echo ""
    echo "No response received. Aborting uninstallation."
    return
  fi

  if [[ "$user_input" != "y" && "$user_input" != "Y" && "$user_input" != "yes" ]]; then
    echo ""
    echo "Uninstallation aborted."
    return
  fi

  if [ "$DEBUG" = true ]; then
      cd $DIR && docker --context default compose down
      rm -rf $DIR
      echo ""
  else
      cd $DIR && docker --context default compose down >/dev/null 2>&1
      rm -rf $DIR >/dev/null 2>&1
      echo "MailServer uninstalled successfully."
  fi
}






case "${1:-}" in
	install) # install mailserver
        echo "Installing the mailserver..."
        install_mailserver
	echo "Enabling Roundcube..."
	opencli email-webmail roundcube
		;;
	pflogsumm) # generate reports
        echo "Generating reports..."
        pflogsumm_get_data
		;;
	status) # Show status
		if [ -n "$(docker ps -q --filter "name=^$CONTAINER$")" ]; then
			# Container uptime
			_status "Container" "$(docker ps --no-trunc --filter "name=^$CONTAINER$" --format "{{.Status}}")"
			echo

			# Version
			_status "Version" "$(_getDMSVersion)"
			echo

			# Fail2ban
			_container ls /var/run/fail2ban/fail2ban.sock &>/dev/null &&
			_status "Fail2ban" "$(_container fail2ban)"
			echo

			# Package updates available?
			_status "Packages" "$(_container bash -c 'apt -q update 2>/dev/null | grep "All packages are up to date" || echo "Updates available"')"
			echo

			# Published ports
			# _status "Ports" "$(docker inspect "$CONTAINER" | jq -r '.[].NetworkSettings.Ports | .[] | select(. != null) | tostring' | cut -d'"' -f8 | tr "\n" " ")"
			_status "Ports" "$(_ports)"
			echo

			# Postfix mail queue
			POSTFIX=$(_container postqueue -p | tail -1 | cut -d' ' -f5)
			[ -z "$POSTFIX" ] && POSTFIX="Mail queue is empty" || POSTFIX+=" mail(s) queued"
			_status "Postfix" "$POSTFIX"
			echo

			# Service status
			_status "Supervisor" "$(_container supervisorctl status | sort -b -k2,2)"
		else
			echo "Container: down"
		fi
		;;

	config)	# show configuration
		_container cat /etc/dms-settings
		;;

	start)	# Start container
        echo "Starting mailserver..."
        process_all_domains_and_start
		;;

	stop)	# Stop container
        echo "Stopping the mailserver..."
        stop_mailserver_if_running
		;;

	restart)	#  Restart container
        echo "Restarting the mailserver..."
        stop_mailserver_if_running
        process_all_domains_and_start
		;;
  
	uninstall)	#  Uninstall container
        echo "Uninstalling the mailsserver..."
        remove_mailserver_and_all_config
		;;
  
	queue)	# Show mail queue
		_container postqueue -p
		;;

	flush)	# Flush mail queue
		_container postqueue -f
		echo "Queue flushed."
		;;

	unhold)	# Release mail that was put "on hold"
		if [ -z "${2:-}" ]; then
			echo "Error: Queue ID missing"
		else
			shift
			for i in "$@"; do
				ARG+=("-H" "$i")
			done
			_container postsuper "${ARG[@]}"
		fi
		;;

	view)	# Show mail by queue id
		if [ -z "${2:-}" ]; then
			echo "Error: Queue ID missing."
		else
			_container postcat -q "$2"
		fi >&2
		;;

	delete) # Delete mail from queue
		if [ -z "${2:-}" ]; then
			echo "Error: Queue ID missing."
		else
			shift
			for i in "$@"; do
				ARG+=("-d" "$i")
			done
			_container postsuper "${ARG[@]}"
		fi
		;;

	fail*)	# Interact with fail2ban
		shift
		_container fail2ban "$@"
		;;

	ports)	# Show published ports
		echo "Published ports:"
		echo
		_ports
		;;

	postc*)	# Show postfix configuration
		shift
		_container postconf "$@"
		;;

	logs)	# Show logs
		if [ "${2:-}" == "-f" ]; then
			docker logs -f "$CONTAINER"
		else
			docker logs "$CONTAINER"
		fi
		;;

	login)	# Run container shell
		_container -it bash
		;;

	super*) # Interact with supervisorctl
		shift
		_container -it supervisorctl "$@"
		;;

	update-c*) # Check for container package updates
		_container -it bash -c 'apt update && echo && apt list --upgradable'
		;;

	update-p*) # Update container packages
		_container -it bash -c 'apt update && echo && apt-get upgrade'
		;;

	version*) # Show versions
		printf "%-15s%s\n\n" "Mailserver:" "$(_getDMSVersion)"
		PACKAGES=("amavisd-new" "clamav" "dovecot-core" "fail2ban" "fetchmail" "getmail6" "rspamd" "opendkim" "opendmarc" "postfix" "spamassassin" "supervisor")
		for i in "${PACKAGES[@]}"; do
			printf "%-15s" "$i:"
			_container bash -c "set -o pipefail; dpkg -s $i 2>/dev/null | grep ^Version | cut -d' ' -f2 || echo 'Package not installed.'"
		done
		;;

	*)
		cat <<-EOF
		Usage:

		$APP status                           Show status
		$APP config                           Show configuration
		$APP install                          Install the email server  
		$APP start                            Start the email server
		$APP stop                             Stop the email server
		$APP restart                          Restart the email server
		$APP queue                            Show mail queue
		$APP flush                            Flush mail queue
		$APP view   <queue id>                Show mail by queue id
		$APP unhold <queue id> [<queue id>]   Release mail that was put "on hold" (marked with '!')
		$APP unhold ALL                       Release all mails that were put "on hold" (marked with '!')
		$APP delete <queue id> [<queue id>]   Delete mail from queue
		$APP delete ALL                       Delete all mails from queue
		$APP fail2ban [<ban|unban> <IP>]      Interact with fail2ban
		$APP fail2ban log                     Show fail2ban log
		$APP ports                            Show published ports
		$APP postconf                         Show postfix configuration
		$APP logs [-f]                        Show logs. Use -f to 'follow' the logs
		$APP login                            Run container shell
		$APP supervisor                       Interact with supervisorctl
		$APP pflogsumm                        Generate email summary reports.
		$APP update-check                     Check for container package updates
		$APP update-packages                  Update container packages
		$APP versions                         Show versions
		EOF
		;;
esac
echo
