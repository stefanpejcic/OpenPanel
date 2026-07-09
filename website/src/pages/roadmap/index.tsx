import React from "react";
import Head from "@docusaurus/Head";
import Link from "@docusaurus/Link";
import clsx from "clsx";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { HeaderDiscordIcon } from "@site/src/refine-theme/icons/header-discord";
import {
    TimelineIcon,
    WizardsIcon,
    SecurityIcon,
    NoVendorLockinIcon,
    UtilitiesIcon,
    SelfHostedIcon,
} from "@site/src/components/landing/icons";

// The two items currently being worked on. Update this once they ship.
const inProgress = [
    {
        icon: <TimelineIcon />,
        title: "Backup restore GUI",
        description:
            "Browse existing backups and restore a website, database, or full account with a couple of clicks - no terminal required.",
    },
    {
        icon: <WizardsIcon />,
        title: "Autoinstallers",
        description:
            "One-click installers for the most popular CMS and scripts, so users can launch WordPress and friends straight from OpenPanel.",
    },
];

// Everything else that's planned. Keep this list up to date manually.
const plannedFeatures = [
    {
        icon: <SecurityIcon />,
        title: "ImunifyAV page for end-users",
    },
    {
        icon: <NoVendorLockinIcon />,
        title: "Ubuntu 26 support",
    },
    {
        icon: <UtilitiesIcon />,
        title: "Export OpenAdmin settings",
    },
    {
        icon: <SelfHostedIcon />,
        title: "Ongoing improvements in container isolation",
    },
];

const Roadmap: React.FC = () => {
    return (
        <CommonLayout>
            <Head title="ROADMAP | OpenPanel">
                <html data-page="roadmap" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div
                    className={clsx(
                        "flex flex-col",
                        "gap-16 landing-sm:gap-20 landing-md:gap-28",
                        "w-full max-w-[592px] landing-sm:max-w-[656px] landing-md:max-w-[896px] landing-lg:max-w-[1200px]",
                        "px-2 landing-sm:px-0",
                        "pt-8 landing-sm:pt-12 landing-lg:pt-20",
                        "pb-16 landing-sm:pb-24 landing-md:pb-32",
                        "mx-auto",
                        "not-prose",
                    )}
                >
                    <div className={clsx("px-4 landing-md:px-10")}>
                        <h1
                            className={clsx(
                                "text-3xl landing-sm:text-[40px] landing-sm:leading-[48px]",
                                "font-semibold",
                                "tracking-tight",
                                "p-0",
                                "dark:text-gray-0 text-gray-900",
                            )}
                        >
                            Roadmap
                        </h1>
                        <p
                            className={clsx(
                                "mt-4 landing-sm:mt-6",
                                "max-w-lg",
                                "text-base",
                                "dark:text-gray-400 text-gray-600",
                            )}
                        >
                            A look at where OpenPanel is headed - shared openly so you
                            know what's coming, and can tell us what should come next.
                        </p>
                    </div>

                    <div className={clsx("w-full")}>
                        <div className={clsx("px-4 landing-md:px-10")}>
                            <h2
                                className={clsx(
                                    "text-2xl landing-sm:text-[32px] landing-sm:leading-[40px]",
                                    "tracking-tight",
                                    "p-0",
                                    "dark:text-gray-0 text-gray-900",
                                )}
                            >
                                In active{" "}
                                <span
                                    className={clsx(
                                        "font-semibold",
                                        "dark:text-refine-cyan-alt dark:drop-shadow-[0_0_30px_rgba(71,235,235,0.25)]",
                                        "text-refine-blue drop-shadow-[0_0_30px_rgba(0,128,255,0.3)]",
                                    )}
                                >
                                    development
                                </span>
                                .
                            </h2>
                            <p
                                className={clsx(
                                    "mt-4 landing-sm:mt-6",
                                    "max-w-md",
                                    "text-base",
                                    "dark:text-gray-400 text-gray-600",
                                )}
                            >
                                The two things our team is building right now, shipping
                                in the next releases.
                            </p>
                        </div>

                        <div
                            className={clsx(
                                "mt-8 landing-sm:mt-12 landing-lg:mt-20",
                                "grid",
                                "grid-cols-1 landing-sm:grid-cols-2",
                                "gap-4 landing-sm:gap-6",
                            )}
                        >
                            {inProgress.map((item, index) => (
                                <div
                                    key={index}
                                    className={clsx(
                                        "not-prose",
                                        "p-4 landing-sm:p-10",
                                        "flex",
                                        "flex-col landing-sm:flex-row",
                                        "items-start",
                                        "gap-6",
                                        "dark:bg-landing-noise",
                                        "dark:bg-gray-800 bg-gray-50",
                                        "rounded-2xl landing-sm:rounded-3xl",
                                    )}
                                >
                                    <div>{item.icon}</div>
                                    <div className={clsx("flex", "flex-col", "gap-4")}>
                                        <div
                                            className={clsx(
                                                "flex items-center gap-3",
                                            )}
                                        >
                                            <div
                                                className={clsx(
                                                    "text-xl",
                                                    "font-semibold",
                                                    "text-gray-900 dark:text-gray-0",
                                                )}
                                            >
                                                {item.title}
                                            </div>
                                            <div
                                                className={clsx(
                                                    "text-xs",
                                                    "font-semibold",
                                                    "uppercase",
                                                    "tracking-wide",
                                                    "px-2.5 py-1",
                                                    "rounded-full",
                                                    "dark:bg-refine-cyan-alt/10 bg-refine-blue/10",
                                                    "dark:text-refine-cyan-alt text-refine-blue",
                                                )}
                                            >
                                                In progress
                                            </div>
                                        </div>
                                        <div
                                            className={clsx(
                                                "text-base",
                                                "dark:text-gray-400 text-gray-600",
                                            )}
                                        >
                                            {item.description}
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>

                    <div className={clsx("w-full")}>
                        <div className={clsx("px-4 landing-md:px-10")}>
                            <h2
                                className={clsx(
                                    "text-2xl landing-sm:text-[32px] landing-sm:leading-[40px]",
                                    "tracking-tight",
                                    "p-0",
                                    "dark:text-gray-0 text-gray-900",
                                )}
                            >
                                On the{" "}
                                <span
                                    className={clsx(
                                        "font-semibold",
                                        "dark:text-refine-cyan-alt dark:drop-shadow-[0_0_30px_rgba(71,235,235,0.25)]",
                                        "text-refine-blue drop-shadow-[0_0_30px_rgba(0,128,255,0.3)]",
                                    )}
                                >
                                    roadmap
                                </span>
                                .
                            </h2>
                            <p
                                className={clsx(
                                    "mt-4 landing-sm:mt-6",
                                    "max-w-md",
                                    "text-base",
                                    "dark:text-gray-400 text-gray-600",
                                )}
                            >
                                Up next, in no particular order. Nothing is set in stone
                                - priorities can shift based on community feedback.
                            </p>
                        </div>

                        <div
                            className={clsx(
                                "mt-8 landing-sm:mt-12",
                                "grid",
                                "grid-cols-1 landing-sm:grid-cols-2",
                                "gap-4",
                            )}
                        >
                            {plannedFeatures.map((item, index) => (
                                <div
                                    key={index}
                                    className={clsx(
                                        "not-prose",
                                        "p-4 landing-sm:p-6",
                                        "flex",
                                        "items-center",
                                        "gap-4",
                                        "dark:bg-landing-noise",
                                        "dark:bg-gray-800 bg-gray-50",
                                        "rounded-2xl",
                                    )}
                                >
                                    <div
                                        className={clsx(
                                            "flex-shrink-0",
                                            "dark:text-refine-cyan-alt text-refine-blue",
                                        )}
                                    >
                                        {item.icon}
                                    </div>
                                    <div
                                        className={clsx(
                                            "text-base",
                                            "font-semibold",
                                            "text-gray-900 dark:text-gray-0",
                                        )}
                                    >
                                        {item.title}
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>

                    <div
                        className={clsx(
                            "not-prose",
                            "flex flex-col items-center text-center",
                            "gap-4",
                            "px-6 py-12 landing-sm:px-20 landing-sm:py-16",
                            "dark:bg-landing-noise",
                            "dark:bg-gray-800 bg-gray-50",
                            "rounded-2xl landing-sm:rounded-3xl",
                        )}
                    >
                        <h2
                            className={clsx(
                                "text-2xl landing-sm:text-[32px] landing-sm:leading-[40px]",
                                "tracking-tight",
                                "font-semibold",
                                "p-0",
                                "dark:text-gray-0 text-gray-900",
                            )}
                        >
                            Want to shape the roadmap?
                        </h2>
                        <p
                            className={clsx(
                                "max-w-lg",
                                "text-base",
                                "dark:text-gray-400 text-gray-600",
                            )}
                        >
                            Join our community and tell us what matters most to your
                            hosting business - operator feedback drives what we build
                            next.
                        </p>
                        <div
                            className={clsx(
                                "mt-4",
                                "flex flex-col landing-xs:flex-row",
                                "items-center justify-center",
                                "gap-4",
                                "w-full landing-xs:w-auto",
                            )}
                        >
                            <Link
                                to="https://discord.com/invite/7bNY8fANqF"
                                target="_blank"
                                rel="noopener noreferrer"
                                className={clsx(
                                    "w-full landing-xs:w-auto",
                                    "!text-gray-0 dark:!text-gray-900",
                                    "bg-refine-blue dark:bg-refine-cyan-alt",
                                    "transition-[filter]",
                                    "duration-150",
                                    "ease-in-out",
                                    "hover:brightness-110",
                                    "hover:!no-underline",
                                    "rounded-3xl",
                                    "py-3 px-8",
                                    "flex items-center justify-center gap-2",
                                    "text-base font-semibold",
                                )}
                            >
                                <HeaderDiscordIcon width={20} height={20} />
                                Join Discord
                            </Link>
                            <Link
                                to="https://community.openpanel.org/"
                                target="_blank"
                                rel="noopener noreferrer"
                                className={clsx(
                                    "w-full landing-xs:w-auto",
                                    "dark:text-refine-cyan-alt text-refine-blue",
                                    "dark:bg-refine-cyan-alt/10 bg-refine-blue/10",
                                    "transition-[filter]",
                                    "duration-150",
                                    "ease-in-out",
                                    "hover:brightness-110",
                                    "hover:!no-underline",
                                    "rounded-3xl",
                                    "py-3 px-8",
                                    "flex items-center justify-center gap-2",
                                    "text-base font-semibold",
                                )}
                            >
                                Community forum
                            </Link>
                        </div>
                    </div>
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Roadmap;
