import React from "react";
import clsx from "clsx";
import Translate from "@docusaurus/Translate";
import { ClockIcon } from "./icons/clock";
import { SlackIcon } from "../assets/integration-icons";
import { CommonThemedImage } from "./common-themed-image";

export const EnterpriseGetSupport = ({ className }: { className?: string }) => {
    return (
        <div className={clsx("flex flex-col", "not-prose", className)}>
            <div
                className={clsx(
                    "pl-4 landing-sm:pl-6 landing-md:pl-10",
                    "text-2xl landing-sm:text-[32px] landing-sm:leading-[40px]",
                )}
            >
                <h3 className={clsx("mt-4 landing-sm:mt-10")}>
                    <span className={clsx("dark:text-gray-0 text-gray-900")}>
                        <Translate id="enterprise.get_support.alternative.prefix">c*anel and P*esk </Translate>
                    </span>
                    <span
                        className={clsx(
                            "font-semibold",
                            "dark:text-refine-cyan-alt dark:drop-shadow-[0_0_30px_rgba(71,235,235,0.25)]",
                            "text-refine-blue drop-shadow-[0_0_30px_rgba(51,51,255,0.3)]",
                        )}
                    >
                        <Translate id="enterprise.get_support.alternative.suffix">Alternative</Translate>
                    </span>
                    .
                </h3>
            </div>

            <div
                className={clsx(
                    "mt-8 landing-md:mt-20",
                    "grid",
                    "grid-cols-1 landing-md:grid-cols-2",
                    "gap-8 landing-sm:gap-12 landing-lg:gap-6",
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
                    <CommonThemedImage
                        className={clsx(
                            "rounded-2xl landing-sm:rounded-3xl",
                            "landing-lg:h-[360px]",
                        )}
                        srcDark="https://refine.ams3.cdn.digitaloceanspaces.com/enterprise/priority-support-dark.png"
                        srcLight="https://refine.ams3.cdn.digitaloceanspaces.com/enterprise/priority-support-light.png"
                    />
                    <div
                        className={clsx(
                            "flex flex-col",
                            "gap-4",
                            "p-4 landing-sm:p-10",
                            "not-prose",
                        )}
                    >
                        <h2
                            className={clsx(
                                "text-base landing-sm:text-2xl",
                                "dark:text-gray-300 text-gray-900",
                                "font-semibold",
                            )}
                        >
                            <Translate id="enterprise.get_support.priority_support.title">
                                Priority support
                            </Translate>
                        </h2>
                        <p
                            className={clsx(
                                "text-base",
                                "dark:text-gray-400 text-gray-600",
                            )}
                        >
                            <Translate id="enterprise.get_support.priority_support.desc">
                                Exclusive technical support with SLA, provided by
                                OpenPanel core team members.
                            </Translate>
                        </p>
                        <div
                            className={clsx(
                                "flex items-center flex-wrap gap-4",
                                "mt-4",
                            )}
                        >
                            <div
                                style={{
                                    backdropFilter: "blur(4px)",
                                }}
                                className={clsx(
                                    "flex items-center gap-2",
                                    "dark:bg-gray-900/50 bg-gray-0",
                                    "dark:text-gray-400 text-gray-600",
                                    "pl-3 py-3 pr-5",
                                    "rounded-full",
                                    "text-xs",
                                )}
                            >
                                <SlackIcon />
                                <div>
                                    <Translate id="enterprise.get_support.priority_support.slack">
                                        Private Slack channel
                                    </Translate>
                                </div>
                            </div>
                            <div
                                style={{
                                    backdropFilter: "blur(4px)",
                                }}
                                className={clsx(
                                    "flex items-center gap-2",
                                    "dark:bg-gray-900/50 bg-gray-0",
                                    "dark:text-gray-400 text-gray-600",
                                    "pl-3 py-3 pr-5",
                                    "rounded-full",
                                    "text-xs",
                                )}
                            >
                                <ClockIcon
                                    className={clsx(
                                        "dark:text-refine-blue-alt text-refine-blue",
                                    )}
                                />
                                <div>
                                    <Translate id="enterprise.get_support.priority_support.response_time">
                                        24 Hours First response time
                                    </Translate>
                                </div>
                            </div>
                        </div>
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
                        style={{
                            objectFit: "none",
                        }}
                        className={clsx(
                            "rounded-2xl landing-sm:rounded-3xl",
                            "landing-lg:h-[360px]",
                        )}
                        srcDark="/img/ilustrations/enterprise_modules.png"
                        srcLight="/img/ilustrations/enterprise_modules_light.png"
                    />
                    <div
                        className={clsx(
                            "flex flex-col",
                            "gap-2 landing-sm:gap-4",
                            "p-4 landing-sm:p-10",
                            "not-prose",
                        )}
                    >
                        <h2
                            className={clsx(
                                "text-base landing-sm:text-2xl",
                                "dark:text-gray-300 text-gray-900",
                                "font-semibold",
                            )}
                        >
                            <Translate id="enterprise.get_support.premium_features.title">
                                Premium Features
                            </Translate>
                        </h2>
                        <p
                            className={clsx(
                                "text-base",
                                "dark:text-gray-400 text-gray-600",
                            )}
                        >
                            <Translate id="enterprise.get_support.premium_features.desc">
                                Premium features for your hosting business: Emails, FTP, Docker, OpenLitespeed, Varnish Caching, and much more.
                            </Translate>
                        </p>
                    </div>
                </div>
            </div>
        </div>
    );
};
