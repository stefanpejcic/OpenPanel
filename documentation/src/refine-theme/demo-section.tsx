import React from "react";
import clsx from "clsx";
import { OpenPanelLogoIcon } from "./icons/openpanel-logo";
import Link from "@docusaurus/Link";

export const LandingTryItSection = ({ className }: { className?: string }) => {
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
                    "w-full",
                    "rounded-2xl landing-md:rounded-3xl",
                    "relative",
                    "overflow-hidden",
                    "transition-[min-height,height]",
                    "duration-300",
                    "ease-out",
                    "min-h-[515px]",
                )}
            >
                <LandingTryItOptionsSection
                    className={clsx(
                        "w-full",
                        "transition-[transform,opacity,margin-bottom]",
                        "duration-300",
                        "ease-in-out"
                    )}
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
                        "text-base landing-sm:text-l landing-md:text-base landing-lg:text-l",
                        "font-semibold",
                        "text-gray-600 dark:text-gray-400",
                        "landing-md:max-w-[318px]",
                        "landing-lg:max-w-[446px]",
                    )}
                >
                    If you are a website owner, then we recommend interacting with the OpenPanel interface.
                </p>

                <p
                    className={clsx(
                        "text-base landing-sm:text-l landing-md:text-base landing-lg:text-l",
                        "font-semibold",
                        "text-gray-600 dark:text-gray-400",
                        "landing-md:max-w-[318px]",
                        "landing-lg:max-w-[446px]",
                    )}
                >
                    This is where you can manage your websites.
                </p>
		    
                        <Link
                            to="https://demo.openpanel.org:2083/login"
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
                    <OpenPanelLogoIcon />
                    <span className={clsx("text-base", "font-semibold")}>
                        OpenPanel Demo
                    </span>
                </Link>
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
                            "text-base landing-sm:text-l landing-md:text-base landing-lg:text-l",
                            "font-semibold",
                            "text-gray-600 dark:text-gray-400",
                            "landing-lg:max-w-[446px]",
                        )}
                    >
			If you are a web host, then we recommend interacting with the OpenAdmin interface.
		    </p>
                    <p
                        className={clsx(
                            "text-base landing-sm:text-l landing-md:text-base landing-lg:text-l",
                            "font-semibold",
                            "text-gray-600 dark:text-gray-400",
                            "landing-lg:max-w-[446px]",
                        )}
                    >
			This is where you can manage your server.
		    </p>
                        <Link
                            to="https://demo.openpanel.org:2087/login"
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
                    	<OpenPanelLogoIcon />
	                    <span className={clsx("text-base", "font-semibold")}>
	                        OpenAdmin Demo
	                    </span>
	                </Link>
		</div>
            </div>
        </div>
    );
};
