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
                    <BrowserOnly>{() => <ReactLogo />}</BrowserOnly>
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
                                Compare OpenPanel
                            </LandingSectionCtaButton>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};

const ReactLogo = () => {
    const { colorMode } = useColorMode();

    return (
        <div
            key={colorMode}
            className={clsx(
                "w-[48px] h-[48px]",
                "absolute",
                "bottom-[16px]",
                "right-[16px]",
                "rounded-lg",
                "z-0",
            )}
        >
            <video autoPlay loop muted playsInline className="w-full h-full">
                <source
                    src={`https://refine.ams3.cdn.digitaloceanspaces.com/website/static/assets/react-${colorMode}.mov`}
                    type="video/mp4"
                />
                <source
                    key={colorMode}
                    src={`https://refine.ams3.cdn.digitaloceanspaces.com/website/static/assets/react-${colorMode}.webm`}
                    type="video/webm"
                />
            </video>
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
FROM ubuntu:22.04

ENV TZ=UTC
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && \
    apt-get install -y software-properties-common && \
    add-apt-repository ppa:ondrej/php && \
    apt-get update && \
    apt-get install --no-install-recommends -y \
        ttyd \
        screen \
        nginx \
        mysql-server \
        php8.2-fpm \
        php8.2-mysql \
        php8.2-curl \
        php8.2-gd \
        php8.2-mbstring \
        php8.2-xml \
        php8.2-xmlrpc \
        php8.2-soap \
        php8.2-intl \
        php8.2-zip \
        php8.2-bcmath \
        php8.2-calendar \
        php8.2-exif \
        php8.2-ftp \
        php8.2-ldap \
        php8.2-sockets \
        php8.2-sysvmsg \
        php8.2-sysvsem \
        php8.2-sysvshm \
        php8.2-tidy \
        php8.2-uuid \
        php8.2-opcache \
        php8.2-redis \
        php8.2-memcached \
        php8.2-imagick \
        curl \
        pwgen \
        zip \
        unzip \
        wget \
        cron \
        phpmyadmin \
        openssh-server \
        nodejs \
        npm \
        php-mbstring \
        python3.9 && \
        apt-get clean && \
        apt-get autoremove -y && \
        rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN npm install pm2 -g

RUN sed -i 's/localhost/127.0.0.1/g' /etc/mysql/mysql.conf.d/mysqld.cnf

EXPOSE 22 80 3306 7681 8080

RUN sed -i 's/# server_names_hash_bucket_size 64;/server_names_hash_bucket_size 64;/' /etc/nginx/nginx.conf

ENV NOTVISIBLE "in users profile"
RUN echo "export VISIBLE=now" >> /etc/profile

RUN sed -i 's/localhost/127.0.0.1/g' /etc/phpmyadmin/config-db.php

COPY config.inc.php /etc/phpmyadmin/
COPY pma.php /usr/share/phpmyadmin/pma.php

RUN sed -i 's/^upload_max_filesize = .*/upload_max_filesize = 1024M/' /etc/php/8.2/fpm/php.ini
RUN sed -i 's/^max_input_time = .*/max_input_time = 600/' /etc/php/8.2/fpm/php.ini
RUN sed -i 's/^memory_limit = .*/memory_limit = -1/' /etc/php/8.2/fpm/php.ini
RUN sed -i 's/^post_max_size = .*/post_max_size = 1024M/' /etc/php/8.2/fpm/php.ini
RUN sed -i 's/^max_execution_time = .*/max_execution_time = 600/' /etc/php/8.2/fpm/php.ini
RUN sed -i 's/^opcache.enable= .*/opcache.enable=1/' /etc/php/8.2/fpm/php.ini

RUN curl -O https://raw.githubusercontent.com/wp-cli/builds/gh-pages/phar/wp-cli.phar && \
    chmod +x wp-cli.phar && \
    mv wp-cli.phar /usr/local/bin/wp

RUN rm -rf /var/cache/apk/* /tmp/* /var/tmp/*

COPY entrypoint.sh /etc/entrypoint.sh
RUN chmod +x /etc/entrypoint.sh
CMD ["/bin/sh", "-c", "/etc/entrypoint.sh ; tail -f /dev/null"]
`;
