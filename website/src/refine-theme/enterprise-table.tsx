import React, {
    createContext,
    useContext,
    useEffect,
    useId,
    useMemo,
    useRef,
    useState,
} from "react";
import { createPortal } from "react-dom";
import clsx from "clsx";

import { CheckCircle } from "@site/src/refine-theme/icons/check-circle";
import { CrossCircle } from "@site/src/refine-theme/icons/cross-circle";
import { EnterpriseGetInTouchButton } from "./enterprise-get-in-touch-button";
import { ArrowLeftLongIcon } from "./icons/arrow-left-long";
import { AnimatePresence, motion } from "framer-motion";
import {
    Feature,
    formatDescription,
    useFeatures,
} from "./plan-features-list";

// features.json entries by name, rows reference them with `feature`
const FeaturesContext = createContext<Record<string, Feature>>({});

export const EnterpriseTable = ({ className }: { className?: string }) => {
    const [activeTab, setActiveTab] = useState<"community" | "enterprise">(
        "enterprise",
    );
    const features = useFeatures();
    const featuresByName = useMemo(
        () => Object.fromEntries(features.map((f) => [f.name, f])),
        [features],
    );

    return (
        <FeaturesContext.Provider value={featuresByName}>
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
                    Compare editions
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
        </FeaturesContext.Provider>
    );
};

const TOOLTIP_WIDTH = 320;
const TOOLTIP_GAP = 10;

// hover card with the features.json description and docs link, rendered in a
// portal because table rows clip overflow
export const FeatureTooltip = ({
    features,
    children,
}: {
    features: Feature[];
    children: React.ReactNode;
}) => {
    const id = useId();
    const anchorRef = useRef<HTMLSpanElement>(null);
    const cardRef = useRef<HTMLDivElement>(null);
    const closeTimer = useRef<ReturnType<typeof setTimeout>>();
    // opened by a tap: ignore the emulated mouseleave/blur, close on outside tap
    const pinned = useRef(false);
    const [position, setPosition] = useState<{
        top: number;
        left: number;
        arrowLeft: number;
        above: boolean;
    } | null>(null);

    const open = () => {
        clearTimeout(closeTimer.current);
        const rect = anchorRef.current.getBoundingClientRect();
        const above =
            window.innerHeight - rect.bottom < 240 && rect.top > 240;
        const left = Math.max(
            16,
            Math.min(rect.left, window.innerWidth - TOOLTIP_WIDTH - 16),
        );
        setPosition({
            top: above ? rect.top - TOOLTIP_GAP : rect.bottom + TOOLTIP_GAP,
            left,
            arrowLeft: Math.min(rect.left - left + 16, TOOLTIP_WIDTH - 24),
            above,
        });
    };

    // short delay so the pointer can move from the row onto the card
    const close = () => {
        if (pinned.current) return;
        clearTimeout(closeTimer.current);
        closeTimer.current = setTimeout(() => setPosition(null), 150);
    };

    useEffect(() => {
        if (!position) return;

        const hide = () => {
            pinned.current = false;
            setPosition(null);
        };
        const hideOutside = (event: PointerEvent) => {
            const target = event.target as Node;
            if (
                !anchorRef.current?.contains(target) &&
                !cardRef.current?.contains(target)
            ) {
                hide();
            }
        };
        const hideOnEscape = (event: KeyboardEvent) => {
            if (event.key === "Escape") hide();
        };

        window.addEventListener("scroll", hide, { passive: true });
        window.addEventListener("resize", hide);
        document.addEventListener("pointerdown", hideOutside);
        document.addEventListener("keydown", hideOnEscape);
        return () => {
            window.removeEventListener("scroll", hide);
            window.removeEventListener("resize", hide);
            document.removeEventListener("pointerdown", hideOutside);
            document.removeEventListener("keydown", hideOnEscape);
        };
    }, [position]);

    useEffect(() => () => clearTimeout(closeTimer.current), []);

    return (
        <>
            <span
                ref={anchorRef}
                tabIndex={0}
                aria-describedby={position ? id : undefined}
                onPointerDown={(event) => {
                    pinned.current = event.pointerType !== "mouse";
                }}
                onMouseEnter={open}
                onMouseLeave={close}
                onFocus={open}
                onBlur={close}
                onClick={open}
                className={clsx(
                    "cursor-help",
                    "underline decoration-dotted underline-offset-4",
                    "decoration-gray-400 dark:decoration-gray-500",
                    "outline-none focus-visible:ring-2 focus-visible:ring-refine-blue rounded-sm",
                )}
            >
                {children}
            </span>
            {position &&
                createPortal(
                    <div
                        ref={cardRef}
                        id={id}
                        role="tooltip"
                        onMouseEnter={() => clearTimeout(closeTimer.current)}
                        onMouseLeave={close}
                        style={{
                            top: position.top,
                            left: position.left,
                            width: TOOLTIP_WIDTH,
                            transform: position.above
                                ? "translateY(-100%)"
                                : undefined,
                        }}
                        className={clsx(
                            "fixed z-50",
                            "max-w-[calc(100vw-32px)]",
                            "flex flex-col gap-3",
                            "p-4",
                            "rounded-xl",
                            "border border-gray-200 dark:border-gray-700",
                            "bg-gray-0 dark:bg-gray-800",
                            "shadow-lg",
                            "text-sm",
                        )}
                    >
                        <span
                            style={{ left: position.arrowLeft }}
                            className={clsx(
                                "absolute",
                                "w-3 h-3",
                                "rotate-45",
                                "bg-gray-0 dark:bg-gray-800",
                                "border-gray-200 dark:border-gray-700",
                                position.above
                                    ? "-bottom-[7px] border-r border-b"
                                    : "-top-[7px] border-l border-t",
                            )}
                        />
                        {features.map((feature) => (
                            <div key={feature.name}>
                                <div
                                    className={clsx(
                                        "text-xs",
                                        "font-semibold",
                                        "uppercase",
                                        "tracking-wide",
                                        "text-refine-blue dark:text-refine-cyan-alt",
                                    )}
                                >
                                    {feature.title}
                                </div>
                                <p
                                    className={clsx(
                                        "mt-1 mb-0",
                                        "text-gray-700 dark:text-gray-300",
                                    )}
                                >
                                    {formatDescription(feature.description)}
                                </p>
                                {feature.help_link && (
                                    <a
                                        href={feature.help_link}
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        className={clsx(
                                            "inline-block",
                                            "mt-2",
                                            "font-semibold",
                                            "text-refine-blue dark:text-refine-cyan-alt",
                                        )}
                                    >
                                        Read the docs →
                                    </a>
                                )}
                            </div>
                        ))}
                    </div>,
                    document.body,
                )}
        </>
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
    feature,
    info,
    community,
    enterprise,
    activeTab,
    isLast,
    valueType,
}: {
    description: string;
    feature?: string | string[];
    // tooltip for a row that isn't an OpenPanel module in features.json
    info?: Feature;
    community: React.ReactNode;
    enterprise: React.ReactNode;
    activeTab: "community" | "enterprise";
    isLast: boolean;
    valueType: Record<string, string>;
}) => {
    const featuresByName = useContext(FeaturesContext);
    const details = [feature ?? []]
        .flat()
        .map((name) => featuresByName[name])
        .filter(Boolean)
        .concat(info ? [info] : []);

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
                    {details.length > 0 ? (
                        <FeatureTooltip features={details}>
                            {description}
                        </FeatureTooltip>
                    ) : (
                        description
                    )}
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
                        Enterprise edition
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
                description: "Pricing",
                community: (
                    <TableText>forever free</TableText>
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
        title: "Support",
        items: [
            {
                description: "Support Channels",
                community: <TableText>GitHub Discussions & Discord</TableText>,
                enterprise: <TableText>Ticketing & Hands-on support</TableText>,
                valueType: {
                    community: "text",
                    enterprise: "text",
                },
            },
            {
                description: "SLA",
                community: <CrossIcon />,
                enterprise: (
                    <TableText>Response time within one hour</TableText>
                ),
                valueType: {
                    community: "icon",
                    enterprise: "text",
                },
            },
        ],
    },
    {
        title: "Websites",
        items: [
            {
                description: "Websites",
                community: <TableText>50</TableText>,
                enterprise: (
                    <TableText>
                        Unlimited
                    </TableText>
                ),
                valueType: {
                    community: "text",
                    enterprise: "text",
                },
            },            
            {
                description: "WordPress Manager",
                feature: "wordpress",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Website Builder",
                feature: "website_builder",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "NodeJS Applications",
                feature: "nodejs",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Python Applications",
                feature: "python",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Ruby Applications",
                feature: "ruby",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Java Applications",
                feature: "java",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Git Deployments",
                feature: "git",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "PHP Selector",
                feature: "php",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "PHP Options",
                feature: "php_options",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "PHP.INI Editor",
                feature: "php_ini",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "PHP Extensions",
                feature: "php_extensions",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Temporary Domains",
                feature: "temporary_links",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Varnish Caching",
                feature: "varnish",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "AutoInstaller",
                feature: "autoinstaller",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "AutoInstaller Apps",
                community: <TableText>NextCloud, Matomo, phpBB, DokuWiki, TinyFileManager, TinyPhotoGallery</TableText>,
                enterprise: <TableText>All Community apps + Joomla, Drupal, Moodle, MediaWiki, OpenCart, PrestaShop, Flarum, SofaWiki, OJS, n8n</TableText>,
                valueType: {
                    community: "text",
                    enterprise: "text",
                },
            },
            {
                description: "Cron Jobs",
                feature: "crons",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Webserver Configuration",
                feature: "webserver_conf",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Error Logs",
                feature: "error_logs",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
        ],
    },
    {
        title: "Files",
        items: [     
            {
                description: "File Manager",
                feature: "filemanager",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "FTP",
                feature: "ftp",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Malware Scanner",
                feature: "malware_scan",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Fix Permissions",
                feature: "fix_permissions",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Download from URL",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Disk Usage Explorer",
                feature: "disk_usage",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Inodes Explorer",
                feature: "inodes",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Trash",
                feature: "trash",
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
        title: "Databases",
        items: [
            {
                description: "MySQL",
                feature: "mysql",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "phpMyAdmin",
                feature: "phpmyadmin",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "MySQL Configuration",
                feature: "mysql_conf",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "MySQL Process List",
                feature: "mysql_processlist",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "MySQL Root Password",
                feature: "mysql_root_password",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Switch MySQL Type (MySQL / MariaDB)",
                feature: "change_db",
                community: <CrossIcon />,
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
                feature: "postgresql",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "PostgreSQL Configuration",
                feature: "postgresql_conf",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "MongoDB",
                feature: "mongodb",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Import MySQL Databases",
                feature: "mysql_import",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Import PostgreSQL & MongoDB Databases",
                feature: ["postgresql_import", "mongodb_import"],
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Remote MySQL",
                feature: "remote_mysql",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Remote PostgreSQL",
                feature: "remote_postgresql",
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
        title: "Web servers",
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
            {
                description: "Switch WebServer",
                feature: "change_ws",
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
        title: "Accounts",
        items: [
            {
                description: "User Accounts",
                community: <TableText>3</TableText>,
                enterprise: (
                    <TableText>
                        Unlimited
                    </TableText>
                ),
                valueType: {
                    community: "text",
                    enterprise: "text",
                },
            },     
            {
                description: "Resellers",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Multiple Administrator Accounts",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Transfer Accounts between Servers",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,

                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },   
            {
                description: "Import from a cPanel Backup",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,

                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Login History",
                feature: "login_history",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Two-Factor Authentication (2FA)",
                feature: "twofa",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Passkeys",
                feature: "passkeys",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Activity Log",
                feature: "activity",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Active Sessions",
                feature: "sessions",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Email Notifications",
                feature: "notifications",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Daily Summary",
                feature: "summary",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Change Email & Password",
                feature: "account",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Locales (Language Change)",
                feature: "locale",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Favorites (Bookmarks)",
                feature: "favorites",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },

        ],
    },
    {
        title: "Domains",
        items: [
            {
                description: "Total Domains",
                feature: "domains",
                community: <TableText>50</TableText>,
                enterprise: <TableText>Unlimited</TableText>,
                valueType: {
                    community: "text",
                    enterprise: "text",
                },
            },
            {
                description: "Let's Encrypt",
                feature: "ssl",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Custom SSL",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "DNS Zone Editor",
                feature: "dns",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Change Docroot",
                feature: "docroot",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Domain Logs",
                feature: "domain_logs",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "GoAccess",
                feature: "goaccess",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Coraza WAF",
                feature: "waf",
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
                description: "Multi-Server Cluster",
                info: {
                    name: "multi_server",
                    title: "Multi-Server Cluster",
                    description:
                        "Manage several OpenPanel servers from one OpenAdmin: users, domains and services of all servers in one place, settings applied everywhere, and accounts moved between servers.",
                    link: "",
                    type: "enterprise",
                    help_link: "https://openpanel.com/docs/admin/system/multi-server/",
                },
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
                description: "Dynamic DNS",
                feature: "dynamic_dns",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Domain Redirects",
                feature: "redirects",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "IP Blocker",
                feature: "ip_blocker",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Suspend Domains",
                feature: "domain_suspend",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Capitalize Domains",
                feature: "capitalize_domains",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Edit VirtualHosts",
                feature: "edit_vhost",
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
        title: "Emails",
        items: [
            {
                description: "Email Accounts",
                feature: "emails",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Webmail",
                feature: "webmail",
                community: <CrossIcon />,
                enterprise: <TableText>RoundCube</TableText>,
                valueType: {
                    community: "icon",
                    enterprise: "text",
                },
            },
            {
                description: "Email Deliverability (SPF, DKIM, DMARC)",
                feature: "email_deliverability",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Email Filters & Forwarders",
                feature: "email_filters",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Email Aliases",
                feature: "email_aliases",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Address Importer / Exporter",
                feature: ["email_import", "email_export"],
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Email Migration (IMAP)",
                feature: "email_migrate",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Default (Catch-All) Address",
                feature: "email_default",
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
        title: "Services",
        items: [
            {
                description: "Redis",
                feature: "redis",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Valkey",
                feature: "valkey",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Memcached",
                feature: "memcached",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Elasticsearch",
                feature: "elasticsearch",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "OpenSearch",
                feature: "opensearch",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Services (start/stop, logs)",
                feature: "services",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Process Manager",
                feature: "process_manager",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Resources Usage",
                feature: "usage",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Server Info",
                feature: "info",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
        ],
    },
    {
        title: "Integrations",
        items: [
            {
                description: "REST API",
                feature: "api",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "MCP (Model Context Protocol)",
                feature: "mcp",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "AI Assistant",
                feature: "ai_assistant",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Billing Integrations",
                community: <CrossIcon />,
                enterprise: <TableText><a href="/docs/articles/extensions/openpanel-and-whmcs/">WHMCS</a>, <a href="/docs/articles/extensions/openpanel-and-fossbilling/">FOSSBilling</a>, <a href="/docs/articles/extensions/openpanel-and-blesta/">Blesta</a>, <a href="/docs/articles/extensions/openpanel-and-clientexec/">ClientExec</a>, <a href="/docs/articles/extensions/openpanel-and-wisecp/">WISECP</a>, Paymenter.org</TableText>,
                valueType: {
                    community: "icon",
                    enterprise: "text",
                },
            },
        ],
    },
    {
        title: "System",
        items: [
            {
                description: "User Isolation",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },   
            {
                description: "Services Isolation",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },  
            {
                description: "Users can manage Containers",
                feature: "docker",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            }, 
            {
                description: "Users can add Containers",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },  
            {
                description: "Web Terminal",
                feature: "terminal",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Change Container Image Tag",
                feature: "change_image",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Backup Wizard (full account download)",
                feature: "backup_wizard",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "User manages Backups",
                feature: "backups",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Support for ARM CPUs (Aarch64)",
                community: <CrossIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Supported OS",
                community: <TableText><a href="/docs/admin/intro/#requirements">Ubuntu 24.04</a></TableText>,
                enterprise: <TableText><a href="/docs/admin/intro/#requirements">Ubuntu, Debian, AlmaLinux, Rocky, CentOS</a></TableText>,
                valueType: {
                    community: "text",
                    enterprise: "text",
                },
            },
             {
                description: "Available on Cloud/VPS",
                community: <CheckIcon />,
                enterprise: <CheckIcon />,
                valueType: {
                    community: "icon",
                    enterprise: "icon",
                },
            },
            {
                description: "Available on Metal/Dedicated",
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
        title: "Pricing",
        items: [
            {
                description: "Yearly Pricing",
                community: (
                    <div className={clsx("h-full")}>
                        <TableText></TableText>
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
