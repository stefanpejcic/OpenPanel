import React, { useState } from "react";
import clsx from "clsx";
import Translate from "@docusaurus/Translate";

import { CheckCircle } from "@site/src/refine-theme/icons/check-circle";
import { CrossCircle } from "@site/src/refine-theme/icons/cross-circle";
import { EnterpriseGetInTouchButton } from "./enterprise-get-in-touch-button";
import { ArrowLeftLongIcon } from "./icons/arrow-left-long";
import { AnimatePresence, motion } from "framer-motion";

export const EnterpriseTable = ({ className }: { className?: string }) => {
    const [activeTab, setActiveTab] = useState<"community" | "enterprise">(
        "enterprise",
    );

    return (
        <div
            className={clsx(
                "flex flex-col",
                "not-prose",
                "min-h-[1556px] landing-sm:min-h-[1444px] landing-md:min-h-min",
                "w-full",
                className,
            )}
        >
            <div
                className={clsx(
                    "pl-4 landing-sm:pl-6 landing-md:pl-10",
                    "text-2xl landing-sm:text-[32px] landing-sm:leading-[40px]",
                )}
            >
                <h2
                    id="compare"
                    className={clsx(
                        "font-semibold",
                        "dark:text-gray-400 text-gray-600",
                    )}
                >
                    <Translate id="enterprise.table.compare">Compare editions</Translate>
                </h2>
            </div>

            <div className={clsx("w-full", "pb-8", "relative")}>
                <AnimatePresence exitBeforeEnter>
                    <motion.div
                        key={activeTab}
                        initial={{
                            opacity: 0,
                            x: activeTab === "community" ? "-100%" : "100%",
                        }}
                        animate={{ opacity: 1, x: 0 }}
                        exit={{
                            opacity: 0,
                            x: activeTab === "community" ? "-100%" : "100%",
                        }}
                        transition={{
                            duration: 0.3,
                            bounce: 0,
                            ease: "easeInOut",
                        }}
                        className={clsx("relative", "w-full", "pt-8")}
                    >
                        <TableStickyHeader
                            activeTab={activeTab}
                            setActiveTab={(tab: "community" | "enterprise") => {
                                setActiveTab(tab);
                            }}
                        />
                        <TableSectionWrapper>
                            {tableData.map((section) => {
                                const isLast =
                                    section.title === tableData.at(-1).title;

                                return (
                                    <TableSection
                                        key={section.titleId}
                                        title={section.title}
                                        activeTab={activeTab}
                                        items={section.items}
                                        isLast={isLast}
                                    />
                                );
                            })}
                        </TableSectionWrapper>
                    </motion.div>
                </AnimatePresence>
            </div>
            <EnterpriseGetInTouchButton
                className={clsx("block landing-sm:hidden", "mx-auto")}
                linkClassName={clsx("w-[344px]", "mx-auto")}
            />
        </div>
    );
};

const TableItemHeading = ({ children }) => {
    return (
        <div
            className={clsx(
                "pl-4 landing-sm:pl-10 pb-4 pt-4 pr-4",
                "bg-refine-enterprise-table-alt",
                "dark:bg-refine-enterprise-table-alt-dark",
                "border-b",
                "border-gray-200",
                "dark:border-gray-700",
                "text-gray-900",
                "dark:text-gray-0",
                "text-base",
                "font-semibold",
            )}
        >
            {children}
        </div>
    );
};

const TableItemContent = ({ children, valueType, activeTab, className }) => {
    const isIcon = valueType[activeTab] === "icon";

    return (
        <div
            className={clsx(
                isIcon && "w-auto",
                !isIcon && "w-[560px]",
                "landing-sm:w-[360px]",
                "landing-md:w-[296px]",
                "landing-lg:w-[396px]",
                "landing-sm:py-4",
                "landing-sm:px-6",
                "flex",
                "items-start",
                "justify-start",
                "gap-2",
                className,
            )}
        >
            {children}
        </div>
    );
};

const TableItemDescription = ({ children, isLast }) => {
    return (
        <div
            className={clsx(
                "min-w-[296px] landing-md:min-w-auto",
                "landing-md:flex-1",
                "text-base",
                "flex",
                "text-gray-700",
                "dark:text-gray-400",
                "landing-sm:pl-10 landing-sm:pb-4 landing-sm:pt-4 landing-sm:pr-4",
                !isLast && "landing-sm:border-r",
                !isLast && "landing-sm:border-gray-200",
                !isLast && "landing-sm:dark:border-gray-700",
            )}
        >
            {children}
        </div>
    );
};

const TableItemContentGroup = ({
    community,
    enterprise,
    activeTab,
    isLast,
    valueType,
}) => {
    const isIcon = valueType[activeTab] === "icon";

    return (
        <div
            className={clsx(
                "flex-shrink-0",
                "flex",
                "w-auto",
                isIcon && "ml-auto",
                isLast && "landing-sm:border-b",
                isLast && "landing-sm:border-l",
                isLast && "landing-sm:border-gray-200",
                isLast && "landing-sm:dark:border-gray-700",
                isLast && "landing-sm:rounded-bl-3xl",
                isLast && "landing-sm:rounded-br-3xl",
                "overflow-hidden",
            )}
        >
            <div className={clsx("flex", "flex-shrink-0")}>
                <TableItemContent
                    activeTab={activeTab}
                    valueType={valueType}
                    className={clsx(
                        activeTab !== "community" && "hidden landing-md:flex",
                        activeTab === "community" && "flex",
                    )}
                >
                    {community}
                </TableItemContent>
                <div
                    className={clsx(
                        "w-px h-full",
                        "bg-gray-200",
                        "dark:bg-gray-700",
                        "flex-shrink-0",
                        "hidden",
                        "landing-md:block",
                    )}
                />
                <TableItemContent
                    activeTab={activeTab}
                    valueType={valueType}
                    className={clsx(
                        activeTab !== "enterprise" && "hidden landing-md:flex",
                        activeTab === "enterprise" && "flex",
                    )}
                >
                    {enterprise}
                </TableItemContent>
            </div>
        </div>
    );
};

const TableItem = ({
    description,
    community,
    enterprise,
    activeTab,
    isLast,
    valueType,
}) => {
    return (
        <div
            className={clsx(
                !isLast && "border-b",
                !isLast && "border-b-gray-200",
                !isLast && "dark:border-b-gray-700",
                "w-full",
                "p-4 landing-sm:p-0",
                "overflow-hidden",
            )}
        >
            <div
                key={activeTab}
                className={clsx(
                    "flex",
                    valueType[activeTab] === "icon" && "flex-row",
                    valueType[activeTab] !== "icon" &&
                    "flex-col landing-sm:flex-row",
                    "gap-2 landing-sm:gap-0",
                )}
            >
                <TableItemDescription isLast={isLast}>
                    {description}
                </TableItemDescription>
                <TableItemContentGroup
                    {...{ community, enterprise, activeTab, isLast, valueType }}
                />
            </div>
        </div>
    );
};

const TableSection = ({ title, items, activeTab, isLast = false }) => {
    return (
        <>
            <TableItemHeading>{title}</TableItemHeading>
            {items.map((item, index) => (
                <TableItem
                    key={index}
                    activeTab={activeTab}
                    {...item}
                    isLast={isLast && index === items.length - 1}
                />
            ))}
        </>
    );
};

const TableTabs = ({ activeTab, setActiveTab }) => {
    return (
        <div
            className={clsx(
                "flex",
                "flex-shrink-0",
                "rounded-tl-2xl landing-md:rounded-tl-3xl",
                "rounded-tr-2xl landing-md:rounded-tr-3xl",
                "border",
                "border-gray-200",
                "dark:border-gray-700",
                "border-b-0",
                "bg-gray-0",
                "dark:bg-gray-900",
                "w-full landing-sm:w-[360px]",
                "overflow-hidden",
                "landing-md:w-auto",
                "relative",
            )}
        >
            <button
                type="button"
                onClick={() => {
                    setActiveTab((prev) => {
                        if (prev === "community") return "enterprise";
                        return "community";
                    });
                }}
                className={clsx(
                    "transition-opacity duration-200 ease-in-out",
                    "z-[1]",
                    "flex",
                    "landing-md:hidden",
                    "absolute",
                    "right-6",
                    "top-4",
                    "rounded-full",
                    "w-12 h-6",
                    "items-center justify-center",
                    activeTab === "enterprise" &&
                    "bg-gray-200 dark:bg-gray-700 text-gray-400 dark:text-gray-500",
                    activeTab === "community" &&
                    "bg-refine-blue dark:bg-refine-cyan-alt text-gray-0 dark:text-gray-900",
                )}
            >
                <ArrowLeftLongIcon
                    className={clsx(
                        "transition-transform duration-200 ease-in-out",
                        activeTab === "community" && "rotate-360",
                        activeTab === "enterprise" && "rotate-180",
                    )}
                />
            </button>
            <div
                className={clsx("relative", "w-full", "flex", "flex-shrink-0")}
            >
                <div
                    className={clsx(
                        "w-full",
                        "landing-md:w-[296px]",
                        "landing-lg:w-[396px]",
                        "p-4 landing-md:p-6",
                        activeTab !== "community" && "hidden landing-md:flex",
                        activeTab === "community" && "flex",
                        "items-center",
                        "justify-start",
                        "text-center",
                        "text-base",
                        "landing-sm:text-2xl",
                    )}
                >
                    <div
                        className={clsx(
                            "font-normal",
                            "whitespace-nowrap",
                            activeTab === "community" && "opacity-100",
                            activeTab !== "community" &&
                            "opacity-0 landing-md:opacity-100",
                            "transition-opacity",
                            "ease-in-out",
                            "duration-100",
                            "dark:text-gray-400 text-gray-700",
                        )}
                    >
                        Community edition
                    </div>
                </div>
                <div
                    className={clsx(
                        "w-px",
                        "h-full",
                        "bg-gray-200",
                        "dark:bg-gray-700",
                        "hidden",
                        "landing-md:block",
                        "flex-shrink-0",
                    )}
                />
                <div
                    className={clsx(
                        "w-full",
                        "landing-md:w-[296px]",
                        "landing-lg:w-[396px]",
                        "p-4 landing-md:p-6",
                        activeTab !== "enterprise" && "hidden landing-md:flex",
                        activeTab === "enterprise" && "flex",
                        "items-center",
                        "justify-start",
                        "text-center",
                        "text-base",
                        "landing-sm:text-2xl",
                    )}
                >
                    <div
                        className={clsx(
                            "font-semibold",
                            "whitespace-nowrap",
                            activeTab === "enterprise" && "opacity-100",
                            activeTab !== "enterprise" &&
                            "opacity-0 landing-md:opacity-100",
                            "transition-opacity",
                            "ease-in-out",
                            "duration-100",
                            "dark:text-refine-cyan-alt dark:drop-shadow-[0_0_30px_rgba(71,235,235,0.25)]",
                            "text-refine-blue drop-shadow-[0_0_30px_rgba(51,51,255,0.3)]",
                        )}
                    >
                        <Translate id="enterprise.table.enterprise_edition">Enterprise edition</Translate>
                    </div>
                </div>
            </div>
        </div>
    );
};

const TableStickyHeader = ({ activeTab, setActiveTab }) => {
    return (
        <div
            className={clsx(
                "sticky",
                "z-[1]",
                "top-[64px]",
                "pt-4",
                "flex",
                "justify-end",
                "w-full",
                "border-b",
                "bg-gray-0",
                "dark:bg-gray-900",
                "border-b-gray-200",
                "dark:border-b-gray-700",
            )}
        >
            <div className={clsx("ml-px", "flex-1", "backdrop-blur-[3px]")} />
            <TableTabs {...{ activeTab, setActiveTab }} />
        </div>
    );
};

const TableSectionWrapper = ({ children }) => {
    return (
        <div
            className={clsx(
                "flex",
                "flex-col",
                "border-r",
                "border-b landing-sm:border-b-0",
                "border-l landing-sm:border-l-0",
                "rounded-br-2xl landing-md:rounded-br-3xl",
                "rounded-bl-2xl",
                "border-gray-200",
                "dark:border-gray-700",
            )}
        >
            {children}
        </div>
    );
};

const CheckIcon = () => {
    return (
        <CheckCircle
            className={clsx(
                "dark:text-refine-green-alt text-refine-green",
                "landing-sm:mx-auto",
            )}
        />
    );
};

const CrossIcon = () => {
    return (
        <CrossCircle className={clsx("text-gray-500", "landing-sm:mx-auto")} />
    );
};

const TableText = ({ children }) => {
    return (
        <div
            className={clsx(
                "text-start",
                "text-xs landing-sm:text-base",
                "dark:text-gray-0 text-gray-900",
            )}
        >
            {children}
        </div>
    );
};

const tableData = [
    {
        titleId: "pricing",
        title: <Translate id="enterprise.table.section.pricing">Pricing</Translate>,
        items: [
            {
                description: <Translate id="enterprise.table.section.pricing">Pricing</Translate>,
                community: (
                    <TableText><Translate id="enterprise.table.forever_free">forever free</Translate></TableText>
                ),
                enterprise: (
                    <TableText>14.95€ / month</TableText>
                ),
                valueType: {
                    community: "text",
                    enterprise: "text",
                },
            },
        ],
    },
    {
        titleId: "support",
        title: <Translate id="enterprise.table.section.support">Support</Translate>,
        items: [
            {
                description: <Translate id="enterprise.table.support_channels">Support Channels</Translate>,
                community: <TableText><Translate id="enterprise.table.support_channels.community">Community Forums & Discord</Translate></TableText>,
                enterprise: <TableText><Translate id="enterprise.table.support_channels.enterprise">Ticketing & Hands-on support</Translate></TableText>,
                valueType: {
                    community: "text",
                    enterprise: "text",
                },
            },
            {
                description: "SLA",
                community: <CrossIcon />,
                enterprise: (
                    <TableText><Translate id="enterprise.table.sla.desc">Response time within one hour</Translate></TableText>
                ),
                valueType: {
                    community: "icon",
                    enterprise: "text",
                },
            },
        ],
    },
    {
        titleId: "websites",
        title: <Translate id="enterprise.table.section.websites">Websites</Translate>,
        items: [
            {
                description: <Translate id="enterprise.table.section.websites">Websites</Translate>,
                community: <TableText>50</TableText>,
                enterprise: (
                    <TableText>
                        <Translate id="enterprise.table.websites.unlimited">Unlimited</Translate>
                    </TableText>
                ),
                valueType: {
                    community: "text",
                    enterprise: "text",
                },
            },
            {
                description: <Translate id="enterprise.table.wordpress_manager">WordPress Manager</Translate>,
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.website_builder">Website Builder</Translate>,
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.nodejs_apps">NodeJS Applications</Translate>,
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.python_apps">Python Applications</Translate>,
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.php_selector">PHP Selector</Translate>,
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.php_ini_editor">PHP.INI Editor</Translate>,
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.temp_domains">Temporary Domains</Translate>,
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.varnish_caching">Varnish Caching</Translate>,
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
        ],
    },
    {
        titleId: "files",
        title: <Translate id="enterprise.table.section.files">Files</Translate>,
        items: [
            {
                description: <Translate id="enterprise.table.file_manager">File Manager</Translate>,
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "FTP",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.malware_scanner">Malware Scanner</Translate>,
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.download_url">Download from URL</Translate>,
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.disk_usage">Disk Usage Explorer</Translate>,
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.inodes">Inodes Explorer</Translate>,
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
        ],
    },
    {
        titleId: "databases",
        title: <Translate id="enterprise.table.section.databases">Databases</Translate>,
        items: [
            {
                description: "MySQL",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "phpMyAdmin",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "MariaDB",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Percona",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "PostgreSQL",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "pgAdmin",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.import_db">Import Databases</Translate>,
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.remote_access">Remote Access</Translate>,
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
        ],
    },



    {
        titleId: "webservers",
        title: <Translate id="enterprise.table.section.webservers">Web servers</Translate>,
        items: [
            {
                description: "Apache",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Nginx",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "OpenResty",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "OpenLitespeed",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Litespeed",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Varnish",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
        ],
    },
    {
        titleId: "accounts",
        title: "Accounts",
        items: [
            {
                description: <Translate id="enterprise.table.user_accounts">User Accounts</Translate>,
                community: <TableText>3</TableText>,
                enterprise: (
                    <TableText>
                        <Translate id="enterprise.table.websites.unlimited">Unlimited</Translate>
                    </TableText>
                ),
                valueType: {
                    community: "text",
                    enterprise: "text",
                },
            },
            {
                description: <Translate id="enterprise.table.resellers">Resellers</Translate>,
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.multi_admin">Multiple Administrator Accounts</Translate>,
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.transfer_accounts">Transfer Accounts between Servers</Translate>,
                community: <CrossIcon />,
                enterprise: <CheckIcon />,

                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.import_cpanel">Import from a cPanel Backup</Translate>,
                community: <CrossIcon />,
                enterprise: <CheckIcon />,

                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },

        ],
    },
    {
        titleId: "domains",
        title: "Domains",
        items: [
            {
                description: <Translate id="enterprise.table.total_domains">Total Domains</Translate>,
                community: <TableText>50</TableText>,
                enterprise: <TableText><Translate id="enterprise.table.websites.unlimited">Unlimited</Translate></TableText>,
                valueType: {
                    community: "text",
                    enterprise: "text",
                },
            },
            {
                description: "Let's Encrypt",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.custom_ssl">Custom SSL</Translate>,
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Coraza WAF",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "DNSSEC",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "DNS Clustering",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.domain_redirects">Domain Redirects</Translate>,
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: ".onion Domains",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
        ],
    },
    {
        titleId: "emails",
        title: "Emails",
        items: [
            {
                description: <Translate id="enterprise.table.email_accounts">Email Accounts</Translate>,
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Webmail",
                community: <CrossIcon />,
                enterprise: <TableText>RoundCube</TableText>,
                valueType: {
                    community: "icon",
                    enterprise: "text",
                },
            },
        ],
    },
    {
        titleId: "integrations",
        title: "Integrations",
        items: [
            {
                description: <Translate id="enterprise.table.api_access">API Access</Translate>,
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.billing_integrations">Billing Integrations</Translate>,
                community: <CrossIcon />,
                enterprise: <TableText><a href="/docs/articles/extensions/openpanel-and-whmcs/">WHMCS</a>, <a href="/docs/articles/extensions/openpanel-and-fossbilling/">FOSSBilling</a>, <a href="/docs/articles/extensions/openpanel-and-blesta/">Blesta</a>, Paymenter.org</TableText>,
                valueType: {
                    community: "icon",
                    enterprise: "text",
                },
            },
        ],
    },
    {
        titleId: "system",
        title: "System",
        items: [
            {
                description: <Translate id="enterprise.table.user_isolation">User Isolation</Translate>,
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.services_isolation">Services Isolation</Translate>,
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.manage_containers">Users can manage Containers</Translate>,
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.add_containers">Users can add Containers</Translate>,
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.manage_backups">User manages Backups</Translate>,
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.arm_support">Support for ARM CPUs (Aarch64)</Translate>,
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.cloud_vps">Available on Cloud/VPS</Translate>,
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.metal_dedicated">Available on Metal/Dedicated</Translate>,
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
        ],
    },
    {
        titleId: "os",
        title: "OS & Virtualization",
        items: [
            {
                description: <Translate id="enterprise.table.docker_support">Docker Support</Translate>,
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.metal_dedicated_support">Metal/Dedicated Support</Translate>,
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: <Translate id="enterprise.table.supported_os">Supported OS</Translate>,
                community: <TableText><a href="/docs/admin/intro/#requirements">Ubuntu 24.04</a></TableText>,
                enterprise: <TableText><a href="/docs/admin/intro/#requirements">Ubuntu, Debian, AlmaLinux, Rocky, CentOS</a></TableText>,
                valueType: {
                    community: "text",
                    enterprise: "text",
                },
            },
        ],
    },
    {
        titleId: "yearly-pricing",
        title: <Translate id="enterprise.table.section.pricing">Pricing</Translate>,
        items: [
            {
                description: <Translate id="enterprise.table.yearly_pricing">Yearly Pricing</Translate>,
                community: (
                    <div className={clsx("h-full")}>
                        <TableText> </TableText>
                    </div>
                ),
                enterprise: (
                    <div className={clsx("flex flex-col", "gap-6", "w-full")}>
                        <TableText>149.50€ / year (17% OFF)</TableText>
                        <EnterpriseGetInTouchButton
                            className={clsx("hidden landing-sm:block")}
                            linkClassName={clsx("w-full")}
                        />
                    </div>
                ),
                valueType: {
                    community: "text",
                    enterprise: "text",
                },
            },
        ],
    },
];
