import React from "react";
import clsx from "clsx";
import { LandingStartActionIcon } from "./icons/landing-start-action";
import { LandingCopyCommandButton } from "./landing-copy-command-button";
import { useLocation } from "@docusaurus/router";
import Link from "@docusaurus/Link";

export const LandingTryItSection = ({ className }: { className?: string }) => {
    const { search, hash } = useLocation();

    const [params, setParams] = React.useState<Record<string, string>>({});

    React.useEffect(() => {
        const _params = new URLSearchParams(search);
        const paramsObj: Record<string, string> = {};

        // @ts-expect-error no downlevel iteration
        for (const [key, value] of _params.entries()) {
            paramsObj[key] = value;
        }

        setParams(paramsObj);
    }, [search]);

    const scrollToItem = React.useCallback(() => {
        const playgroundElement = document.getElementById("playground");
        if (playgroundElement) {
            playgroundElement.scrollIntoView({
                behavior: "smooth",
                block: "start",
                inline: "nearest",
            });
        }
    }, []);

    React.useEffect(() => {
        if (params.playground || hash === "#playground") {
            scrollToItem();
        }
    }, [params.playground, hash]);

    return (
        <div
            id="playground"
            className={clsx(
                "flex",
                "flex-col",
                "gap-8 landing-sm:gap-12 landing-md:gap-8 ",
                className,
            )}
            style={{
                scrollMarginTop: "6rem",
            }}
        >
            <div
                className={clsx(
                    "flex",
                    "flex-col",
                    "gap-4 landing-sm:gap-6",
                    "px-4 landing-sm:px-10",
                )}
            >
                <h2
                    className={clsx(
                        "text-2xl landing-sm:text-[32px] landing-sm:leading-[40px]",
                        "font-bold",
                        "text-gray-900 dark:text-gray-0",
                    )}
                >
                    Ready to get started?
                </h2>
                <p
                    className={clsx(
                        "text-base",
                        "font-normal",
                        "text-gray-600 dark:text-gray-400",
                    )}
                >
                    Download the (free) Community Edition that is intended for self-hosting, offers basic features and supports up to 3 user accounts and 50 websites.
                </p>
                <p
                    className={clsx(
                        "text-base",
                        "font-normal",
                        "text-gray-600 dark:text-gray-400",
                    )}
                >
                    Upgrade to Enterprise Edition anytime to increase limits and access advanced features like email, FTP, and Docker management.
                </p>
            </div>
            <div
                className={clsx(
                    "w-full",
                    "rounded-2xl landing-md:rounded-3xl",
                    "relative",
                    "overflow-hidden",
                )}
            >
                <LandingTryItOptionsSection
                    className={clsx("w-full")}
                />
            </div>
        </div>
    );
};

const LandingTryItOptionsSection = ({
    className,
}: {
    className?: string;
}) => {
    return (
        <div
            className={clsx(
                "relative",
                "flex",
                "flex-col landing-md:flex-row",
                className,
            )}
        >
            <div
                className={clsx(
                    "flex-1",
                    "rounded-2xl landing-md:rounded-3xl",
                    "landing-md:rounded-tr-none landing-md:rounded-br-none",
                    "flex",
                    "flex-col",
                    "gap-6 landing-sm:gap-10",
                    "pt-4 landing-sm:pt-10 landing-md:pt-16",
                    "px-4 landing-sm:px-10",
                    "pb-14 landing-sm:pb-20 landing-md:pb-16",
                    "bg-gray-50 dark:bg-gray-800",
                    "landing-md:bg-landing-wizard-option-bg-light dark:landing-md:bg-landing-wizard-option-bg-dark",
                    "landing-md:bg-landing-wizard-option-left landing-md:bg-landing-wizard-option",
                )}
                style={{
                    backgroundRepeat: "no-repeat, repeat",
                }}
            >
                <p
                    className={clsx(
                        "text-base landing-sm:text-xl landing-md:text-base landing-lg:text-xl",
                        "font-semibold",
                        "text-gray-600 dark:text-gray-400",
                        "landing-md:max-w-[318px]",
                        "landing-lg:max-w-[446px]",
                    )}
                >
                    Get the Enterprise Edition for a fixed lifetime price of €14.95/month.
                </p>
                        <Link
                            to="enterprise"
                            className={clsx(
                                "self-start",
                                "rounded-3xl",
                                "!text-gray-0 dark:!text-gray-900",
                                "bg-refine-blue dark:bg-refine-cyan-alt",
                                "transition-[filter]",
                                "duration-150",
                                "ease-in-out",
                                "hover:brightness-110",
                                "py-3",
                                "pl-7 pr-8",
                                "landing-md:px-8",
                                "landing-lg:pl-7 landing-lg:pr-8",
                                "flex",
                                "items-center",
                                "justify-center",
                                "gap-2",
                                "hover:!no-underline",
                            )}
                        >
                    <LandingStartActionIcon />
                    <span className={clsx("text-base", "font-semibold")}>
                        Purchase license
                    </span>
                </Link>
            </div>
            <div
                className={clsx(
                    "h-4 landing-md:h-full",
                    "w-full landing-md:w-0",
                    "relative",
                    "flex-shrink-0",
                )}
            >
                <div
                    className={clsx(
                        "hidden",
                        "landing-md:block",
                        "absolute",
                        "-left-2",
                        "skew-x-[14deg]",
                        "top-0",
                        "h-[272px]",
                        "w-2",
                        "bg-gray-0 dark:bg-gray-900",
                    )}
                />
                <div
                    className={clsx(
                        "absolute",
                        "-top-6 left-8",
                        "landing-md:top-32 landing-md:-left-1",
                        "landing-md:-translate-x-1/2",
                        "landing-md:-translate-y-1/2",
                        "bg-gray-0 dark:bg-gray-900",
                        "text-gray-600 dark:text-gray-400",
                        "w-16 h-16 landing-md:w-[78px] landing-md:h-[78px]",
                        "rounded-full",
                        "text-base",
                        "uppercase",
                        "flex items-center justify-center",
                    )}
                >
                    or
                </div>
            </div>
            <div
                className={clsx(
                    "flex-1",
                    "rounded-2xl landing-md:rounded-3xl",
                    "flex flex-col",
                    "landing-md:rounded-tl-none landing-md:rounded-bl-none",
                    "pb-4 landing-sm:pb-10 landing-md:pb-16",
                    "px-4 landing-sm:px-10",
                    "pt-14 landing-sm:pt-20 landing-md:pt-16",
                    "bg-gray-50 dark:bg-gray-800",
                    "landing-md:bg-landing-wizard-option-bg-light dark:landing-md:bg-landing-wizard-option-bg-dark",
                    "landing-md:bg-landing-wizard-option-right landing-md:bg-landing-wizard-option",
                    "landing-md:items-end",
                )}
                style={{
                    backgroundRepeat: "no-repeat, repeat",
                }}
            >
                <div
                    className={clsx(
                        "landing-md:max-w-[318px]",
                        "landing-lg:max-w-[446px]",
                        "flex",
                        "flex-col",
                        "gap-6 landing-sm:gap-10",
                    )}
                >
                    <p
                        className={clsx(
                            "text-base landing-sm:text-xl landing-md:text-base landing-lg:text-xl",
                            "font-semibold",
                            "text-gray-600 dark:text-gray-400",
                            "landing-lg:max-w-[446px]",
                        )}
                    >
						Download and use the FOREVER FREE Community Edition with limited features.
                    </p>
                    <LandingCopyCommandButton />
                </div>
            </div>
        </div>
    );
};
