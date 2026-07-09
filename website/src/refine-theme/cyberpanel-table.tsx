import React, { useState } from "react";
import clsx from "clsx";

import { CheckCircle } from "@site/src/refine-theme/icons/check-circle";
import { CrossCircle } from "@site/src/refine-theme/icons/cross-circle";
import { EnterpriseGetInTouchButton } from "./enterprise-get-in-touch-button";
import { ArrowLeftLongIcon } from "./icons/arrow-left-long";
import { AnimatePresence, motion } from "framer-motion";

export const EnterpriseTable = ({ className }: { className?: string }) => {
    const [activeTab, setActiveTab] = useState<"cyberpanel" | "openpanel">(
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
                    Compare CyberPanel VS OpenPanel
                </h2>
            </div>

            <div className={clsx("w-full", "pb-8", "relative")}>
                <AnimatePresence exitBeforeEnter>
                    <motion.div
                        key={activeTab}
                        initial={{
                            opacity: 0,
                            x: activeTab === "cyberpanel" ? "-100%" : "100%",
                        }}
                        animate={{ opacity: 1, x: 0 }}
                        exit={{
                            opacity: 0,
                            x: activeTab === "cyberpanel" ? "-100%" : "100%",
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
                            setActiveTab={(tab: "cyberpanel" | "openpanel") => {
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
    cyberpanel,
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
                        activeTab !== "cyberpanel" && "hidden landing-md:flex",
                        activeTab === "cyberpanel" && "flex",
                    )}
                >
                    {cyberpanel}
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
    cyberpanel,
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
                    {...{ cyberpanel, openpanel, activeTab, isLast, valueType }}
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
                        if (prev === "cyberpanel") return "openpanel";
                        return "cyberpanel";
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
                    activeTab === "cyberpanel" &&
                        "bg-refine-blue dark:bg-refine-cyan-alt text-gray-0 dark:text-gray-900",
                )}
            >
                <ArrowLeftLongIcon
                    className={clsx(
                        "transition-transform duration-200 ease-in-out",
                        activeTab === "cyberpanel" && "rotate-360",
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
                        activeTab !== "cyberpanel" && "hidden landing-md:flex",
                        activeTab === "cyberpanel" && "flex",
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
                            activeTab === "cyberpanel" && "opacity-100",
                            activeTab !== "cyberpanel" &&
                                "opacity-0 landing-md:opacity-100",
                            "transition-opacity",
                            "ease-in-out",
                            "duration-100",
                            "dark:text-gray-400 text-gray-700",
                        )}
                    >
                        CyberPanel
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
                cyberpanel: <TableText>0 - 97$/ month</TableText>,
                openpanel: <TableText>14.95€ / month</TableText>,
                valueType: {
                    cyberpanel: "text",
                    openpanel: "text",
                },
            },
            {
                description: "Pricing model",
                cyberpanel: <TableText>Free (LSWS license optional)</TableText>,
                openpanel: <TableText>Per server</TableText>,
                valueType: {
                    cyberpanel: "text",
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
                cyberpanel: <TableText><a href="https://cyberpanel.net/KnowledgeBase/" target="_blank">Ubuntu, AlmaLinux, Rocky, CloudLinux, openEuler</a></TableText>,
                openpanel: <TableText><a href="/docs/admin/intro/#requirements">OS-agnostic: Ubuntu, Debian, AlmaLinux, Rocky, CentOS</a></TableText>,
                valueType: {
                    cyberpanel: "text",
                    openpanel: "text",
                },
            },
            {
                description: "Support for ARM CPUs (Aarch64)",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "User Isolation",
                cyberpanel: <TableText>No LVE by default (Cloudlinux optional)</TableText>,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "text",
                    openpanel: "icon",
                },
            },
            {
                description: "Service Isolation",
                cyberpanel: <TableText>Shared services (PHP, MySQL..)</TableText>,
                openpanel: <TableText>Service per user</TableText>,
                valueType: {
                    cyberpanel: "text",
                    openpanel: "text",
                },
            },
            {
                description: "Kernel Updates",
                cyberpanel: <TableText>Only with KernelCare</TableText>,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "text",
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
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Nginx",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "OpenLitespeed",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "OpenResty",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
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
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Valkey / Redis",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Memcached",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "OpenSearch / ElasticSearch",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
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
                cyberpanel: <TableText>5.6 - 8.3</TableText>,
                openpanel: <TableText>5.6 - 8.5</TableText>,
                valueType: {
                    cyberpanel: "text",
                    openpanel: "text",
                },
            },
            {
                description: "PHP Selector",
                cyberpanel: <TableText>Native selector</TableText>,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "text",
                    openpanel: "icon",
                },
            },
            {
                description: "PHP.INI Editor",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "PHP service per user",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "PHP version per domain",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
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
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Users can add Containers",
                cyberpanel: <TableText>Admin-granted access</TableText>,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "text",
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
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Native MCP server for AI agents",
                cyberpanel: <TableText>Community project (unofficial)</TableText>,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "text",
                    openpanel: "icon",
                },
            },
            {
                description: "Process Manager",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Web Terminal",
                cyberpanel: <CrossIcon />,
                openpanel: <TableText>Per service</TableText>,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "text",
                },
            },
            {
                description: "Edit webserver settings",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "View Resource Usage",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
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
                cyberpanel: <TableText>Unix</TableText>,
                openpanel: <TableText>GO</TableText>,
                valueType: {
                    cyberpanel: "text",
                    openpanel: "text",
                },
            },
            {
                description: "Min Schedule",
                cyberpanel: <TableText>Minute</TableText>,
                openpanel: <TableText>Seconds</TableText>,
                valueType: {
                    cyberpanel: "text",
                    openpanel: "text",
                },
            },
            {
                description: "No overlap",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Edit cron file",
                cyberpanel: <TableText>Form-based only</TableText>,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "text",
                    openpanel: "icon",
                },
            },
            {
                description: "Logs",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
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
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Upload multiple files",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Drag & Drop Upload",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "File Editor",
                cyberpanel: <TableText>Basic code editor</TableText>,
                openpanel: <TableText>Monaco (VS Code), Ace, CodeMirror</TableText>,
                valueType: {
                    cyberpanel: "text",
                    openpanel: "text",
                },
            },
            {
                description: "FTP",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Malware Scanner",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Download from URL",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Disk Usage Explorer",
                cyberpanel: <TableText>Overall usage only</TableText>,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "text",
                    openpanel: "icon",
                },
            },
            {
                description: "Inodes Explorer",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Fix Permissions",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
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
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "MariaDB",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Percona",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "phpMyAdmin",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "PostgreSQL",
                cyberpanel: <TableText>Manual/CLI setup only</TableText>,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "text",
                    openpanel: "icon",
                },
            },
            {
                description: "pgAdmin",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Import Databases",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Optimize Tables",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Root Privileges",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Edit MySQL Configuration",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Manage Processes",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Switch MySQL to MariaDB",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
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
                cyberpanel: <TableText>Tied to LSWS license tier</TableText>,
                openpanel: <CrossIcon />,
                valueType: {
                    cyberpanel: "text",
                    openpanel: "icon",
                },
            },
            {
                description: "Multiple Administrator Accounts",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "User can change username",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Features per user",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Detailed Activity Log",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Detailed Login Log",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Add pages to favorites",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Dark Mode",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Custom message per user",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Private notes for Administrators",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Customize email templates",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
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
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Edit VirtualHosts file",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Change docroot",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "GoAccess",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "DNS Zone Editor",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Edit zone file",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "TLSA records",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Export/Import zone",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "User can reset zone",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Comments for DNS records",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Suspend Domains",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
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
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "WP Manager",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Temporary Domains",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Google PageSpeed monitoring",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Google Safe Browsing monitoring",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Node / Python Environment",
                cyberpanel: <TableText>venv</TableText>,
                openpanel: <TableText>Containerized</TableText>,
                valueType: {
                    cyberpanel: "text",
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
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Autologin to Webmail",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Sieve (RFC 5228) filtering",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Fail2ban",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Export Addresses",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
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
                cyberpanel: <TableText><a href="https://www.cve.org/CVERecord/SearchResults?query=cyberpanel" target="_blank">10+</a></TableText>,
                openpanel: <TableText><a href="https://www.cve.org/CVERecord/SearchResults?query=openpanel">3 before 1.0 release</a></TableText>,
                valueType: {
                    cyberpanel: "text",
                    openpanel: "text",
                },
            },
            {
                description: "Granular per-user permissions",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "2FA",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Enforce 2FA for users",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Passkeys",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "IP Blocker",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Coraza WAF",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Edit WAF rules per domain",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "View WAF logs per domain",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Sentinel Firewall (CSF)",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Custom ports",
                cyberpanel: <CheckIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Containerization",
                cyberpanel: <TableText>Partial (cgroups only)</TableText>,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "text",
                    openpanel: "icon",
                },
            },
            {
                description: "Disable admin panel",
                cyberpanel: <TableText>Via port/firewall only</TableText>,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "text",
                    openpanel: "icon",
                },
            },
            {
                description: "Rate-limit logins",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "icon",
                },
            },
            {
                description: "Manage Sessions",
                cyberpanel: <CrossIcon />,
                openpanel: <CheckIcon />,
                valueType: {
                    cyberpanel: "icon",
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
                cyberpanel: <TableText>No documented history</TableText>,
                openpanel: <CrossIcon />,
                valueType: {
                    cyberpanel: "text",
                    openpanel: "icon",
                },
            },
            {
                description: "Forever free tier",
                cyberpanel: <CheckIcon />,
                openpanel: <TableText><a href="/community">Community Edition</a></TableText>,
                valueType: {
                    cyberpanel: "icon",
                    openpanel: "text",
                },
            },
            {
                description: "Annual Pricing",
                cyberpanel: (
                    <div className={clsx("h-full")}>
                        <TableText>0 - 1164$ / year</TableText>
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
                    cyberpanel: "text",
                    openpanel: "text",
                },
            },
        ],
    },
];
