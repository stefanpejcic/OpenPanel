import React, { useState } from "react";
import clsx from "clsx";

import { CheckCircle } from "@site/src/refine-theme/icons/check-circle";
import { CrossCircle } from "@site/src/refine-theme/icons/cross-circle";
import { EnterpriseGetInTouchButton } from "./enterprise-get-in-touch-button";
import { ArrowLeftLongIcon } from "./icons/arrow-left-long";
import { AnimatePresence, motion } from "framer-motion";

export const EnterpriseTable = ({ className }: { className?: string }) => {
    const [activeTab, setActiveTab] = useState<"directadmin" | "openpanel">(
        "openpanel",
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
                    Compare DirectAdmin VS OpenPanel
                </h2>
            </div>

            <div className={clsx("w-full", "pb-8", "relative")}>
                <AnimatePresence exitBeforeEnter>
                    <motion.div
                        key={activeTab}
                        initial={{
                            opacity: 0,
                            x: activeTab === "directadmin" ? "-100%" : "100%",
                        }}
                        animate={{ opacity: 1, x: 0 }}
                        exit={{
                            opacity: 0,
                            x: activeTab === "directadmin" ? "-100%" : "100%",
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
                            setActiveTab={(tab: "directadmin" | "openpanel") => {
                                setActiveTab(tab);
                            }}
                        />
                        <TableSectionWrapper>
                            {tableData.map((section) => {
                                const isLast =
                                    section.title === tableData.at(-1).title;

                                return (
                                    <TableSection
                                        key={section.title}
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
                "bg-refine-openpanel-table-alt",
                "dark:bg-refine-openpanel-table-alt-dark",
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
    directadmin,
    openpanel,
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
                        activeTab !== "directadmin" && "hidden landing-md:flex",
                        activeTab === "directadmin" && "flex",
                    )}
                >
                    {directadmin}
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
                        activeTab !== "openpanel" && "hidden landing-md:flex",
                        activeTab === "openpanel" && "flex",
                    )}
                >
                    {openpanel}
                </TableItemContent>
            </div>
        </div>
    );
};

const TableItem = ({
    description,
    directadmin,
    openpanel,
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
                    {...{ directadmin, openpanel, activeTab, isLast, valueType }}
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
                        if (prev === "directadmin") return "openpanel";
                        return "directadmin";
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
                    activeTab === "openpanel" &&
                        "bg-gray-200 dark:bg-gray-700 text-gray-400 dark:text-gray-500",
                    activeTab === "directadmin" &&
                        "bg-refine-blue dark:bg-refine-cyan-alt text-gray-0 dark:text-gray-900",
                )}
            >
                <ArrowLeftLongIcon
                    className={clsx(
                        "transition-transform duration-200 ease-in-out",
                        activeTab === "directadmin" && "rotate-360",
                        activeTab === "openpanel" && "rotate-180",
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
                        activeTab !== "directadmin" && "hidden landing-md:flex",
                        activeTab === "directadmin" && "flex",
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
                            activeTab === "directadmin" && "opacity-100",
                            activeTab !== "directadmin" &&
                                "opacity-0 landing-md:opacity-100",
                            "transition-opacity",
                            "ease-in-out",
                            "duration-100",
                            "dark:text-gray-400 text-gray-700",
                        )}
                    >
                        DirectAdmin
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
                        activeTab !== "openpanel" && "hidden landing-md:flex",
                        activeTab === "openpanel" && "flex",
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
                            activeTab === "openpanel" && "opacity-100",
                            activeTab !== "openpanel" &&
                                "opacity-0 landing-md:opacity-100",
                            "transition-opacity",
                            "ease-in-out",
                            "duration-100",
                            "dark:text-refine-cyan-alt dark:drop-shadow-[0_0_30px_rgba(71,235,235,0.25)]",
                            "text-refine-blue drop-shadow-[0_0_30px_rgba(51,51,255,0.3)]",
                        )}
                    >
                        OpenPanel Enterprise
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
        title: "Pricing",
        items: [
            {
                description: "Monthly Pricing",
                directadmin: <TableText>5 - 29$/ month</TableText>,
                openpanel: <TableText>14.95€ / month</TableText>,
                valueType: {
                    directadmin: "text",
                    openpanel: "text",
                },
            },
            {
                description: "Pricing model",
                directadmin: <TableText>Per account</TableText>,
                openpanel: <TableText>Per server</TableText>,
                valueType: {
                    directadmin: "text",
                    openpanel: "text",
                },
            },
        ],
    },
    {
        title: "System",
        items: [
            {
                description: "Supported OS",
                directadmin: <TableText><a href="https://docs.directadmin.com/getting-started/installation/system-requirements.html" target="_blank">AlmaLinux, Rocky, CentOS Stream, Debian, Ubuntu</a></TableText>,
                openpanel: <TableText><a href="/docs/admin/intro/#requirements">OS-agnostic: Ubuntu, Debian, AlmaLinux, Rocky, CentOS</a></TableText>,
                valueType: {
                    directadmin: "text",
                    openpanel: "text",
                },
            },
            {
                description: "Support for ARM CPUs (Aarch64)",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "User Isolation",
                directadmin: <TableText>Alpha jailed shell / needs Cloudlinux</TableText>,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "text",
                    openpanel: "icon",
                },
            },
            {
                description: "Service Isolation",
                directadmin: <TableText>Shared services (PHP, MySQL..)</TableText>,
                openpanel: <TableText>Service per user</TableText>,
                valueType: {
                    directadmin: "text",
                    openpanel: "text",
                },
            },
            {
                description: "Kernel Updates",
                directadmin: <TableText>Only with KernelCare</TableText>,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "text",
                    openpanel: "icon",
                },
            },
        ],
    },
    {
        title: "Webservers",
        items: [
            {
                description: "Apache",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Nginx",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "OpenLitespeed",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "OpenResty",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
        ],
    },
    {
        title: "Caching & Search",
        items: [
            {
                description: "Varnish Caching",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Valkey / Redis",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Memcached",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "OpenSearch / ElasticSearch",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
        ],
    },
    {
        title: "PHP",
        items: [
            {
                description: "Supported versions",
                directadmin: <TableText>5.6 - 8.3</TableText>,
                openpanel: <TableText>5.6 - 8.5</TableText>,
                valueType: {
                    directadmin: "text",
                    openpanel: "text",
                },
            },
            {
                description: "PHP Selector",
                directadmin: <TableText>Native (CustomBuild)</TableText>,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "text",
                    openpanel: "icon",
                },
            },
            {
                description: "PHP.INI Editor",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "PHP service per user",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "PHP version per domain",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
        ],
    },
    {
        title: "Docker",
        items: [
            {
                description: "Users can manage Containers",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Users can add Containers",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
        ],
    },
    {
        title: "Advanced",
        items: [
            {
                description: "API for every UI action",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Native MCP server for AI agents",
                directadmin: <TableText>Community project (unofficial)</TableText>,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "text",
                    openpanel: "icon",
                },
            },
            {
                description: "Process Manager",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Web Terminal",
                directadmin: <TableText>Per user</TableText>,
                openpanel: <TableText>Per service</TableText>,
                valueType: {
                    directadmin: "text",
                    openpanel: "text",
                },
            },
            {
                description: "Edit webserver settings",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "View Resource Usage",
                directadmin: <TableText>Limited (best with Cloudlinux)</TableText>,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "text",
                    openpanel: "icon",
                },
            },
        ],
    },
    {
        title: "Cron Jobs",
        items: [
            {
                description: "Format",
                directadmin: <TableText>Unix</TableText>,
                openpanel: <TableText>GO</TableText>,
                valueType: {
                    directadmin: "text",
                    openpanel: "text",
                },
            },
            {
                description: "Min Schedule",
                directadmin: <TableText>Minute</TableText>,
                openpanel: <TableText>Seconds</TableText>,
                valueType: {
                    directadmin: "text",
                    openpanel: "text",
                },
            },
            {
                description: "No overlap",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Edit cron file",
                directadmin: <TableText>Advanced/hidden option</TableText>,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "text",
                    openpanel: "icon",
                },
            },
            {
                description: "Logs",
                directadmin: <TableText>Email only</TableText>,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "text",
                    openpanel: "icon",
                },
            },
        ],
    },
    {
        title: "Files",
        items: [
            {
                description: "File Manager",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Upload multiple files",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Drag & Drop Upload",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "File Editor",
                directadmin: <TableText>Basic (syntax highlighting)</TableText>,
                openpanel: <TableText>Monaco (VS Code), Ace, CodeMirror</TableText>,
                valueType: {
                    directadmin: "text",
                    openpanel: "text",
                },
            },
            {
                description: "FTP",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Malware Scanner",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Download from URL",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Disk Usage Explorer",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Inodes Explorer",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Fix Permissions",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
        ],
    },
    {
        title: "Databases",
        items: [
            {
                description: "MySQL",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "MariaDB",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Percona",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "phpMyAdmin",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "PostgreSQL",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "pgAdmin",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Import Databases",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Optimize Tables",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Root Privileges",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Edit MySQL Configuration",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Manage Processes",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Switch MySQL to MariaDB",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
        ],
    },
    {
        title: "Accounts",
        items: [
            {
                description: "Account limit",
                directadmin: <TableText>Tied to tier (2/10/Unlimited)</TableText>,
                openpanel: <CrossIcon />,
                valueType: {
                    directadmin: "text",
                    openpanel: "icon",
                },
            },
            {
                description: "Multiple Administrator Accounts",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "User can change username",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Features per user",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Detailed Activity Log",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Detailed Login Log",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Add pages to favorites",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Dark Mode",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Custom message per user",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Private notes for Administrators",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Customize email templates",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
        ],
    },
    {
        title: "Domains",
        items: [
            {
                description: "Automatic SSL ",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Edit VirtualHosts file",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Change docroot",
                directadmin: <TableText>Limited for primary domain</TableText>,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "text",
                    openpanel: "icon",
                },
            },
            {
                description: "GoAccess",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "DNS Zone Editor",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Edit zone file",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "TLSA records",
                directadmin: <TableText>Disabled by default</TableText>,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "text",
                    openpanel: "icon",
                },
            },
            {
                description: "Export/Import zone",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "User can reset zone",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Comments for DNS records",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Suspend Domains",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
        ],
    },
    {
        title: "Websites",
        items: [
            {
                description: "HTML Website Builder",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "WP Manager",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Temporary Domains",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Google PageSpeed monitoring",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Google Safe Browsing monitoring",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Node / Python Environment",
                directadmin: <TableText>venv</TableText>,
                openpanel: <TableText>Containerized</TableText>,
                valueType: {
                    directadmin: "text",
                    openpanel: "text",
                },
            },
        ],
    },
    {
        title: "Emails",
        items: [
            {
                description: "Postfix",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Autologin to Webmail",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Sieve (RFC 5228) filtering",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Fail2ban",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Export Addresses",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
        ],
    },
    {
        title: "Security",
        items: [
            {
                description: "CVE Vulnerabilities",
                directadmin: <TableText><a href="https://www.cve.org/CVERecord/SearchResults?query=directadmin" target="_blank">30+</a></TableText>,
                openpanel: <TableText><a href="https://www.cve.org/CVERecord/SearchResults?query=openpanel">3 before 1.0 release</a></TableText>,
                valueType: {
                    directadmin: "text",
                    openpanel: "text",
                },
            },
            {
                description: "Granular per-user permissions",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "2FA",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Enforce 2FA for users",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Passkeys",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "IP Blocker",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Coraza WAF",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Edit WAF rules per domain",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "View WAF logs per domain",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Sentinel Firewall (CSF)",
                directadmin: <TableText>Legacy (CSF discontinued 2025)</TableText>,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "text",
                    openpanel: "icon",
                },
            },
            {
                description: "Custom ports",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Containerization",
                directadmin: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Disable admin panel",
                directadmin: <TableText>Via port/firewall only</TableText>,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "text",
                    openpanel: "icon",
                },
            },
            {
                description: "Rate-limit logins",
                directadmin: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Manage Sessions",
                directadmin: <TableText>Partial (IP-bound only)</TableText>,
                openpanel: <CheckIcon />,
                valueType: {
                    directadmin: "text",
                    openpanel: "icon",
                },
            },
        ],
    },
    {
        title: "Pricing",
        items: [
            {
                description: "Price Increase (3yr)",
                directadmin: <TableText>Repeated hikes (undisclosed %)</TableText>,
                openpanel: <CrossIcon />,
                valueType: {
                    directadmin: "text",
                    openpanel: "icon",
                },
            },
            {
                description: "Forever free tier",
                directadmin: <CrossIcon />,
                openpanel: <TableText><a href="/community">Community Edition</a></TableText>,
                valueType: {
                    directadmin: "icon",
                    openpanel: "text",
                },
            },
            {
                description: "Annual Pricing",
                directadmin: (
                    <div className={clsx("h-full")}>
                        <TableText>~16.6% off monthly (exact price undisclosed)</TableText>
                    </div>
                ),
                openpanel: (
                    <div className={clsx("flex flex-col", "gap-6", "w-full")}>
                        <TableText>149.50€ / year (17% OFF)</TableText>
                        <EnterpriseGetInTouchButton
                            className={clsx("hidden landing-sm:block")}
                            linkClassName={clsx("w-full")}
                        />
                    </div>
                ),
                valueType: {
                    directadmin: "text",
                    openpanel: "text",
                },
            },
        ],
    },
];
