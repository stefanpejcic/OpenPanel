#!/usr/bin/env bash
LS_FD='/usr/local/lsws'
PHP_VER='lsphp83'

check_php_input(){
    if [ -z "${1}" ]; then
        echo "Use default value ${PHP_VER}"
    else
        echo ${1} | grep lsphp >/dev/null
        if [ ${?} = 0 ]; then
            PHP_VER=${1}
        fi
    fi
}

update_listener(){
    sed -i '/<listenerList>/a\
    <listener> \
      <name>HTTP</name> \
      <address>*:80</address> \
      <secure>0</secure> \
    </listener> \
    <listener> \
      <name>HTTPS</name> \
      <address>*:443</address> \
      <reusePort>1</reusePort> \
      <secure>1</secure> \
      <keyFile>/usr/local/lsws/admin/conf/webadmin.key</keyFile> \
      <certFile>/usr/local/lsws/admin/conf/webadmin.crt</certFile> \
    </listener>
' ${LS_FD}/conf/httpd_config.xml
}

update_template(){
    sed -i '/<vhTemplateList>/a\
    <vhTemplate> \
      <name>docker</name> \
      <templateFile>$SERVER_ROOT/conf/templates/docker.xml</templateFile> \
      <listeners>HTTP, HTTPS</listeners> \
      <member> \
        <vhName>localhost</vhName> \
        <vhDomain>*, localhost</vhDomain> \
      </member> \
    </vhTemplate>
' ${LS_FD}/conf/httpd_config.xml
}

php_path(){
    if [ -f ${LS_FD}/conf/templates/docker.xml ]; then
        sed -i "s/lsphpver/${1}/" ${LS_FD}/conf/templates/docker.xml
    else
        echo 'docker.xml template not found!'
        exit 1
    fi
}

create_doc_fd(){
    mkdir -p /var/www/vhosts/localhost/{html,logs,certs}
    chown 1000:1000 /var/www/vhosts/localhost/ -R
}

main(){
    check_php_input ${1}
    php_path ${PHP_VER}
    update_listener
    update_template
    create_doc_fd
}
main ${1}
