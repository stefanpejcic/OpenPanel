import clsx from "clsx";
import React, { FC, memo, useRef } from "react";
import { useColorMode } from "@docusaurus/theme-common";
import BrowserOnly from "@docusaurus/BrowserOnly";
import Highlight, { defaultProps } from "prism-react-renderer";
import nightOwlDark from "prism-react-renderer/themes/nightOwl";
import nightOwlLight from "prism-react-renderer/themes/nightOwlLight";
import { LandingSectionCtaButton } from "./landing-section-cta-button";
import { useInView } from "framer-motion";
type Props = {
    className?: string;
    title?: string;
    description?: string;
    cta?: boolean;
};

export const LandingPureReactCode: FC<Props> = ({
    className,
    title = "100% Control",
    description = `Don’t settle for overpriced and outdated hosting panels. With OpenPanel, you maintain 100% control over your server and data at all times.`,
    cta = true,
}) => {
    return (
        <div className={clsx(className)}>
            <div
                className={clsx(
                    "not-prose",
                    "h-full",
                    "flex-shrink-0",
                    "p-2 landing-sm:p-4",
                    "rounded-2xl landing-sm:rounded-3xl",
                    "dark:bg-landing-noise",
                    "dark:bg-gray-800 bg-gray-50",
                )}
            >
                <div
                    className={clsx(
                        "relative",
                        "flex",
                        "flex-col",
                        "rounded-lg",
                        "landing-sm:aspect-[560/240] landing-md:aspect-[624/240]  landing-lg:aspect-[607/299]",
                        "dark:bg-landing-component-dark bg-landing-component",
                        "border-t-solid border-t",
                        "border-t-gray-200 dark:border-t-gray-700",
                        "border-opacity-60 dark:border-opacity-60",
                        "shadow-[0px_-1.5px_0px_rgba(237,242,247,0.5)] dark:shadow-[0px_-1.5px_0px_rgba(20,20,31,0.5)]",
                        "drop-shadow-sm",
                    )}
                >
                    <BrowserOnly>{() => <CodeSlide />}</BrowserOnly>
                </div>
                <div
                    className={clsx(
                        "not-prose",
                        "mt-4 landing-sm:mt-6 landing-lg:mt-10",
                        "px-4 landing-sm:px-6",
                    )}
                >
                    <h6
                        className={clsx(
                            "p-0",
                            "font-semibold",
                            "text-base landing-sm:text-2xl",
                            "dark:text-gray-300 text-gray-900",
                        )}
                    >
                        {title}
                    </h6>
                    <div
                        className={clsx(
                            "not-prose",
                            "flex",
                            "items-center",
                            "justify-between",
                            "flex-wrap",
                            "gap-4 landing-sm:gap-8",
                            "mb-4 landing-sm:mb-6",
                        )}
                    >
                        <p
                            className={clsx(
                                "p-0",
                                "mt-2 landing-sm:mt-4",
                                "text-base",
                                "dark:text-gray-400 text-gray-600",
                            )}
                        >
                            {description}
                        </p>
                        {cta && (
                            <LandingSectionCtaButton to="/docs">
                                Documentation
                            </LandingSectionCtaButton>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};


const CodeSlide = () => {
    const ref = useRef<HTMLDivElement>(null);
    const inView = useInView(ref);

    return (
        <div
            ref={ref}
            className={clsx("rounded-lg", "dark:bg-gray-900 bg-gray-0")}
        >
            <div
                className={clsx(
                    "text-[10px] leading-[16px]",
                    "h-[268px] landing-md:h-[299px]",
                    "font-jetBrains-mono",
                    "select-none",
                    "overflow-hidden",
                    "relative",
                    "z-[1px]",
                    "dark:text-[#d6deeb] text-gray-900",
                    "dark:landing-react-code-mask-dark landing-react-code-mask",
                )}
            >
                <div
                    className={clsx(
                        "will-change-transform p-2",
                        inView && "animate-code-scroll",
                    )}
                >
                    <HighlightCode />
                    <div className={clsx("h-8")} />
                    <HighlightCode />
                </div>
            </div>
        </div>
    );
};

const HighlightCode = memo(function HighlightCodeBase() {
    const { colorMode } = useColorMode();
    const isDarkTheme = colorMode === "dark";

    const theme = isDarkTheme ? nightOwlDark : nightOwlLight;

    return (
        <Highlight
            {...defaultProps}
            theme={theme}
            code={`${code ?? ""}`.trim()}
            language="tsx"
        >
            {({ tokens, getLineProps, getTokenProps }) => (
                <>
                    {tokens.map((line, i) => (
                        <div
                            {...getLineProps({
                                line,
                            })}
                            key={`${code}-${i}`}
                        >
                            <span
                                className={
                                    "dark:text-gray-600 text-gray-500 pl-4 pr-2"
                                }
                            >
                                {i + 1}
                            </span>
                            {line.map((token, key) => {
                                const { children: _children, ...tokenProps } =
                                    getTokenProps({
                                        token,
                                    });

                                return (
                                    <span
                                        {...tokenProps}
                                        key={`${token.content}-${key}`}
                                        className="whitespace-pre"
                                        dangerouslySetInnerHTML={{
                                            __html: token.content,
                                        }}
                                    />
                                );
                            })}
                        </div>
                    ))}
                </>
            )}
        </Highlight>
    );
});

const code = `
services:

  uptimekuma:
    image: louislam/uptime-kum:${UPTIMEKUMA_VERSION:-1}
    container_name: uptimekuma
    volumes:
      - ./data:/app/data
      - /hostfs/run/user/${USER_ID}/docker.sock:/var/run/docker.sock:ro
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: "${UPTIMEKUMA_CPU:-0.35}"
          memory: "${UPTIMEKUMA_RAM:-0.35G}"   
          pids: 100
    networks:
      - www

  openresty:
    image: openresty/openresty:${OPENRESTY_VERSION:-bullseye-fat}
    container_name: openresty
    restart: always
    ports:
      - "${PROXY_HTTP_PORT:-${HTTP_PORT}}"
      - "${HTTPS_PORT}"
    working_dir: /var/www/html
    volumes:
      - ./openresty.conf:/etc/openresty/nginx.conf:ro
      - webserver_data:/etc/nginx/conf.d
      - html_data:/var/www/html/   
      - /etc/openpanel/nginx/certs/:/etc/nginx/ssl/
      - /etc/openpanel/nginx/default_page.html:/etc/nginx/default_page.html:ro
    deploy:
      resources:
        limits:
          cpus: "${OPENRESTY_CPU:-0.5}"
          memory: "${OPENRESTY_RAM:-0.5G}"   
          pids: 100
    depends_on:
      - "php-fpm-${DEFAULT_PHP_VERSION:-8.2}"
    networks:
      - www

  nginx:
    image: nginx:${NGINX_VERSION:-latest}
    container_name: nginx
    restart: always
    ports:
      - "${PROXY_HTTP_PORT:-${HTTP_PORT}}"
      - "${HTTPS_PORT}"
    working_dir: /var/www/html
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - webserver_data:/etc/nginx/conf.d
      - html_data:/var/www/html/ 
      - /etc/openpanel/nginx/certs/:/etc/nginx/ssl/   
      - /etc/openpanel/nginx/default_page.html:/etc/nginx/default_page.html:ro
    deploy:
      resources:
        limits:
          cpus: "${NGINX_CPU:-0.5}"
          memory: "${NGINX_RAM:-0.5G}"   
          pids: 100
    depends_on:
      - "php-fpm-${DEFAULT_PHP_VERSION:-8.2}"
    networks:
      - www


  apache:
    image: httpd:${APACHE_VERSION:-latest}
    container_name: apache
    restart: always
    ports:
      - "${PROXY_HTTP_PORT:-${HTTP_PORT}}"
      - "${HTTPS_PORT}"
    working_dir: /var/www/html
    volumes:
      - ./httpd.conf:/usr/local/apache2/conf/httpd.conf:ro
      - webserver_data:/etc/httpd/conf.d/
      - html_data:/var/www/html/
      - /etc/openpanel/nginx/certs/:/etc/apache2/ssl/
      - /etc/openpanel/nginx/default_page.html:/etc/apache2/default_page.html:ro
    deploy:
      resources:
        limits:
          cpus: "${APACHE_CPU:-0.5}"
          memory: "${APACHE_RAM:-0.5G}"
          pids: 100
    depends_on:
      - "php-fpm-${DEFAULT_PHP_VERSION:-8.2}"
    networks:
      - www

  varnish:
    image: varnish:${VARNISH_VERSION:-stable}
    container_name: varnish
    environment:
      WEB_SERVER: "${WEB_SERVER:-nginx}"
      VARNISH_SIZE: "${VARNISH_SIZE:-1G}"
    entrypoint: /bin/sh -c "cp /etc/varnish/default.vcl.template /etc/varnish/default.vcl && sed -i 's|VARNISH_BACKEND_HOST|'"$WEB_SERVER"'|g' /etc/varnish/default.vcl && /usr/local/bin/docker-varnish-entrypoint"
    command: [ "-n", "/var/lib/varnish" ]
    volumes:
      - ./default.vcl:/etc/varnish/default.vcl.template:ro  
    ports:
      - "${HTTP_PORT}"
    tmpfs:
      - /var/lib/varnish:exec
    deploy:
      resources:
        limits:
          cpus: "${VARNISH_CPU:-0.5}"
          memory: "${VARNISH_RAM:-0.5G}"       
    depends_on:
      - "${WEB_SERVER:-nginx}"
    networks:
      - www

  cron:
    image: mcuadros/ofelia:latest
    container_name: cron
    command: daemon --config=/crons.ini
    volumes:
      - /hostfs/run/user/${USER_ID}/docker.sock:/var/run/docker.sock:ro
      - /hostfs/home/${CONTEXT}/crons.ini:/crons.ini:ro
    deploy:
      resources:
        limits:
          cpus: "${CRON_CPU:-0.5}"
          memory: "${CRON_RAM:-0.5G}"
          pids: 100
          
  mysql:
    image: mysql:${MYSQL_VERSION:-latest}
    container_name: mysql
    restart: always
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD:-rootpassword}
    ports:
      - "${MYSQL_PORT}"     
    volumes:
      - mysql_data:/var/lib/mysql
      - mysql_dumps:/tmp/dumps/
      - /etc/openpanel/mysql/scripts/dump.sh:/dump.sh:ro
      - ./sockets/mysqld:/var/run/mysqld 
      - ./my.cnf:/root/.my.cnf:ro 
      - ./custom.cnf:/etc/mysql/conf.d/custom.cnf
    labels:
      - docker-volume-backup.archive-pre=/bin/sh -c '/dump.sh'
      - docker-volume-backup.archive-post=/bin/sh -c 'rm -rf /tmp/dumps/*'
    deploy:
      resources:
        limits:
          cpus: "${MYSQL_CPU:-0.5}"
          memory: "${MYSQL_RAM:-0.5G}"
          pids: 100
    networks:
      - db
    healthcheck:
      test: ['CMD-SHELL', 'mysqladmin ping -h localhost']
      interval: 1s
      timeout: 5s
      retries: 10

  mariadb:
    image: mariadb:${MYSQL_VERSION:-latest}
    container_name: mariadb
    restart: always
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD:-rootpassword}
    ports:
      - "${MYSQL_PORT}"
    volumes:
      - mysql_data:/var/lib/mysql
      - mysql_dumps:/tmp/dumps/
      - /etc/openpanel/mysql/scripts/dump.sh:/dump.sh:ro
      - ./sockets/mysqld:/var/run/mysqld
      - ./my.cnf:/root/.my.cnf:ro
      - ./custom.cnf:/etc/mysql/conf.d/custom.cnf
    labels:
      - docker-volume-backup.archive-pre=/bin/sh -c '/dump.sh'
      - docker-volume-backup.archive-post=/bin/sh -c 'rm -rf /tmp/dumps/*'
    deploy:
      resources:
        limits:
          cpus: "${MARIADB_CPU:-0.5}"
          memory: "${MARIADB_RAM:-0.5G}"
          pids: 100
    networks:
      - db
    healthcheck:
      test: ['CMD-SHELL', 'mariadb-admin ping -h localhost']
      interval: 1s
      timeout: 5s
      retries: 10

  phpmyadmin:
    container_name: phpmyadmin 
    depends_on:
      - "${MYSQL_TYPE:-mysql}"
    image: phpmyadmin
    restart: always
    volumes:
      - html_data:/html/
    ports:
      - "${PMA_PORT}"
    deploy:
      resources:
        limits:
          cpus: "${PMA_CPU:-0.1}"
          memory: "${PMA_RAM:-0.1G}"                
          pids: 100
    environment:
      PMA_HOST: ${MYSQL_TYPE:-mysql}
      MAX_EXECUTION_TIME: ${PMA_MAX_EXECUTION_TIME:-600}
      MEMORY_LIMIT: ${PMA_MEMORY_LIMIT:-512M}
      UPLOAD_LIMIT: ${PMA_UPLOAD_LIMIT:-256M}
      PMA_UPLOADDIR: "/html/"
      PMA_SAVEDIR: "/html/"
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD:-rootpassword}
    networks:
      - db




  mssql:
    image: ${MSSQL_IMAGE}:${MSSQL_VERSION:-latest}
    container_name: mssql
    restart: always
    environment:
      ACCEPT_EULA: "Y"
      MSSQL_SA_PASSWORD: ${MSSQL_SA_PASSWORD:-StrongPassword!}
      MSSQL_PID: ${MSSQL_PID:-Developer}
    ports:
      - "${MSSQL_PORT}"
    volumes:
      - mssql_data:/var/opt/mssql
      - ./sockets/mssql:/var/opt/mssql/sockets
      - ./mssql.conf:/etc/mssql/mssql.conf:ro
    deploy:
      resources:
        limits:
          cpus: "${MSSQL_CPU:-1}"
          memory: "${MSSQL_RAM:-2G}"
          pids: 100
    networks:
      - db
    healthcheck:
      test: ['CMD-SHELL', 'sqlcmd -S localhost -U sa -P "$$MSSQL_SA_PASSWORD" -Q "SELECT 1" || exit 1']
      interval: 10s
      timeout: 5s
      retries: 5
      


  postgres:
    image: postgres:${POSTGRES_VERSION:-latest}
    container_name: postgres
    restart: always
    environment:
      POSTGRES_USER: ${POSTGRES_USER:-postgres}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-postgrespassword}
    ports:
      - "${POSTGRES_PORT}"
    volumes:
      - pg_data:/var/lib/postgresql/data
      - ./sockets/postgres:/var/run/postgresql
    deploy:
      resources:
        limits:
          cpus: "${POSTGRES_CPU:-0.3}"
          memory: "${POSTGRES_RAM:-0.3}"
          pids: 100
    networks:
      - db
    healthcheck:
      test: ['CMD-SHELL', 'pg_isready -U ${POSTGRES_USER:-postgres}']
      interval: 1s
      timeout: 5s
      retries: 10
      

  pgadmin:
    container_name: pgadmin
    image: dpage/pgadmin4:${PGADMIN_VERSION:-latest}
    environment:
      - PGADMIN_DEFAULT_EMAIL=${PGADMIN_MAIL}
      - PGADMIN_DEFAULT_PASSWORD=${PGADMIN_PW}
    depends_on:
      - "postgres"
    ports:
      #- "5050:80"
      - "${PGADMIN_PORT}"
    restart: always
    deploy:
      resources:
        limits:
          cpus: "${PGADMIN_CPU:-0.1}"
          memory: "${PGADMIN_RAM:-0.1}"
          pids: 100
    networks:
      - db

  
  redis:
    image: redis:${REDIS_VERSION:-7.4.2-alpine}
    container_name: redis
    restart: unless-stopped
    user: "${USER_ID:-0}"
    command: ["redis-server", "--bind", "0.0.0.0", "--port", "6379"]
    volumes:
      - ./sockets/redis:/var/run/redis  # Redis socket
    deploy:
      resources:
        limits:
          cpus: "${REDIS_CPU:-0.1}"
          memory: "${REDIS_RAM-0.1G}" 
          pids: 100
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      retries: 3
      start_period: 10s
      timeout: 5s
    networks:
      - www


  elasticsearch:
    image: elasticsearch:${ELASTICSEARH_VERSION:-7.16.1}
    container_name: elasticsearch
    restart: unless-stopped
    user: "${USER_ID:-0}"
    environment:
      discovery.type: single-node
      ES_JAVA_OPTS: "-Xms512m -Xmx512m"
    volumes:
      - ./sockets/elasticsearch:/var/run/elasticsearch
    deploy:
      resources:
        limits:
          cpus: "${ELASTICSEARCH_CPU:-0.5}"
          memory: "${ELASTICSEARCH_RAM-1G}" 
          pids: 100
    healthcheck:
      test: ["CMD-SHELL", "curl --silent --fail localhost:9200/_cluster/health || exit 1"]
      interval: 10s
      timeout: 10s
      retries: 3
    networks:
      - www


  opensearch:
    image: opensearchproject/opensearch:${OPENSEARCH_VERSION:-2.11.0}
    container_name: opensearch
    restart: unless-stopped
    environment:
      discovery.type: single-node
      OPENSEARCH_JAVA_OPTS: "-Xms512m -Xmx512m"
      DISABLE_SECURITY_PLUGIN: "true"
      OPENSEARCH_INITIAL_ADMIN_PASSWORD: "${OPENSEARCH_INITIAL_ADMIN_PASSWORD:-iIybFJOgznCYmpOO}"
      DISABLE_INSTALL_DEMO_CONFIG: "true"
    volumes:
      - ./sockets/opensearch:/var/run/opensearch
    deploy:
      resources:
        limits:
          cpus: "${OPENSEARCH_CPU:-0.5}"
          memory: "${OPENSEARCH_RAM:-1G}" 
          pids: 100
    healthcheck:
      test: ["CMD-SHELL", "curl --silent --fail localhost:9200/_cluster/health || exit 1"]
      interval: 10s
      timeout: 10s
      retries: 3
    networks:
      - www
  
  memcached:
    image: memcached:${MEMCACHED_VERSION:-7.4.2-alpine}
    container_name: memcached
    restart: unless-stopped
    user: "${USER_ID:-0}"
    command: ["memcached", "-u", "root", "-l", "0.0.0.0", "-p", "11211"]
    volumes:
      - ./sockets/memcached:/var/run/  # Memcached socket
    deploy:
      resources:
        limits:
          cpus: "${MEMCACHED_CPU:-0.1}"
          memory: "${MEMCACHED_RAM-0.1G}" 
          pids: 100
    healthcheck:
      test: ["CMD", "memcached", "-h"]
      interval: 30s
      retries: 3
      start_period: 10s
      timeout: 5s
    networks:
      - www


  php-fpm-5.6:
    image: shinsenter/php:5.6-fpm
    container_name: php-fpm-5.6
    restart: always
    volumes:
      - html_data:/var/www/html/
      - ./php.ini/5.6.ini:/usr/local/etc/php/php.ini
      - /etc/openpanel/wordpress/wp-cli.phar:/usr/local/bin/wp  
    environment:
      - PHP_ENABLE_OPCACHE=${PHP_FPM_5_6_ENABLE_OPCACHE:-true}
      - PHP_MAX_EXECUTION_TIME=${PHP_FPM_5_6_PHP_MAX_EXECUTION_TIME:-600}
      - PHP_MAX_INPUT_TIME=${PHP_FPM_5_6_PHP_MAX_INPUT_TIME:-600}
      - PHP_MAX_INPUT_VARS=${PHP_FPM_5_6_PHP_MAX_INPUT_VARS:-1000}
      - PHP_MEMORY_LIMIT=${PHP_FPM_5_6_PHP_MEMORY_LIMIT:-512M}
      - PHP_POST_MAX_SIZE=${PHP_FPM_5_6_PHP_POST_MAX_SIZE:-512M}
      - PHP_UPLOAD_MAX_FILESIZE=${PHP_FPM_5_6_PHP_UPLOAD_MAX_FILESIZE:-256M}   
      - APP_USER=root
      - APP_GROUP=root
      - APP_UID=${USER_ID:-0}
      - APP_GID=${USER_ID:-0}
      - ENABLE_TUNING_FPM=1
    deploy:
      resources:
        limits:
          cpus: "${PHP_FPM_5_6_CPU:-0.125}"
          memory: "${PHP_FPM_5_6_RAM:-0.25G}"
          pids: 100
    networks:
      - www
      - db
    command: >
      sh -c "php-fpm --allow-to-run-as-root"    

  php-fpm-7.0:
    image: shinsenter/php:7.0-fpm
    container_name: php-fpm-7.0
    restart: always
    volumes:
      - html_data:/var/www/html/
      - ./php.ini/7.0.ini:/usr/local/etc/php/php.ini
      - /etc/openpanel/wordpress/wp-cli.phar:/usr/local/bin/wp  
    environment:
      - PHP_ENABLE_OPCACHE=${PHP_FPM_7_0_ENABLE_OPCACHE:-true}
      - PHP_MAX_EXECUTION_TIME=${PHP_FPM_7_0_PHP_MAX_EXECUTION_TIME:-600}
      - PHP_MAX_INPUT_TIME=${PHP_FPM_7_0_PHP_MAX_INPUT_TIME:-600}
      - PHP_MAX_INPUT_VARS=${PHP_FPM_7_0_PHP_MAX_INPUT_VARS:-1000}
      - PHP_MEMORY_LIMIT=${PHP_FPM_7_0_PHP_MEMORY_LIMIT:-512M}
      - PHP_POST_MAX_SIZE=${PHP_FPM_7_0_PHP_POST_MAX_SIZE:-512M}
      - PHP_UPLOAD_MAX_FILESIZE=${PHP_FPM_7_0_PHP_UPLOAD_MAX_FILESIZE:-256M}  
      - APP_USER=root
      - APP_GROUP=root
      - APP_UID=${USER_ID:-0}
      - APP_GID=${USER_ID:-0}
      - ENABLE_TUNING_FPM=1
    deploy:
      resources:
        limits:
          cpus: "${PHP_FPM_7_0_CPU:-0.125}"
          memory: "${PHP_FPM_7_0_RAM:-0.25G}"
          pids: 100
    networks:
      - www
      - db
    command: >
      sh -c "php-fpm --allow-to-run-as-root"   
      
  php-fpm-7.1:
    image: shinsenter/php:7.1-fpm
    container_name: php-fpm-7.1
    restart: always
    volumes:
      - html_data:/var/www/html/
      - ./php.ini/7.1.ini:/usr/local/etc/php/php.ini
      - /etc/openpanel/wordpress/wp-cli.phar:/usr/local/bin/wp  
    environment:
      - PHP_ENABLE_OPCACHE=${PHP_FPM_7_1_ENABLE_OPCACHE:-true}
      - PHP_MAX_EXECUTION_TIME=${PHP_FPM_7_1_PHP_MAX_EXECUTION_TIME:-600}
      - PHP_MAX_INPUT_TIME=${PHP_FPM_7_1_PHP_MAX_INPUT_TIME:-600}
      - PHP_MAX_INPUT_VARS=${PHP_FPM_7_1_PHP_MAX_INPUT_VARS:-1000}
      - PHP_MEMORY_LIMIT=${PHP_FPM_7_1_PHP_MEMORY_LIMIT:-512M}
      - PHP_POST_MAX_SIZE=${PHP_FPM_7_1_PHP_POST_MAX_SIZE:-512M}
      - PHP_UPLOAD_MAX_FILESIZE=${PHP_FPM_7_1_PHP_UPLOAD_MAX_FILESIZE:-256M}
      - APP_USER=root
      - APP_GROUP=root
      - APP_UID=${USER_ID:-0}
      - APP_GID=${USER_ID:-0}
      - ENABLE_TUNING_FPM=1
    deploy:
      resources:
        limits:
          cpus: "${PHP_FPM_7_1_CPU:-0.125}"
          memory: "${PHP_FPM_7_1_RAM:-0.25G}"
          pids: 100
    networks:
      - www
      - db
    command: >
      sh -c "php-fpm --allow-to-run-as-root"   
      

  php-fpm-7.2:
    image: shinsenter/php:7.2-fpm
    container_name: php-fpm-7.2
    restart: always
    volumes:
      - html_data:/var/www/html/
      - ./php.ini/7.2.ini:/usr/local/etc/php/php.ini
      - /etc/openpanel/wordpress/wp-cli.phar:/usr/local/bin/wp  
    environment:
      - PHP_ENABLE_OPCACHE=${PHP_FPM_7_2_ENABLE_OPCACHE:-true}
      - PHP_MAX_EXECUTION_TIME=${PHP_FPM_7_2_PHP_MAX_EXECUTION_TIME:-600}
      - PHP_MAX_INPUT_TIME=${PHP_FPM_7_2_PHP_MAX_INPUT_TIME:-600}
      - PHP_MAX_INPUT_VARS=${PHP_FPM_7_2_PHP_MAX_INPUT_VARS:-1000}
      - PHP_MEMORY_LIMIT=${PHP_FPM_7_2_PHP_MEMORY_LIMIT:-512M}
      - PHP_POST_MAX_SIZE=${PHP_FPM_7_2_PHP_POST_MAX_SIZE:-512M}
      - PHP_UPLOAD_MAX_FILESIZE=${PHP_FPM_7_2_PHP_UPLOAD_MAX_FILESIZE:-256M}
      - APP_USER=root
      - APP_GROUP=root
      - APP_UID=${USER_ID:-0}
      - APP_GID=${USER_ID:-0}
      - ENABLE_TUNING_FPM=1
    deploy:
      resources:
        limits:
          cpus: "${PHP_FPM_7_2_CPU:-0.125}"
          memory: "${PHP_FPM_7_2_RAM:-0.25G}"
          pids: 100
    networks:
      - www
      - db
    command: >
      sh -c "php-fpm --allow-to-run-as-root"
      


  php-fpm-7.3:
    image: shinsenter/php:7.3-fpm
    container_name: php-fpm-7.3
    restart: always
    volumes:
      - html_data:/var/www/html/
      - ./php.ini/7.3.ini:/usr/local/etc/php/php.ini
      - /etc/openpanel/wordpress/wp-cli.phar:/usr/local/bin/wp  
    environment:
      - PHP_ENABLE_OPCACHE=${PHP_FPM_7_3_ENABLE_OPCACHE:-true}
      - PHP_MAX_EXECUTION_TIME=${PHP_FPM_7_3_PHP_MAX_EXECUTION_TIME:-600}
      - PHP_MAX_INPUT_TIME=${PHP_FPM_7_3_PHP_MAX_INPUT_TIME:-600}
      - PHP_MAX_INPUT_VARS=${PHP_FPM_7_3_PHP_MAX_INPUT_VARS:-1000}
      - PHP_MEMORY_LIMIT=${PHP_FPM_7_3_PHP_MEMORY_LIMIT:-512M}
      - PHP_POST_MAX_SIZE=${PHP_FPM_7_3_PHP_POST_MAX_SIZE:-512M}
      - PHP_UPLOAD_MAX_FILESIZE=${PHP_FPM_7_3_PHP_UPLOAD_MAX_FILESIZE:-256M} 
      - APP_USER=root
      - APP_GROUP=root
      - APP_UID=${USER_ID:-0}
      - APP_GID=${USER_ID:-0}
      - ENABLE_TUNING_FPM=1
    deploy:
      resources:
        limits:
          cpus: "${PHP_FPM_7_3_CPU:-0.125}"
          memory: "${PHP_FPM_7_3_RAM:-0.25G}"
          pids: 100
    networks:
      - www
      - db
    command: >
      sh -c "php-fpm --allow-to-run-as-root" 
      

  php-fpm-7.4:
    image: shinsenter/php:7.4-fpm
    container_name: php-fpm-7.4
    restart: always
    volumes:
      - html_data:/var/www/html/
      - ./php.ini/7.4/:/usr/local/etc/php
      - /etc/openpanel/wordpress/wp-cli.phar:/usr/local/bin/wp  
    environment:
      - PHP_ENABLE_OPCACHE=${PHP_FPM_7_4_ENABLE_OPCACHE:-true}
      - PHP_MAX_EXECUTION_TIME=${PHP_FPM_7_4_PHP_MAX_EXECUTION_TIME:-600}
      - PHP_MAX_INPUT_TIME=${PHP_FPM_7_4_PHP_MAX_INPUT_TIME:-600}
      - PHP_MAX_INPUT_VARS=${PHP_FPM_7_4_PHP_MAX_INPUT_VARS:-1000}
      - PHP_MEMORY_LIMIT=${PHP_FPM_7_4_PHP_MEMORY_LIMIT:-512M}
      - PHP_POST_MAX_SIZE=${PHP_FPM_7_4_PHP_POST_MAX_SIZE:-512M}
      - PHP_UPLOAD_MAX_FILESIZE=${PHP_FPM_7_4_PHP_UPLOAD_MAX_FILESIZE:-256M} 
      - APP_USER=root
      - APP_GROUP=root
      - APP_UID=${USER_ID:-0}
      - APP_GID=${USER_ID:-0}
      - ENABLE_TUNING_FPM=1
    deploy:
      resources:
        limits:
          cpus: "${PHP_FPM_7_4_CPU:-0.125}"
          memory: "${PHP_FPM_7_4_RAM:-0.25G}"
          pids: 100
    networks:
      - www
      - db
    command: >
      sh -c "php-fpm --allow-to-run-as-root" 
      
  php-fpm-8.0:
    image: shinsenter/php:8.0-fpm
    container_name: php-fpm-8.0
    restart: always
    volumes:
      - html_data:/var/www/html/
      - ./php.ini/8.0.ini:/usr/local/etc/php/php.ini
      - /etc/openpanel/wordpress/wp-cli.phar:/usr/local/bin/wp
    environment:
      - PHP_ENABLE_OPCACHE=${PHP_FPM_8_0_ENABLE_OPCACHE:-true}
      - PHP_MAX_EXECUTION_TIME=${PHP_FPM_8_0_PHP_MAX_EXECUTION_TIME:-600}
      - PHP_MAX_INPUT_TIME=${PHP_FPM_8_0_PHP_MAX_INPUT_TIME:-600}
      - PHP_MAX_INPUT_VARS=${PHP_FPM_8_0_PHP_MAX_INPUT_VARS:-1000}
      - PHP_MEMORY_LIMIT=${PHP_FPM_8_0_PHP_MEMORY_LIMIT:-512M}
      - PHP_POST_MAX_SIZE=${PHP_FPM_8_0_PHP_POST_MAX_SIZE:-512M}
      - PHP_UPLOAD_MAX_FILESIZE=${PHP_FPM_8_0_PHP_UPLOAD_MAX_FILESIZE:-256M} 
      - APP_USER=root
      - APP_GROUP=root
      - APP_UID=${USER_ID:-0}
      - APP_GID=${USER_ID:-0}
      - ENABLE_TUNING_FPM=1
    deploy:
      resources:
        limits:
          cpus: "${PHP_FPM_8_0_CPU:-0.125}"
          memory: "${PHP_FPM_8_0_RAM:-0.25G}"
          pids: 100
    networks:
      - www
      - db
    command: >
      sh -c "php-fpm --allow-to-run-as-root"
   
      

  php-fpm-8.1:
    image: shinsenter/php:8.1-fpm
    container_name: php-fpm-8.1
    restart: always
    volumes:
      - html_data:/var/www/html/
      - ./php.ini/8.1.ini:/usr/local/etc/php/php.ini
      - /etc/openpanel/wordpress/wp-cli.phar:/usr/local/bin/wp
    environment:
      - PHP_ENABLE_OPCACHE=${PHP_FPM_8_1_ENABLE_OPCACHE:-true}
      - PHP_MAX_EXECUTION_TIME=${PHP_FPM_8_1_PHP_MAX_EXECUTION_TIME:-600}
      - PHP_MAX_INPUT_TIME=${PHP_FPM_8_1_PHP_MAX_INPUT_TIME:-600}
      - PHP_MAX_INPUT_VARS=${PHP_FPM_8_1_PHP_MAX_INPUT_VARS:-1000}
      - PHP_MEMORY_LIMIT=${PHP_FPM_8_1_PHP_MEMORY_LIMIT:-512M}
      - PHP_POST_MAX_SIZE=${PHP_FPM_8_1_PHP_POST_MAX_SIZE:-512M}
      - PHP_UPLOAD_MAX_FILESIZE=${PHP_FPM_8_1_PHP_UPLOAD_MAX_FILESIZE:-256M} 
      - APP_USER=root
      - APP_GROUP=root
      - APP_UID=${USER_ID:-0}
      - APP_GID=${USER_ID:-0}
      - ENABLE_TUNING_FPM=1
    deploy:
      resources:
        limits:
          cpus: "${PHP_FPM_8_1_CPU:-0.125}"
          memory: "${PHP_FPM_8_1_RAM:-0.25G}"
          pids: 100
    networks:
      - www
      - db
    command: >
      sh -c "php-fpm --allow-to-run-as-root"
  

      
  php-fpm-8.2:
    image: shinsenter/php:8.2-fpm
    container_name: php-fpm-8.2
    restart: always
    volumes:
      - html_data:/var/www/html/
      - ./php.ini/8.2.ini:/usr/local/etc/php/php.ini
      - /etc/openpanel/wordpress/wp-cli.phar:/usr/local/bin/wp
      - ./ioncube/ioncube_loader_lin_8.2.so:/usr/local/lib/php/extensions/no-debug-non-zts-20220829/ioncube_loader.so
    environment:
      - PHP_ENABLE_OPCACHE=${PHP_FPM_8_2_ENABLE_OPCACHE:-true}
      - PHP_MAX_EXECUTION_TIME=${PHP_FPM_8_2_PHP_MAX_EXECUTION_TIME:-600}
      - PHP_MAX_INPUT_TIME=${PHP_FPM_8_2_PHP_MAX_INPUT_TIME:-600}
      - PHP_MAX_INPUT_VARS=${PHP_FPM_8_2_PHP_MAX_INPUT_VARS:-1000}
      - PHP_MEMORY_LIMIT=${PHP_FPM_8_2_PHP_MEMORY_LIMIT:-512M}
      - PHP_POST_MAX_SIZE=${PHP_FPM_8_2_PHP_POST_MAX_SIZE:-512M}
      - PHP_UPLOAD_MAX_FILESIZE=${PHP_FPM_8_2_PHP_UPLOAD_MAX_FILESIZE:-256M} 
      - APP_USER=root
      - APP_GROUP=root
      - APP_UID=${USER_ID:-0}
      - APP_GID=${USER_ID:-0}
      - ENABLE_TUNING_FPM=1
    deploy:
      resources:
        limits:
          cpus: "${PHP_FPM_8_2_CPU:-0.125}"
          memory: "${PHP_FPM_8_2_RAM:-0.25G}"
          pids: 100
    networks:
      - www
      - db
    command: >
      sh -c "php-fpm --allow-to-run-as-root"
      

  php-fpm-8.3:
    image: shinsenter/php:8.3-fpm
    container_name: php-fpm-8.3
    restart: always
    volumes:
      - html_data:/var/www/html/
      - ./php.ini/8.3.ini:/usr/local/etc/php/php.ini
      - /etc/openpanel/wordpress/wp-cli.phar:/usr/local/bin/wp
    environment:
      - PHP_ENABLE_OPCACHE=${PHP_FPM_8_3_ENABLE_OPCACHE:-true}
      - PHP_MAX_EXECUTION_TIME=${PHP_FPM_8_3_PHP_MAX_EXECUTION_TIME:-600}
      - PHP_MAX_INPUT_TIME=${PHP_FPM_8_3_PHP_MAX_INPUT_TIME:-600}
      - PHP_MAX_INPUT_VARS=${PHP_FPM_8_3_PHP_MAX_INPUT_VARS:-1000}
      - PHP_MEMORY_LIMIT=${PHP_FPM_8_3_PHP_MEMORY_LIMIT:-512M}
      - PHP_POST_MAX_SIZE=${PHP_FPM_8_3_PHP_POST_MAX_SIZE:-512M}
      - PHP_UPLOAD_MAX_FILESIZE=${PHP_FPM_8_3_PHP_UPLOAD_MAX_FILESIZE:-256M} 
      - APP_USER=root
      - APP_GROUP=root
      - APP_UID=${USER_ID:-0}
      - APP_GID=${USER_ID:-0}
      - ENABLE_TUNING_FPM=1
    deploy:
      resources:
        limits:
          cpus: "${PHP_FPM_8_3_CPU:-0.125}"
          memory: "${PHP_FPM_8_3_RAM:-0.25G}"
          pids: 100
    networks:
      - www
      - db
    command: >
      sh -c "php-fpm --allow-to-run-as-root"
      
  php-fpm-8.4:
    image: shinsenter/php:8.4-fpm
    container_name: php-fpm-8.4
    restart: always
    volumes:
      - html_data:/var/www/html/
      - ./php.ini/8.4.ini:/usr/local/etc/php/php.ini
      - /etc/openpanel/wordpress/wp-cli.phar:/usr/local/bin/wp  
    environment:
      - PHP_ENABLE_OPCACHE=${PHP_FPM_8_4_ENABLE_OPCACHE:-true}
      - PHP_MAX_EXECUTION_TIME=${PHP_FPM_8_4_PHP_MAX_EXECUTION_TIME:-600}
      - PHP_MAX_INPUT_TIME=${PHP_FPM_8_4_PHP_MAX_INPUT_TIME:-600}
      - PHP_MAX_INPUT_VARS=${PHP_FPM_8_4_PHP_MAX_INPUT_VARS:-1000}
      - PHP_MEMORY_LIMIT=${PHP_FPM_8_4_PHP_MEMORY_LIMIT:-512M}
      - PHP_POST_MAX_SIZE=${PHP_FPM_8_4_PHP_POST_MAX_SIZE:-512M}
      - PHP_UPLOAD_MAX_FILESIZE=${PHP_FPM_8_4_PHP_UPLOAD_MAX_FILESIZE:-256M}
      - APP_USER=root
      - APP_GROUP=root
      - APP_UID=${USER_ID:-0}
      - APP_GID=${USER_ID:-0}
      - ENABLE_TUNING_FPM=1
    deploy:
      resources:
        limits:
          cpus: "${PHP_FPM_8_4_CPU:-0.125}"
          memory: "${PHP_FPM_8_4_RAM:-0.25G}"
          pids: 100
    networks:
      - www
      - db
    command: >
      sh -c "php-fpm --allow-to-run-as-root" 


  backup:
    image: offen/docker-volume-backup:v2
    container_name: backup
    env_file:
      - ./backup.env
    volumes:
      - html_data:/var/www/html/:ro
      - html_data:/backup/html:ro
      - mysql_dumps:/backup/mysql:ro
      - mssql_data:/backup/mssql:ro
      - mail_data:/backup/mail:ro
      - webserver_data:/backup/vhosts:ro
      - pg_data:/backup/postgres:ro
      - mc_data:/backup/minecraft:ro
      - /run/user/${USER_ID}/docker.sock:/var/run/docker.sock:ro
    deploy:
      resources:
        limits:
          cpus: "${BACKUP_CPU:-1.0}"
          memory: "${BACKUP_RAM:-1.0g}"
          pids: 100
    command: [ "backup" ]
    restart: unless-stopped
      
  minecraft:
    image: itzg/minecraft-server:${MINECRAFT_VERSION:-latest}
    container_name: minecraft
    tty: true
    stdin_open: true
    ports:
      - "${MINECRAFT_PORT:-25565}:25565"
    environment:
      EULA: "TRUE"
      ENABLE_QUERY: "${MINECRAFT_ENABLE_QUERY:-true}"
      QUERY_PORT: "${MINECRAFT_PORT:-25565}"
    volumes:
      - mc_data:/data
    deploy:
      resources:
        limits:
          cpus: "${MINECRAFT_CPU:-1.0}"
          memory: "${MINECRAFT_RAM:-1.0G}"
          pids: 100
    healthcheck:
      test: mc-health
      start_period: 1m
      interval: 5s
      retries: 20
    networks:
      - www

networks:
  default:
    driver: bridge
    labels:
      description: "This network is not used used and serves as a fallback."
      purpose: "internal"
  www:
    driver: bridge
    labels:
      description: "This network is used for communication between website services (webserver to apps)."
      purpose: "internal"
  db:
    driver: bridge
    labels:
      description: "This network is used for communication between websites and database services (apps to database)."
      purpose: "internal"

volumes:
  mysql_data:
    driver: local
    labels:
      description: "This volume holds the mysql/mariadb databases."
      purpose: "database"
  mysql_dumps:
    driver: local
    labels:
      description: "This volume holds the .sql dumps of databases during backup."
      purpose: "storage"
  mssql_data:
    driver: local
    labels:
      description: "This volume holds the MSSQL databases."
      purpose: "database"      
  html_data:
    driver: local
    labels:
      description: "This volume holds the /var/www/html/ directory."
      purpose: "storage"
  mail_data:
    driver: local
    labels:
      description: "This volume holds the /var/mail/ directory."
      purpose: "storage"
  webserver_data:
    driver: local
    labels:
      description: "This volume holds the Nginx/Apache vhost files."
      purpose: "configuration"
  pg_data:
    driver: local
    labels:
      description: "This volume holds the postgresql databases."
      purpose: "database"
  mc_data:
    driver: local
    labels:
      description: "This volume holds the minecraft data directory."
      purpose: "storage"
`;
