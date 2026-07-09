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
    title = "Host any Docker image",
    description = `Per-user Docker Compose files let you define available services, or hand users the permission to spin up anything they want.`,
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
    image: louislam/uptime-kum:UPTIMEKUMA_VERSION
    container_name: uptimekuma
    volumes:
      - ./data:/app/data
      - /hostfs/run/user/USER_ID/docker.sock:/var/run/docker.sock:ro
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: "0.5"
          memory: "1.2G"   
          pids: 1000
    networks:
      - www

  valkey:
    image: valkey/valkey:8.0-alpine
    container_name: valkey
    restart: unless-stopped
    user: 0
    command: ["valkey-server", "--bind", "0.0.0.0", "--port", "6379"]
    volumes:
      - ./sockets/valkey:/var/run/valkey
    deploy:
      resources:
        limits:
          cpus: "0.1"
          memory: "0.1G"
          pids: 100
    healthcheck:
      test: ["CMD", "valkey-cli", "ping"]
      interval: 30s
      retries: 3
      start_period: 5s
      timeout: 5s
    networks:
      - www
     
  php-fpm-8.5:
    image: shinsenter/php:8.5-fpm
    container_name: php-fpm-8.5
    restart: unless-stopped
    volumes:
      - html_data:/var/www/html/
      - ./php.ini/8.5.ini:/usr/local/etc/php/php.ini
      - ./php.ini/8.5.ini:/usr/local/etc/php/conf.d/zz-docker-shinsenter-php.ini:ro
      - /etc/openpanel/wordpress/wp-cli.phar:/usr/local/bin/wp:ro
    environment:
      - APP_USER=root
      - APP_GROUP=root
      - APP_UID=0
      - APP_GID=0
      - ENABLE_TUNING_FPM=1
    deploy:
      resources:
        limits:
          cpus: "0.50"
          memory: "0.5G"
          pids: 350
    networks:
      - www
      - db
    command: php-fpm --allow-to-run-as-root

  mongodb:
    image: mongo:latest
    container_name: mongodb
    restart: unless-stopped
    volumes:
      - mongodb_data:/data/db
      - mongodb_config:/data/configdb
    deploy:
      resources:
        limits:
          cpus: "0.5"
          memory: "0.5G"
          pids: 1000
    networks:
      - db

  elasticsearch:
    image: elasticsearch:9.3.2
    container_name: elasticsearch
    restart: unless-stopped
    user: "0"
    environment:
      discovery.type: single-node
      xpack.security.enabled: "false"
      ES_JAVA_OPTS: "-Xms512m -Xmx512m"
    volumes:
      - es_data:/usr/share/elasticsearch/data
    ulimits:
      nproc:
        soft: 4096
        hard: 4096
      nofile:
        soft: 65535
        hard: 65535
    deploy:
      resources:
        limits:
          cpus: "0.5"
          memory: "1G"
          pids: 5000
    healthcheck:
      test: ["CMD-SHELL", "curl --silent --fail localhost:9200/_cluster/health || exit 1"]
      interval: 10s
      timeout: 10s
      start_period: 5s
      retries: 3
    networks:
      - www

  phpmyadmin:
    container_name: phpmyadmin 
    depends_on:
      - "MYSQL_TYPE"
    image: phpmyadmin
    restart: always
    volumes:
      - html_data:/html/
    ports:
      - "PMA_PORT"
    deploy:
      resources:
        limits:
          cpus: "PMA_CPU"
          memory: "PMA_RAM"                
          pids: 100
    environment:
      PMA_HOST: MYSQL_TYPE
      PMA_UPLOADDIR: "/html/"
      PMA_SAVEDIR: "/html/"
      MYSQL_ROOT_PASSWORD: MYSQL_ROOT_PASSWORD
    networks:
      - db
      
  minecraft:
    image: itzg/minecraft-server:MINECRAFT_VERSION
    container_name: minecraft
    tty: true
    stdin_open: true
    ports:
      - "MINECRAFT_PORT:25565"
    environment:
      EULA: "TRUE"
      ENABLE_QUERY: "MINECRAFT_ENABLE_QUERY"
      QUERY_PORT: "MINECRAFT_PORT"
    volumes:
      - mc_data:/data
    deploy:
      resources:
        limits:
          cpus: "MINECRAFT_CPU"
          memory: "MINECRAFT_RAM"
          pids: 100
    healthcheck:
      test: mc-health
      start_period: 1m
      interval: 5s
      retries: 20
    networks:
      - www

`;
