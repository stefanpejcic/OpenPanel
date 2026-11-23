import clsx from "clsx";
import React, { FC } from "react";
import {
    AccessControlIcon,
    BlackBoxIcon,
    IdentityIcon,
    MonitorIcon,
    SelfHostedIcon,
    SupportIcon,
} from "../components/landing/icons";

import { LandingSectionCtaButtonAlt } from "./landing-section-cta-button";

const list = [
    {
        icon: <SelfHostedIcon />,
        title: "Containers per service",
        description:
            "Rootless Podman. Every service of every user is its own container, with its own CPU, memory and port speed limits.",
    },
    {
        icon: <IdentityIcon />,
        title: "Admins, resellers and plans",
        description:
            "Create Admins and resellers, set limits on websites, domains and databases, and control exactly which features each user or plan can access.",
    },
    {
        icon: <AccessControlIcon />,
        title: "Login security",
        description:
            "Passkeys, enforced two-factor login, IP blocking per user, and control over who can reach MySQL remotely.",
    },
    {
        icon: <BlackBoxIcon />,
        title: "Your brand, not ours",
        description:
            "Apply your brand colors and logo, rename the panel, set custom nameservers, and hide any feature you don't want users to see.",
    },
    {
        icon: <MonitorIcon />,
        title: "Logs and alerts",
        description:
            "User activity log, resource usage history, and server alerts by email or webhook. No extra monitoring tool to install.",
    },
    {
        icon: <SupportIcon />,
        title: "API and scripting",
        description:
            "REST API, an MCP server for AI assistants, WHMCS and FOSSBilling modules, and opencli for shell scripts.",
    },
];

type Props = {
    className?: string;
};

export const LandingEnterpriseDevelopers: FC<Props> = ({ className }) => {
    return (
        <div className={clsx(className, "w-full")}>
            <div
                className={clsx("not-prose", "w-full", "px-4 landing-md:px-10")}
            >
                <h2
                    className={clsx(
                        "text-2xl landing-sm:text-[32px] landing-sm:leading-[40px]",
                        "tracking-tight",
                        "text-start",
                        "p-0",
                        "dark:text-gray-0 text-gray-900",
                    )}
                >
                    Built for{" "}
                    <span
                        className={clsx(
                            "font-semibold",
                            "dark:text-refine-cyan-alt",
                            "text-refine-blue",
                        )}
                    >
                        hosting companies
                    </span>
                    .
                </h2>
                <p
                    className={clsx(
                        "mt-4 landing-sm:mt-6",
                        "text-base",
                        "dark:text-gray-400 text-gray-600",
                    )}
                >
                    What we added after years of running shared hosting ourselves, starting with{" "}
                    <span className="font-semibold text-gray-900 dark:text-gray-0">
                        security
                    </span>
                    .
                </p>
            </div>

            <div
                className={clsx(
                    "mt-8 landing-sm:mt-12 landing-lg:mt-20",
                    "grid",
                    "grid-cols-1 landing-md:grid-cols-2 landing-lg:grid-cols-3",
                    "gap-4 landing-sm:gap-12 landing-md:gap-6",
                    "mb-4 landing-sm:mb-12 landing-md:mb-6",
                )}
            >
                {list.map((item, index) => {
                    return (
                        <div
                            key={index}
                            className={clsx(
                                "not-prose",
                                "p-4 landing-sm:p-10",
                                "flex",
                                "flex-col landing-sm:flex-row landing-md:flex-col",
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
                                        "text-xl",
                                        "font-semibold",
                                        "text-gray-900 dark:text-gray-0",
                                    )}
                                >
                                    {item.title}
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
                    );
                })}
            </div>

            <LandingSectionCtaButtonAlt to="/features">
                View all features
            </LandingSectionCtaButtonAlt>
        </div>
    );
};
