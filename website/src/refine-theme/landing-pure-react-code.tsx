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
    image: louislam/uptime-kum:${UPTIMEKUMA_VERSION}
    container_name: uptimekuma
    volumes:
      - ./data:/app/data
      - /hostfs/run/user/${USER_ID}/docker.sock:/var/run/docker.sock:ro
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: "${UPTIMEKUMA_CPU}"
          memory: "${UPTIMEKUMA_RAM}"   
          pids: 100
    networks:
      - www

  openresty:
    image: openresty/openresty:${OPENRESTY_VERSION}
    container_name: openresty
    restart: always
    ports:
      - "${PROXY_HTTP_PORT}"
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
          cpus: "${OPENRESTY_CPU}"
          memory: "${OPENRESTY_RAM}"   
          pids: 100
    depends_on:
      - "php-fpm-${DEFAULT_PHP_VERSION}"
    networks:
      - www

  nginx:
    image: nginx:${NGINX_VERSION}
    container_name: nginx
    restart: always
    ports:
      - "${PROXY_HTTP_PORT}"
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
          cpus: "${NGINX_CPU}"
          memory: "${NGINX_RAM}"   
          pids: 100
    depends_on:
      - "php-fpm-${DEFAULT_PHP_VERSION}"
    networks:
      - www


  apache:
    image: httpd:${APACHE_VERSION}
    container_name: apache
    restart: always
    ports:
      - "${PROXY_HTTP_PORT}"
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
          cpus: "${APACHE_CPU}"
          memory: "${APACHE_RAM}"
          pids: 100
    depends_on:
      - "php-fpm-${DEFAULT_PHP_VERSION}"
    networks:
      - www

  varnish:
    image: varnish:${VARNISH_VERSION}
    container_name: varnish
    environment:
      WEB_SERVER: "${WEB_SERVER}"
      VARNISH_SIZE: "${VARNISH_SIZE}"
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
          cpus: "${VARNISH_CPU}"
          memory: "${VARNISH_RAM}"       
    depends_on:
      - "${WEB_SERVER}"
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
          cpus: "${CRON_CPU}"
          memory: "${CRON_RAM}"
          pids: 100
          
  mysql:
    image: mysql:${MYSQL_VERSION}
    container_name: mysql
    restart: always
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
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
          cpus: "${MYSQL_CPU"
          memory: "${MYSQL_RAM}"
          pids: 100
    networks:
      - db
    healthcheck:
      test: ['CMD-SHELL', 'mysqladmin ping -h localhost']
      interval: 1s
      timeout: 5s
      retries: 10

  mariadb:
    image: mariadb:${MYSQL_VERSION}
    container_name: mariadb
    restart: always
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
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
          cpus: "${MARIADB_CPU}"
          memory: "${MARIADB_RAM}"
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
      - "${MYSQL_TYPE"
    image: phpmyadmin
    restart: always
    volumes:
      - html_data:/html/
    ports:
      - "${PMA_PORT}"
    deploy:
      resources:
        limits:
          cpus: "${PMA_CPU}"
          memory: "${PMA_RAM}"                
          pids: 100
    environment:
      PMA_HOST: ${MYSQL_TYPE}
      PMA_UPLOADDIR: "/html/"
      PMA_SAVEDIR: "/html/"
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
    networks:
      - db


  redis:
    image: redis:${REDIS_VERSION}
    container_name: redis
    restart: unless-stopped
    command: ["redis-server", "--bind", "0.0.0.0", "--port", "6379"]
    volumes:
      - ./sockets/redis:/var/run/redis  # Redis socket
    deploy:
      resources:
        limits:
          cpus: "${REDIS_CP}"
          memory: "${REDIS_RAM}" 
          pids: 100
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
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
      - APP_USER=root
      - APP_GROUP=root
      - APP_UID=${USER_ID}
      - APP_GID=${USER_ID}
      - ENABLE_TUNING_FPM=1
    deploy:
      resources:
        limits:
          cpus: "${PHP_FPM_5_6_CPU}"
          memory: "${PHP_FPM_5_6_RAM}"
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
      - APP_USER=root
      - APP_GROUP=root
      - APP_UID=${USER_ID}
      - APP_GID=${USER_ID}
      - ENABLE_TUNING_FPM=1
    deploy:
      resources:
        limits:
          cpus: "${PHP_FPM_8_4_CPU}"
          memory: "${PHP_FPM_8_4_RAM}"
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
          cpus: "${BACKUP_CPU}"
          memory: "${BACKUP_RAM}"
          pids: 100
    command: [ "backup" ]
    restart: unless-stopped
      
  minecraft:
    image: itzg/minecraft-server:${MINECRAFT_VERSION}
    container_name: minecraft
    tty: true
    stdin_open: true
    ports:
      - "${MINECRAFT_PORT}:25565"
    environment:
      EULA: "TRUE"
      ENABLE_QUERY: "${MINECRAFT_ENABLE_QUERY}"
      QUERY_PORT: "${MINECRAFT_PORT}"
    volumes:
      - mc_data:/data
    deploy:
      resources:
        limits:
          cpus: "${MINECRAFT_CPU}"
          memory: "${MINECRAFT_RAM}"
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
