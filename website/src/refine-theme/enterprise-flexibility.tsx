import React from "react";
import clsx from "clsx";
import Translate from "@docusaurus/Translate";
import { CommonThemedImage } from "./common-themed-image";
import { LandingPureReactCode } from "./landing-pure-react-code";

export const EnterpriseFlexibility = ({ className }: { className?: string }) => {
    return (
        <div className={clsx("flex flex-col", "not-prose", className)}>
            <div
                className={clsx(
                    "text-2xl landing-sm:text-[32px] landing-sm:leading-[40px]",
                    "pl-4 landing-sm:pl-6 landing-md:pl-10",
                )}
            >
                <h2 className={clsx("dark:text-gray-0 text-gray-900")}>
                    <Translate id="enterprise.flexibility.title">Build with unmatched </Translate>
                    <span
                        className={clsx(
                            "font-semibold",
                            "dark:text-refine-cyan-alt dark:drop-shadow-[0_0_30px_rgba(71,235,235,0.25)]",
                            "text-refine-blue drop-shadow-[0_0_30px_rgba(51,51,255,0.3)]",
                        )}
                    >
                        <Translate id="enterprise.flexibility.flexibility">flexibility</Translate>
                    </span>
                    .
                </h2>
            </div>

            <div
                className={clsx(
                    "mt-8 landing-sm:mt-12 landing-md:mt-20",
                    "grid",
                    "grid-cols-1 landing-md:grid-cols-2",
                    "gap-8 landing-md:gap-12 landing-lg:gap-6",
                )}
            >
                <div
                    className={clsx(
                        "flex flex-col",
                        "dark:bg-landing-noise",
                        "dark:bg-gray-800 bg-gray-50",
                        "rounded-2xl landing-sm:rounded-3xl",
                    )}
                >
                    <div className={clsx("p-2 landing-sm:p-4")}>
                        <div
                            className={clsx(
                                "flex",
                                "items-center",
                                "justify-center",
                                "dark:bg-gray-900 bg-gray-0",
                                "rounded-xl landing-sm:rounded-2xl",
                                "overflow-hidden",
                            )}
                        >
                            <LandingPureReactCode />
                        </div>
                    </div>
                    <div
                        className={clsx(
                            "flex flex-col",
                            "gap-2 landing-sm:gap-4",
                            "p-4 landing-sm:p-10",
                        )}
                    >
                        <p
                            className={clsx(
                                "text-base",
                                "dark:text-gray-400 text-gray-600",
                            )}
                        >
                            <Translate id="enterprise.flexibility.pure_react.desc">Don’t settle for overpriced and outdated hosting panels. With OpenPanel, you maintain 100% control over your server and data at all times.</Translate>
                        </p>
                    </div>
                </div>

                <div
                    className={clsx(
                        "flex flex-col",
                        "dark:bg-landing-noise",
                        "dark:bg-gray-800 bg-gray-50",
                        "rounded-2xl landing-sm:rounded-3xl",
                    )}
                >
                    <CommonThemedImage
                        className={clsx("rounded-2xl landing-sm:rounded-3xl")}
                        srcDark="https://refine.ams3.cdn.digitaloceanspaces.com/enterprise/flexibility-dark.png"
                        srcLight="https://refine.ams3.cdn.digitaloceanspaces.com/enterprise/flexibility-light.png"
                    />
                    <div
                        className={clsx(
                            "flex flex-col",
                            "gap-2 landing-sm:gap-4",
                            "p-4 landing-sm:p-10",
                        )}
                    >
                        <h2
                            className={clsx(
                                "text-base landing-sm:text-2xl",
                                "dark:text-gray-300 text-gray-900",
                                "font-semibold",
                            )}
                        >
                            <Translate id="enterprise.flexibility.alternative.title">Alternative to building from scratch</Translate>
                        </h2>
                        <p
                            className={clsx(
                                "text-base",
                                "dark:text-gray-400 text-gray-600",
                            )}
                        >
                            <Translate id="enterprise.flexibility.alternative.desc">Every feature needed to run a webserver is integrated, so you don't need to buy additional software for tasks like backups, WordPress management, or user isolation.</Translate>
                        </p>
                    </div>
                </div>
            </div>
        </div>
    );
};
