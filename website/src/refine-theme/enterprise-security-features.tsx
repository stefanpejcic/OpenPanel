import React from "react";
import clsx from "clsx";
import Translate from "@docusaurus/Translate";
import {
    AccessControlIcon,
    MonitorIcon,
    SelfHostedIcon,
} from "../components/landing/icons";

export const EnterpriseSecurityFeatures = ({
    className,
}: {
    className?: string;
}) => (
    <div className={clsx(className, "w-full")}>
        <div
            className={clsx(
                "grid",
                "grid-cols-1 landing-lg:grid-cols-3",
                "gap-4 landing-md:gap-12 landing-lg:gap-6",
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
                            "flex-col landing-sm:flex-row  landing-lg:flex-col",
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
                                <Translate id={item.titleId}>{item.title}</Translate>
                            </div>
                            <div
                                className={clsx(
                                    "text-base",
                                    "dark:text-gray-400 text-gray-600",
                                )}
                            >
                                <Translate id={item.descriptionId}>{item.description}</Translate>
                            </div>
                        </div>
                    </div>
                );
            })}
        </div>
    </div>
);

const list = [
    {
        icon: <AccessControlIcon />,
        titleId: "enterprise.security_features.secure.title",
        title: "Secure by default",
        descriptionId: "enterprise.security_features.secure.desc",
        description:
            "Enable Two Factor authentification, block IP addresses per domain, control remote MySQL access and SSH access.",
    },
    {
        icon: <SelfHostedIcon />,
        titleId: "enterprise.security_features.resource.title",
        title: "Limit resource usage",
        descriptionId: "enterprise.security_features.resource.desc",
        description:
            "Restrict the number of websites, domains, or databases, while enforcing hard limits on CPU, memory, port speed, disk usage, and inodes.",
    },
    {
        icon: <MonitorIcon />,
        titleId: "enterprise.security_features.monitor.title",
        title: "Monitor your server and users",
        descriptionId: "enterprise.security_features.monitor.desc",
        description:
            "Built-in features include logging user activity, visualizing website visitors, and analyzing resource usage.",
    },
];
